package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var url string
var clientset *kubernetes.Clientset
var readch = make(chan map[string]float64)
var writech = make(chan map[string]float64)
var readLatencych = make(chan map[string]float64)
var writeLatencych = make(chan map[string]float64)
var readThroughputch = make(chan map[string]float64)
var writeThroughputch = make(chan map[string]float64)

// Plugin groups the methods a plugin needs
type Plugin struct {
	pvs     map[string]pvdata
	Latpvs  map[string]pvLatdata
	Tputpvs map[string]pvTputdata
}

type request struct {
	NodeID string
}

type response struct {
	ShortcutReport *report `json:"shortcutReport,omitempty"`
}

type report struct {
	PersistentVolume topology
	Plugins          []pluginSpec
}

type topology struct {
	Nodes           map[string]node           `json:"nodes"`
	MetricTemplates map[string]metricTemplate `json:"metric_templates"`
}

type node struct {
	Metrics map[string]metric `json:"metrics"`
}

type metric struct {
	Samples []sample `json:"samples,omitempty"`
	Min     float64  `json:"min"`
	Max     float64  `json:"max"`
}

type sample struct {
	Date  time.Time `json:"date"`
	Value float64   `json:"value"`
}

type metricTemplate struct {
	ID       string  `json:"id"`
	Label    string  `json:"label,omitempty"`
	Format   string  `json:"format,omitempty"`
	Priority float64 `json:"priority,omitempty"`
}

type pluginSpec struct {
	ID          string   `json:"id"`
	Label       string   `json:"label"`
	Description string   `json:"description,omitempty"`
	Interfaces  []string `json:"interfaces"`
	APIVersion  string   `json:"api_version,omitempty"`
}

// Iops struct
type Iops struct {
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

// main function
func main() {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	// creates the clientset
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// gets Url
	url = os.Getenv("CORTEXAGENT")
	if url == "" {
		panic("Unable to retrieve the URL")
	}

	// we put the in a sub-directory to have more control on the permissions
	const socketPath = "/var/run/scope/plugins/iops/iops.sock"

	// Handle the exit signal
	setupSignals(socketPath)
	listener, err := setupSocket(socketPath)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		listener.Close()
		os.RemoveAll(filepath.Dir(socketPath))
	}()

	plugin, err := NewPlugin()
	if err != nil {
		log.Fatalf("Failed to create a plugin: %v", err)
	}
	http.HandleFunc("/report", plugin.Report)
	if err := http.Serve(listener, nil); err != nil {
		log.Printf("error: %v", err)
	}
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

// NewPlugin instantiates a new plugin
func NewPlugin() (*Plugin, error) {
	plugin := &Plugin{
		pvs:     getPVs(),
		Latpvs:  getLatPVs(),
		Tputpvs: getTputPVs(),
	}
	return plugin, nil
}

func getValue(body []byte) (*Iops, error) {
	storeBefore := new(Iops)
	err := json.Unmarshal(body, &storeBefore)
	if err != nil {
		fmt.Println("whoops:")
	}
	return storeBefore, err
}

func getLatValue(body []byte) (*Iops, error) {
	s1 := new(Iops)
	err := json.Unmarshal(body, &s1)
	if err != nil {
		fmt.Println("whoops:")
	}
	return s1, err
}

func getTputValue(body []byte) (*Iops, error) {
	s2 := new(Iops)
	err := json.Unmarshal(body, &s2)
	if err != nil {
		fmt.Println("whoops:")
	}
	return s2, err
}

func (p *Plugin) makeReport() (*report, error) {
	go p.updatePVs()
	go p.updateLatPVs()
	go p.updateTputPVs()
	metrics := make(map[string][]float64)
	resource := make(map[string]node)
	for k, v := range p.pvs {
		metrics[p.getTopologyPv(k)] = append(metrics[p.getTopologyPv(k)], v.read, v.write)
	}
	for x, y := range p.Latpvs {
		metrics[p.getTopologyPv1(x)] = append(metrics[p.getTopologyPv1(x)], y.readLatency, y.writeLatency)
	}
	for c, d := range p.Tputpvs {
		metrics[p.getTopologyPv2(c)] = append(metrics[p.getTopologyPv2(c)], d.readThroughput, d.writeThroughput)
	}
	for a, _ := range metrics {
		resource[a] = node{
			Metrics: p.metrics(metrics[a]),
		}
	}
	rpt := &report{
		PersistentVolume: topology{
			Nodes:           resource,
			MetricTemplates: p.metricTemplates(),
		},
		Plugins: []pluginSpec{
			{
				ID:          "iops",
				Label:       "iops",
				Description: "Adds a graph of read and write IOPS to PV",
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
		"r": {
			Samples: []sample{
				{
					Date:  time.Now(),
					Value: data[0],
				},
			},
			Min: 0,
			Max: 20,
		},
		"w": {
			Samples: []sample{
				{
					Date:  time.Now(),
					Value: data[1],
				},
			},
			Min: 0,
			Max: 20,
		},
		"r1": {
			Samples: []sample{
				{
					Date:  time.Now(),
					Value: data[2],
				},
			},
			Min: 0,
			Max: 20,
		},
		"w1": {
			Samples: []sample{
				{
					Date:  time.Now(),
					Value: data[3],
				},
			},
			Min: 0,
			Max: 20,
		},
		"r2": {
			Samples: []sample{
				{
					Date:  time.Now(),
					Value: data[4],
				},
			},
			Min: 0,
			Max: 20,
		},
		"w2": {
			Samples: []sample{
				{
					Date:  time.Now(),
					Value: data[5],
				},
			},
			Min: 0,
			Max: 20,
		},
	}
	return metrics
}

func (p *Plugin) metricTemplates() map[string]metricTemplate {
	return map[string]metricTemplate{
		"r": {
			ID:       "r",
			Label:    "Iops(R)",
			Format:   "",
			Priority: 0.1,
		},
		"w": {
			ID:       "w",
			Label:    "Iops(W)",
			Format:   "",
			Priority: 0.2,
		},
		"r1": {
			ID:       "r1",
			Label:    "Latency(R)",
			Format:   "millisecond",
			Priority: 0.3,
		},
		"w1": {
			ID:       "w1",
			Label:    "Latency(W)",
			Format:   "millisecond",
			Priority: 0.4,
		},
		"r2": {
			ID:       "r2",
			Label:    "Tput(R)",
			Format:   "bytes",
			Priority: 0.5,
		},
		"w2": {
			ID:       "w2",
			Label:    "Tput(W)",
			Format:   "bytes",
			Priority: 0.6,
		},
	}
}

// Report is called by scope when a new report is needed. It is part of the
// "reporter" interface, which all plugins must implement.
func (p *Plugin) Report(w http.ResponseWriter, r *http.Request) {
	rpt, err := p.makeReport()
	if err != nil {
		log.Printf("error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	raw, err := json.Marshal(*rpt)
	if err != nil {
		log.Printf("error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(raw)
}

func (p *Plugin) metricIDAndName() (string, string) {
	return "iops", "Iops"
}
