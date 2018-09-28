package main

import (
	"io/ioutil"
	"net/http"
	"strconv"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	latencyReadChan  = make(chan map[string]float64)
	latencyWriteChan = make(chan map[string]float64)
)

var (
	latencyReadQuery  = "increase(openebs_read_time[5m])/1000000"
	latencyWriteQuery = "increase(openebs_write_time[5m])/1000000"
)

func getLatencyValues(query string) {
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

	latencyMetrics := make(map[string]float64)

	for _, result := range response.Data.Result {
		latencyMetrics[result.Metric.OpenebsPv], _ = strconv.ParseFloat(result.Value[1].(string), 32)
	}

	if query == latencyReadQuery {
		latencyReadChan <- latencyMetrics
	} else {
		latencyWriteChan <- latencyMetrics
	}
}

func getLatencyMetrics() map[string]PVMetrics {
	queries := []string{latencyReadQuery, latencyWriteQuery}
	for _, query := range queries {
		go getLatencyValues(query)
	}

	readLatency := make(map[string]float64)
	writeLatency := make(map[string]float64)

	for i := 0; i < len(queries); i++ {
		select {
		case readLatency = <-latencyReadChan:
		case writeLatency = <-latencyWriteChan:
		}
	}

	latency := make(map[string]PVMetrics)
	if len(readLatency) > 0 && len(writeLatency) > 0 {
		for pvName, latencyRead := range readLatency {
			meta, err := ClientSet.CoreV1().PersistentVolumes().Get(pvName, metav1.GetOptions{})
			if err != nil {
				log.Errorf("error in fetching PV: %+v", err)
				continue
			}

			latency[string(meta.UID)] = PVMetrics{
				ReadLatency:  latencyRead,
				WriteLatency: writeLatency[pvName],
			}
		}
	}
	return latency
}

func (p *Plugin) updateLatency() {
	latencyMetrics := getLatencyMetrics()
	if len(latencyMetrics) > 0 {
		p.Latency = latencyMetrics
	}
}
