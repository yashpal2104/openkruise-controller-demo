# Kubernetes Controller Demo: API Versioning & Migration

A comprehensive Kubernetes controller demonstration showcasing API versioning and migration patterns used in real projects like OpenKruise. This project demonstrates how to evolve APIs from `v1alpha1` to `v1beta1` while maintaining backward compatibility.

## What This Demo Does

Creates a simple Kubernetes controller that automatically manages pods with two different update strategies:
- **RollingUpdate** (safe) - Updates pods one at a time with zero downtime
- **Recreate** (fast) - Replaces all pods simultaneously with brief downtime

## Architecture Overview

```
User Creates YAML → Controller Watches → Creates/Updates Pods → Reports Status
     ↓                    ↓                      ↓               ↓
  v1alpha1/v1beta1    Reconcile Loop        Pod Management    Status Update
```

### Key Components

- **MiniCloneSet CRD**: Custom resource users create to define desired state
- **Controller**: Watches for changes and reconciles actual state with desired state
- **Conversion System**: Automatically converts between `v1alpha1` ↔ `v1beta1`
- **Update Strategies**: Two approaches for application updates

## API Evolution Showcase

This project demonstrates how APIs evolve in production Kubernetes projects:

### v1alpha1 (Simple Structure)
```go
type MiniCloneSetSpec struct {
    Replicas       int                `json:"replicas"`
    Image          string             `json:"image"`           // Flat structure
    UpdateStrategy UpdateStrategyType `json:"updateStrategy"`
}
```

### v1beta1 (Enhanced Structure)
```go
type MiniCloneSetSpec struct {
    Replicas       int            `json:"replicas"`
    Container      Container      `json:"container"`        // Nested structure
    UpdateStrategy UpdateStrategy `json:"updateStrategy"`
}

type Container struct {
    Image string `json:"image"`
}

type UpdateStrategy struct {
    Type           UpdateStrategyType `json:"type,omitempty"`
    MaxUnavailable *string           `json:"maxUnavailable,omitempty"`  // New field!
}
```

## Quick Start

### Prerequisites
- Go 1.19+
- Kubernetes cluster (local or remote)
- kubectl configured
- kubebuilder installed

### Installation & Running

1. **Install CRDs**
   ```bash
   make install
   ```

2. **Run Controller Locally**
   ```bash
   make run
   ```

3. **Create Test Application (v1alpha1)**
   ```bash
   kubectl apply -f - <<EOF
   apiVersion: apps.example.com.my.domain/v1alpha1
   kind: MiniCloneSet
   metadata:
     name: demo-app
   spec:
     replicas: 2
     image: nginx:1.20
     updateStrategy: RollingUpdate
   EOF
   ```

4. **Watch Pods Being Created**
   ```bash
   kubectl get pods -l app=demo-app -w
   ```

5. **Test Rolling Update**
   ```bash
   kubectl patch minicloneset demo-app -p '{"spec":{"image":"nginx:1.21"}}'
   ```

6. **Test API Conversion with v1beta1**
   ```bash
   kubectl apply -f - <<EOF
   apiVersion: apps.example.com.my.domain/v1beta1
   kind: MiniCloneSet
   metadata:
     name: beta-app
   spec:
     replicas: 2
     container:
       image: httpd:2.4
     updateStrategy:
       type: RollingUpdate
       maxUnavailable: "50%"
   EOF
   ```

## Update Strategies Explained

### RollingUpdate Strategy
- **Behavior**: Updates pods one at a time
- **Advantage**: Zero downtime
- **Use Case**: Production applications that can't afford downtime
- **Process**: Creates new pod → Waits for readiness → Deletes old pod → Repeats

### Recreate Strategy  
- **Behavior**: Deletes all pods, then creates new ones
- **Advantage**: Faster updates, simpler logic
- **Use Case**: Development environments, applications that can handle brief downtime
- **Process**: Deletes all pods simultaneously → Creates all new pods

## Development

### Phase 1: Project Setup & Initial API
- Created kubebuilder project with domain structure
- Generated initial `v1alpha1` API with basic fields
- Implemented simple reconciliation with pod management
- Added type-safe enums for update strategies
- Fixed code generation issues

### Phase 2: API Evolution Strategy
- Established `v1alpha1` → `v1beta1` migration path
- Set `v1alpha1` as storage version, `v1beta1` as hub version
- Designed clear versioning strategy

### Phase 3: Multi-Version API Implementation
- Created enhanced `v1beta1` API structure
- Added nested container specification
- Introduced `maxUnavailable` field for advanced control
- Configured storage and serving versions

