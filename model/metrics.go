package model

type AppMetric struct {
	App    string  `json:"app"`
	Cpu    float32 `json:"cpu"`
	Memory int64   `json:"memory"`

	Request          int64
	Response         int64
	Response2xx      int64
	Response4xx      int64
	Response5xx      int64
	Response5xxRoute int64
}
