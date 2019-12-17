package etcd

import (
	"context"
	"fmt"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/echool/go-echool/config"
	"google.golang.org/grpc/resolver"
)

type Builder struct {
	name      string
	endpoints []string
	config.Auth
}

type Resolver struct {
	conn    resolver.ClientConn
	client3 *clientv3.Client
	name    string
}

// NewResolver initialize an etcd client
func NewResolver(conf config.EtcdCliConf) string {
	builder := &Builder{
		name:      conf.ServiceName,
		endpoints: conf.Endpoints,
		Auth:      conf.Auth,
	}
	resolver.Register(builder)

	return "etcd:///" + conf.ServiceName
}

func (r *Builder) Build(target resolver.Target, clientConn resolver.ClientConn, opts resolver.BuildOption) (resolver.Resolver, error) {
	cli3, err := clientv3.New(
		clientv3.Config{
			Endpoints:   r.endpoints,
			Username:    r.Auth.UserName,
			Password:    r.Auth.PassWord,
			DialTimeout: 5 * time.Second,
		})
	if nil != err {
		return nil, err
	}

	rl := &Resolver{conn: clientConn, client3: cli3, name: r.name}
	go rl.watch(fmt.Sprintf("%s/%s/", target.Scheme, target.Endpoint))
	return rl, nil
}

func (r *Builder) Scheme() string {
	return "etcd"
}

func (r *Resolver) watch(keyPrefix string) {
	var addrList []resolver.Address
	resp, err := r.client3.Get(context.Background(), keyPrefix, clientv3.WithPrefix())
	if err != nil {
		panic(err)
	}
	serverName := r.name
	for i := range resp.Kvs {
		addrList = append(
			addrList,
			resolver.Address{
				Addr:       string(resp.Kvs[i].Value),
				ServerName: serverName,
			})
	}
	r.conn.UpdateState(resolver.State{Addresses: addrList})
	watchChan := r.client3.Watch(context.Background(), keyPrefix, clientv3.WithPrefix())
	for n := range watchChan {
		for _, ev := range n.Events {
			addr := string(ev.Kv.Value)
			switch ev.Type {
			case mvccpb.PUT:
				if !r.exist(addrList, addr) {
					addrList = append(
						addrList,
						resolver.Address{
							Addr:       addr,
							ServerName: serverName,
						})
					r.conn.UpdateState(resolver.State{Addresses: addrList})
				}
			case mvccpb.DELETE:
				if s, ok := r.remove(addrList, addr); ok {
					addrList = s
					r.conn.UpdateState(resolver.State{Addresses: addrList})
				}
			}
		}
	}
}

// ResolveNow will be called by gRPC to try to resolve the target name
// again. It's just a hint, resolver can ignore this if it's not necessary.
//
// It could be called multiple times concurrently.
func (r *Resolver) ResolveNow(opt resolver.ResolveNowOption) {
}

// It's just a hint, resolver can ignore this if it's not necessary.
func (r *Resolver) Close() {
}

func (r *Resolver) exist(l []resolver.Address, addr string) bool {
	for i := range l {
		if l[i].Addr == addr {
			return true
		}
	}
	return false
}

func (r *Resolver) remove(s []resolver.Address, addr string) ([]resolver.Address, bool) {
	for i := range s {
		if s[i].Addr == addr {
			s[i] = s[len(s)-1]
			return s[:len(s)-1], true
		}
	}
	return nil, false
}
