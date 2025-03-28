/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1 "k8s.io/api/storage/v1"
	storagev1 "k8s.io/client-go/applyconfigurations/storage/v1"
	gentype "k8s.io/client-go/gentype"
	typedstoragev1 "k8s.io/client-go/kubernetes/typed/storage/v1"
)

// fakeCSIDrivers implements CSIDriverInterface
type fakeCSIDrivers struct {
	*gentype.FakeClientWithListAndApply[*v1.CSIDriver, *v1.CSIDriverList, *storagev1.CSIDriverApplyConfiguration]
	Fake *FakeStorageV1
}

func newFakeCSIDrivers(fake *FakeStorageV1) typedstoragev1.CSIDriverInterface {
	return &fakeCSIDrivers{
		gentype.NewFakeClientWithListAndApply[*v1.CSIDriver, *v1.CSIDriverList, *storagev1.CSIDriverApplyConfiguration](
			fake.Fake,
			"",
			v1.SchemeGroupVersion.WithResource("csidrivers"),
			v1.SchemeGroupVersion.WithKind("CSIDriver"),
			func() *v1.CSIDriver { return &v1.CSIDriver{} },
			func() *v1.CSIDriverList { return &v1.CSIDriverList{} },
			func(dst, src *v1.CSIDriverList) { dst.ListMeta = src.ListMeta },
			func(list *v1.CSIDriverList) []*v1.CSIDriver { return gentype.ToPointerSlice(list.Items) },
			func(list *v1.CSIDriverList, items []*v1.CSIDriver) { list.Items = gentype.FromPointerSlice(items) },
		),
		fake,
	}
}
