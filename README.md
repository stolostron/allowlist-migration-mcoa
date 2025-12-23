# Allowlist Migration MCOA
Converts an allowlist ConfigMap (used by the multicluster-observability-operator) into equivalent ScrapeConfig and PrometheusRule resources (used by the multicluster-observability-addon).

## Build from source
To build `allowlist-migration` and output to `bin/`:
```
make build
```
To build binaries for all OS/architectures and output to `build_output/`:
```
make build-all
```

## Using the migration tool
### Prerequisites
ScrapeConfig and PrometheusRule CRDs must be installed on the hub cluster. CRDs can be found in the [multicluster-observability-addon](https://github.com/stolostron/multicluster-observability-addon) repo and installed with `make install-crds`.

### Instructions
1. Build or download the binary and ensure it is executable.

2. Run the binary, providing the allowlist ConfigMap as input and a directory for output:
   ```
   allowlist-migration <path/to/allowlist.yaml> <output/dir/> [targets.yaml]
   ```
   - The optional third argument specifies which key in the ConfigMap contains the metrics list (defaults to `metrics_list.yaml`).
   - Output files are named `<configmap-name>-scrapeconfig.yaml` and `<configmap-name>-prometheusrule.yaml`.

3. Replace the `<MCO-UID>` placeholder in both output files with the actual MCO UID.

   i. Get the MCO UID:
      ```
      kubectl get multiclusterobservability observability -o jsonpath='{.metadata.uid}'
      ```

   ii. In each output file, replace `<MCO-UID>` with the value returned:
      ```yaml
      ownerReferences:
      - apiVersion: observability.open-cluster-management.io/v1beta2
        controller: true
        kind: MultiClusterObservability
        name: observability
        uid: <MCO-UID>      # <-- REPLACE THIS
      ```

4. Apply the generated resources to the hub cluster:
   ```
   oc apply -f <output/dir>/<configmap-name>-prometheusrule.yaml
   oc apply -f <output/dir>/<configmap-name>-scrapeconfig.yaml
   ```

5. Verify configs have been applied:
   - Check that the new ScrapeConfig and PrometheusRule exist on the hub cluster:
     ```
     kubectl get ScrapeConfig -n open-cluster-management-observability
     kubectl get PrometheusRule -n open-cluster-management-observability
     ```
   - Check that the configs have propagated to a managed cluster (switch kubecontext first):
     ```
     kubectl get ScrapeConfig -n open-cluster-management-agent-addon
     kubectl get PrometheusRule -n open-cluster-management-agent-addon
     ```
