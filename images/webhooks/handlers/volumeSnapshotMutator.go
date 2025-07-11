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
	"errors"
	"fmt"
	"log/slog"

	snapshotv1 "github.com/kubernetes-csi/external-snapshotter/client/v6/apis/volumesnapshot/v1"
	"github.com/slok/kubewebhook/v2/pkg/model"
	kwhmutating "github.com/slok/kubewebhook/v2/pkg/webhook/mutating"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	sv1 "k8s.io/api/storage/v1"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	d8commonapi "github.com/deckhouse/sds-common-lib/api/v1alpha1"
	"github.com/deckhouse/sds-common-lib/kubeclient"
	"github.com/deckhouse/sds-common-lib/slogh"
)

const (
	storageClassVolumeSnapshotAnnotationName = "storage.deckhouse.io/volumesnapshotclass"
	storageClassManagedbyLabelName           = "storage.deckhouse.io/managed-by"
)

func VolumeSnapshotMutate(ctx context.Context, _ *model.AdmissionReview, obj metav1.Object) (*kwhmutating.MutatorResult, error) {
	log := slog.New(slogh.NewHandler(slogh.Config{}))

	log.Debug("VolumeSnapshotMutate called")
	snapshot, ok := obj.(*snapshotv1.VolumeSnapshot)
	if !ok {
		return &kwhmutating.MutatorResult{}, nil
	}

	log.Info("VolumeSnapshotMutate: object is VolumeSnapshot", "name", snapshot.Name, "namespace", snapshot.Namespace)

	client, err := kubeclient.New(d8commonapi.AddToScheme,
		corev1.AddToScheme,
		storagev1.AddToScheme,
		snapshotv1.AddToScheme,
		clientgoscheme.AddToScheme,
		extv1.AddToScheme,
		v1.AddToScheme,
		sv1.AddToScheme,
	)
	if err != nil {
		log.Error("VolumeSnapshotMutate: failed to create kube client", "error", err)
		return &kwhmutating.MutatorResult{}, err
	}

	if snapshot.Spec.Source.PersistentVolumeClaimName == nil {
		log.Warn("VolumeSnapshotMutate: VolumeSnapshot PersistentVolumeClaimName is nil. Skipping mutation", "snapshot", snapshot.Name)
		return &kwhmutating.MutatorResult{}, nil
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

	if pvc.Spec.StorageClassName == nil {
		log.Error("VolumeSnapshotMutate: PVC StorageClassName is nil", "pvc", pvc.Name)
		return &kwhmutating.MutatorResult{}, errors.New("PVC StorageClassName is nil")
	}

	sc := &storagev1.StorageClass{}
	err = client.Get(ctx, types.NamespacedName{Name: *pvc.Spec.StorageClassName, Namespace: namespace}, sc)
	if err != nil {
		log.Error("VolumeSnapshotMutate: failed to get StorageClass", "name", *pvc.Spec.StorageClassName, "namespace", namespace, "error", err)
		return &kwhmutating.MutatorResult{}, err
	}

	log.Info("VolumeSnapshotMutate: found StorageClass", "name", sc.Name, "provisioner", sc.Provisioner)

	if managedBy, ok := sc.Labels[storageClassManagedbyLabelName]; ok {
		log.Info("VolumeSnapshotMutate: StorageClass is managed by module", "storageClass", sc.Name, "managedBy", managedBy)
		if volumeSnapshotClassName, ok := sc.Annotations[storageClassVolumeSnapshotAnnotationName]; ok {
			if snapshot.Spec.VolumeSnapshotClassName != nil && *snapshot.Spec.VolumeSnapshotClassName != volumeSnapshotClassName {
				log.Error("VolumeSnapshotMutate: if VolumeSnapshotClassName is set, it must match the StorageClass annotation", "snapshotVolumeSnapshotClassName", *snapshot.Spec.VolumeSnapshotClassName, "storageClassAnnotation", volumeSnapshotClassName)
				return &kwhmutating.MutatorResult{}, fmt.Errorf("VolumeSnapshotMutate: if VolumeSnapshotClassName is set, it must match the StorageClass annotation, snapshotVolumeSnapshotClassName %s, storageClassAnnotation %s", *snapshot.Spec.VolumeSnapshotClassName, volumeSnapshotClassName)
			}

			log.Info("VolumeSnapshotMutate: StorageClass has volume snapshot class name annotation, set it in VolumeSnapshot", "name", volumeSnapshotClassName)
			snapshot.Spec.VolumeSnapshotClassName = &volumeSnapshotClassName
			return &kwhmutating.MutatorResult{
				MutatedObject: snapshot,
			}, nil
		} else {
			log.Error("VolumeSnapshotMutate: StorageClass does not have volume snapshot class name annotation", "name", sc.Name)
			return &kwhmutating.MutatorResult{}, errors.New("StorageClass does not have volume snapshot class name annotation")
		}
	} else {
		log.Info("VolumeSnapshotMutate: StorageClass", sc.Name, " not managed by Deckhouse")

		if snapshot.Spec.VolumeSnapshotClassName == nil {
			if volumeSnapshotClassName, ok := sc.Annotations[storageClassVolumeSnapshotAnnotationName]; ok {
				log.Info("VolumeSnapshotMutate: StorageClass has volume snapshot class name annotation, set it in VolumeSnapshot", "name", volumeSnapshotClassName)
				snapshot.Spec.VolumeSnapshotClassName = &volumeSnapshotClassName
				return &kwhmutating.MutatorResult{
					MutatedObject: snapshot,
				}, nil
			}
		}
		return &kwhmutating.MutatorResult{}, nil
	}
}
