/*
Copyright 2025 Flant JSC

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

package handlers

import (
	"context"

	snapshotv1 "github.com/kubernetes-csi/external-snapshotter/client/v6/apis/volumesnapshot/v1"
	"github.com/slok/kubewebhook/v2/pkg/model"
	kwhmutating "github.com/slok/kubewebhook/v2/pkg/webhook/mutating"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	storagev1 "k8s.io/api/storage/v1"
	corev1 "k8s.io/api/core/v1"
	types "k8s.io/apimachinery/pkg/types"
)

func VolumeSnapshotMutate(ctx context.Context, _ *model.AdmissionReview, obj metav1.Object) (*kwhmutating.MutatorResult, error) {
	snapshot, ok := obj.(*snapshotv1.VolumeSnapshot)
	if !ok {
		return &kwhmutating.MutatorResult{}, nil
	}

	if snapshot.Spec.VolumeSnapshotClassName != nil {		
		return &kwhmutating.MutatorResult{}, nil
	}

	client, err := NewKubeClient("")
	if err != nil {
		return &kwhmutating.MutatorResult{}, err
	}

	namespace := snapshot.ObjectMeta.Namespace
	pvcName := snapshot.Spec.Source.PersistentVolumeClaimName

	pvc := &corev1.PersistentVolumeClaim{}
	err = client.Get(ctx, types.NamespacedName{Name: *pvcName, Namespace: namespace}, pvc)
	if err != nil {
		return &kwhmutating.MutatorResult{}, err
	}

	sc := &storagev1.StorageClass{}
	err = client.Get(ctx, types.NamespacedName{Name: *pvc.Spec.StorageClassName, Namespace: namespace}, sc)
	if err != nil {
		return &kwhmutating.MutatorResult{}, err
	}

	vscList := &snapshotv1.VolumeSnapshotClassList{}
	err = client.List(ctx, vscList)
	if err != nil {
		return &kwhmutating.MutatorResult{}, err
	}

	vscDict := make(map[string][]string)
	for _, vscItem := range vscList.Items {
		if vscDict[vscItem.Driver] == nil {
			vscDict[vscItem.Driver] = []string{}
		}
		vscDict[vscItem.Driver] = append(vscDict[vscItem.Driver], vscItem.Name)

		if vscItem.Parameters["storageClassName"] == *pvc.Spec.StorageClassName {
			snapshot.Spec.VolumeSnapshotClassName = &vscItem.Name
			return &kwhmutating.MutatorResult{
				MutatedObject: snapshot,
			}, nil
		}
	}

	if len(vscDict[sc.Provisioner]) == 1 {
		snapshot.Spec.VolumeSnapshotClassName = &vscDict[sc.Provisioner][0]
		return &kwhmutating.MutatorResult{
			MutatedObject: snapshot,
		}, nil
	}

	return &kwhmutating.MutatorResult{
		MutatedObject: snapshot,
	}, nil
}
