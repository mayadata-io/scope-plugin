package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var Latency = make(map[string]float64)

type pvLatdata struct {
	readLatency  float64
	writeLatency float64
}

func getLatValues(urlpassed string, query string) {
	res, err := http.Get(urlpassed)
	if err != nil {
		panic(err.Error())
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err.Error())
	}

	s1, err := getValue([]byte(body))
	if err != nil {
		panic(err.Error())
	}

	rand.Seed(time.Now().Unix())

	for _, result := range s1.Data.Result {
		Latency[result.Metric.OpenebsPv], _ = strconv.ParseFloat(result.Value[1].(string), 32)
	}

	if query == "OpenEBS_read_latency" {
		readLatencych <- Latency
	} else if query == "OpenEBS_write_latency" {
		writeLatencych <- Latency
	}
}

func getLatPVs() map[string]pvLatdata {
	queries := []string{"OpenEBS_read_latency", "OpenEBS_write_latency"}
	for _, query := range queries {
		go getLatValues(url+query, query)
	}
	var readLatency, writeLatency map[string]float64
	for i := 0; i < len(queries); i++ {
		select {
		case readLatency = <-readLatencych:
		case writeLatency = <-writeLatencych:
		}
	}
	Latency := make(map[string]pvLatdata)
	if len(readLatency) == len(writeLatency) {
		for k1, v1 := range readLatency {
			metaLat, err := clientset.CoreV1().PersistentVolumes().Get(k1, metav1.GetOptions{})
			if err != nil {
				continue
			}
			Latency[string(metaLat.UID)] = pvLatdata{
				readLatency:  v1,
				writeLatency: writeLatency[k1],
			}
		}
	}
	return Latency
}

func (p *Plugin) updateLatPVs() {
	m1 := getLatPVs()
	if len(m1) > 0 {
		p.Latpvs = m1
	}
}

func (p *Plugin) getTopologyPv1(str string) string {
	return fmt.Sprintf("%s;<persistent_volume>", str)
}
