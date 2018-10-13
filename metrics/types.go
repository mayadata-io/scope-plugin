package metrics

// PVMetrics groups the data a plugin needs.
type PVMetrics struct {
	ReadIops        float64
	WriteIops       float64
	ReadLatency     float64
	WriteLatency    float64
	ReadThroughput  float64
	WriteThroughput float64
}

// Metrics stores the json provided by cortex agent.
type Metrics struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric struct {
				Name              string `json:"__name__"`
				Instance          string `json:"instance"`
				Job               string `json:"job"`
				KubernetesPodName string `json:"kubernetes_pod_name"`
				OpenebsPv         string `json:"openebs_pv"`
			} `json:"metric"`
			Value []interface{} `json:"value"`
		} `json:"result"`
	} `json:"data"`
}
