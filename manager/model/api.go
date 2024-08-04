package model

type Service struct {
	Host   string
	Name   string
	Params map[string]string
}

type ProxyApi struct {
	Service        Service
	Id             string
	UpdateInterval float64
	Name           string
	ProviderId     string
}
