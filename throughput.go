package main

import (
	"io/ioutil"
	"net/http"
	"strconv"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	throughputReadChan  = make(chan map[string]float64)
	throughputWriteChan = make(chan map[string]float64)
)

var (
	throughputReadQuery  = "increase(openebs_read_block_count[5m])/(1024*1024*60*5)"
	throughputWriteQuery = "increase(openebs_write_block_count[5m])/(1024*1024*60*5)"
)

func getThroughputValues(query string) {
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

	throughputMetrics := make(map[string]float64)

	for _, result := range response.Data.Result {
		throughputMetrics[result.Metric.OpenebsPv], _ = strconv.ParseFloat(result.Value[1].(string), 32)
	}

	if query == throughputReadQuery {
		throughputReadChan <- throughputMetrics
	} else {
		throughputWriteChan <- throughputMetrics
	}
}

func getThroughputMetrics() map[string]PVMetrics {
	queries := []string{throughputReadQuery, throughputWriteQuery}
	for _, query := range queries {
		go getThroughputValues(query)
	}

	readThroughput := make(map[string]float64)
	writeThroughput := make(map[string]float64)

	for i := 0; i < len(queries); i++ {
		select {
		case readThroughput = <-throughputReadChan:
		case writeThroughput = <-throughputWriteChan:
		}
	}

	throughput := make(map[string]PVMetrics)
	if len(readThroughput) > 0 && len(writeThroughput) > 0 {
		for pvName, throughputRead := range readThroughput {
			meta, err := ClientSet.CoreV1().PersistentVolumes().Get(pvName, metav1.GetOptions{})
			if err != nil {
				log.Errorf("error in fetching PV: %+v", err)
				continue
			}

			throughput[string(meta.UID)] = PVMetrics{
				ReadThroughput:  throughputRead,
				WriteThroughput: writeThroughput[pvName],
			}
		}
	}
	return throughput
}

func (p *Plugin) updateThroughput() {
	throughputMetrics := getThroughputMetrics()
	if len(throughputMetrics) > 0 {
		p.Throughput = throughputMetrics
	}
}
