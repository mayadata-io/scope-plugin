package metrics

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/openebs/scope-plugin/k8s"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// URL is the address of cortex agent.
var (
	URL = "http://cortex-agent-service.maya-system.svc.cluster.local:80/api/v1/query?query="
)

// Mutex is used to lock over metrics structure.
var Mutex = &sync.Mutex{}

// NewMetrics will return an object of PVMetrics struct initialized with the queries.
func NewMetrics() PVMetrics {
	return PVMetrics{
		Queries: map[string]string{
			"iopsReadQuery":        "increase(openebs_reads[5m])/300",
			"iopsWriteQuery":       "increase(openebs_writes[5m])/300",
			"latencyReadQuery":     "((increase(openebs_read_time[5m]))/(increase(openebs_reads[5m])))/1000000",
			"latencyWriteQuery":    "((increase(openebs_write_time[5m]))/(increase(openebs_writes[5m])))/1000000",
			"throughputReadQuery":  "increase(openebs_read_block_count[5m])/(1024*1024*60*5)",
			"throughputWriteQuery": "increase(openebs_write_block_count[5m])/(1024*1024*60*5)",
		},
		PVList:    nil,
		Data:      nil,
		ClientSet: k8s.NewClientSet(),
	}
}

// UpdateMetrics will update the metrics data and PV list
func (p *PVMetrics) UpdateMetrics() {
	for {
		p.UpdatePVMetrics()
		time.Sleep(2 * time.Second)
	}
}

// UpdatePVMetrics will update the PVMetrics struct object with the required data
func (p *PVMetrics) UpdatePVMetrics() {
	data := make(map[string]map[string]float64)
	for queryName, query := range p.Queries {
		pvMetricsvalue, err := p.GetMetrics(query)
		if err != nil {
			log.Error(err)
		}

		if pvMetricsvalue == nil {
			data = nil
			log.Infof("Failed to fetch metrics for %s", queryName)
			break
		}
		data[queryName] = pvMetricsvalue
	}

	if data != nil {
		Mutex.Lock()
		p.Data = data
		Mutex.Unlock()
	}

	p.GetPVList()
}

// GetMetrics will return the metrics for the given query.
func (p *PVMetrics) GetMetrics(query string) (map[string]float64, error) {
	response, err := http.Get(URL + query)
	if err != nil {
		return nil, err
	}

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	pvMetrics, err := p.UnmarshalResponse([]byte(responseBody))
	if err != nil {
		return nil, err
	}

	if len(pvMetrics.Data.Result) == 0 {
		return nil, errors.New("Result is empty")
	}

	pvMetricsValue := make(map[string]float64)
	for _, pvMetric := range pvMetrics.Data.Result {
		if pvMetric.Value[1].(string) == "NaN" {
			pvMetricsValue[pvMetric.Metric.OpenebsPv] = 0
		} else {
			metric, err := strconv.ParseFloat(pvMetric.Value[1].(string), 64)
			if err != nil {
				log.Error(err)
				pvMetricsValue[pvMetric.Metric.OpenebsPv] = 0
			} else {
				pvMetricsValue[pvMetric.Metric.OpenebsPv] = metric
			}
		}
	}

	return pvMetricsValue, nil
}

// GetPVList fetch and update the list of PV.
func (p *PVMetrics) GetPVList() {
	pvList, err := p.ClientSet.CoreV1().PersistentVolumes().List(metav1.ListOptions{})
	if err != nil {
		log.Error(err)
		return
	}

	p.PVList = p.PVNameAndUID(pvList.Items)
}

// PVNameAndUID returns the name and UID of all the PVs.
func (p *PVMetrics) PVNameAndUID(pvListItems []corev1.PersistentVolume) map[string]string {
	pvList := make(map[string]string)
	for _, pv := range pvListItems {
		pvList[pv.GetName()] = string(pv.GetUID())
	}
	return pvList
}

// UnmarshalResponse unmarshal the obtained Metric json.
func (p *PVMetrics) UnmarshalResponse(response []byte) (*Metrics, error) {
	metric := new(Metrics)
	err := json.Unmarshal(response, &metric)
	return metric, err
}
