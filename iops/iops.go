package main

import (
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var pv = make(map[string]float64)
var storeAfter *Iops

type pvdata struct {
	read  float64
	write float64
}

func getValues(urlpassed string, query string) {
	res, err := http.Get(urlpassed)
	if err != nil {
		panic(err.Error())
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err.Error())
	}

	storeBefore, err := getValue([]byte(body))
	if err != nil {
		panic(err.Error())
	}

	rand.Seed(time.Now().Unix())
	time.Sleep(1 * time.Second)

	res, err = http.Get(urlpassed)
	if err != nil {
		panic(err.Error())
	}

	body, err = ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err.Error())
	}

	storeAfter, err = getValue([]byte(body))
	if err != nil {
		panic(err.Error())
	}

	rand.Seed(time.Now().Unix())

	for _, result := range storeBefore.Data.Result {
		for _, newresult := range storeAfter.Data.Result {
			if result.Metric.OpenebsPv == newresult.Metric.OpenebsPv {
				pv[result.Metric.OpenebsPv] = math.Abs(result.Value[0].(float64) - newresult.Value[0].(float64))
			}
		}
	}

	if query == "OpenEBS_read_iops" {
		readch <- pv
	} else if query == "OpenEBS_write_iops" {
		writech <- pv
	}
}

func getPVs() map[string]pvdata {
	// Get request to url
	queries := []string{"OpenEBS_read_iops", "OpenEBS_write_iops"}
	for _, query := range queries {
		go getValues(url+query, query)
	}
	var read, write map[string]float64
	for i := 0; i < len(queries); i++ {
		select {
		case read = <-readch:
		case write = <-writech:
		}
	}
	pv := make(map[string]pvdata)
	if len(read) == len(write) {
		for k, v := range read {
			meta, err := clientset.CoreV1().PersistentVolumes().Get(k, metav1.GetOptions{})
			if err != nil {
				continue
			}
			pv[string(meta.UID)] = pvdata{
				read:  v,
				write: write[k],
			}
		}
	}
	return pv
}

func (p *Plugin) updatePVs() {
	m := getPVs()
	if len(m) > 0 {
		p.pvs = m
	}
}

func (p *Plugin) getTopologyPv(str string) string {
	return fmt.Sprintf("%s;<persistent_volume>", str)
}

