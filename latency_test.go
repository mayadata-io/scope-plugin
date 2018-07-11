package main

import (
	"net/http/httptest"
	"reflect"
	"testing"

	utiltesting "k8s.io/client-go/util/testing"
)

var (
	resp1 = `{"status":"success","data":{"resultType":"vector","result":[{"metric":{"__name__":"OpenEBS__iops","instance":"172.17.0.2:9500","job":"cluster_uuid_9aba2480-a180-41ca-b5cb-f4a099376a16_openebs-volumes","kubernetes_pod_name":"pvc-4fa13b09-6242-11e8-a310-1458d00e6b83-ctrl-745784bb48-z9pl8","openebs_pv":"pvc-4fa13b09-6242-11e8-a310-1458d00e6b83"},"value":[1528354477.902,"0"]}]}}`
)

func TestGetLatValues(t *testing.T) {
	cases := map[string]*struct {
		fakeHandler    utiltesting.FakeHandler
		queryType      string
		channel        string
		ExpectedOutput map[string]int
	}{
		"When getting data for OpenEBS_read_latency:": {
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   200,
				ResponseBody: string(resp1),
				T:            t,
			},
			queryType: "OpenEBS_read_latency",
			channel:   "read",
			ExpectedOutput: map[string]int{
				"pvc-4fa13b09-6242-11e8-a310-1458d00e6b83": 0,
			},
		},
		"When getting data for OpenEBS_write_latency:": {
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   200,
				ResponseBody: string(resp1),
				T:            t,
			},
			queryType: "OpenEBS_write_latency",
			channel:   "write",
			ExpectedOutput: map[string]int{
				"pvc-4fa13b09-6242-11e8-a310-1458d00e6b83": 0,
			},
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(&tt.fakeHandler)
			go getLatValues(server.URL, tt.queryType)
			if tt.channel == "read" {
				readLatency := <-readLatencych
				if eq := reflect.DeepEqual(readLatency, tt.ExpectedOutput); eq {
					t.Errorf("Test Name :%v\nExpected :%v but got :%v", name, tt.ExpectedOutput, readLatency)
				}
			} else if tt.channel == "write" {
				writeLatency := <-writeLatencych
				if eq := reflect.DeepEqual(writeLatency, tt.ExpectedOutput); eq {
					t.Errorf("Test Name :%v\nExpected :%v but got :%v", name, tt.ExpectedOutput, writeLatency)
				}
			}
		})
	}
}

func TestPlugin_getTopologyPv1(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Persistent Volume",
			args: args{
				str: "123456",
			},
			want: "123456;<persistent_volume>",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Plugin{
				Latpvs: nil,
			}
			if got := p.getTopologyPv1(tt.args.str); got != tt.want {
				t.Errorf("Plugin.getTopologyPv1() = %v, want %v", got, tt.want)
			}
		})
	}
}
