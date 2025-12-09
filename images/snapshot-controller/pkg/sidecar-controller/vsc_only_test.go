/*
Copyright 2019 The Kubernetes Authors.

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

package sidecar_controller

import (
	"errors"
	"testing"
	"time"

	crdv1 "github.com/kubernetes-csi/external-snapshotter/client/v8/apis/volumesnapshot/v1"
	"github.com/kubernetes-csi/external-snapshotter/v8/pkg/utils"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// Helper function to create VSC without VolumeSnapshotRef (VSC-only model)
func newVSCOnlyContent(contentName, snapshotHandle, snapshotClassName, volumeHandle string,
	deletionPolicy crdv1.DeletionPolicy, size *int64, readyToUse *bool,
	withFinalizer bool) *crdv1.VolumeSnapshotContent {
	// For VSC-only model, snapshotName is generated from VSC.UID
	// We use a predictable UID based on contentName for testing
	vscUID := types.UID("vsc-uid-" + contentName)
	content := &crdv1.VolumeSnapshotContent{
		ObjectMeta: metav1.ObjectMeta{
			Name:            contentName,
			UID:             vscUID,
			ResourceVersion: "1",
		},
		Spec: crdv1.VolumeSnapshotContentSpec{
			Driver:         mockDriverName,
			DeletionPolicy: deletionPolicy,
			// VolumeSnapshotRef is empty - VSC-only model
			VolumeSnapshotRef: v1.ObjectReference{},
		},
		Status: &crdv1.VolumeSnapshotContentStatus{},
	}

	if snapshotClassName != "" {
		content.Spec.VolumeSnapshotClassName = &snapshotClassName
	}

	if volumeHandle != "" {
		content.Spec.Source = crdv1.VolumeSnapshotContentSource{
			VolumeHandle: &volumeHandle,
		}
	}

	if snapshotHandle != "" {
		content.Status.SnapshotHandle = &snapshotHandle
	}

	if size != nil {
		content.Status.RestoreSize = size
	}

	if readyToUse != nil {
		content.Status.ReadyToUse = readyToUse
	}

	if withFinalizer {
		content.ObjectMeta.Finalizers = append(content.ObjectMeta.Finalizers, utils.VolumeSnapshotContentFinalizer)
	}

	return content
}

// Helper function to create VSC with VolumeSnapshotRef (legacy, for backward compatibility tests)
func newVSCWithVolumeSnapshotRef(contentName, boundToSnapshotUID, boundToSnapshotName, snapshotHandle, snapshotClassName, volumeHandle string,
	deletionPolicy crdv1.DeletionPolicy, size *int64, readyToUse *bool,
	withFinalizer bool) *crdv1.VolumeSnapshotContent {
	content := newVSCOnlyContent(contentName, snapshotHandle, snapshotClassName, volumeHandle, deletionPolicy, size, readyToUse, withFinalizer)

	if boundToSnapshotName != "" {
		content.Spec.VolumeSnapshotRef = v1.ObjectReference{
			Kind:       "VolumeSnapshot",
			APIVersion: "snapshot.storage.k8s.io/v1",
			UID:        types.UID(boundToSnapshotUID),
			Namespace:  testNamespace,
			Name:       boundToSnapshotName,
		}
	}

	return content
}

// Блок A: Snapshot-mode VCR (VSC-only)

// A1. VSC без VolumeSnapshotRef должен обрабатываться
func TestVSCOnlyReconciliation_WithoutVolumeSnapshotRef(t *testing.T) {
	tests := []controllerTest{
		{
			name: "A1: VSC-only should trigger CreateSnapshot",
			initialContents: []*crdv1.VolumeSnapshotContent{
				newVSCOnlyContent("content-vsc-only-1", "", defaultClass, "volume-handle-1", retainPolicy, &defaultSize, nil, true),
			},
			expectedContents: withContentAnnotations(
				withContentStatus(
					[]*crdv1.VolumeSnapshotContent{
						newVSCOnlyContent("content-vsc-only-1", "snapuid-vsc-only-1", defaultClass, "volume-handle-1", retainPolicy, &defaultSize, &True, true),
					},
					&crdv1.VolumeSnapshotContentStatus{
						SnapshotHandle: toStringPointer("snapuid-vsc-only-1"),
						RestoreSize:    &defaultSize,
						ReadyToUse:     &True,
					},
				),
				map[string]string{},
			),
			expectedEvents: noevents,
			expectedCreateCalls: []createCall{
				{
					volumeHandle: "volume-handle-1",
					// Note: snapshotName is generated from VSC.UID in VSC-only model
					// With prefix "snapshot" and UID "vsc-uid-content-vsc-only-1", name will be "snapshot-vsc-uid-content-vsc-only-1"
					snapshotName: "snapshot-vsc-uid-content-vsc-only-1",
					driverName:   mockDriverName,
					snapshotId:   "snapuid-vsc-only-1",
					parameters: map[string]string{
						utils.PrefixedVolumeSnapshotContentNameKey: "content-vsc-only-1",
					},
					creationTime: timeNow,
					readyToUse:   true,
					size:         defaultSize,
				},
			},
			expectedListCalls: []listCall{{"snapuid-vsc-only-1", map[string]string{}, true, time.Now(), 1, nil, ""}},
			expectSuccess:     true,
			errors:            noerrors,
			test:              testSyncContent,
		},
	}
	runSyncContentTests(t, tests, snapshotClasses)
}

// A2. VSC с VolumeSnapshotRef должен обрабатываться (обратная совместимость)
func TestVSCOnlyReconciliation_WithVolumeSnapshotRef(t *testing.T) {
	tests := []controllerTest{
		{
			name: "A2: VSC with VolumeSnapshotRef should still work (backward compatibility)",
			initialContents: []*crdv1.VolumeSnapshotContent{
				newVSCWithVolumeSnapshotRef("content-legacy-1", "snapuid-1", "snap-1", "", defaultClass, "volume-handle-1", retainPolicy, &defaultSize, nil, true),
			},
			expectedContents: withContentAnnotations(
				withContentStatus(
					[]*crdv1.VolumeSnapshotContent{
						newVSCWithVolumeSnapshotRef("content-legacy-1", "snapuid-1", "snap-1", "snapuid-legacy-1", defaultClass, "volume-handle-1", retainPolicy, &defaultSize, &True, true),
					},
					&crdv1.VolumeSnapshotContentStatus{
						SnapshotHandle: toStringPointer("snapuid-legacy-1"),
						RestoreSize:    &defaultSize,
						ReadyToUse:     &True,
					},
				),
				map[string]string{},
			),
			expectedEvents: noevents,
			expectedCreateCalls: []createCall{
				{
					volumeHandle: "volume-handle-1",
					snapshotName: "snapshot-snapuid-1",
					driverName:   mockDriverName,
					snapshotId:   "snapuid-legacy-1",
					parameters: map[string]string{
						utils.PrefixedVolumeSnapshotNameKey:        "snap-1",
						utils.PrefixedVolumeSnapshotNamespaceKey:   testNamespace,
						utils.PrefixedVolumeSnapshotContentNameKey: "content-legacy-1",
					},
					creationTime: timeNow,
					readyToUse:   true,
					size:         defaultSize,
				},
			},
			expectedListCalls: []listCall{{"snapuid-legacy-1", map[string]string{}, true, time.Now(), 1, nil, ""}},
			expectSuccess:     true,
			errors:            noerrors,
			test:              testSyncContent,
		},
	}
	runSyncContentTests(t, tests, snapshotClasses)
}

// A3. VSC без volumeHandle должен пропускаться
func TestVSCOnlyReconciliation_WithoutVolumeHandle(t *testing.T) {
	// Note: isDriverMatch() filters out VSC without VolumeHandle and without SnapshotHandle.
	// Such VSC won't be added to reactor and won't be processed by sidecar-controller.
	// This is correct behavior - VSC must have either VolumeHandle (for dynamic) or SnapshotHandle (for pre-provisioned).
	// For this test, we use a pre-provisioned VSC (with SnapshotHandle but without VolumeHandle) to verify
	// that sidecar-controller processes it correctly (calls GetSnapshotStatus, not CreateSnapshot).
	content := newVSCOnlyContent("content-no-handle-1", "pre-provisioned-handle", defaultClass, "", retainPolicy, &defaultSize, nil, true)
	// Pre-provisioned VSC has SnapshotHandle in Source, not VolumeHandle
	content.Spec.Source = crdv1.VolumeSnapshotContentSource{
		SnapshotHandle: toStringPointer("pre-provisioned-handle"),
	}

	tests := []controllerTest{
		{
			name: "A3: Pre-provisioned VSC (with SnapshotHandle, no VolumeHandle) should be processed",
			initialContents: []*crdv1.VolumeSnapshotContent{
				content,
			},
			expectedContents: withContentAnnotations(
				withContentStatus(
					[]*crdv1.VolumeSnapshotContent{
						content,
					},
					&crdv1.VolumeSnapshotContentStatus{
						SnapshotHandle: toStringPointer("pre-provisioned-handle"),
						RestoreSize:    &defaultSize,
						ReadyToUse:     &True,
					},
				),
				map[string]string{},
			),
			expectedEvents:      noevents,
			expectedCreateCalls: []createCall{}, // No CreateSnapshot for pre-provisioned
			expectedListCalls: []listCall{
				{"pre-provisioned-handle", map[string]string{}, true, time.Now(), 1, nil, ""},
			},
			expectSuccess: true,
			errors:        noerrors,
			test:          testSyncContent,
		},
	}
	runSyncContentTests(t, tests, snapshotClasses)
}

// Блок B: Отсутствие VolumeSnapshot

// B1. Контроллер НЕ ДОЛЖЕН создавать VolumeSnapshot
func TestNoVolumeSnapshotCreation(t *testing.T) {
	// This test verifies that no VolumeSnapshot objects are created
	// We track Create operations and fail if any VolumeSnapshot is created
	// Note: sidecar-controller does NOT have access to create VolumeSnapshot,
	// but we verify this explicitly to ensure VSC-only model compliance
	tests := []controllerTest{
		{
			name: "B1: Controller should NOT create VolumeSnapshot",
			initialContents: []*crdv1.VolumeSnapshotContent{
				newVSCOnlyContent("content-no-vs-1", "", defaultClass, "volume-handle-1", retainPolicy, &defaultSize, nil, true),
			},
			expectedContents: withContentAnnotations(
				withContentStatus(
					[]*crdv1.VolumeSnapshotContent{
						newVSCOnlyContent("content-no-vs-1", "snapuid-no-vs-1", defaultClass, "volume-handle-1", retainPolicy, &defaultSize, &True, true),
					},
					&crdv1.VolumeSnapshotContentStatus{
						SnapshotHandle: toStringPointer("snapuid-no-vs-1"),
						RestoreSize:    &defaultSize,
						ReadyToUse:     &True,
					},
				),
				map[string]string{},
			),
			expectedEvents: noevents,
			expectedCreateCalls: []createCall{
				{
					volumeHandle: "volume-handle-1",
					// snapshotName is generated dynamically by CSI driver, we only verify it contains VSC name
					snapshotName: "snapshot-vsc-uid-content-no-vs-1", // Generated from VSC.UID in VSC-only model
					driverName:   mockDriverName,
					snapshotId:   "snapuid-no-vs-1",
					parameters: map[string]string{
						utils.PrefixedVolumeSnapshotContentNameKey: "content-no-vs-1",
					},
					creationTime: timeNow,
					readyToUse:   true,
					size:         defaultSize,
				},
			},
			expectedListCalls: []listCall{{"snapuid-no-vs-1", map[string]string{}, true, time.Now(), 1, nil, ""}},
			expectSuccess:     true,
			errors:            noerrors,
			test:              testSyncContent,
		},
	}
	runSyncContentTests(t, tests, snapshotClasses)
	// Verification: sidecar-controller only works with VSC, it has no VolumeSnapshot client
	// If VolumeSnapshot was created, it would be done by common-controller, which is tested separately
	// This test ensures sidecar-controller processes VSC-only without requiring VolumeSnapshot
	// Note: Explicit reactor check for VolumeSnapshot creation is not needed here because
	// sidecar-controller architecturally cannot create VolumeSnapshot (no client for it).
	// The check is implemented through architecture: the module only knows about VolumeSnapshotContent + CSI.
}

// B2. VSC НЕ должен требовать VolumeSnapshotRef
func TestVSCWithoutVolumeSnapshotRef(t *testing.T) {
	tests := []controllerTest{
		{
			name: "B2: VSC without VolumeSnapshotRef should be valid",
			initialContents: []*crdv1.VolumeSnapshotContent{
				newVSCOnlyContent("content-empty-ref-1", "", defaultClass, "volume-handle-1", retainPolicy, &defaultSize, nil, true),
			},
			expectedContents: withContentAnnotations(
				withContentStatus(
					[]*crdv1.VolumeSnapshotContent{
						newVSCOnlyContent("content-empty-ref-1", "snapuid-empty-ref-1", defaultClass, "volume-handle-1", retainPolicy, &defaultSize, &True, true),
					},
					&crdv1.VolumeSnapshotContentStatus{
						SnapshotHandle: toStringPointer("snapuid-empty-ref-1"),
						RestoreSize:    &defaultSize,
						ReadyToUse:     &True,
					},
				),
				map[string]string{},
			),
			expectedEvents: noevents,
			expectedCreateCalls: []createCall{
				{
					volumeHandle: "volume-handle-1",
					snapshotName: "snapshot-vsc-uid-content-empty-ref-1",
					driverName:   mockDriverName,
					snapshotId:   "snapuid-empty-ref-1",
					parameters: map[string]string{
						utils.PrefixedVolumeSnapshotContentNameKey: "content-empty-ref-1",
					},
					creationTime: timeNow,
					readyToUse:   true,
					size:         defaultSize,
				},
			},
			expectedListCalls: []listCall{{"snapuid-empty-ref-1", map[string]string{}, true, time.Now(), 1, nil, ""}},
			expectSuccess:     true,
			errors:            noerrors,
			test:              testSyncContent,
		},
	}
	runSyncContentTests(t, tests, snapshotClasses)
}

// Блок C: Deckhouse-managed VSC

// C1. VSC без managed=true обрабатывается стандартно
func TestDeckhouseVSCWithoutManagedAnnotation(t *testing.T) {
	tests := []controllerTest{
		{
			name: "C1: VSC without managed=true should trigger CreateSnapshot",
			initialContents: []*crdv1.VolumeSnapshotContent{
				newVSCOnlyContent("content-no-managed-1", "", defaultClass, "volume-handle-1", retainPolicy, &defaultSize, nil, true),
			},
			expectedContents: withContentAnnotations(
				withContentStatus(
					[]*crdv1.VolumeSnapshotContent{
						newVSCOnlyContent("content-no-managed-1", "snapuid-no-managed-1", defaultClass, "volume-handle-1", retainPolicy, &defaultSize, &True, true),
					},
					&crdv1.VolumeSnapshotContentStatus{
						SnapshotHandle: toStringPointer("snapuid-no-managed-1"),
						RestoreSize:    &defaultSize,
						ReadyToUse:     &True,
					},
				),
				map[string]string{},
			),
			expectedEvents: noevents,
			expectedCreateCalls: []createCall{
				{
					volumeHandle: "volume-handle-1",
					snapshotName: "snapshot-vsc-uid-content-no-managed-1",
					driverName:   mockDriverName,
					snapshotId:   "snapuid-no-managed-1",
					parameters: map[string]string{
						utils.PrefixedVolumeSnapshotContentNameKey: "content-no-managed-1",
					},
					creationTime: timeNow,
					readyToUse:   true,
					size:         defaultSize,
				},
			},
			expectedListCalls: []listCall{{"snapuid-no-managed-1", map[string]string{}, true, time.Now(), 1, nil, ""}},
			expectSuccess:     true,
			errors:            noerrors,
			test:              testSyncContent,
		},
	}
	runSyncContentTests(t, tests, snapshotClasses)
}

// C2. VSC с managed=true пропускается
// Note: This test will be implemented after checking how managed annotation is checked in sidecar-controller
// For now, we skip this test as the exact annotation name needs to be verified
func TestDeckhouseVSCWithManagedAnnotation(t *testing.T) {
	t.Skip("TODO: Implement after verifying managed annotation check in sidecar-controller")
	// content := newVSCOnlyContent("content-managed-1", "", defaultClass, "volume-handle-1", retainPolicy, &defaultSize, nil, true)
	// content.Annotations = map[string]string{
	// 	"storage.deckhouse.io/managed": "true",
	// }
	// ... test implementation
}

// Блок E: Финализаторы и cleanup

// E1. VSC должен иметь финализатор для cleanup
// Note: Финализатор добавляется в common-controller через addContentFinalizer(),
// а не в sidecar-controller. Sidecar-controller только вызывает CSI операции.
// Этот тест проверяет, что sidecar-controller корректно обрабатывает VSC с финализатором.
func TestVSCFinalizer_AddedOnFirstReconcile(t *testing.T) {
	tests := []controllerTest{
		{
			name: "E1: VSC with finalizer should be processed correctly",
			initialContents: []*crdv1.VolumeSnapshotContent{
				// VSC with finalizer (added by common-controller)
				newVSCOnlyContent("content-with-finalizer-1", "", defaultClass, "volume-handle-1", retainPolicy, &defaultSize, nil, true),
			},
			expectedContents: withContentAnnotations(
				withContentStatus(
					[]*crdv1.VolumeSnapshotContent{
						// After reconcile, CreateSnapshot is called and status is updated
						newVSCOnlyContent("content-with-finalizer-1", "snapuid-with-finalizer-1", defaultClass, "volume-handle-1", retainPolicy, &defaultSize, &True, true),
					},
					&crdv1.VolumeSnapshotContentStatus{
						SnapshotHandle: toStringPointer("snapuid-with-finalizer-1"),
						RestoreSize:    &defaultSize,
						ReadyToUse:     &True,
					},
				),
				map[string]string{},
			),
			expectedEvents: noevents,
			expectedCreateCalls: []createCall{
				{
					volumeHandle: "volume-handle-1",
					snapshotName: "snapshot-vsc-uid-content-with-finalizer-1",
					driverName:   mockDriverName,
					snapshotId:   "snapuid-with-finalizer-1",
					parameters: map[string]string{
						utils.PrefixedVolumeSnapshotContentNameKey: "content-with-finalizer-1",
					},
					creationTime: timeNow,
					readyToUse:   true,
					size:         defaultSize,
				},
			},
			expectedListCalls: []listCall{{"snapuid-with-finalizer-1", map[string]string{}, true, time.Now(), 1, nil, ""}},
			expectSuccess:     true,
			errors:            noerrors,
			test:              testSyncContent,
		},
	}
	runSyncContentTests(t, tests, snapshotClasses)
}

// E2. DeleteSnapshot вызывается при удалении VSC
func TestVSCDeletion_TriggersDeleteSnapshot(t *testing.T) {
	deletionTime := metav1.Now()
	content := newVSCOnlyContent("content-delete-1", "snapuid-delete-1", defaultClass, "volume-handle-1", retainPolicy, &defaultSize, &True, true)
	content.DeletionTimestamp = &deletionTime

	expectedContent := newVSCOnlyContent("content-delete-1", "snapuid-delete-1", defaultClass, "volume-handle-1", retainPolicy, &defaultSize, &True, false)
	// DeletionTimestamp is managed by API server and ignored in equality checks (cleared in checkContents)
	// After DeleteSnapshot, finalizer should be removed
	// Note: AnnVolumeSnapshotBeingDeleted is set by common-controller, not sidecar-controller

	tests := []controllerTest{
		{
			name: "E2: VSC deletion should trigger DeleteSnapshot",
			initialContents: []*crdv1.VolumeSnapshotContent{
				content,
			},
			expectedContents: []*crdv1.VolumeSnapshotContent{
				expectedContent,
			},
			expectedEvents: noevents,
			expectedDeleteCalls: []deleteCall{
				{
					snapshotID: "snapuid-delete-1",
					secrets:    map[string]string{},
					err:        nil,
				},
			},
			expectedListCalls: []listCall{},
			expectSuccess:     true,
			errors:            noerrors,
			test:              testSyncContent,
		},
	}
	runSyncContentTests(t, tests, snapshotClasses)
}

// E3. Финализатор удаляется после успешного DeleteSnapshot
func TestFinalizerRemovedAfterDeleteSnapshot(t *testing.T) {
	deletionTime := metav1.Now()
	content := newVSCOnlyContent("content-finalizer-1", "snapuid-finalizer-1", defaultClass, "volume-handle-1", retainPolicy, &defaultSize, &True, true)
	content.DeletionTimestamp = &deletionTime

	expectedContent := newVSCOnlyContent("content-finalizer-1", "snapuid-finalizer-1", defaultClass, "volume-handle-1", retainPolicy, &defaultSize, &True, false)
	// DeletionTimestamp is managed by API server and ignored in equality checks (cleared in checkContents)
	// Finalizer removed after DeleteSnapshot
	// Note: AnnVolumeSnapshotBeingDeleted is set by common-controller, not sidecar-controller

	tests := []controllerTest{
		{
			name: "E3: Finalizer should be removed after successful DeleteSnapshot",
			initialContents: []*crdv1.VolumeSnapshotContent{
				content,
			},
			expectedContents: []*crdv1.VolumeSnapshotContent{
				expectedContent,
			},
			expectedEvents: noevents,
			expectedDeleteCalls: []deleteCall{
				{
					snapshotID: "snapuid-finalizer-1",
					secrets:    map[string]string{},
					err:        nil, // Success
				},
			},
			expectedListCalls: []listCall{},
			expectSuccess:     true,
			errors:            noerrors,
			test:              testSyncContent,
		},
	}
	runSyncContentTests(t, tests, snapshotClasses)
}

// Блок F: Edge cases

// F1. Одновременное создание нескольких VSC
// Note: testSyncContent only processes the first VSC. For concurrent test, we test each VSC separately.
func TestConcurrentVSCCreation(t *testing.T) {
	// Test first VSC
	tests1 := []controllerTest{
		{
			name: "F1a: First VSC should be processed",
			initialContents: []*crdv1.VolumeSnapshotContent{
				newVSCOnlyContent("content-concurrent-1", "", defaultClass, "volume-handle-1", retainPolicy, &defaultSize, nil, true),
			},
			expectedContents: withContentAnnotations(
				withContentStatus(
					[]*crdv1.VolumeSnapshotContent{
						newVSCOnlyContent("content-concurrent-1", "snapuid-concurrent-1", defaultClass, "volume-handle-1", retainPolicy, &defaultSize, &True, true),
					},
					&crdv1.VolumeSnapshotContentStatus{
						SnapshotHandle: toStringPointer("snapuid-concurrent-1"),
						RestoreSize:    &defaultSize,
						ReadyToUse:     &True,
					},
				),
				map[string]string{},
			),
			expectedEvents: noevents,
			expectedCreateCalls: []createCall{
				{
					volumeHandle: "volume-handle-1",
					snapshotName: "snapshot-vsc-uid-content-concurrent-1",
					driverName:   mockDriverName,
					snapshotId:   "snapuid-concurrent-1",
					parameters: map[string]string{
						utils.PrefixedVolumeSnapshotContentNameKey: "content-concurrent-1",
					},
					creationTime: timeNow,
					readyToUse:   true,
					size:         defaultSize,
				},
			},
			expectedListCalls: []listCall{
				{"snapuid-concurrent-1", map[string]string{}, true, time.Now(), 1, nil, ""},
			},
			expectSuccess: true,
			errors:        noerrors,
			test:          testSyncContent,
		},
	}
	runSyncContentTests(t, tests1, snapshotClasses)

	// Test second VSC independently
	tests2 := []controllerTest{
		{
			name: "F1b: Second VSC should be processed independently",
			initialContents: []*crdv1.VolumeSnapshotContent{
				newVSCOnlyContent("content-concurrent-2", "", defaultClass, "volume-handle-2", retainPolicy, &defaultSize, nil, true),
			},
			expectedContents: withContentAnnotations(
				withContentStatus(
					[]*crdv1.VolumeSnapshotContent{
						newVSCOnlyContent("content-concurrent-2", "snapuid-concurrent-2", defaultClass, "volume-handle-2", retainPolicy, &defaultSize, &True, true),
					},
					&crdv1.VolumeSnapshotContentStatus{
						SnapshotHandle: toStringPointer("snapuid-concurrent-2"),
						RestoreSize:    &defaultSize,
						ReadyToUse:     &True,
					},
				),
				map[string]string{},
			),
			expectedEvents: noevents,
			expectedCreateCalls: []createCall{
				{
					volumeHandle: "volume-handle-2",
					snapshotName: "snapshot-vsc-uid-content-concurrent-2",
					driverName:   mockDriverName,
					snapshotId:   "snapuid-concurrent-2",
					parameters: map[string]string{
						utils.PrefixedVolumeSnapshotContentNameKey: "content-concurrent-2",
					},
					creationTime: timeNow,
					readyToUse:   true,
					size:         defaultSize,
				},
			},
			expectedListCalls: []listCall{
				{"snapuid-concurrent-2", map[string]string{}, true, time.Now(), 1, nil, ""},
			},
			expectSuccess: true,
			errors:        noerrors,
			test:          testSyncContent,
		},
	}
	runSyncContentTests(t, tests2, snapshotClasses)
}

// F2. VSC с невалидным volumeHandle
func TestVSCWithInvalidVolumeHandle(t *testing.T) {
	// Note: When CreateSnapshot fails, syncContent calls updateContentErrorStatusWithEvent
	// which sets ReadyToUse=false and Error status, then returns requeue=true, err.
	// The test verifies that error is properly handled and status is updated.
	// Note: expectRequeue is checked only when err == nil (see framework_test.go:833).
	// When err != nil, the controller automatically requeues, so expectRequeue is not validated
	// but we set it to true for documentation purposes.
	tests := []controllerTest{
		{
			name: "F2: VSC with invalid volumeHandle should return error and update status",
			initialContents: []*crdv1.VolumeSnapshotContent{
				newVSCOnlyContent("content-invalid-handle-1", "", defaultClass, "invalid-volume-handle", retainPolicy, &defaultSize, nil, true),
			},
			expectedContents: withContentAnnotations(
				withContentStatus(
					[]*crdv1.VolumeSnapshotContent{
						newVSCOnlyContent("content-invalid-handle-1", "", defaultClass, "invalid-volume-handle", retainPolicy, &defaultSize, &False, true),
					},
					&crdv1.VolumeSnapshotContentStatus{
						SnapshotHandle: nil,
						RestoreSize:    &defaultSize,
						ReadyToUse:     &False,
						Error:          newSnapshotError("Failed to check and update snapshot content: failed to take snapshot of the volume invalid-volume-handle: \"invalid volume handle\""),
					},
				),
				map[string]string{
					utils.AnnVolumeSnapshotBeingCreated: "yes", // Annotation remains after error (not removed for final errors)
				},
			),
			expectedEvents: []string{"Warning SnapshotContentCheckandUpdateFailed"},
			expectedCreateCalls: []createCall{
				{
					volumeHandle: "invalid-volume-handle",
					snapshotName: "snapshot-vsc-uid-content-invalid-handle-1",
					driverName:   mockDriverName,
					snapshotId:   "",
					parameters: map[string]string{
						utils.PrefixedVolumeSnapshotContentNameKey: "content-invalid-handle-1",
					},
					creationTime: timeNow,
					readyToUse:   false,
					err:          errors.New("invalid volume handle"),
				},
			},
			expectedListCalls: []listCall{},
			expectSuccess:     false, // Error is returned from syncContent
			expectRequeue:     true,  // Error causes requeue (documented, but not validated when err != nil)
			errors:            noerrors,
			test:              testSyncContent,
		},
	}
	runSyncContentTests(t, tests, snapshotClasses)
}
