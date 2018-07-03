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

var Throughput = make(map[string]float64)

type pvTputdata struct {
	readThroughput  float64
	writeThroughput float64
}

func getTputValues(urlpassed string, query string) {
	res, err := http.Get(urlpassed)
	if err != nil {
		panic(err.Error())
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err.Error())
	}

	s2, err := getValue([]byte(body))
	if err != nil {
		panic(err.Error())
	}
	rand.Seed(time.Now().Unix())

	for _, result := range s2.Data.Result {
		Throughput[result.Metric.OpenebsPv], _ = strconv.ParseFloat(result.Value[1].(string), 32)
		Throughput[result.Metric.OpenebsPv] = Throughput[result.Metric.OpenebsPv] / (1024 * 1024)
	}

	if query == "OpenEBS_read_block_count_per_second" {
		readThroughputch <- Throughput
	} else if query == "OpenEBS_write_block_count_per_second" {
		writeThroughputch <- Throughput
	}
}

func getTputPVs() map[string]pvTputdata {
	queries := []string{"OpenEBS_read_block_count_per_second", "OpenEBS_write_block_count_per_second"}
	for _, query := range queries {
		go getTputValues(url+query, query)
	}
	var readThroughput, writeThroughput map[string]float64
	for i := 0; i < len(queries); i++ {
		select {
		case readThroughput = <-readThroughputch:
		case writeThroughput = <-writeThroughputch:
		}
	}
	Throughput := make(map[string]pvTputdata)
	if len(readThroughput) == len(writeThroughput) {
		for k2, v2 := range readThroughput {
			metaTput, err := clientset.CoreV1().PersistentVolumes().Get(k2, metav1.GetOptions{})
			if err != nil {
				continue
			}
			Throughput[string(metaTput.UID)] = pvTputdata{
				readThroughput:  v2,
				writeThroughput: writeThroughput[k2],
			}
		}
	}
	return Throughput
}

func (p *Plugin) updateTputPVs() {
	m2 := getTputPVs()
	if len(m2) > 0 {
		p.Tputpvs = m2
	}
}

func (p *Plugin) getTopologyPv2(str string) string {
	return fmt.Sprintf("%s;<persistent_volume>", str)
}
