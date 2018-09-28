package main

import (
	"io/ioutil"
	"net/http"
	"strconv"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	iopsReadChan  = make(chan map[string]float64)
	iopsWriteChan = make(chan map[string]float64)
)

var (
	iopsReadQuery  = "increase(openebs_reads[5m])/300"
	iopsWriteQuery = "increase(openebs_writes[5m])/300"
)

func getValues(query string) {
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

	iopsMetrics := make(map[string]float64)

	for _, result := range response.Data.Result {
		iopsMetrics[result.Metric.OpenebsPv], _ = strconv.ParseFloat(result.Value[1].(string), 32)
	}

	if query == iopsReadQuery {
		iopsReadChan <- iopsMetrics
	} else {
		iopsWriteChan <- iopsMetrics
	}
}

func getIopsMetrics() map[string]PVMetrics {
	queries := []string{iopsReadQuery, iopsWriteQuery}
	for _, query := range queries {
		go getValues(query)
	}

	readIops := make(map[string]float64)
	writeIops := make(map[string]float64)

	for i := 0; i < len(queries); i++ {
		select {
		case readIops = <-iopsReadChan:
		case writeIops = <-iopsWriteChan:
		}
	}

	iops := make(map[string]PVMetrics)
	if len(readIops) > 0 && len(writeIops) > 0 {
		for pvName, iopsRead := range readIops {
			meta, err := ClientSet.CoreV1().PersistentVolumes().Get(pvName, metav1.GetOptions{})
			if err != nil {
				log.Errorf("error in fetching PV: %+v", err)
				continue
			}

			iops[string(meta.UID)] = PVMetrics{
				ReadIops:  iopsRead,
				WriteIops: writeIops[pvName],
			}
		}
	}
	return iops
}

func (p *Plugin) updateIops() {
	iopsMetrics := getIopsMetrics()
	if len(iopsMetrics) > 0 {
		p.Iops = iopsMetrics
	}
}
