# Cluster Autoscaler

The implementation lives in [`pkg/addons/autoscaler`](../../../pkg/addons/autoscaler).
It integrates the separately installed
[minikube-cluster-autoscaler-addon](https://github.com/astrivant/minikube-cluster-autoscaler-addon)
project with the standard enable/disable commands. The handbook contains the
[short user guide](../../../site/content/en/docs/handbook/addons/cluster-autoscaler.md).

## Installation discovery

Minikube looks for the bundle in `~/.minikube/addons/cluster-autoscaler`, beside a
provider binary on `PATH` (in the bundle's `bin` directory), or in the sibling
`minikube-cluster-autoscaler-addon` checkout during development. The optional
`MINIKUBE_AUTOSCALER_ADDON_PATH` overrides discovery.

## Lifecycle

On first enable, Minikube builds missing host components, discovers the host
address, and generates a configuration. It allows zero to two additional workers,
reducing the maximum to keep node allocations within 75% of physical host memory.
Initialization records existing nodes and credentials once; existing journals are
never automatically replaced. The native bridge runs in a detached process with
logs at `<state-dir>/bridge.log`.

The installed project's `scripts/addon.sh` (macOS/Linux) or `scripts/addon.ps1`
(Windows) owns the Helm release, provider container, TLS Secret and shared CRD.
There are no guest manifests to embed here. Disable removes the release and
provider, stops the managed bridge, and retains nodes and ownership state.
Cluster stop/delete stop the host processes before changing nodes; restart retains
the addon setting. If a cluster is replaced, its old journals must be retired
using the project's recovery workflow rather than reused for new nodes.

## Optional overrides

| Variable | Default |
| --- | --- |
| `MINIKUBE_AUTOSCALER_STATE_DIR` | `$XDG_STATE_HOME/minikube-cluster-autoscaler-addon/<profile>`, falling back to `$HOME/.local/state/...` |
| `MINIKUBE_AUTOSCALER_BINARY` | `<bundle>/bin/minikube-cluster-autoscaler-addon` (`.exe` on Windows) |
| `MINIKUBE_AUTOSCALER_IMAGE` | `minikube-cluster-autoscaler-addon:local` |

Paths and image selection are saved per Minikube profile after successful enable.
Saved settings take precedence while enabled and during disable. On the next
enable after disabling, environment overrides can select a new installation.
The selected `-p` profile always wins over `MINIKUBE_AUTOSCALER_PROFILE`.
The initialized `<state-dir>/config.json` is authoritative, including when
`MINIKUBE_AUTOSCALER_CONFIG` points elsewhere.

The external project retains its license and release lifecycle; its Go
implementation is not vendored or linked into minikube.
