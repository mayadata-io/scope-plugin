package main

import (
	"encoding/json"
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
				"readIops": {
					Samples: []sample{
						{
							Date:  time.Now(),
							Value: 12.6,
						},
					},
					Min: 0,
					Max: 20,
				},
				"writeIops": {
					Samples: []sample{
						{
							Date:  time.Now(),
							Value: 5.6,
						},
					},
					Min: 0,
					Max: 20,
				},
				"readLatency": {
					Samples: []sample{
						{
							Date:  time.Now(),
							Value: 5.4,
						},
					},
					Min: 0,
					Max: 20,
				},
				"writeLatency": {
					Samples: []sample{
						{
							Date:  time.Now(),
							Value: 7.8,
						},
					},
					Min: 0,
					Max: 20,
				},
				"readThroughput": {
					Samples: []sample{
						{
							Date:  time.Now(),
							Value: 19.2,
						},
					},
					Min: 0,
					Max: 20,
				},
				"writeThroughput": {
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
				"readIops": {
					ID:       "readIops",
					Label:    "Iops(R)",
					Format:   "",
					Priority: 0.1,
				},
				"writeIops": {
					ID:       "writeIops",
					Label:    "Iops(W)",
					Format:   "",
					Priority: 0.2,
				},
				"readLatency": {
					ID:       "readLatency",
					Label:    "Latency(R)",
					Format:   "millisecond",
					Priority: 0.3,
				},
				"writeLatency": {
					ID:       "writeLatency",
					Label:    "Latency(W)",
					Format:   "millisecond",
					Priority: 0.4,
				},
				"readThroughput": {
					ID:       "readThroughput",
					Label:    "Throughput(R)",
					Format:   "bytes",
					Priority: 0.5,
				},
				"writeThroughput": {
					ID:       "writeThroughput",
					Label:    "Throughput(W)",
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
			id:   "OpenEBS Plugin",
			name: "OpenEBS Plugin",
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
func TestResponse(t *testing.T) {
	dummyJSONResponse := []byte(`{
		"status": "ready",
		"data": {
			"resultType": "dummy",
			"result": [{
				"metric": {
					"__name__": "test",
					"instance": "test",
					"job": "testjob",
					"kubernetes_pod_name": "testpod",
					"openebs_pv": "testpv"
				},
				"value": []
			}]
		}
	}`)
	wantResponse := new(Metrics)
	_ = json.Unmarshal(dummyJSONResponse, &wantResponse)
	tests := []struct {
		name     string
		response []byte
		want     *Metrics
		wantErr  bool
	}{
		{
			name:     "Test response",
			response: dummyJSONResponse,
			want:     wantResponse,
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Response(tt.response)
			if (err != nil) != tt.wantErr {
				t.Errorf("Response() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Response() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestPlugin_getPVTopology(t *testing.T) {
	type fields struct {
		HostID     string
		Iops       map[string]PVMetrics
		Latency    map[string]PVMetrics
		Throughput map[string]PVMetrics
	}
	type args struct {
		PVName string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "Test getPVTopology",
			fields: fields{
				HostID:     "test",
				Iops:       nil,
				Latency:    nil,
				Throughput: nil,
			},
			args: args{
				PVName: "testpv",
			},
			want: "testpv;<persistent_volume>",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Plugin{
				HostID:     tt.fields.HostID,
				Iops:       tt.fields.Iops,
				Latency:    tt.fields.Latency,
				Throughput: tt.fields.Throughput,
			}
			if got := p.getPVTopology(tt.args.PVName); got != tt.want {
				t.Errorf("Plugin.getPVTopology() = %v, want %v", got, tt.want)
			}
		})
	}
}
