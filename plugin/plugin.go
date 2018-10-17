package plugin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/openebs/scope-plugin/metrics"
	log "github.com/sirupsen/logrus"
)

// UpdateMetrics calls the Getvalues every 60seconds and stores 
// the metrics received from the cortex agent.
func (p *Plugin) UpdateMetrics() {
	for {
		Metrics := metrics.GetValues()
		if len(Metrics) > 0 {
			metrics.Mutex.Lock()
			p.PVData = Metrics
			metrics.Mutex.Unlock()
		}
		time.Sleep(60 * time.Second)
	}
}

// Report is called by scope when a new report is needed. It is part of the
// "reporter" interface, which all plugins must implement.
func (p *Plugin) Report(w http.ResponseWriter, r *http.Request) {
	rpt, err := p.makeReport()
	if err != nil {
		log.Errorf("error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	raw, err := json.Marshal(*rpt)
	if err != nil {
		log.Errorf("error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(raw)
}

func ReportWrapper(p *Plugin, w http.ResponseWriter, r *http.Request) {
	p.Report(w, r)
}

func (p *Plugin) getPVTopology(PVName string) string {
	return fmt.Sprintf("%s;<persistent_volume>", PVName)
}

func (p *Plugin) makeReport() (*report, error) {

	metrics := make(map[string][]float64)
	resource := make(map[string]node)

	for pvUID, values := range p.PVData {
		metrics[p.getPVTopology(pvUID)] = append(metrics[p.getPVTopology(pvUID)], values.ReadIops, values.WriteIops, values.ReadLatency, values.WriteLatency, values.ReadThroughput, values.WriteThroughput)
	}

	for pvName := range metrics {
		resource[pvName] = node{
			Metrics: p.metrics(metrics[pvName]),
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
				Label:       "OpenEBS Plugin",
				Description: "Adds graphs of metrics of OpenEBS PV",
				Interfaces:  []string{"reporter"},
				APIVersion:  "1",
			},
		},
	}
	return rpt, nil
}

// Create the Metrics type on top-left side
func (p *Plugin) metrics(data []float64) map[string]metric {
	metrics := map[string]metric{
		"readIops": {
			Samples: []sample{
				{
					Date:  time.Now(),
					Value: data[0],
				},
			},
			Min: 0,
			Max: 100,
		},
		"writeIops": {
			Samples: []sample{
				{
					Date:  time.Now(),
					Value: data[1],
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

func (p *Plugin) metricTemplates() map[string]metricTemplate {
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

func (p *Plugin) metricIDAndName() (string, string) {
	return "OpenEBS Plugin", "OpenEBS Plugin"
}
