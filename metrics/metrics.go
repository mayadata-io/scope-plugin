package metrics

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Query parameters for cortex agent.
const (
	iopsReadQuery        = "increase(openebs_reads[5m])/300"
	iopsWriteQuery       = "increase(openebs_writes[5m])/300"
	latencyReadQuery     = "((increase(openebs_read_time[5m]))/(increase(openebs_reads[5m])))/1000000"
	latencyWriteQuery    = "((increase(openebs_write_time[5m]))/(increase(openebs_writes[5m])))/1000000"
	throughputReadQuery  = "increase(openebs_read_block_count[5m])/(1024*1024*60*5)"
	throughputWriteQuery = "increase(openebs_write_block_count[5m])/(1024*1024*60*5)"
	// URL is the address of cortex agent.
	URL = "http://cortex-agent-service.maya-system.svc.cluster.local:80/api/v1/query?query="
)

// Map to store the query response.
var (
	readIops        = make(map[string]float64)
	writeIops       = make(map[string]float64)
	readLatency     = make(map[string]float64)
	writeLatency    = make(map[string]float64)
	readThroughput  = make(map[string]float64)
	writeThroughput = make(map[string]float64)
	// Clientset contains kubernetes client.
	ClientSet *kubernetes.Clientset
)

// Mutex is used to lock over metrics structure.
var Mutex = &sync.Mutex{}

// Response unmarshal the obtained Metric json
func Response(response []byte) (*Metrics, error) {
	result := new(Metrics)
	err := json.Unmarshal(response, &result)
	return result, err
}

//GetValues will get the values from the cortex agent.
func GetValues() map[string]PVMetrics {
	queries := []string{iopsReadQuery, iopsWriteQuery, latencyReadQuery, latencyWriteQuery, throughputReadQuery, throughputWriteQuery}
	for _, query := range queries {
		res, err := http.Get(URL + query)
		if err != nil {
			log.Error(err)
		}

		responseBody, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Error(err)
		}

		response, err := Response([]byte(responseBody))
		if err != nil {
			log.Error(err)
		}

		metrics := make(map[string]float64)

		for _, result := range response.Data.Result {
			if result.Value[1].(string) != "NaN" {
				floatVal, err := strconv.ParseFloat(result.Value[1].(string), 32)
				if err == nil {
					metrics[result.Metric.OpenebsPv] = floatVal
				} else {
					log.Error(err)
				}
			} else {
				metrics[result.Metric.OpenebsPv] = 0
			}
		}

		if query == iopsReadQuery {
			readIops = metrics
		} else if query == iopsWriteQuery {
			writeIops = metrics
		} else if query == latencyReadQuery {
			readLatency = metrics
		} else if query == latencyWriteQuery {
			writeLatency = metrics
		} else if query == throughputReadQuery {
			readThroughput = metrics
		} else if query == throughputWriteQuery {
			writeThroughput = metrics
		}
	}

	data := make(map[string]PVMetrics)
	if len(readIops) > 0 && len(writeIops) > 0 && len(readLatency) > 0 && len(writeLatency) > 0 && len(readThroughput) > 0 && len(writeThroughput) > 0 {
		for pvName, iopsRead := range readIops {
			meta, err := ClientSet.CoreV1().PersistentVolumes().Get(pvName, metav1.GetOptions{})
			if err != nil {
				log.Errorf("error in fetching PV: %+v", err)
				continue
			}

			metrics := PVMetrics{
				ReadIops: iopsRead,
			}

			if val, ok := writeIops[pvName]; ok {
				metrics.WriteIops = val
			}
			if val, ok := readLatency[pvName]; ok {
				metrics.ReadLatency = val
			}
			if val, ok := writeLatency[pvName]; ok {
				metrics.WriteLatency = val
			}
			if val, ok := readThroughput[pvName]; ok {
				metrics.ReadThroughput = val
			}
			if val, ok := writeThroughput[pvName]; ok {
				metrics.WriteThroughput = val
			}
			data[string(meta.UID)] = metrics
		}
	}
	return data
}
