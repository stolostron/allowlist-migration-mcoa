package migration

import (
	"reflect"
	"testing"

	promv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	prometheusalpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestMigrateAllowlist(t *testing.T) {

	allowlist := corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "custom-allowlist",
			Namespace: "open-cluster-management-observability",
		},
		Data: map[string]string{
			"metrics_list.yaml": `
  names:
    - up
  matches:
    - __name__="container_memory_cache",container!=""
  recording_rules:
    - record: container_memory_rss:sum
      expr: sum(container_memory_rss) by (container, namespace)
`,
		},
	}

	// Expected ScrapeConfig output
	jobName := "custom-job"
	metricsPath := "/federate"
	controller := true
	expectedScrapeConfig := &prometheusalpha1.ScrapeConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ScrapeConfig",
			APIVersion: "monitoring.rhobs/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "custom-allowlist-scrapeconfig",
			Namespace: "open-cluster-management-observability",
			Labels: map[string]string{
				"app.kubernetes.io/component":  "platform-metrics-collector",
				"app.kubernetes.io/part-of":    "multicluster-observability-addon",
				"app.kubernetes.io/managed-by": "multicluster-observability-operator",
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "observability.open-cluster-management.io/v1beta2",
					Kind:       "MultiClusterObservability",
					Name:       "observability",
					UID:        "<MCO-UID>",
					Controller: &controller,
				},
			},
		},
		Spec: prometheusalpha1.ScrapeConfigSpec{
			JobName:     &jobName,
			MetricsPath: &metricsPath,
			Params: map[string][]string{
				"match[]": {
					`{__name__="up"}`,
					`{__name__="container_memory_cache",container!=""}`,
				},
			},
		},
	}

	// Expected PrometheusRule output
	expectedPrometheusRule := &promv1.PrometheusRule{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PrometheusRule",
			APIVersion: "monitoring.coreos.com/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "custom-allowlist-prometheusrule",
			Namespace: "open-cluster-management-observability",
			Labels: map[string]string{
				"app.kubernetes.io/component":  "platform-metrics-collector",
				"app.kubernetes.io/part-of":    "multicluster-observability-addon",
				"app.kubernetes.io/managed-by": "multicluster-observability-operator",
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "observability.open-cluster-management.io/v1beta2",
					Kind:       "MultiClusterObservability",
					Name:       "observability",
					UID:        "<MCO-UID>",
					Controller: &controller,
				},
			},
		},
		Spec: promv1.PrometheusRuleSpec{
			Groups: []promv1.RuleGroup{
				{
					Name: "custom-allowlist.rules",
					Rules: []promv1.Rule{
						{
							Record: "container_memory_rss:sum",
							Expr:   intstr.FromString("sum(container_memory_rss) by (container, namespace)"),
						},
					},
				},
			},
		},
	}

	scrapeConfig, prometheusRule, err := MigrateAllowlist(allowlist, "metrics_list.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Check specific TypeMeta fields
	if scrapeConfig.Kind != expectedScrapeConfig.Kind {
		t.Errorf("expected Kind '%s', got '%s'", expectedScrapeConfig.Kind, scrapeConfig.Kind)
	}
	if scrapeConfig.APIVersion != expectedScrapeConfig.APIVersion {
		t.Errorf("expected APIVersion '%s', got '%s'", expectedScrapeConfig.APIVersion, scrapeConfig.APIVersion)
	}

	if scrapeConfig.Namespace != expectedScrapeConfig.Namespace {
		t.Errorf("expected Namespace '%s', got '%s'", expectedScrapeConfig.Namespace, scrapeConfig.Namespace)
	}

	// Check OwnerReferences for ScrapeConfig
	if len(scrapeConfig.OwnerReferences) != len(expectedScrapeConfig.OwnerReferences) {
		t.Fatalf("expected %d OwnerReference(s), got %d", len(expectedScrapeConfig.OwnerReferences), len(scrapeConfig.OwnerReferences))
	}
	if scrapeConfig.OwnerReferences[0].Kind != expectedScrapeConfig.OwnerReferences[0].Kind {
		t.Errorf("expected OwnerReference Kind '%s', got '%s'", expectedScrapeConfig.OwnerReferences[0].Kind, scrapeConfig.OwnerReferences[0].Kind)
	}
	if scrapeConfig.OwnerReferences[0].Name != expectedScrapeConfig.OwnerReferences[0].Name {
		t.Errorf("expected OwnerReference Name '%s', got '%s'", expectedScrapeConfig.OwnerReferences[0].Name, scrapeConfig.OwnerReferences[0].Name)
	}
	if scrapeConfig.OwnerReferences[0].Controller == nil || !*scrapeConfig.OwnerReferences[0].Controller {
		t.Errorf("expected OwnerReference Controller to be true")
	}

	if scrapeConfig.Spec.MetricsPath == nil || *scrapeConfig.Spec.MetricsPath != *expectedScrapeConfig.Spec.MetricsPath {
		t.Errorf("expected MetricsPath '%s', got '%v'", *expectedScrapeConfig.Spec.MetricsPath, scrapeConfig.Spec.MetricsPath)
	}

	if !reflect.DeepEqual(scrapeConfig.Spec.Params, expectedScrapeConfig.Spec.Params) {
		t.Errorf("ScrapeConfig.Spec.Params mismatch:\ngot:  %+v\nwant: %+v", scrapeConfig.Spec.Params, expectedScrapeConfig.Spec.Params)
	}

	if prometheusRule.Kind != expectedPrometheusRule.Kind {
		t.Errorf("expected Kind '%s', got '%s'", expectedPrometheusRule.Kind, prometheusRule.Kind)
	}
	if prometheusRule.APIVersion != expectedPrometheusRule.APIVersion {
		t.Errorf("expected APIVersion '%s', got '%s'", expectedPrometheusRule.APIVersion, prometheusRule.APIVersion)
	}

	if prometheusRule.Namespace != expectedPrometheusRule.Namespace {
		t.Errorf("expected Namespace '%s', got '%s'", expectedPrometheusRule.Namespace, prometheusRule.Namespace)
	}

	// Check OwnerReferences for PrometheusRule
	if len(prometheusRule.OwnerReferences) != len(expectedPrometheusRule.OwnerReferences) {
		t.Fatalf("expected %d OwnerReference(s), got %d", len(expectedPrometheusRule.OwnerReferences), len(prometheusRule.OwnerReferences))
	}
	if prometheusRule.OwnerReferences[0].Kind != expectedPrometheusRule.OwnerReferences[0].Kind {
		t.Errorf("expected OwnerReference Kind '%s', got '%s'", expectedPrometheusRule.OwnerReferences[0].Kind, prometheusRule.OwnerReferences[0].Kind)
	}
	if prometheusRule.OwnerReferences[0].Name != expectedPrometheusRule.OwnerReferences[0].Name {
		t.Errorf("expected OwnerReference Name '%s', got '%s'", expectedPrometheusRule.OwnerReferences[0].Name, prometheusRule.OwnerReferences[0].Name)
	}
	if prometheusRule.OwnerReferences[0].Controller == nil || !*prometheusRule.OwnerReferences[0].Controller {
		t.Errorf("expected OwnerReference Controller to be true")
	}
	// Compare entire Spec using DeepEqual
	if !reflect.DeepEqual(prometheusRule.Spec, expectedPrometheusRule.Spec) {
		t.Errorf("PrometheusRule.Spec mismatch:\ngot:  %+v\nwant: %+v", prometheusRule.Spec, expectedPrometheusRule.Spec)
	}

}
