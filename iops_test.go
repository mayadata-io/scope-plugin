package main

import (
	"net/http/httptest"
	"reflect"
	"testing"

	utiltesting "k8s.io/client-go/util/testing"
)

var (
	resp = `{"status":"success","data":{"resultType":"vector","result":[{"metric":{"__name__":"OpenEBS__iops","instance":"172.17.0.2:9500","job":"cluster_uuid_9aba2480-a180-41ca-b5cb-f4a099376a16_openebs-volumes","kubernetes_pod_name":"pvc-4fa13b09-6242-11e8-a310-1458d00e6b83-ctrl-745784bb48-z9pl8","openebs_pv":"pvc-4fa13b09-6242-11e8-a310-1458d00e6b83"},"value":[1528354477.902,"0"]}]}}`
)

func TestGetValues(t *testing.T) {
	cases := map[string]*struct {
		fakeHandler    utiltesting.FakeHandler
		queryType      string
		channel        string
		ExpectedOutput map[string]int
	}{
		"When getting data for OpenEBS_read_iops:": {
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   200,
				ResponseBody: string(resp),
				T:            t,
			},
			queryType: "OpenEBS_read_iops",
			channel:   "read",
			ExpectedOutput: map[string]int{
				"pvc-4fa13b09-6242-11e8-a310-1458d00e6b83": 0,
			},
		},
		"When getting data for OpenEBS_write_iops:": {
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   200,
				ResponseBody: string(resp),
				T:            t,
			},
			queryType: "OpenEBS_write_iops",
			channel:   "write",
			ExpectedOutput: map[string]int{
				"pvc-4fa13b09-6242-11e8-a310-1458d00e6b83": 0,
			},
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(&tt.fakeHandler)
			go getValues(server.URL, tt.queryType)
			if tt.channel == "read" {
				read := <-readch
				if eq := reflect.DeepEqual(read, tt.ExpectedOutput); eq {
					t.Errorf("Test Name :%v\nExpected :%v but got :%v", name, tt.ExpectedOutput, read)
				}
			} else if tt.channel == "write" {
				write := <-writech
				if eq := reflect.DeepEqual(write, tt.ExpectedOutput); eq {
					t.Errorf("Test Name :%v\nExpected :%v but got :%v", name, tt.ExpectedOutput, write)
				}
			}
		})
	}
}

func TestPlugin_getTopologyPv(t *testing.T) {
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
				pvs: nil,
			}
			if got := p.getTopologyPv(tt.args.str); got != tt.want {
				t.Errorf("Plugin.getTopologyPv() = %v, want %v", got, tt.want)
			}
		})
	}
}
