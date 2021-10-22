package validate

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"

	"github.com/operator-framework/operator-registry/alpha/declcfg"
	"github.com/operator-framework/operator-registry/alpha/model"
	"github.com/operator-framework/operator-registry/alpha/property"
)

func TestValidate(t *testing.T) {
	requireMarshal := func(i interface{}) []byte {
		out, err := json.Marshal(i)
		require.NoError(t, err)
		return out
	}

	tests := []struct {
		name    string
		config  model.Model
		wantErr error
	}{
		{
			name: "failWithoutHeadBundleCSV",
			config: model.Model{
				"testpkg": {
					Name:           "testpkg",
					DefaultChannel: &model.Channel{Name: "stable"},
					Channels: map[string]*model.Channel{
						"stable": {
							Package: &model.Package{Name: "testpkg"},
							Name:    "stable",
							Bundles: map[string]*model.Bundle{
								"head": {
									Package: &model.Package{Name: "testpkg"},
									Channel: &model.Channel{Name: "stable"},
									Name:    "head",
									Image:   "head:image",
									Properties: []property.Property{
										{
											Type: property.TypePackage,
											Value: requireMarshal(&property.Package{
												PackageName: "testpkg",
												Version:     "1.1.1",
											}),
										},
									},
								},
							},
						},
					},
				},
			},
			wantErr: fmt.Errorf("missing head CSV on package testpkg, channel stable head head: ensure valid csv under 'olm.bundle.object' properties"),
		},
		{
			name: "passWithHeadBundleCSV",
			config: model.Model{
				"testpkg": {
					Name:           "testpkg",
					DefaultChannel: &model.Channel{Name: "stable"},
					Channels: map[string]*model.Channel{
						"stable": {
							Package: &model.Package{Name: "testpkg"},
							Name:    "stable",
							Bundles: map[string]*model.Bundle{
								"head": {
									Package:  &model.Package{Name: "testpkg"},
									Channel:  &model.Channel{Name: "stable"},
									Name:     "head",
									Image:    "head:image",
									Replaces: "non-head",
									Properties: []property.Property{
										{
											Type: property.TypePackage,
											Value: requireMarshal(&property.Package{
												PackageName: "testpkg",
												Version:     "1.1.1",
											}),
										},
										{
											Type:  property.TypeBundleObject,
											Value: json.RawMessage(fmt.Sprintf(`{"data":"%s"}`, base64.StdEncoding.EncodeToString([]byte(`{"kind":"ClusterServiceVersion"}`)))),
										},
									},
								},
								"non-head": {
									Package: &model.Package{Name: "testpkg"},
									Channel: &model.Channel{Name: "stable"},
									Name:    "non-head",
									Image:   "non-head:image",
									Properties: []property.Property{
										{
											Type: property.TypePackage,
											Value: requireMarshal(&property.Package{
												PackageName: "testpkg",
												Version:     "1.1.0",
											}),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := bytes.Buffer{}
			require.NoError(t, declcfg.WriteJSON(declcfg.ConvertFromModel(tt.config), &b))
			testFs := fstest.MapFS{
				"catalog.json": {
					Data: b.Bytes(),
					Mode: 0755,
				},
			}

			err := Validate(testFs)

			if tt.wantErr == nil {
				require.NoError(t, err)
				return
			}
			require.EqualError(t, err, tt.wantErr.Error())
		})

	}

}

func TestValidatePackageManifest(t *testing.T) {
	tests := []struct {
		name    string
		model   model.Model
		wantErr error
	}{
		{
			name: "FailOnNoHeadCSV",
			model: model.Model{
				"pkg": {
					Name: "pkg",
					Channels: map[string]*model.Channel{
						"alpha": {
							Name: "alpha",
							Bundles: map[string]*model.Bundle{
								"a": {
									Name: "a",
								},
							},
						},
					},
				},
			},
			wantErr: fmt.Errorf("missing head CSV on package pkg, channel alpha head a: ensure valid csv under 'olm.bundle.object' properties"),
		},
		{
			name: "FailOnInvalidHeadCSV",
			model: model.Model{
				"pkg": {
					Name: "pkg",
					Channels: map[string]*model.Channel{
						"alpha": {
							Name: "alpha",
							Bundles: map[string]*model.Bundle{
								"a": {
									Name:    "a",
									Package: &model.Package{},
									Channel: &model.Channel{},
									Properties: []property.Property{
										{
											Type:  "olm.package",
											Value: json.RawMessage(`{"packageName":"","version":""}`),
										},
									},
									CsvJSON: `invalid-csv`,
								},
							},
						},
					},
				},
			},
			wantErr: fmt.Errorf("invalid head CSV on package pkg, channel alpha head a: failed to unmarshal any 'olm.bundle.object' property as CSV JSON: invalid character 'i' looking for beginning of value"),
		},
		{
			name: "PassWithHeadCSV",
			model: model.Model{
				"pkg": {
					Name: "pkg",
					Channels: map[string]*model.Channel{
						"alpha": {
							Name: "alpha",
							Bundles: map[string]*model.Bundle{
								"a": {
									Name:     "a",
									Replaces: "aa",
									Package:  &model.Package{},
									Channel:  &model.Channel{},
									Properties: []property.Property{
										{
											Type:  "olm.package",
											Value: json.RawMessage(`{"packageName":"","version":""}`),
										},
									},
									CsvJSON: `{}`,
								},
								"aa": {
									Name: "aa",
								},
							},
						},
						"beta": {
							Name: "beta",
							Bundles: map[string]*model.Bundle{
								"b": {
									Name:    "b",
									Package: &model.Package{},
									Channel: &model.Channel{},
									Properties: []property.Property{
										{
											Type:  "olm.package",
											Value: json.RawMessage(`{"packageName":"","version":""}`),
										},
									},
									CsvJSON: `{}`,
								},
							},
						},
					},
				},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePackageManifest(tt.model)
			if tt.wantErr == nil {
				require.NoError(t, err)
				return
			}
			require.EqualError(t, err, tt.wantErr.Error())
		})
	}
}
