/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package storageclass

import (
	"context"
	"strconv"

	"github.com/pkg/errors"
	v1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	storagev1 "k8s.io/client-go/kubernetes/typed/storage/v1"
	"k8s.io/minikube/pkg/kapi"
)

func annotateDefaultStorageClass(storage storagev1.StorageV1Interface, class *v1.StorageClass, enable bool) error {
	isDefault := strconv.FormatBool(enable)
	metav1.SetMetaDataAnnotation(&class.ObjectMeta, "storageclass.kubernetes.io/is-default-class", isDefault)
	_, err := storage.StorageClasses().Update(context.Background(), class, metav1.UpdateOptions{})

	return err
}

// DisableDefaultStorageClass disables the default storage class provisioner
// The addon-manager and kubectl apply cannot delete storageclasses
func DisableDefaultStorageClass(storage storagev1.StorageV1Interface, class string) error {
	sc, err := storage.StorageClasses().Get(context.Background(), class, metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "Error getting storage class %s", class)
	}

	err = annotateDefaultStorageClass(storage, sc, false)
	if err != nil {
		return errors.Wrapf(err, "Error marking storage class %s as non-default", class)
	}
	return nil
}

// SetDefaultStorageClass makes sure only the class with @name is marked as
// default.
func SetDefaultStorageClass(storage storagev1.StorageV1Interface, name string) error {
	scList, err := storage.StorageClasses().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "Error listing StorageClasses")
	}
	for _, sc := range scList.Items {
		err = annotateDefaultStorageClass(storage, &sc, sc.Name == name)
		if err != nil {
			isDefault := "non-default"
			if sc.Name == name {
				isDefault = "default"
			}
			return errors.Wrapf(err, "Error while marking storage class %s as %s", sc.Name, isDefault)
		}
	}
	return nil
}

// GetStoragev1 return storage v1 interface for client
func GetStoragev1(context string) (storagev1.StorageV1Interface, error) {
	client, err := kapi.Client(context)
	if err != nil {
		return nil, err
	}
	sv1 := client.StorageV1()
	return sv1, nil
}
