package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Plugin groups the methods a plugin needs
type Plugin struct {
	pvs map[string]float64
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

var url string
var clientset *kubernetes.Clientset
var ch chan map[string]float64

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
	const socketPath = "/var/run/scope/plugins/iowait/iowait.sock"
	// Handle the exit signal
	setupSignals(socketPath)

	listener, err := setupSocket(socketPath)
	if err != nil {
		log.Fatal(err)
	}
	ch = make(chan map[string]float64)
	defer func() {
		listener.Close()
		os.RemoveAll(filepath.Dir(socketPath))
	}()

	// var uid []string

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

func getPVS() map[string]float64 {

	// Get request to url
	res, err := http.Get(url)
	if err != nil {
		panic(err.Error())
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err.Error())
	}
	pv := make(map[string]float64)
	s, err := getValue([]byte(body))
	rand.Seed(time.Now().Unix())
	for _, pvName := range s.Data.Result {
		meta, err := clientset.CoreV1().PersistentVolumes().Get(pvName.Metric.OpenebsPv, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			fmt.Printf("Pod not found\n")
		} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
			fmt.Printf("Error getting pod %v\n", statusError.ErrStatus.Message)
		} else if err != nil {
			panic(err.Error())
		} else {
			log.Printf("Found pod\n")
		}
		pv[string(meta.UID)] = pvName.Value[0].(float64)
	}
	// fmt.Printf("%+v\n", pv)
	return pv
}

// NewPlugin instantiates a new plugin
func NewPlugin() (*Plugin, error) {
	plugin := &Plugin{
		pvs: getPVS(),
	}
	return plugin, nil
}

func getValue(body []byte) (*Iops, error) {
	var s = new(Iops)
	err := json.Unmarshal(body, &s)
	if err != nil {
		fmt.Println("whoops:")
	}
	return s, err
}
func (p *Plugin) updatePVS() {
	p.pvs = getPVS()
}
func (p *Plugin) makeReport() (*report, error) {
	go p.updatePVS()
	nod := make(map[string]node)
	for k, v := range p.pvs {
		nod[p.getTopologyPv(k)] = node{
			Metrics: p.metrics(v),
		}
	}
	rpt := &report{
		PersistentVolume: topology{
			Nodes:           nod,
			MetricTemplates: p.metricTemplates(),
		},
		Plugins: []pluginSpec{
			{
				ID:          "iowait",
				Label:       "iowait",
				Description: "Adds a graph of CPU IO Wait to Pod",
				Interfaces:  []string{"reporter"},
				APIVersion:  "1",
			},
		},
	}
	return rpt, nil
}

func (p *Plugin) metrics(value float64) map[string]metric {
	id, _ := p.metricIDAndName()
	metrics := map[string]metric{
		id: {
			Samples: []sample{
				{
					Date:  time.Now(),
					Value: value,
				},
			},
			Min: 0,
			Max: 100,
		},
	}
	return metrics
}

func (p *Plugin) metricTemplates() map[string]metricTemplate {
	id, name := p.metricIDAndName()
	return map[string]metricTemplate{
		id: {
			ID:       id,
			Label:    name,
			Format:   "percentage",
			Priority: 0.1,
		},
	}
}

// Report is called by scope when a new report is needed. It is part of the
// "reporter" interface, which all plugins must implement.
func (p *Plugin) Report(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL.String())
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

func (p *Plugin) getTopologyPv(str string) string {
	return fmt.Sprintf("%s;<persistent_volume>", str)
}

func (p *Plugin) metricIDAndName() (string, string) {
	return "iops", "Iops"
}
