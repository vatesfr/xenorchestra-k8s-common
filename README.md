# xenorchestra-k8s-common

Shared Go module for Xen Orchestra Kubernetes integrations.

This project centralizes logic used by multiple Kubernetes components
(such as Xen Orchestra Cloud Controller Manager and CSI-related code), so
ProviderID parsing, cloud config validation, XO lookups, and label keys stay
consistent across repositories.

## 🧐 Why This Exists

Without a shared module, the same logic tends to be duplicated and drifts over
time. `xenorchestra-k8s-common` provides one source of truth for:

- Xen Orchestra cloud config parsing and validation
- Kubernetes ProviderID generation/parsing for Xen Orchestra
- Node-to-VM lookup helpers through the XO SDK
- Shared Kubernetes node label constants

### What should stay out of this module

- Consumer-specific business logic (for example, CCM controller loops or
  CSI attach/detach workflows).
- Code used by only one consumer until a second consumer needs it.

## 🛠️ Installation

```bash
go get github.com/vatesfr/xenorchestra-k8s-common
```

## 🧩 Configuration

Example cloud config (`xo-config.yaml`):

```yaml
url: https://xo.example.com
insecure: false
token: "YOUR_API_TOKEN"
```

Validation rules:

- `url` is required and must start with `http`
- Authentication is required:
  - either `token`
  - or both `username` and `password`
- `token` cannot be set together with `username/password`

## 🧑🏻‍💻 Usage Examples

### Read Config

```go
cfg, err := xok8scommon.ReadCloudConfigFromFile("/etc/kubernetes/xo-config.yaml")
if err != nil {
    return err
}
```

### Build and Parse ProviderID

```go
providerID := xok8scommon.GetProviderID(poolID, vm)
// xenorchestra://<poolUUID>/<vmUUID> or xenorchestra:///<vmUUID>

vmID, err := xok8scommon.GetVMID(providerID)
if err != nil {
    return err
}
_ = vmID
```

### Create XO Client and Resolve VM From Node

```go
client, err := xok8scommon.NewXOClient(&cfg)
if err != nil {
    return err
}

if err := client.CheckClient(ctx); err != nil {
    return err
}

vm, poolID, err := client.FindVMByNode(ctx, node)
if err != nil {
    return err
}
_, _ = vm, poolID
```


## ➤ License

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.