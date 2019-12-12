package echool

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"

	sd "github.com/echool/go-echool/discovery"
)

var (
	deadSignal = []os.Signal{
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGKILL,
		syscall.SIGHUP,
		syscall.SIGQUIT,
	}
)

type (
	// ServiceFunc ...
	ServiceFunc func(server *grpc.Server)
	// GrpcServer ...
	GrpcServer struct {
		register    sd.Register
		serviceFunc ServiceFunc
	}
)

// NewServer ...
func NewServer(discovery sd.Register, serviceFunc ServiceFunc) *GrpcServer {
	return &GrpcServer{
		register:    discovery,
		serviceFunc: serviceFunc,
	}
}

// Run ...
func (s *GrpcServer) Run(serverOptions ...grpc.ServerOption) error {
	_, port, err := net.SplitHostPort(s.register.GetServiceAddress())
	if err != nil {
		return err
	}
	listen, err := net.Listen("tcp", ":"+port)
	if nil != err {
		return err
	}
	if err := s.register.Register(); err != nil {
		return err
	}
	server := grpc.NewServer(serverOptions...)
	s.serviceFunc(server)
	s.stopNotify()

	fmt.Printf(
		"[%s] of %s gRpc server has started\n",
		s.register.GetServiceAddress(),
		s.register.GetServiceName(),
	)
	if err := server.Serve(listen); nil != err {
		return err
	}
	return nil
}

func (s *GrpcServer) stopNotify() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, deadSignal...)
	go func() {
		fmt.Printf(
			"[%s] of %s gRpc server has existed with got signal [%v]\n",
			s.register.GetServiceAddress(),
			s.register.GetServiceName(),
			<-ch,
		)
		if err := s.register.Deregister(); err != nil {
			fmt.Printf(
				"[%s] of %s gRpc server Deregister fail and  err %v\n",
				s.register.GetServiceAddress(),
				s.register.GetServiceName(),
				err,
			)
		}
		os.Exit(0)
	}()
}
