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

package consts

const (
	ModuleName      			string = "snapshotController"
	ModuleNamespace 			string = "d8-snapshot-controller"
	ModulePluralName			string = "snapshot-controller"
	ValidatingWebhookCertCn   	string = "snapshot-validation-webhook"
	WebhookCertCn   			string = "webhooks"
)

var AllowedProvisioners = []string{}

var WebhookConfigurationsToDelete = []string{}

var CRGVKsForFinalizerRemoval = []CRGVK{
	{Group: "snapshot.storage.k8s.io", Version: "v1", Kind: "VolumeSnapshot", Namespaced: true},
	{Group: "snapshot.storage.k8s.io", Version: "v1", Kind: "VolumeSnapshotContent", Namespaced: false},
	{Group: "snapshot.storage.k8s.io", Version: "v1", Kind: "VolumeSnapshotClass", Namespaced: false},
}

type CRGVK struct {
	Group      string
	Version    string
	Kind       string
	Namespaced bool
}
