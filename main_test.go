package main

import (
	"reflect"
	"testing"
	"time"
)

func TestMetrics(t *testing.T) {
	testCases := map[string]*struct {
		inputData      []float64
		expectedOutput map[string]metric
	}{
		"data containing random values between 0-20": {
			inputData: []float64{12.6, 5.6, 5.4, 7.8, 19.2, 11.0},
			expectedOutput: map[string]metric{
				"r": {
					Samples: []sample{
						{
							Date:  time.Now(),
							Value: 12.6,
						},
					},
					Min: 0,
					Max: 20,
				},
				"w": {
					Samples: []sample{
						{
							Date:  time.Now(),
							Value: 5.6,
						},
					},
					Min: 0,
					Max: 20,
				},
				"r1": {
					Samples: []sample{
						{
							Date:  time.Now(),
							Value: 5.4,
						},
					},
					Min: 0,
					Max: 20,
				},
				"w1": {
					Samples: []sample{
						{
							Date:  time.Now(),
							Value: 7.8,
						},
					},
					Min: 0,
					Max: 20,
				},
				"r2": {
					Samples: []sample{
						{
							Date:  time.Now(),
							Value: 19.2,
						},
					},
					Min: 0,
					Max: 20,
				},
				"w2": {
					Samples: []sample{
						{
							Date:  time.Now(),
							Value: 11.0,
						},
					},
					Min: 0,
					Max: 20,
				},
			},
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			var p Plugin
			got := p.metrics(tt.inputData)
			for k, v := range got {
				if v.Samples[0].Value != tt.expectedOutput[k].Samples[0].Value {
					t.Errorf("Test Name :%v\nExpected :%v but got :%v", name, tt.expectedOutput[k], v)
				}
			}
		})
	}
}

func TestMetricTemplate(t *testing.T) {
	testCases := map[string]*struct {
		inputData      []float64
		expectedOutput map[string]metricTemplate
	}{
		"data containing random values between 0-20": {
			inputData: []float64{12.6, 5.6, 5.4, 7.8, 19.2, 11.0},
			expectedOutput: map[string]metricTemplate{
				"r": {
					ID:       "r",
					Label:    "Iops(R)",
					Format:   "",
					Priority: 0.1,
				},
				"w": {
					ID:       "w",
					Label:    "Iops(W)",
					Format:   "",
					Priority: 0.2,
				},
				"r1": {
					ID:       "r1",
					Label:    "Latency(R)",
					Format:   "millisecond",
					Priority: 0.3,
				},
				"w1": {
					ID:       "w1",
					Label:    "Latency(W)",
					Format:   "millisecond",
					Priority: 0.4,
				},
				"r2": {
					ID:       "r2",
					Label:    "Tput(R)",
					Format:   "bytes",
					Priority: 0.5,
				},
				"w2": {
					ID:       "w2",
					Label:    "Tput(W)",
					Format:   "bytes",
					Priority: 0.6,
				},
			},
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			var p Plugin
			got := p.metricTemplates()
			if eq := reflect.DeepEqual(got, tt.expectedOutput); !eq {
				t.Errorf("Test Name :%v\nExpected :%v but got :%v", name, tt.expectedOutput, got)
			}
		})
	}
}

func TestMetricIDAndName(t *testing.T) {
	testCases := map[string]*struct {
		id, name string
	}{
		"id and name ": {
			id:   "iops",
			name: "Iops",
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			var p Plugin
			gid, gname := p.metricIDAndName()
			if gid != tt.id || gname != tt.name {
				t.Errorf("Test Name :%v\nExpected gid :%v , Expected gname :%v but got gid :%v, got gname :%v", name, tt.id, tt.name, gid, gname)
			}
		})
	}
}
