package config

import "time"

type (
	EtcdSrvConf struct {
		ServiceName    string
		ServiceAddress string
		Endpoints      []string
		Auth
	}

	EtcdCliConf struct {
		ServiceName string
		Endpoints   []string
		Auth
	}

	Auth struct {
		UserName string
		PassWord string
	}

	ConsulSrvConf struct {
		ServiceName    string
		ServiceAddress string
		Address        string
		Token          string
	}

	ConsulCliConf struct {
		ServiceName string
		Address     string
		Token       string
	}
)

const (
	InvokeTimeout = 3 * time.Second
)
