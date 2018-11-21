package metrics

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"k8s.io/client-go/kubernetes/fake"
)

var FieldsWithOnePV = &fields{
	Queries: nil,
	PVList: map[string]string{
		"testPV": "abcdef1234",
	},
	Data:      nil,
	ClientSet: fake.NewSimpleClientset(),
}

var FieldsWithOnePVAndSinglePVData = &fields{
	Queries: nil,
	PVList: map[string]string{
		"testPV": "abcdef1234",
	},
	Data: map[string]map[string]float64{
		"iopsWriteQuery": map[string]float64{
			"testPV": 5,
		},
	},
	ClientSet: fake.NewSimpleClientset(),
}

var FieldsWithOnePVAndMultiplePVData = &fields{
	Queries: nil,
	PVList: map[string]string{
		"testPV": "abcdef1234",
	},
	Data: map[string]map[string]float64{
		"iopsWriteQuery": map[string]float64{
			"testPV":  5,
			"testPV1": 6,
		},
	},
	ClientSet: fake.NewSimpleClientset(),
}

var FieldsWithNoPV = &fields{
	Queries: nil,
	PVList:  nil,
	Data: map[string]map[string]float64{
		"iopsWriteQuery": map[string]float64{
			"testPV":  5,
			"testPV1": 6,
		},
	},
	ClientSet: fake.NewSimpleClientset(),
}

var testMetricTemplate = map[string]metricTemplate{
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
}

