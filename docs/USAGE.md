---
title: "The snapshot-controller module: configuration examples"
---

### Using snapshots

Specify a [VolumeSnapshotClass](/modules/snapshot-controller/cr.html#volumesnapshotclass) to use snapshots.
To get a list of available [VolumeSnapshotClass](/modules/snapshot-controller/cr.html#volumesnapshotclass) in your cluster, run:

```bash
d8 k get volumesnapshotclasses.snapshot.storage.k8s.io
```

Use [VolumeSnapshotClass](/modules/snapshot-controller/cr.html#volumesnapshotclass) to create a snapshot from an existing PersistentVolumeClaim (PVC):

```yaml
apiVersion: snapshot.storage.k8s.io/v1
kind: VolumeSnapshot
metadata:
  name: my-first-snapshot
spec:
  volumeSnapshotClassName: sds-replicated-volume
  source:
    persistentVolumeClaimName: my-first-volume
```

After a short wait, check that the snapshot is ready:

```bash
d8 k describe volumesnapshots.snapshot.storage.k8s.io my-first-snapshot
```

Example output:

```console
...
Spec:
  Source:
    Persistent Volume Claim Name:  my-first-snapshot
  Volume Snapshot Class Name:      sds-replicated-volume
Status:
  Bound Volume Snapshot Content Name:  snapcontent-b6072ab7-6ddf-482b-a4e3-693088136d2c
  Creation Time:                       2020-06-04T13:02:28Z
  Ready To Use:                        true
  Restore Size:                        500Mi
```

Restore the content of this snapshot by creating a new PVC with the snapshot as source:

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: my-first-volume-from-snapshot
spec:
  storageClassName: sds-replicated-volume-data-r2
  dataSource:
    name: my-first-snapshot
    kind: VolumeSnapshot
    apiGroup: snapshot.storage.k8s.io
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 500Mi
```

### CSI Volume Cloning

You can also clone Persistent Volume (PV) using the snapshot concept (clone existing PVC).
Note that the CSI specification has restrictions when cloning PVCs in different namespaces and StorageClass than the original PVC.
See [Kubernetes documentation](https://kubernetes.io/docs/concepts/storage/volume-pvc-datasource/) for details.

To clone a volume, create a new PVC and specify the origin PVC in the `dataSource`:

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: my-cloned-pvc
spec:
  storageClassName: sds-replicated-volume-data-r2
  dataSource:
    name: my-origin-pvc
    kind: PersistentVolumeClaim
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 500Mi
```
