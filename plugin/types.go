package plugin

import (
	"time"

	"github.com/openebs/scope-plugin/metrics"
)

// Plugin struct groups the methods a plugin needs.
type Plugin struct {
	HostID string
	PVData map[string]metrics.PVMetrics
}

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
