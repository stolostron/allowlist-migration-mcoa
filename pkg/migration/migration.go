package migration

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/util/intstr"

	pkgyaml "gopkg.in/yaml.v2"

	promv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	prometheusalpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	operatorconfig "github.com/stolostron/multicluster-observability-operator/operators/pkg/config"
)

func MigrateAllowlist(cm corev1.ConfigMap, targets string) (*prometheusalpha1.ScrapeConfig, *promv1.PrometheusRule, error) {
	allowlist := &operatorconfig.MetricsAllowlist{}
	err := pkgyaml.Unmarshal([]byte(cm.Data[targets]), allowlist)
	if err != nil {
		return nil, nil, err
	}

	objMeta := cm.ObjectMeta
	if objMeta.Name == "" {
		objMeta.Name = "allowlist"
	}
	if objMeta.Namespace == "" {
		objMeta.Namespace = "open-cluster-management-observability"
	}
	scrapeConfig, err := createScrapeConfig(*allowlist, objMeta)
	if err != nil {
		return nil, nil, err
	}

	prometheusRule, err := createPrometheusRule(*allowlist, objMeta)
	if err != nil {
		return nil, nil, err
	}

	return scrapeConfig, prometheusRule, nil
}

func createScrapeConfig(allowlist operatorconfig.MetricsAllowlist, objMeta metav1.ObjectMeta) (*prometheusalpha1.ScrapeConfig, error) {
	matches := []string{}
	nameCnt := 0
	// names entries
	for _, name := range allowlist.NameList {
		// Skip empty names
		if strings.TrimSpace(name) == "" {
			continue
		}
		matches = append(matches, fmt.Sprintf(`{__name__="%s"}`, name))
		nameCnt++
	}

	// Add matches entries (already in the right format, just need to wrap with {})
	matchCnt := 0
	for _, match := range allowlist.MatchList {
		if strings.TrimSpace(match) == "" {
			continue
		}
		// We need to wrap MatchList entries with {}
		matches = append(matches, fmt.Sprintf(`{%s}`, match))
		matchCnt++
	}

	fmt.Printf("  - NameList: %d entries\n", nameCnt)
	fmt.Printf("  - MatchList: %d entries\n", matchCnt)

	// Add recording rule name
	for _, rule := range allowlist.RecordingRuleList {
		matches = append(matches, fmt.Sprintf("{__name__=\"%s\"}", rule.Record))
	}
	labels := map[string]string{}
	labels["app.kubernetes.io/component"] = "platform-metrics-collector"
	labels["app.kubernetes.io/part-of"] = "multicluster-observability-addon"
	labels["app.kubernetes.io/managed-by"] = "multicluster-observability-operator"

	jobName := "custom-job"
	metricsPath := "/federate"
	scrapeConfig := &prometheusalpha1.ScrapeConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ScrapeConfig",
			APIVersion: "monitoring.rhobs/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "custom-scrapeconfig",
			Namespace: objMeta.Namespace,
			Labels:    labels,
		},
		Spec: prometheusalpha1.ScrapeConfigSpec{
			JobName:     &jobName,
			MetricsPath: &metricsPath,
			Params: map[string][]string{
				"match[]": matches,
			},
		},
	}

	return scrapeConfig, nil
}

// adding recording rule to rule group. Does not process collectRuleGroup as dynamic metrics not supported
func createPrometheusRule(allowlist operatorconfig.MetricsAllowlist, objMeta metav1.ObjectMeta) (*promv1.PrometheusRule, error) {
	// recording rules are like:
	// - name: "apiserver_request_duration_seconds:histogram_quantile_99"
	//   expr: "histogram_quantile(0.99,sum(rate(apiserver_request_duration_seconds_bucket{job=\"apiserver\", verb!=\"WATCH\"}[5m])) by (le))"

	recordingRules := []promv1.Rule{}
	ruleCnt := 0
	for _, rule := range allowlist.RecordingRuleList {
		recordingRules = append(recordingRules, promv1.Rule{
			Record: rule.Record,
			Expr:   intstr.IntOrString{Type: intstr.String, StrVal: strings.ReplaceAll(rule.Expr, `\"`, `"`)},
		})
		ruleCnt++
	}
	fmt.Printf("  - RecordingRuleList: %d entries\n", ruleCnt)

	labels := map[string]string{}
	labels["app.kubernetes.io/component"] = "platform-metrics-collector"
	labels["app.kubernetes.io/part-of"] = "multicluster-observability-addon"
	labels["app.kubernetes.io/managed-by"] = "multicluster-observability-operator"
	prometheusRule := &promv1.PrometheusRule{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PrometheusRule",
			APIVersion: "monitoring.coreos.com/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "custom-prometheusrule",
			Namespace: objMeta.Namespace,
			Labels:    labels,
		},
		Spec: promv1.PrometheusRuleSpec{
			Groups: []promv1.RuleGroup{
				{
					Name:  "custom-prometheusrule.rules",
					Rules: recordingRules,
				},
			},
		},
	}
	return prometheusRule, nil

}
