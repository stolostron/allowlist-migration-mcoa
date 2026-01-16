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
   - Output files are named `custom-scrapeconfig.yaml` and `custom-prometheusrule.yaml`.

3. Apply the generated resources to the hub cluster:
   ```
   oc apply -f <output/dir>/custom-prometheusrule.yaml -n open-cluster-management-observability
   oc apply -f <output/dir>/custom-scrapeconfig.yaml -n open-cluster-management-observability
   ```

4. Add the generated resources to the ClusterManagementAddOn resource

   ```
   kubectl patch clustermanagementaddon multicluster-observability-addon --type=json -p='[
   {
      "op": "add",
      "path": "/spec/installStrategy/placements/0/configs/-",
      "value": {
         "group": "monitoring.rhobs",
         "resource": "scrapeconfigs",
         "name": "custom-scrapeconfig",
         "namespace": "open-cluster-management-observability"
      }
   },
   {
      "op": "add",
      "path": "/spec/installStrategy/placements/0/configs/-",
      "value": {
         "group": "monitoring.coreos.com",
         "resource": "prometheusrules",
         "name": "custom-prometheusrule",
         "namespace": "open-cluster-management-observability"
      }
   }
   ]'
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

   - Should now be able to verify that your custom metrics are being collected through querying grafana or perses UI. For example, if in your custom ScrapeConfig matches the metric `up`, you can query for `up{job="my-app"}` to verify the ScrapeConfig is being used. For PrometheusRule, if you have a recording rule for `avg_response_time` you can quickly verify it is working by running a query `avg_response_time`. Should also check that any alerts or dashboards that use these custom metrics are now being populated.