# Minikube Doctor

`minikube doctor` is a centralized diagnostic command for troubleshooting Minikube clusters and host environments. Instead of manually running multiple commands (`minikube status`, `minikube profile list`, `kubectl`, etc.), `minikube doctor` aggregates system, configuration, and cluster status checks into a single unified report.

---

## Features & Diagnostic Phases

The diagnostics run in five distinct phases:

### 1. Configuration (Informational)
Displays passive configuration details for the currently active profile:
- **Active Profile**: The name of the profile currently checked out.
- **Driver**: The active VM/Container driver configured (e.g., `docker`, `podman`, `virtualbox`).
- **Kubernetes Version**: The configured version of Kubernetes for this profile.
- **Container Runtime**: The target runtime (e.g., `docker`, `containerd`).

### 2. Validation
Scans and checks configuration file integrity:
- Validates active and inactive profiles to flag corrupted settings, indicating which files are invalid.

### 3. Environment
Verifies host-level tools and services based on the active driver (Driver-Aware Environment Checks):
- Checks that the configured driver's binary is installed on the host's `PATH` (e.g., looking up `vboxmanage` for VirtualBox, or `podman` for Podman).
- Verifies that the corresponding daemon or system service is running (e.g., Docker daemon, Podman machine).
- Checks that `kubectl` is installed and ready.

### 4. Cluster Health
Queries the running state of the cluster using Minikube's internal APIs:
- **Cluster Status**: Assesses if the cluster container/VM host is active.
- **API Server**: Verifies the status of the Kubernetes API Server.
- **Nodes Ready**: Scans all cluster nodes to verify that the `kubelet` is active.
- **Kubeconfig**: Validates that `kubectl` can communicate with the cluster.

### 5. Resources
Inspects hardware resource allocations to ensure stability:
- **CPU**: Checks allocated CPUs (minimum 2 required, 4 recommended).
- **Memory**: Checks allocated RAM (minimum 2048 MB required, 4096 MB recommended).
- **Disk**: Checks allocated disk space (minimum 20000 MB recommended).

---

## Commands & Usage

### 1. Run Standard Diagnostics
Generates a polished, categorized text summary of all checks:
```bash
minikube doctor
```

### 2. Run Verbose Diagnostics
Outputs detailed terminal stack traces/raw errors under failing checks and details about warnings:
```bash
minikube doctor --verbose
```

### 3. JSON Output (Automation Mode)
Outputs the diagnostic results in structured JSON format with camelCase keys:
```bash
minikube doctor --output=json
# or
minikube doctor -o json
```

Example JSON element:
```json
[
  {
    "name": "Cluster Status",
    "status": "FAIL",
    "message": "Cluster is not running",
    "details": "docker container inspect minikube --format={{.State.Status}}: exit status 1 ...",
    "recommendation": "Start the cluster by running: minikube start"
  }
]
```

---

## Developer Guide: Adding a New Diagnostic Check

Each check resides in its own source file in the `pkg/minikube/doctor` directory.

1. **Define the Check**: Create a check function returning a `Result`:
   ```go
   package doctor

   func MyNewCheck() Result {
       // logic here
       if failed {
           return Result{
               Name:           "My Check",
               Status:         "FAIL",
               Message:        "Reason for failure",
               Details:        "Raw technical trace",
               Recommendation: "How to fix it",
           }
       }
       return Result{
           Name:    "My Check",
           Status:  "PASS",
           Message: "Healthy status description",
       }
   }
   ```

2. **Register the Check**: Add your check to the orchestrator `Run()` list in `pkg/minikube/doctor/doctor.go`.
