package metrics

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	queries = []string{
		"iopsReadQuery",
		"iopsWriteQuery",
		"latencyReadQuery",
		"latencyWriteQuery",
		"throughputReadQuery",
		"throughputWriteQuery",
	}
)

// Report is called by scope when a new report is needed. It is part of the
// "reporter" interface, which all plugins must implement.
func (p *PVMetrics) Report(w http.ResponseWriter, r *http.Request) {
	rpt := p.makeReport()
	raw, err := json.Marshal(*rpt)
	if err != nil {
		log.Errorf("error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(raw)
}

// getPVTopology will create a UID by appending the UID with resource name.
func (p *PVMetrics) getPVTopology(PersistentVolumeUID string) string {
	return fmt.Sprintf("%s;<persistent_volume>", PersistentVolumeUID)
}

// makeReport will create the report.
func (p *PVMetrics) makeReport() *report {
	metrics := make(map[string][]float64)
	updatedMetrics := make(map[string][]float64)
	resource := make(map[string]node)

	for pvName := range p.PVList {
		metrics[pvName] = []float64{0, 0, 0, 0, 0, 0}
	}

	if p.Data != nil && p.PVList != nil && len(metrics) > 0 {
		for index, queryName := range queries {
			if p.Data[queryName] == nil {
				for k := range metrics {
					if _, ok := metrics[k]; ok {
						metrics[k][index] = 0
					}
				}
			} else {
				for k, v := range p.Data[queryName] {
					if _, ok := metrics[k]; ok {
						metrics[k][index] = v
					}
				}
			}
		}

		for pvName, pvUID := range p.PVList {
			updatedMetrics[p.getPVTopology(pvUID)] = metrics[pvName]
		}

		for pvNodeID := range updatedMetrics {
			resource[pvNodeID] = node{
				Metrics: p.metrics(updatedMetrics[pvNodeID]),
			}
		}
		rpt := &report{
			PersistentVolume: topology{
				Nodes:           resource,
				MetricTemplates: p.metricTemplates(),
			},
			Plugins: []pluginSpec{
				{
					ID:          "openebs",
					Label:       "OpenEBS Monitor Plugin",
					Description: "OpenEBS Monitor Plugin: Monitor OpeneEBS volumes",
					Interfaces:  []string{"reporter"},
					APIVersion:  "1",
				},
			},
		}
		return rpt
	}

	rpt := &report{
		PersistentVolume: topology{
			Nodes:           nil,
			MetricTemplates: p.metricTemplates(),
		},
		Plugins: []pluginSpec{
			{
				ID:          "openebs",
				Label:       "OpenEBS Monitor Plugin",
				Description: "OpenEBS Monitor Plugin: Monitor OpeneEBS volumes",
				Interfaces:  []string{"reporter"},
				APIVersion:  "1",
			},
		},
	}
	return rpt
}

// Create the Metrics type on top-left side
func (p *PVMetrics) metrics(data []float64) map[string]metric {
	metrics := map[string]metric{
		"readIops": {
			Samples: []sample{
				{
					Date:  time.Now(),
					Value: float64(int(data[0] + 0.5)),
				},
			},
			Min: 0,
			Max: 100,
		},
		"writeIops": {
			Samples: []sample{
				{
					Date:  time.Now(),
					Value: float64(int(data[1] + 0.5)),
				},
			},
			Min: 0,
			Max: 100,
		},
		"readLatency": {
			Samples: []sample{
				{
					Date:  time.Now(),
					Value: data[2],
				},
			},
			Min: 0,
			Max: 100,
		},
		"writeLatency": {
			Samples: []sample{
				{
					Date:  time.Now(),
					Value: data[3],
				},
			},
			Min: 0,
			Max: 100,
		},
		"readThroughput": {
			Samples: []sample{
				{
					Date:  time.Now(),
					Value: data[4],
				},
			},
			Min: 0,
			Max: 100,
		},
		"writeThroughput": {
			Samples: []sample{
				{
					Date:  time.Now(),
					Value: data[5],
				},
			},
			Min: 0,
			Max: 100,
		},
	}
	return metrics
}

func (p *PVMetrics) metricTemplates() map[string]metricTemplate {
	return map[string]metricTemplate{
		"readIops": {
			ID:       "readIops",
			Label:    "Iops(R)",
			Format:   "",
			Priority: 0.1,
		},
		"writeIops": {
			ID:       "writeIops",
			Label:    "Iops(W)",
			Format:   "",
			Priority: 0.2,
		},
		"readLatency": {
			ID:       "readLatency",
			Label:    "Latency(R)",
			Format:   "millisecond",
			Priority: 0.3,
		},
		"writeLatency": {
			ID:       "writeLatency",
			Label:    "Latency(W)",
			Format:   "millisecond",
			Priority: 0.4,
		},
		"readThroughput": {
			ID:       "readThroughput",
			Label:    "Throughput(R)",
			Format:   "bytes",
			Priority: 0.5,
		},
		"writeThroughput": {
			ID:       "writeThroughput",
			Label:    "Throughput(W)",
			Format:   "bytes",
			Priority: 0.6,
		},
	}
}

func (p *PVMetrics) metricIDAndName() (string, string) {
	return "OpenEBS Plugin", "OpenEBS Plugin"
}
