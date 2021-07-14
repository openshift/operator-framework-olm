package main

import (
	"reflect"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSeparateConfigMapsIntoNewestAndExpired(t *testing.T) {
	nowTime := metav1.Now()
	futureTime := metav1.NewTime(nowTime.Add(time.Hour))
	var tests = []struct {
		name                         string
		arg, wantNewest, wantExpired []corev1.ConfigMap
	}{
		{
			name:        "empty",
			arg:         []corev1.ConfigMap{},
			wantNewest:  []corev1.ConfigMap{},
			wantExpired: []corev1.ConfigMap{},
		},
		{
			name: "single newest configMap no expired configMaps",
			arg: []corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						GenerateName:      "foo-",
						CreationTimestamp: futureTime,
					},
				},
			},
			wantNewest: []corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						GenerateName:      "foo-",
						CreationTimestamp: futureTime,
					},
				},
			},
			wantExpired: []corev1.ConfigMap{},
		},
		{
			name: "multiple newest configMap no expired configMaps",
			arg: []corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						GenerateName:      "foo-",
						CreationTimestamp: futureTime,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						GenerateName:      "bar-",
						CreationTimestamp: futureTime,
					},
				},
			},
			wantNewest: []corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						GenerateName:      "foo-",
						CreationTimestamp: futureTime,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						GenerateName:      "bar-",
						CreationTimestamp: futureTime,
					},
				},
			},
			wantExpired: []corev1.ConfigMap{},
		},
		{
			name: "single newest configMap one expired configMap",
			arg: []corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						GenerateName:      "foo-",
						CreationTimestamp: futureTime,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						GenerateName:      "foo-",
						CreationTimestamp: nowTime,
					},
				},
			},
			wantNewest: []corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						GenerateName:      "foo-",
						CreationTimestamp: futureTime,
					},
				},
			},
			wantExpired: []corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						GenerateName:      "foo-",
						CreationTimestamp: nowTime,
					},
				},
			},
		},
		{
			name: "multiple newest configMaps and multiple expired configMaps",
			arg: []corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						GenerateName:      "foo-",
						CreationTimestamp: futureTime,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						GenerateName:      "foo-",
						CreationTimestamp: nowTime,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						GenerateName:      "bar-",
						CreationTimestamp: futureTime,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						GenerateName:      "bar-",
						CreationTimestamp: nowTime,
					},
				},
			},
			wantNewest: []corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						GenerateName:      "foo-",
						CreationTimestamp: futureTime,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						GenerateName:      "bar-",
						CreationTimestamp: futureTime,
					},
				},
			},
			wantExpired: []corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						GenerateName:      "foo-",
						CreationTimestamp: nowTime,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						GenerateName:      "bar-",
						CreationTimestamp: nowTime,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotNewest, gotExpired := separateConfigMapsIntoNewestAndExpired(tt.arg)

			if !equal(gotNewest, tt.wantNewest) {
				t.Errorf("separateConfigMapsIntoNewestAndExpired contains unexpected newest CMs. Got %v, want %v", gotNewest, tt.wantNewest)
			}

			if !equal(gotExpired, tt.wantExpired) {
				t.Errorf("separateConfigMapsIntoNewestAndExpired contains unexpected expired CMs. Got %v, want %v", gotExpired, tt.wantExpired)
			}

		})
	}
}

func equal(a, b []corev1.ConfigMap) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		found := false
		for j := range b {
			if reflect.DeepEqual(a[i], b[j]) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
