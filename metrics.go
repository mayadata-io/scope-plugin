package main

import (
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	iopsReadQuery        = "increase(openebs_reads[5m])/300"
	iopsWriteQuery       = "increase(openebs_writes[5m])/300"
	latencyReadQuery     = "((increase(openebs_read_time[5m]))/(increase(openebs_reads[5m])))/1000000"
	latencyWriteQuery    = "((increase(openebs_write_time[5m]))/(increase(openebs_writes[5m])))/1000000"
	throughputReadQuery  = "increase(openebs_read_block_count[5m])/(1024*1024*60*5)"
	throughputWriteQuery = "increase(openebs_write_block_count[5m])/(1024*1024*60*5)"
)

var (
	readIops        = make(map[string]float64)
	writeIops       = make(map[string]float64)
	readLatency     = make(map[string]float64)
	writeLatency    = make(map[string]float64)
	readThroughput  = make(map[string]float64)
	writeThroughput = make(map[string]float64)
)

var mutex = &sync.Mutex{}

func getValues() map[string]PVMetrics {
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

		Metrics := make(map[string]float64)

		for _, result := range response.Data.Result {
			if result.Value[1].(string) != "NaN" {
				floatVal, err := strconv.ParseFloat(result.Value[1].(string), 32)
				if err == nil {
					Metrics[result.Metric.OpenebsPv] = floatVal
				} else {
					log.Error(err)
				}
			} else {
				Metrics[result.Metric.OpenebsPv] = 0
			}
		}

		if query == iopsReadQuery {
			readIops = Metrics
		} else if query == iopsWriteQuery {
			writeIops = Metrics
		} else if query == latencyReadQuery {
			readLatency = Metrics
		} else if query == latencyWriteQuery {
			writeLatency = Metrics
		} else if query == throughputReadQuery {
			readThroughput = Metrics
		} else if query == throughputWriteQuery {
			writeThroughput = Metrics
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

func (p *Plugin) updateIops() {
	for {
		Metrics := getValues()
		if len(Metrics) > 0 {
			mutex.Lock()
			p.Iops = Metrics
			mutex.Unlock()
		}
		time.Sleep(60 * time.Second)
	}
}
