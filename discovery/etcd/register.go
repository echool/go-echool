package etcd

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/echool/go-echool/config"
	"github.com/echool/go-echool/util"
	uuid "github.com/satori/go.uuid"
)

// Register ...
type Register struct {
	client3   *clientv3.Client
	stop      chan bool
	interval  time.Duration
	leaseTime int64
	name      string
	address   string
	fullAddr  string
}

// NewRegister naming.Update{Op: naming.Add, Addr: "1.2.3.4", Metadata: "..."})
func NewRegister(conf config.EtcdSrvConf) (*Register, error) {
	client3, err := clientv3.New(
		clientv3.Config{
			Endpoints: conf.Endpoints,
			Username:  conf.Auth.UserName,
			Password:  conf.Auth.PassWord,
		})
	if nil != err {
		return nil, err
	}

	host, pt, err := net.SplitHostPort(conf.ServiceAddress)
	if err != nil {
		return nil, err
	}
	addr, err := util.ExtractIP(host)
	if err != nil {
		return nil, err
	}

	return &Register{
		name:      conf.ServiceName,
		address:   addr + ":" + pt,
		client3:   client3,
		interval:  3 * time.Second,
		leaseTime: 6,
		stop:      make(chan bool, 1),
		fullAddr:  fmt.Sprintf("etcd/%s/%s", conf.ServiceName, uuid.NewV4().String()),
	}, nil
}

// Register register service with name as prefix to etcd, multi etcd addr should use ; to split
func (r *Register) Register() error {
	ticker := time.NewTicker(r.interval)
	go func() {
		for {
			if getResp, err := r.client3.Get(context.Background(), r.fullAddr); err != nil {
				fmt.Println("error:", err)
			} else if getResp.Count == 0 {
				if err = r.withAlive(); err != nil {
					fmt.Println("error:", err)
				}
			}
			select {
			case <-ticker.C:
			case <-r.stop:
				return
			}
		}
	}()
	return nil
}

// Deregister remove service from etcd
func (r *Register) Deregister() error {
	if r.client3 != nil {
		r.stop <- true
		_, err := r.client3.Delete(context.Background(), r.fullAddr)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Register) GetServiceName() string {
	return r.name
}

func (r *Register) GetServiceAddress() string {
	return r.address
}

func (r *Register) GetFullAddress() string {
	return r.fullAddr
}

func (r *Register) SetLeaseTime(leaseTime int64) {
	r.leaseTime = leaseTime
}

func (r *Register) SetInterval(interval time.Duration) {
	r.interval = interval
}

func (r *Register) withAlive() error {
	leaseResp, err := r.client3.Grant(context.Background(), r.leaseTime)
	if err != nil {
		return err
	}
	_, err = r.client3.Put(context.Background(), r.fullAddr, r.address, clientv3.WithLease(leaseResp.ID))
	if err != nil {
		return err
	}
	_, err = r.client3.KeepAlive(context.Background(), leaseResp.ID)
	if err != nil {
		return err
	}
	return nil
}
