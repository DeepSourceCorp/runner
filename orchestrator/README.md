# Orchestrator

The `orchestrator` module is responsible for triggering analysis,
autofix and transformer jobs on the configured architecture —
Kubernetes, Nomad, etc.

## Flows

```mermaid
flowchart LR
    http(HTTP) & grpc(gRPC) & rmq(RMQ) --> protocol("protocol (Interface)")
    protocol --> command(command)
    command --> driver("driver(Interface)")
    driver --> Kubernetes(Kubernetes) & Nomad(Nomad)
```

## Implementation Specs

- The `orchestrator` module will be used for initializing a server that is configured via the runner config file.

- The config file will specify which protocol is being used to communicate with the server - HTTP, gRPC, etc., and what infrastructure the jobs will be spawned on - Kubernetes, Nomad, etc.

- The orchestrator will initialize and split off into a separate goroutine and keep listening for requests for the configured protocol.

- The requests will be of the following types (as currently configured in `atlas`):

  - `analysis_run`
  - `autofix_run`
  - `transforer_run`
  - `ssh_verify_run`
  - `cancel_check_run`

- These requests will be translated via the `protocol` interface into their respective implementation; HTTP, gRPC, etc.

- The request having been validated and sanitized, will call the respective command in the `command` module.

- The `command` module will then call the `driver` module to execute the command on the configured driver: Kubernetes, Nomad, etc.

## SAMPLE PAYLOAD
Analysis /tasks/analysis
```
{"run_id": "f231fc3d-d586-4d65-a478-402005e1736f", "run_serial": "0", "config": {"version": 1, "analyzers": [{"name": "python", "enabled": true, "meta": {"runtime_version": "3.x.x"}}], "transformers": [{"name": "autopep8", "enabled": true}, {"name": "isort", "enabled": true}, {"name": "black", "enabled": true}]}, "ds_config_updated": false, "vcs_meta": {"remote_url": "https://***:***@github.com/t3st-org/monitors-test-repo.git", "is_for_default_analysis_branch": false, "base_branch": "main", "base_oid": "5cd7c55b65720b5251d85e53dfb1dd7c6fb6561e", "checkout_oid": "7d8dd9f8f0b1d1c7aaafda6f4adde8ca13db185c", "clone_submodules": false}, "keys": {}, "checks": [{"check_seq": "1", "artifacts": [], "analyzer_meta": {"name": "python", "analyzer_type": "core", "command": "/app/marvin-python", "version": "dev", "cpu_limit": "2400", "memory_limit": "6501"}, "processors": ["source_code_load", "skip_cq"], "diff_meta_commits": []}], "_meta": {"resource_pool": "plan-pf"}}
```

Autofix /tasks/autofix
```
{"run_id": "277adfe1-f7ea-4767-bace-8e84e2556b69", "run_serial": "0", "config": {"version": 1, "analyzers": [{"meta": {"runtime_version": "3.x.x"}, "name": "python", "enabled": true}], "transformers": [{"name": "autopep8", "enabled": true}]}, "vcs_meta": {"remote_url": "https://***:***@github.com/t3st-org/demo-python.git", "checkout_oid": "2b9caeb6f21d9f13bb1cc1bdb984ebdccb1956f2", "base_branch": "master", "clone_submodules": false}, "keys": {}, "autofixer": {"autofix_meta": {"name": "python", "command": "/app/autofix.sh", "version": "dev", "cpu_limit": "2400", "memory_limit": "6501"}, "autofixes": [{"issue_code": "PTC-W0014", "occurrences": [{"path": "demo_code.py", "position": {"begin": {"line": 56, "column": 4}, "end": {"line": 56, "column": 4}}}]}]}, "_meta": {"resource_pool": "plan-pf"}}
```