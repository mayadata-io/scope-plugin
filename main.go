package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	URL       string
	ClientSet *kubernetes.Clientset
)

type PVMetrics struct {
	ReadIops        float64
	WriteIops       float64
	ReadLatency     float64
	WriteLatency    float64
	ReadThroughput  float64
	WriteThroughput float64
}

type Plugin struct {
	HostID     string
	Iops       map[string]PVMetrics
	Latency    map[string]PVMetrics
	Throughput map[string]PVMetrics
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

type Metrics struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric struct {
				Name              string `json:"__name__"`
				Instance          string `json:"instance"`
				Job               string `json:"job"`
				KubernetesPodName string `json:"kubernetes_pod_name"`
				OpenebsPv         string `json:"openebs_pv"`
			} `json:"metric"`
			Value []interface{} `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

func setupSocket(socketPath string) (net.Listener, error) {
	os.RemoveAll(filepath.Dir(socketPath))
	if err := os.MkdirAll(filepath.Dir(socketPath), 0700); err != nil {
		return nil, fmt.Errorf("failed to create directory %q: %v", filepath.Dir(socketPath), err)
	}
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %q: %v", socketPath, err)
	}

	log.Printf("Listening on: unix://%s", socketPath)
	return listener, nil
}

func setupSignals(socketPath string) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-interrupt
		os.RemoveAll(filepath.Dir(socketPath))
		os.Exit(0)
	}()
}

func Response(response []byte) (*Metrics, error) {
	result := new(Metrics)
	err := json.Unmarshal(response, &result)
	return result, err
}

func main() {
	// creating in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Error(err)
	}

	// create clientset of kubernetes
	ClientSet, err = kubernetes.NewForConfig(config)
	if err != nil {
		log.Error(err)
	}

	URL = os.Getenv("CORTEXAGENT")
	if URL == "" {
		log.Info("Unable to get cortex agent URL")
	}

	// Put socket in sub-directory to have more control on permissions
	const socketPath = "/var/run/scope/plugins/openebs/openebs.sock"

	// Handle the exit signal
	setupSignals(socketPath)
	listener, err := setupSocket(socketPath)
	if err != nil {
		log.Error(err)
	}

	defer func() {
		listener.Close()
		os.RemoveAll(filepath.Dir(socketPath))
	}()

	plugin := &Plugin{
		HostID: "host",
	}

	http.HandleFunc("/report", plugin.Report)
	if err := http.Serve(listener, nil); err != nil {
		log.Errorf("error: %v", err)
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

func (p *Plugin) getPVTopology(PVName string) string {
	return fmt.Sprintf("%s;<persistent_volume>", PVName)
}

func (p *Plugin) makeReport() (*report, error) {
	go p.updateIops()
	go p.updateLatency()
	go p.updateThroughput()

	metrics := make(map[string][]float64)
	resource := make(map[string]node)

	for pvUID, iopsMetrics := range p.Iops {
		metrics[p.getPVTopology(pvUID)] = append(metrics[p.getPVTopology(pvUID)], iopsMetrics.ReadIops, iopsMetrics.WriteIops)
	}

	for pvUID, latencyMetrics := range p.Latency {
		metrics[p.getPVTopology(pvUID)] = append(metrics[p.getPVTopology(pvUID)], latencyMetrics.ReadLatency, latencyMetrics.WriteLatency)
	}

	for pvUID, throughputMetrics := range p.Throughput {
		metrics[p.getPVTopology(pvUID)] = append(metrics[p.getPVTopology(pvUID)], throughputMetrics.ReadThroughput, throughputMetrics.WriteThroughput)
	}

	for pvName, _ := range metrics {
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
