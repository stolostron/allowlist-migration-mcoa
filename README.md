# Allowlist Migration MCOA
Builds binary to convert allowlist used for multicluster obervability operator into equivalent scrapeconfig and prometheusrule used for multicluster observability addon

## Build from source
To build `allowlist-migration` and output it to `bin/`
```
make build
```
To build binaries for all OS and architectures and output it to `build_output/`
```
make build-all
```

## Using the migration tool
### Prerequisites
ScrapeConfig and Prometheus rule CRDs must be installed on the hub cluster. CRDs can be found in the [multicluster-observability-addon](https://github.com/stolostron/multicluster-observability-addon) repo and installed on the cluster with the command `make install-crds`

### Instructions
1. Install binary and make binary executable
2. Binary takes an allowlist.yaml as input and creates equivalent ScrapeConfig and PrometheusRule in given output directory. Example execution of the binary is 
  ```
  path/to/allowlist-migration <path/to/allowlist.yaml> <path/to/write/dir/>
  ```
3. In the prometheusrule.yaml and scrapeconfig.yaml that were created, the field `uid` needs to have the uid of mco.

   i. Get the mco uid using the following command:
      ```
      kubectl get multiclusterobservability observability -o jsonpath='{.metadata.uid}'
      ```

   ii. Replace `<MCO-UID>` with the value returned:
      ```
      metadata:
        labels:
          ...
        name: custom-scrapeconfig
        namespace: open-cluster-management-observability
        ownerReferences:
        - apiVersion: observability.open-cluster-management.io/v1beta2
          controller: true
          kind: MultiClusterObservability
          name: observability
          uid: <MCO-UID>      <-- REPLACE THIS 
      ```
4. Apply new prometheusrule and scrapeconfig on hub cluster
  ```
  oc apply -f custom-prometheusrule.yaml
  oc apply -f custom-scrapeconfig.yaml
  ```

5. Verify configs have been applied
* Check that the new ScrapeConfig and PrometheusRule are on hub cluster
  ```
  kubectl get ScrapeConfig -n open-cluster-management-observability
  kubectl get PrometheusRule -n open-cluster-management-observability
  ```
* Check that the new PrometheusRule and ScrapeConfig have been applied on a managed cluster. Change the kubecontext to a managed cluster and check that configs are present in the output of the following commands.
  ```
  kubectl get ScrapeConfig -n open-cluster-management-agent-addon
  kubectl get PrometheusRule -n open-cluster-management-agent-addon
  ```