# runner

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 1.16.0](https://img.shields.io/badge/AppVersion-1.16.0-informational?style=flat-square)

Helm Chart for DeepSource Runner

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| DeepSource | <support@deepsource.io> |  |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | pod affinity configuration for the runner |
| autoscaling | object | `{"enabled":false,"maxReplicas":100,"minReplicas":1,"targetCPUUtilizationPercentage":80,"targetMemoryUtilizationPercentage":80}` | the autoscaling configuration for the runner |
| autoscaling.enabled | bool | `false` | whether to enable autoscaling for the runner |
| autoscaling.maxReplicas | int | `100` | the maximum number of replicas to scale up to |
| autoscaling.minReplicas | int | `1` | the minimum number of replicas to scale down to |
| autoscaling.targetCPUUtilizationPercentage | int | `80` | the target CPU utilization percentage to scale up to |
| autoscaling.targetMemoryUtilizationPercentage | int | `80` | the target memory utilization percentage to scale up to |
| config.apps.myapp.github.api_host | string | `""` |  |
| config.apps.myapp.github.app_id | string | `""` |  |
| config.apps.myapp.github.client_id | string | `""` |  |
| config.apps.myapp.github.client_secret | string | `""` |  |
| config.apps.myapp.github.host | string | `""` |  |
| config.apps.myapp.github.private_key | string | `""` |  |
| config.apps.myapp.github.slug | string | `""` |  |
| config.apps.myapp.github.webhook_secret | string | `""` |  |
| config.deepsource.host | string | `""` |  |
| config.deepsource.public_key | string | `""` |  |
| config.object_storage.backend | string | `""` |  |
| config.object_storage.bucket | string | `""` |  |
| config.object_storage.credential | string | `""` |  |
| config.runner.client_id | string | `""` |  |
| config.runner.client_secret | string | `""` |  |
| config.runner.host | string | `""` |  |
| config.runner.id | string | `""` |  |
| config.runner.private_key | string | `""` |  |
| config.runner.webhook_secret | string | `""` |  |
| config.saml.certificate | string | `""` |  |
| config.saml.enabled | bool | `false` |  |
| config.saml.key | string | `""` |  |
| config.saml.metadata_url | string | `""` |  |
| fullnameOverride | string | `""` | the fullname override for the runner |
| image | object | `{"pullPolicy":"Always","repository":"us.gcr.io/deepsource-dev/runner","tag":""}` | the image to use for the runner |
| image.pullPolicy | string | `"Always"` | the image pull policy to use |
| image.repository | string | `"us.gcr.io/deepsource-dev/runner"` | the repository to pull the image from |
| image.tag | string | `""` | the image tag to use |
| imagePullSecrets | string | `""` | the image pull secrets to use |
| ingress | object | `{"annotations":{"cert-manager.io/cluster-issuer":"letsencrypt-prod","ingress.kubernetes.io/service-upstream":"true","ingress.kubernetes.io/ssl-redirect":"true","kubernetes.io/ingress.class":"nginx","kubernetes.io/tls-acme":"true"},"className":"nginx","enabled":true,"hosts":[{"host":"runner.example.com","paths":[{"path":"/","pathType":"ImplementationSpecific"}]}],"tls":[{"hosts":["runner.example.com"],"secretName":"tls-runner-deepsource-com"}]}` | ingress configuration for the runner |
| ingress.annotations | object | `{"cert-manager.io/cluster-issuer":"letsencrypt-prod","ingress.kubernetes.io/service-upstream":"true","ingress.kubernetes.io/ssl-redirect":"true","kubernetes.io/ingress.class":"nginx","kubernetes.io/tls-acme":"true"}` | annotations to add to the ingress |
| ingress.className | string | `"nginx"` | the ingress class to use |
| ingress.enabled | bool | `true` | whether to enable ingress for the runner |
| ingress.hosts | list | `[{"host":"runner.example.com","paths":[{"path":"/","pathType":"ImplementationSpecific"}]}]` | hosts to add to the ingress |
| ingress.tls | list | `[{"hosts":["runner.example.com"],"secretName":"tls-runner-deepsource-com"}]` | tls configuration for the ingress |
| nameOverride | string | `""` | the name override for the runner |
| nodeSelector | object | `{}` | the node selector to use for the runner |
| podAnnotations | object | `{}` | annotations to add to the runner pod |
| podSecurityContext | object | `{}` | pod security context to be added to runner pods |
| replicaCount | int | `1` | number of replicas to deploy |
| resources | object | `{"limits":{"cpu":"200m","memory":"1Gi"},"requests":{"cpu":"100m","memory":"128Mi"}}` | the resources to allocate for the runner |
| rqlite | object | `{"image":{"pullPolicy":"IfNotPresent","repository":"rqlite/rqlite","tag":"7.20.6"},"replicaCount":1,"resources":{"limits":{"cpu":"200m","memory":"1Gi"},"requests":{"cpu":"100m","memory":"128Mi"}},"storageSize":"1Gi"}` | the configuration for the rqlite database |
| rqlite.replicaCount | int | `1` | the number of replicas to deploy for the rqlite database |
| rqlite.resources | object | `{"limits":{"cpu":"200m","memory":"1Gi"},"requests":{"cpu":"100m","memory":"128Mi"}}` | the resources to allocate for the rqlite database |
| rqlite.storageSize | string | `"1Gi"` | the storage size for the rqlite database |
| securityContext | object | `{}` | security context to be added to containers |
| service | object | `{"port":80,"type":"ClusterIP"}` | service configuration for the runner |
| service.port | int | `80` | the port to use for the service |
| service.type | string | `"ClusterIP"` | the type of service to use |
| serviceAccount | object | `{"annotations":{},"create":true,"name":"runner"}` | the service account configuration to use for the runner |
| serviceAccount.annotations | object | `{}` | annotations to add to the service account |
| serviceAccount.create | bool | `true` | specifies whether a service account should be created |
| serviceAccount.name | string | `"runner"` | The name of the service account to use. If not set and create is true, a name is generated using the fullname template |
| taskNamespace | string | `"atlas-jobs"` | the namespace to run the analysis jobs in |
| tolerations | list | `[]` | node tolerations for server scheduling to nodes with taints |
| useExistingConfig | bool | `false` | whether to use existing secret for the runner |

----------------------------------------------
Autogenerated from chart metadata using [helm-docs v1.11.0](https://github.com/norwoodj/helm-docs/releases/v1.11.0)
