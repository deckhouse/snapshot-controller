---
title: "Модуль snapshot-controller: примеры конфигурации"
---

## Использование снимков

Для использования снимков укажите [VolumeSnapshotClass](/modules/snapshot-controller/cr.html#volumesnapshotclass).
Чтобы получить список доступных [VolumeSnapshotClass](/modules/snapshot-controller/cr.html#volumesnapshotclass) в кластере, выполните:

```bash
d8 k get volumesnapshotclasses.snapshot.storage.k8s.io
```

Используйте [VolumeSnapshotClass](/modules/snapshot-controller/cr.html#volumesnapshotclass) для создания снимка из существующего PersistentVolumeClaim (PVC):

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

Спустя небольшой промежуток времени проверьте, что снимок готов:

```bash
d8 k describe volumesnapshots.snapshot.storage.k8s.io my-first-snapshot
```

Пример вывода:

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

Восстановите содержимое снимка, создав новый PVC и указав снимок в качестве источника:

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

## Клонирование CSI-томов

Также можно клонировать Persistent Volume (PV) на основе концепции снимков, а именно существующие PVC.
Обратите внимание, что спецификация CSI имеет ограничения при клонировании PVC в неймспейсах и StorageClass, отличных от исходного PVC.
См. [документацию Kubernetes](https://kubernetes.io/docs/concepts/storage/volume-pvc-datasource/) для подробностей.

Для клонирования тома создайте новый PVC и укажите исходный PVC в `dataSource`:

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
