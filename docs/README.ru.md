---
title: "Модуль snapshot-controller"
---

{% alert level="warning" %}
Модуль **устарел (deprecated)**: поддержку снимков CSI теперь предоставляет модуль
`storage-foundation`. Новые кластеры получают её из `storage-foundation` (автовключение);
`snapshot-controller` больше не включается автоматически ни в одном бандле. Пока
`storage-foundation` включён, этот модуль ничего не разворачивает и только выставляет алерт
`D8SnapshotControllerModuleDeprecated` с просьбой выключить его. Выключите `snapshot-controller`
и используйте `storage-foundation`.
{% endalert %}

Модуль `snapshot-controller` включает поддержку снимков для совместимых CSI-драйверов в кластере Kubernetes.

Список CSI-драйверов в Deckhouse Kubernetes Platform, поддерживающих работу со снимками:

- [cloud-provider-openstack](/modules/cloud-provider-openstack/)
- [cloud-provider-vsphere](/modules/cloud-provider-vsphere/)
- [cloud-provider-aws](/modules/cloud-provider-aws/)
- [cloud-provider-azure](/modules/cloud-provider-azure/)
- [cloud-provider-gcp](/modules/cloud-provider-gcp/)
- [sds-local-volume](/modules/sds-local-volume/stable/)
- [sds-replicated-volume](/modules/sds-replicated-volume/stable/)
- [csi-ceph](/modules/csi-ceph/stable/)
- [csi-nfs](/modules/csi-nfs/stable/)
- [csi-hpe](/modules/csi-hpe/stable/)
- [csi-huawei](/modules/csi-huawei/stable/)
- [csi-yadro-tatlin-unified](/modules/csi-yadro-tatlin-unified/stable/)
