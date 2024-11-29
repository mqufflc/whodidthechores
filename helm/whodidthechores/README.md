# Who Did The Chores

Who did the Chores is a web application to record and visualize who did the chores in an household.

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Prerequisites](#prerequisites)
- [Adding the Chart Repo](#adding-the-chart-repo)
- [Installing the Chart](#installing-the-chart)
- [Uninstalling the chart](#uninstalling-the-chart)
- [Parameters](#parameters)
  - [Common parameters](#common-parameters)
  - [Who Did The Chores Parameters](#who-did-the-chores-parameters)
  - [Postgres Parameters](#postgres-parameters)
  - [Expose Application Parameters](#expose-application-parameters)
  - [Other Parameters](#other-parameters)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

<!--- TODO once helm repo live
## TL;DR

```console
helm repo add whodidthechores https://
helm install my-release whodidthechores/whodidthechores
```
-->

## Prerequisites

- Kubernetes 1.19+
- Helm 3.1.0

## Adding the Chart Repo

To add Who Did The Chores helm repository with the name `whodidthechores`:

```console
helm repo add whodidthechores https://
```

## Installing the Chart

To install the chart with the release name `my-release`:

```console
helm install my-release whodidthechores/whodidthechores
```

This command deploys whodidthechores on a Kubernetes cluster with the default configuration. Head over to the [Parameters](#parameters) section to get the available configurations for your deployment.

## Uninstalling the chart

To remove the `my-release` deployment from a kubernetes cluster:

```console
helm delete my-release
```

This command removes all the components from the Kubernetes cluster and removes the release.

## Parameters

### Common parameters

| Name                | Description                                           | Value |
| ------------------- | ----------------------------------------------------- | ----- |
| `nameOverride`      | String to partially override whodidthechores.fullname | `""`  |
| `fullnameOverride`  | String to fully override whodidthechores.fullname     | `""`  |
| `commonAnnotations` | Annotations to add to all deployed resources          | `{}`  |
| `commonLabels`      | Labels to add to all deployed resources               | `{}`  |

### Who Did The Chores Parameters

| Name                                              | Description                                                                                           | Value                     |
| ------------------------------------------------- | ----------------------------------------------------------------------------------------------------- | ------------------------- |
| `image.registry`                                  | Who Did The Chores image registry                                                                     | `docker.io`               |
| `image.repository`                                | Who Did The Chores image repository                                                                   | `mqufflc/whodidthechores` |
| `image.tag`                                       | Who Did The Chores image tag                                                                          | `0.2.0`                   |
| `image.pullPolicy`                                | Who Did The Chores image pull policy                                                                  | `IfNotPresent`            |
| `image.pullSecrets`                               | Who Did The Chores image pull secrets                                                                 | `[]`                      |
| `revisionHistoryLimit`                            | Number of old history to retain to allow rollback (If not set, default Kubernetes value is set to 10) | `""`                      |
| `startupProbe.enabled`                            | Enable startupProbe on Who Did The Chores containers                                                  | `false`                   |
| `startupProbe.initialDelaySeconds`                | Initial delay seconds for startupProbe                                                                | `0`                       |
| `startupProbe.periodSeconds`                      | Period seconds for startupProbe                                                                       | `10`                      |
| `startupProbe.timeoutSeconds`                     | Timeout seconds for startupProbe                                                                      | `1`                       |
| `startupProbe.failureThreshold`                   | Failure threshold for startupProbe                                                                    | `3`                       |
| `startupProbe.successThreshold`                   | Success threshold for startupProbe                                                                    | `1`                       |
| `resources.limits`                                | The resources limits for the Who Did The Chores containers                                            | `{}`                      |
| `resources.requests`                              | The requested resources for the Who Did The Chores containers                                         | `{}`                      |
| `podSecurityContext.enabled`                      | Enabled Who Did The Chores pods' Security Context                                                     | `true`                    |
| `podSecurityContext.fsGroup`                      | Set Who Did The Chores pod's Security Context fsGroup                                                 | `1001`                    |
| `containerSecurityContext.enabled`                | Enabled Who Did The Chores containers' Security Context                                               | `true`                    |
| `containerSecurityContext.readOnlyRootFilesystem` | Whether the Who Did The Chores container has a read-only root filesystem                              | `true`                    |
| `containerSecurityContext.runAsNonRoot`           | Indicates that the Who Did The Chores container must run as a non-root user                           | `true`                    |
| `containerSecurityContext.runAsUser`              | Set Who Did The Chores containers' Security Context runAsUser                                         | `1001`                    |
| `containerSecurityContext.capabilities`           | Adds and removes POSIX capabilities from running containers (see `values.yaml`)                       |                           |
| `podLabels`                                       | Extra labels for Who Did The Chores pods                                                              | `{}`                      |
| `podAnnotations`                                  | Annotations for Who Did The Chores pods                                                               | `{}`                      |
| `priorityClassName`                               | Who Did The Chores pods' priorityClassName                                                            | `""`                      |
| `runtimeClassName`                                | Who Did The Chores pods' runtimeClassName                                                             | `""`                      |
| `affinity`                                        | Affinity for Who Did The Chores pods assignment                                                       | `{}`                      |
| `nodeSelector`                                    | Node labels for Who Did The Chores pods assignment                                                    | `{}`                      |
| `tolerations`                                     | Tolerations for Who Did The Chores pods assignment                                                    | `[]`                      |
| `additionalVolumes`                               | Extra Volumes for the Who Did The Chores Deployment                                                   | `{}`                      |
| `additionalVolumeMounts`                          | Extra volumeMounts for the Who Did The Chores container                                               | `{}`                      |
| `hostNetwork`                                     | Who Did The Chores pods' hostNetwork                                                                  | `false`                   |
| `containerPorts.http`                             | HTTP Port on the Host and Container                                                                   | `8080`                    |
| `containerPorts.metrics`                          | Metrics HTTP Port on the Host and Container                                                           | `8081`                    |
| `hostPorts.http`                                  | HTTP Port on the Host                                                                                 | `""`                      |
| `hostPorts.metrics`                               | Metrics HTTP Port on the Host                                                                         | `""`                      |
| `dnsPolicy`                                       | Who Did The Choress pods' dnsPolicy                                                                   | `""`                      |

### Postgres Parameters

| Name                                    | Description                                                                                             | Value  |
| --------------------------------------- | ------------------------------------------------------------------------------------------------------- | ------ |
| `postgres.enabled`                      | Wether to deploy a postgresql statefulset with Who Did The Chores deployment                            | `true` |
| `postgres.createPvc`                    | Wether to create a pvc for postgres data                                                                | `true` |
| `postgres.pvcName`                      | The name of the pvc for postgres statefulset that will be created or the name of an existing pvc to use | `""`   |
| `postgres.storageClassName`             | The Storage Class Name to use for the PVC                                                               | `""`   |
| `postgres.storageRequest`               | The storage request for the pvc                                                                         | `10Gi` |
| `postgres.annotations`                  | Additional custom annotations for postgres statefulset                                                  | `{}`   |
| `postgres.podSecurityContext.enabled`   | Enabled postgres' Security Context                                                                      | `true` |
| `postgres.podSecurityContext.fsGroup`   | Set postgres pod's Security Context fsGroup                                                             | `1001` |
| `postgres.podSecurityContext.runAsUser` | Set postgres pod's Security Context runAsUser                                                           | `1001` |

### Expose Application Parameters

| Name                        | Description                                                                                                                      | Value                    |
| --------------------------- | -------------------------------------------------------------------------------------------------------------------------------- | ------------------------ |
| `service.type`              | Who Did The Chores service type                                                                                                  | `ClusterIP`              |
| `service.loadBalancerClass` | Who Did The Chores service loadBalancerClass                                                                                     | `""`                     |
| `service.port`              | Who Did The Chores service HTTP port                                                                                             | `8080`                   |
| `service.nodePort`          | Node port for HTTP                                                                                                               | `""`                     |
| `service.annotations`       | Additional custom annotations for Who Did The Chores service                                                                     | `{}`                     |
| `ingress.enabled`           | Enable ingress record generation for Who Did The Chores                                                                          | `false`                  |
| `ingress.pathType`          | Ingress path type                                                                                                                | `ImplementationSpecific` |
| `ingress.apiVersion`        | Force Ingress API version (automatically detected if not set)                                                                    | `""`                     |
| `ingress.ingressClassName`  | IngressClass that will be be used to implement the Ingress                                                                       | `""`                     |
| `ingress.hostname`          | Default host for the ingress record                                                                                              | `whodidthechores.local`  |
| `ingress.path`              | Default path for the ingress record                                                                                              | `/`                      |
| `ingress.annotations`       | Additional annotations for the Ingress resource. To enable certificate autogeneration, place here your cert-manager annotations. | `{}`                     |
| `ingress.tls`               | Enable TLS configuration for the host defined at `ingress.hostname` parameter                                                    | `false`                  |
| `ingress.secretName`        |                                                                                                                                  | `""`                     |

### Other Parameters

| Name                         | Description                                          | Value  |
| ---------------------------- | ---------------------------------------------------- | ------ |
| `serviceAccount.annotations` | Annotations for Who Did The Chores service account   | `{}`   |
| `serviceAccount.create`      | Specifies whether a ServiceAccount should be created | `true` |
| `serviceAccount.labels`      | Extra labels to be added to the ServiceAccount       | `{}`   |
| `serviceAccount.name`        | The name of the ServiceAccount to use.               | `""`   |
