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
	"log/slog"

	snapshotv1 "github.com/kubernetes-csi/external-snapshotter/client/v6/apis/volumesnapshot/v1"
	"github.com/slok/kubewebhook/v2/pkg/model"
	kwhmutating "github.com/slok/kubewebhook/v2/pkg/webhook/mutating"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"

	"github.com/deckhouse/sds-common-lib/slogh"
)

func VolumeSnapshotMutate(ctx context.Context, _ *model.AdmissionReview, obj metav1.Object) (*kwhmutating.MutatorResult, error) {
	log := slog.New(slogh.NewHandler(slogh.Config{}))

	log.Debug("VolumeSnapshotMutate called")
	snapshot, ok := obj.(*snapshotv1.VolumeSnapshot)
	if !ok {
		return &kwhmutating.MutatorResult{}, nil
	}

	log.Info("VolumeSnapshotMutate: object is VolumeSnapshot", "name", snapshot.Name, "namespace", snapshot.Namespace)

	if snapshot.Spec.VolumeSnapshotClassName != nil {
		log.Info("VolumeSnapshotMutate: VolumeSnapshotClassName is already set", "name", *snapshot.Spec.VolumeSnapshotClassName)
		return &kwhmutating.MutatorResult{}, nil
	}

	client, err := NewKubeClient("")
	if err != nil {
		log.Error("VolumeSnapshotMutate: failed to create kube client", "error", err)
		return &kwhmutating.MutatorResult{}, err
	}

	namespace := snapshot.ObjectMeta.Namespace
	pvcName := snapshot.Spec.Source.PersistentVolumeClaimName

	pvc := &corev1.PersistentVolumeClaim{}
	err = client.Get(ctx, types.NamespacedName{Name: *pvcName, Namespace: namespace}, pvc)
	if err != nil {
		log.Error("VolumeSnapshotMutate: failed to get PVC", "name", *pvcName, "namespace", namespace, "error", err)
		return &kwhmutating.MutatorResult{}, err
	}

	log.Info("VolumeSnapshotMutate: found PVC", "name", pvc.Name, "namespace", pvc.Namespace, "storageClassName", pvc.Spec.StorageClassName)

	sc := &storagev1.StorageClass{}
	err = client.Get(ctx, types.NamespacedName{Name: *pvc.Spec.StorageClassName, Namespace: namespace}, sc)
	if err != nil {
		log.Error("VolumeSnapshotMutate: failed to get StorageClass", "name", *pvc.Spec.StorageClassName, "namespace", namespace, "error", err)
		return &kwhmutating.MutatorResult{}, err
	}

	log.Info("VolumeSnapshotMutate: found StorageClass", "name", sc.Name, "provisioner", sc.Provisioner)

	vscList := &snapshotv1.VolumeSnapshotClassList{}
	err = client.List(ctx, vscList)
	if err != nil {
		log.Error("VolumeSnapshotMutate: failed to list VolumeSnapshotClasses", "error", err)
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
			log.Info("VolumeSnapshotMutate: found matching VolumeSnapshotClass", "name", vscItem.Name, "driver", vscItem.Driver)
			return &kwhmutating.MutatorResult{
				MutatedObject: snapshot,
			}, nil
		}
	}

	if len(vscDict[sc.Provisioner]) == 1 {
		snapshot.Spec.VolumeSnapshotClassName = &vscDict[sc.Provisioner][0]
		log.Info("VolumeSnapshotMutate: setting VolumeSnapshotClassName to the only available class", "name", snapshot.Spec.VolumeSnapshotClassName)
		return &kwhmutating.MutatorResult{
			MutatedObject: snapshot,
		}, nil
	}

	log.Info("VolumeSnapshotMutate: no matching VolumeSnapshotClass found", "availableClasses", vscDict[sc.Provisioner])

	return &kwhmutating.MutatorResult{
		MutatedObject: snapshot,
	}, nil
}