func TestPVMetrics_metricTemplates(t *testing.T) {
	tests := []struct {
		name   string
		fields *fields
		want   map[string]metricTemplate
	}{
		{
			name:   "Test metricTemplates method",
			fields: FieldsWithNilValue,
			want:   testMetricTemplate,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PVMetrics{
				Queries:   tt.fields.Queries,
				PVList:    tt.fields.PVList,
				Data:      tt.fields.Data,
				ClientSet: tt.fields.ClientSet,
			}
			if got := p.metricTemplates(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PVMetrics.metricTemplates() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPVMetrics_metricIDAndName(t *testing.T) {
	tests := []struct {
		name   string
		fields *fields
		want   string
		want1  string
	}{
		{
			name:   "Test metricIDAndName method",
			fields: FieldsWithNilValue,
			want:   "OpenEBS Plugin",
			want1:  "OpenEBS Plugin",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PVMetrics{
				Queries:   tt.fields.Queries,
				PVList:    tt.fields.PVList,
				Data:      tt.fields.Data,
				ClientSet: tt.fields.ClientSet,
			}
			got, got1 := p.metricIDAndName()
			if got != tt.want {
				t.Errorf("PVMetrics.metricIDAndName() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("PVMetrics.metricIDAndName() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestPVMetrics_metrics(t *testing.T) {
	type args struct {
		data []float64
	}
	tests := []struct {
		name   string
		fields *fields
		args   args
		want   map[string]metric
	}{
		{
			name:   "When each metrics is 0",
			fields: FieldsWithNilValue,
			args: args{
				data: []float64{0, 0, 0, 0, 0, 0},
			},
			want: map[string]metric{
				"readIops": {
					Samples: []sample{
						{
							Date:  time.Now(),
							Value: 0,
						},
					},
					Min: 0,
					Max: 100,
				},
				"writeIops": {
					Samples: []sample{
						{
							Date:  time.Now(),
							Value: 0,
						},
					},
					Min: 0,
					Max: 100,
				},
				"readLatency": {
					Samples: []sample{
						{
							Date:  time.Now(),
							Value: 0,
						},
					},
					Min: 0,
					Max: 100,
				},
				"writeLatency": {
					Samples: []sample{
						{
							Date:  time.Now(),
							Value: 0,
						},
					},
					Min: 0,
					Max: 100,
				},
				"readThroughput": {
					Samples: []sample{
						{
							Date:  time.Now(),
							Value: 0,
						},
					},
					Min: 0,
					Max: 100,
				},
				"writeThroughput": {
					Samples: []sample{
						{
							Date:  time.Now(),
							Value: 0,
						},
					},
					Min: 0,
					Max: 100,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PVMetrics{
				Queries:   tt.fields.Queries,
				PVList:    tt.fields.PVList,
				Data:      tt.fields.Data,
				ClientSet: tt.fields.ClientSet,
			}
			got := p.metrics(tt.args.data)
			for k, v := range got {
				if !reflect.DeepEqual(v.Samples[0].Value, tt.want[k].Samples[0].Value) {
					t.Errorf("PVMetrics.metrics() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestPVMetrics_getPVTopology(t *testing.T) {
	type args struct {
		PersistentVolumeUID string
	}
	tests := []struct {
		name   string
		fields *fields
		args   args
		want   string
	}{
		{
			name:   "When argument is empty",
			fields: FieldsWithNilValue,
			args: args{
				PersistentVolumeUID: "",
			},
			want: ";<persistent_volume>",
		},
		{
			name:   "When argument is some UUID",
			fields: FieldsWithNilValue,
			args: args{
				PersistentVolumeUID: "abcdef-1234-pqrst-123",
			},
			want: "abcdef-1234-pqrst-123;<persistent_volume>",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PVMetrics{
				Queries:   tt.fields.Queries,
				PVList:    tt.fields.PVList,
				Data:      tt.fields.Data,
				ClientSet: tt.fields.ClientSet,
			}
			if got := p.getPVTopology(tt.args.PersistentVolumeUID); got != tt.want {
				t.Errorf("PVMetrics.getPVTopology() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPVMetrics_makeReport(t *testing.T) {
	tests := []struct {
		name   string
		fields *fields
		want   *report
	}{
		{
			name:   "when data and PV list are empty",
			fields: FieldsWithNilValue,
			want: &report{
				PersistentVolume: topology{
					Nodes:           nil,
					MetricTemplates: testMetricTemplate,
				},
				Plugins: []pluginSpec{
					{
						ID:          "openebs",
						Label:       "OpenEBS Monitor Plugin",
						Description: "OpenEBS Monitor Plugin: Monitor OpeneEBS volumes",
						Interfaces:  []string{"reporter"},
						APIVersion:  "1",
					},
				},
			},
		},
		{
			name:   "when data is nil and PV list has one PV",
			fields: FieldsWithOnePV,
			want: &report{
				PersistentVolume: topology{
					Nodes:           nil,
					MetricTemplates: testMetricTemplate,
				},
				Plugins: []pluginSpec{
					{
						ID:          "openebs",
						Label:       "OpenEBS Monitor Plugin",
						Description: "OpenEBS Monitor Plugin: Monitor OpeneEBS volumes",
						Interfaces:  []string{"reporter"},
						APIVersion:  "1",
					},
				},
			},
		},
		{
			name:   "when data has one PV value and PV list has one PV",
			fields: FieldsWithOnePVAndSinglePVData,
			want: &report{
				PersistentVolume: topology{
					Nodes: map[string]node{
						"abcdef1234;<persistent_volume>": node{
							Metrics: map[string]metric{
								"readIops": {
									Samples: []sample{
										{
											Date:  time.Now(),
											Value: 0,
										},
									},
									Min: 0,
									Max: 100,
								},
								"writeIops": {
									Samples: []sample{
										{
											Date:  time.Now(),
											Value: 5,
										},
									},
									Min: 0,
									Max: 100,
								},
								"readLatency": {
									Samples: []sample{
										{
											Date:  time.Now(),
											Value: 0,
										},
									},
									Min: 0,
									Max: 100,
								},
								"writeLatency": {
									Samples: []sample{
										{
											Date:  time.Now(),
											Value: 0,
										},
									},
									Min: 0,
									Max: 100,
								},
								"readThroughput": {
									Samples: []sample{
										{
											Date:  time.Now(),
											Value: 0,
										},
									},
									Min: 0,
									Max: 100,
								},
								"writeThroughput": {
									Samples: []sample{
										{
											Date:  time.Now(),
											Value: 0,
										},
									},
									Min: 0,
									Max: 100,
								},
							},
						},
					},
					MetricTemplates: testMetricTemplate,
				},
				Plugins: []pluginSpec{
					{
						ID:          "openebs",
						Label:       "OpenEBS Monitor Plugin",
						Description: "OpenEBS Monitor Plugin: Monitor OpeneEBS volumes",
						Interfaces:  []string{"reporter"},
						APIVersion:  "1",
					},
				},
			},
		},
		{
			name:   "when data has multiple PV value and PV list has one PV",
			fields: FieldsWithOnePVAndMultiplePVData,
			want: &report{
				PersistentVolume: topology{
					Nodes: map[string]node{
						"abcdef1234;<persistent_volume>": node{
							Metrics: map[string]metric{
								"readIops": {
									Samples: []sample{
										{
											Date:  time.Now(),
											Value: 0,
										},
									},
									Min: 0,
									Max: 100,
								},
								"writeIops": {
									Samples: []sample{
										{
											Date:  time.Now(),
											Value: 5,
										},
									},
									Min: 0,
									Max: 100,
								},
								"readLatency": {
									Samples: []sample{
										{
											Date:  time.Now(),
											Value: 0,
										},
									},
									Min: 0,
									Max: 100,
								},
								"writeLatency": {
									Samples: []sample{
										{
											Date:  time.Now(),
											Value: 0,
										},
									},
									Min: 0,
									Max: 100,
								},
								"readThroughput": {
									Samples: []sample{
										{
											Date:  time.Now(),
											Value: 0,
										},
									},
									Min: 0,
									Max: 100,
								},
								"writeThroughput": {
									Samples: []sample{
										{
											Date:  time.Now(),
											Value: 0,
										},
									},
									Min: 0,
									Max: 100,
								},
							},
						},
					},
					MetricTemplates: testMetricTemplate,
				},
				Plugins: []pluginSpec{
					{
						ID:          "openebs",
						Label:       "OpenEBS Monitor Plugin",
						Description: "OpenEBS Monitor Plugin: Monitor OpeneEBS volumes",
						Interfaces:  []string{"reporter"},
						APIVersion:  "1",
					},
				},
			},
		},
		{
			name:   "when data has multiple PV value and PV list has one PV",
			fields: FieldsWithNoPV,
			want: &report{
				PersistentVolume: topology{
					Nodes:           nil,
					MetricTemplates: testMetricTemplate,
				},
				Plugins: []pluginSpec{
					{
						ID:          "openebs",
						Label:       "OpenEBS Monitor Plugin",
						Description: "OpenEBS Monitor Plugin: Monitor OpeneEBS volumes",
						Interfaces:  []string{"reporter"},
						APIVersion:  "1",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PVMetrics{
				Queries:   tt.fields.Queries,
				PVList:    tt.fields.PVList,
				Data:      tt.fields.Data,
				ClientSet: tt.fields.ClientSet,
			}
			got := p.makeReport()
			if tt.want.PersistentVolume.Nodes == nil {
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("PVMetrics.makeReport() = %v, want %v", got, tt.want)
				}
			} else {
				for k, v := range got.PersistentVolume.Nodes {
					for x, y := range v.Metrics {
						for i, val := range y.Samples {
							if !reflect.DeepEqual(val.Value, tt.want.PersistentVolume.Nodes[k].Metrics[x].Samples[i].Value) {
								t.Errorf("PVMetrics.makeReport() %s, %s = %v, want %v", k, x, val.Value, tt.want.PersistentVolume.Nodes[k].Metrics[x].Samples[i].Value)
							}
						}
					}
				}
			}
		})
	}
}

func TestPVMetrics_Report(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name   string
		fields *fields
		args   args
	}{
		{
			name:   "Test report method",
			fields: FieldsWithNilValue,
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest("GET", "/report", nil),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PVMetrics{
				Queries:   tt.fields.Queries,
				PVList:    tt.fields.PVList,
				Data:      tt.fields.Data,
				ClientSet: tt.fields.ClientSet,
			}
			p.Report(tt.args.w, tt.args.r)
		})
	}
}
