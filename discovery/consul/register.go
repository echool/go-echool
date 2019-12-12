package consul

import (
	"net"
	"strconv"

	"github.com/echool/go-echool/util"

	"github.com/echool/go-echool/config"
	"github.com/hashicorp/consul/api"
	uuid "github.com/satori/go.uuid"
)

// Register ...
type Register struct {
	cli           *api.Client
	consulAddress string
	consulToken   string
	name          string
	address       string
	svcID         string
}

// NewRegister ...
func NewRegister(conf config.ConsulSrvConf) (*Register, error) {
	host, pt, err := net.SplitHostPort(conf.ServiceAddress)
	if err != nil {
		return nil, err
	}
	addr, err := util.ExtractIP(host)
	if err != nil {
		return nil, err
	}

	return &Register{
		name:          conf.ServiceName,
		address:       addr + ":" + pt,
		consulAddress: conf.Address,
		consulToken:   conf.Token,
	}, nil
}

// Register 注册
func (r *Register) Register() error {
	config := api.DefaultConfig()
	config.Address = r.consulAddress
	config.Token = r.consulToken
	consulCli, err := api.NewClient(config)
	if err != nil {
		return err
	}
	r.cli = consulCli

	check := api.AgentServiceCheck{
		TCP:                            r.address,
		Interval:                       "5s",
		Timeout:                        "3s",
		DeregisterCriticalServiceAfter: "60s",
	}
	host, pt, err := net.SplitHostPort(r.address)
	if err != nil {
		return err
	}
	port, _ := strconv.Atoi(pt)

	r.svcID = r.name + "-" + uuid.NewV4().String()
	asr := &api.AgentServiceRegistration{
		ID:      r.svcID,
		Name:    r.name,
		Address: host,
		Port:    port,
		//Tags:    []string{"v1.01"},
		Check: &check,
	}

	if err := r.cli.Agent().ServiceRegister(asr); err != nil {
		return err
	}

	return nil
}

// Deregister 注销
func (r *Register) Deregister() error {
	if err := r.cli.Agent().ServiceDeregister(r.svcID); err != nil {
		return err
	}

	return nil
}

// GetServiceName ...
func (r *Register) GetServiceName() string {
	return r.name
}

// GetServiceAddress ...
func (r *Register) GetServiceAddress() string {
	return r.address
}