### Phase 4: Conversion Implementation
- Added Hub method to `v1beta1`
- Implemented bidirectional conversion methods
- Handled field mapping and default value assignment
- Ensured data integrity across versions

### Phase 5: Controller Architecture
- Controller works with `v1alpha1` (storage version)
- Kubernetes handles conversion automatically
- Single controller manages both API versions

### Phase 6: Controller Logic Implementation
- Main reconciliation loop
- Rolling update strategy implementation
- Recreate strategy implementation
- Pod lifecycle management
- Status reporting

## Demo Scenarios

### Scenario 1: Simple v1alpha1 Usage
```yaml
apiVersion: apps.example.com.my.domain/v1alpha1
kind: MiniCloneSet
metadata:
  name: simple-app
spec:
  replicas: 3
  image: nginx:1.20
  updateStrategy: RollingUpdate
```

### Scenario 2: Advanced v1beta1 Usage
```yaml
apiVersion: apps.example.com.my.domain/v1beta1
kind: MiniCloneSet
metadata:
  name: advanced-app
spec:
  replicas: 3
  container:
    image: nginx:1.20
  updateStrategy:
    type: RollingUpdate
    maxUnavailable: "50%"
```

### Scenario 3: Strategy Comparison
```bash
# Rolling Update - Safe, gradual updates
kubectl patch minicloneset simple-app -p '{"spec":{"image":"nginx:1.21"}}'

# Recreate - Fast, complete replacement  
kubectl patch minicloneset recreate-app -p '{"spec":{"updateStrategy":"Recreate"}}'
kubectl patch minicloneset recreate-app -p '{"spec":{"image":"nginx:1.21"}}'
```

## Technical Concepts Demonstrated

### 1. API Versioning Patterns
- **Alpha → Beta Progression**: Standard Kubernetes API evolution
- **Storage vs Served Versions**: Internal version management
- **Conversion Webhooks**: Automatic field transformation

### 2. Controller Architecture  
- **Version-agnostic Reconciliation**: Works with storage version
- **Watch Patterns**: Event-driven state management
- **Owner References**: Automatic resource cleanup

### 3. Code Generation
- **DeepCopy Methods**: Required for Kubernetes types
- **CRD Generation**: Automatic OpenAPI schema creation  
- **RBAC Generation**: Automatic permission setup

### 4. Kubernetes Operators
- **Custom Resource Management**: Extending Kubernetes
- **Reconciliation Loops**: Core controller pattern
- **Status Reporting**: User feedback mechanism

## Project Value

### For Learning
- Demonstrates deep Kubernetes internals knowledge
- Shows API design and evolution expertise
- Highlights backward compatibility considerations
- Showcases production-ready patterns
- Complete controller development lifecycle
- Real-world API evolution strategies
- Hands-on kubebuilder experience
- Testing and validation approaches

### Technical Skills Showcased
- **Go Programming**: Advanced patterns and Kubernetes client libraries
- **Kubernetes Expertise**: API machinery and controller internals
- **Software Architecture**: API design and versioning strategies
- **DevOps Practices**: Operator development and testing

### Real-World Applications
- **OpenKruise CloneSets**: Simplified version of production patterns
- **Kubernetes Operators**: Foundation for operator development
- **API Evolution**: Critical for long-running Kubernetes projects
- **Cloud Native Development**: Essential cloud-native patterns

## Testing

### Unit Tests
```bash
make test
```

### Integration Tests
```bash
make test-integration
```

### End-to-End Tests
```bash
make test-e2e
```

## Project Structure

```
.
├── api/
│   ├── v1alpha1/          # Initial API version
│   │   ├── minicloneset_types.go
│   │   └── zz_generated.deepcopy.go
│   └── v1beta1/           # Enhanced API version  
│       ├── minicloneset_types.go
│       └── zz_generated.deepcopy.go
├── controllers/
│   └── minicloneset_controller.go
├── config/
│   ├── crd/               # Custom Resource Definitions
│   ├── rbac/              # Role-Based Access Control
│   └── samples/           # Example resources
├── Makefile
└── main.go
```

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Inspired by [OpenKruise](https://github.com/openkruise/kruise) CloneSet patterns
- Built with [Kubebuilder](https://kubebuilder.io/)
- Follows [Kubernetes API Conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md)

---

*This project demonstrates the exact patterns used by mature Kubernetes projects for API evolution, making it an excellent showcase of cloud-native development expertise!* 🚀
