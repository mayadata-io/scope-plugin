package metrics

import (
	"time"

	"k8s.io/client-go/kubernetes"
)

type sample struct {
	Date  time.Time `json:"date"`
	Value float64   `json:"value"`
}

type metric struct {
	Samples []sample `json:"samples,omitempty"`
	Min     float64  `json:"min"`
	Max     float64  `json:"max"`
}

type metricTemplate struct {
	ID       string  `json:"id"`
	Label    string  `json:"label,omitempty"`
	Format   string  `json:"format,omitempty"`
	Priority float64 `json:"priority,omitempty"`
}

type node struct {
	Metrics map[string]metric `json:"metrics"`
}

type topology struct {
	Nodes           map[string]node           `json:"nodes"`
	MetricTemplates map[string]metricTemplate `json:"metric_templates"`
}

type pluginSpec struct {
	ID          string   `json:"id"`
	Label       string   `json:"label"`
	Description string   `json:"description,omitempty"`
	Interfaces  []string `json:"interfaces"`
	APIVersion  string   `json:"api_version,omitempty"`
}

type request struct {
	NodeID string
}

type report struct {
	PersistentVolume topology
	Plugins          []pluginSpec
}

type response struct {
	ShortcutReport *report `json:"shortcutReport,omitempty"`
}

// PVMetrics will store all the queries and data.
type PVMetrics struct {
	Queries   map[string]string
	PVList    map[string]string
	Data      map[string]map[string]float64
	ClientSet kubernetes.Interface
}

type Metric struct {
	Name              string `json:"__name__"`
	Instance          string `json:"instance"`
	Job               string `json:"job"`
	KubernetesPodName string `json:"kubernetes_pod_name"`
	OpenebsPv         string `json:"openebs_pv"`
	OpenebsPvc        string `json:"openebs_pvc"`
}

type Result struct {
	Metric Metric        `json:"metric"`
	Value  []interface{} `json:"value"`
}

type Data struct {
	ResultType string   `json:"resultType"`
	Result     []Result `json:"result"`
}

// Metrics stores the json provided by cortex agent.
type Metrics struct {
	Status string `json:"status"`
	Data   Data   `json:"data"`
}
