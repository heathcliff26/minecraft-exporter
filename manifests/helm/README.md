# minecraft-exporter Helm Chart

This Helm chart deploys the minecraft-exporter - monitor the state of your Minecraft server with Prometheus and Grafana

## Prerequisites

- Kubernetes 1.32+
- Helm 3.19+
- FluxCD installed in the cluster (recommended)

## Installation

### Installing from OCI Registry (GitHub Packages)

```bash
# Install the chart
helm install minecraft-exporter oci://ghcr.io/heathcliff26/manifests/minecraft-exporter --version <version>
```

## Configuration

### Minimal Configuration (No Ingress)

You need to specify a volume mount for your world folder.
```
volumes:
  - name: world
    hostPath:
      path: /path/to/world
      type: Directory
volumeMounts:
  - mountPath: /world
    name: world
    readOnly: true
```

## Values Reference

See [values.yaml](./values.yaml) for all available configuration options.

### Key Parameters

| Parameter                | Description                                         | Default                                   |
| ------------------------ | --------------------------------------------------- | ----------------------------------------- |
| `image.repository`       | Container image repository                          | `ghcr.io/heathcliff26/minecraft-exporter` |
| `image.tag`              | Container image tag                                 | Same as chart version                     |
| `ingress.enabled`        | Enable ingress                                      | `false`                                   |
| `servicemonitor.enabled` | Create a ServiceMonitor for the Prometheus Operator | `false`                                   |

## Support

For more information, visit: https://github.com/heathcliff26/minecraft-exporter
