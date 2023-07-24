## Parameters

### Runner configuration Parameters

| Name                                          | Description                                                    | Value   |
| --------------------------------------------- | -------------------------------------------------------------- | ------- |
| `config.apps`                                 | Configuration values for the VCS apps to be used by the runner | `[]`    |
| `config.deepsource`                           | Configuration values for the deepsource remote host            | `{}`    |
| `config.kubernetes.namespace`                 | The namespace to schedule the tasks in                         | `""`    |
| `config.kubernetes.nodeSelector`              | The node selector to use for the tasks                         | `{}`    |
| `config.kubernetes.imageRegistry.registryUrl` | The registry url to use for the task images                    | `""`    |
| `config.kubernetes.imageRegistry.username`    | The username to use for the image registry                     | `""`    |
| `config.kubernetes.imageRegistry.password`    | The password to use for the image registry                     | `""`    |
| `config.objectStorage.backend`                | The backend to use for the object storage (e.g gcs)            | `""`    |
| `config.objectStorage.bucket`                 | The bucket to use for the object storage                       | `""`    |
| `config.objectStorage.credential`             | The credentials value to use for the object storage            | `""`    |
| `config.runner.id`                            | The id of the runner                                           | `""`    |
| `config.runner.host`                          | The host of the runner to use                                  | `""`    |
| `config.runner.clientId`                      | The client id to use for the runner                            | `""`    |
| `config.runner.clientSecret`                  | The client secret to use for the runner                        | `""`    |
| `config.runner.privateKey`                    | The private key to use for the runner                          | `""`    |
| `config.runner.webhookSecret`                 | The webhook secret to use for the runner                       | `""`    |
| `config.saml.enabled`                         | Whether to enable SAML2.0 authentication                       | `false` |
| `config.saml.certificate`                     | The certificate to use for the runner as service provider      | `""`    |
| `config.saml.key`                             | The private key to use for the runner as service provider      | `""`    |
| `config.saml.metadataUrl`                     | The metadata url to use for the identity provider              | `""`    |

### Common Parameters

| Name                                            | Description                                                                                                                      | Value                             |
| ----------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------- | --------------------------------- |
| `replicaCount`                                  | Number of deepsource runner replicas to deploy                                                                                   | `1`                               |
| `image.repository`                              | deepsource runner image repository                                                                                               | `us.gcr.io/deepsource-dev/runner` |
| `image.pullPolicy`                              | deepsource runner image pull policy                                                                                              | `Always`                          |
| `image.tag`                                     | deepsource runner image tag                                                                                                      | `""`                              |
| `imagePullSecrets`                              | deepsource runner image pull secrets                                                                                             | `[]`                              |
| `nameOverride`                                  | String to partially override runner.name                                                                                         | `""`                              |
| `fullnameOverride`                              | String to partially override runner.name                                                                                         | `""`                              |
| `serviceAccount.create`                         | Specifies whether a ServiceAccount should be created                                                                             | `true`                            |
| `serviceAccount.annotations`                    | Annotations for service account. Evaluated as a template. Only used if `create` is `true`.                                       | `{}`                              |
| `serviceAccount.name`                           | Name of the service account to use. If not set and create is true, a name is generated using the fullname template.              | `""`                              |
| `podAnnotations`                                | Annotations for the deepsource runner pods                                                                                       | `{}`                              |
| `podSecurityContext`                            | Security context policies to add to the deepsource runner pods                                                                   | `{}`                              |
| `securityContext`                               | Security context policies to add to the containers                                                                               | `{}`                              |
| `service.type`                                  | deepsource runner service type                                                                                                   | `ClusterIP`                       |
| `service.port`                                  | deepsource runner service HTTP port                                                                                              | `80`                              |
| `ingress.enabled`                               | Enable ingress record generation for deepsource runner                                                                           | `false`                           |
| `ingress.className`                             | IngressClass that will be be used to implement the Ingress (Kubernetes 1.18+)                                                    | `""`                              |
| `ingress.annotations`                           | Additional annotations for the Ingress resource. To enable certificate autogeneration, place here your cert-manager annotations. | `{}`                              |
| `ingress.hosts`                                 | Deepsource runner Ingress hosts                                                                                                  | `[]`                              |
| `ingress.tls`                                   | Deepsource runner Ingress TLS configuration                                                                                      | `[]`                              |
| `resources.limits.cpu`                          | The resources limits for the deepsource runner containers                                                                        | `200m`                            |
| `resources.limits.memory`                       | The resources limits for the deepsource runner containers                                                                        | `1Gi`                             |
| `resources.requests.cpu`                        | The requested cpu for the deepsource runner containers                                                                           | `100m`                            |
| `resources.requests.memory`                     | The requested memory for the deepsource runner containers                                                                        | `128Mi`                           |
| `autoscaling.enabled`                           | Enable Horizontal POD autoscaling for deepsource runner                                                                          | `false`                           |
| `autoscaling.minReplicas`                       | Minimum number of deepsource runner replicas                                                                                     | `1`                               |
| `autoscaling.maxReplicas`                       | Maximum number of deepsource runner replicas                                                                                     | `100`                             |
| `autoscaling.targetCPUUtilizationPercentage`    | Target CPU utilization percentage                                                                                                | `80`                              |
| `autoscaling.targetMemoryUtilizationPercentage` | Target Memory utilization percentage                                                                                             | `80`                              |
| `nodeSelector`                                  | Node labels for deepsource runner pods assignment                                                                                | `{}`                              |
| `tolerations`                                   | Tolerations for deepsource runner pods assignment                                                                                | `[]`                              |
| `affinity`                                      | Affinity for deepsource runner pods assignment                                                                                   | `{}`                              |

### RQLite configuration parameters

| Name                               | Description                                                      | Value           |
| ---------------------------------- | ---------------------------------------------------------------- | --------------- |
| `rqlite.image.repository`          | RQLite image repository                                          | `rqlite/rqlite` |
| `rqlite.image.pullPolicy`          | RQLite image pull policy                                         | `IfNotPresent`  |
| `rqlite.image.tag`                 | RQLite image tag                                                 | `7.20.6`        |
| `rqlite.replicaCount`              | Number of rqlite replicas to deploy                              | `1`             |
| `rqlite.storageSize`               | The size of the persistent volume to use for the rqlite database | `1Gi`           |
| `rqlite.resources.limits.cpu`      | The resources limits for the rqlite containers                   | `200m`          |
| `rqlite.resources.limits.memory`   | The resources limits for the rqlite containers                   | `1Gi`           |
| `rqlite.resources.requests.cpu`    | The requested cpu for the rqlite containers                      | `100m`          |
| `rqlite.resources.requests.memory` | The requested memory for the rqlite containers                   | `128Mi`         |
