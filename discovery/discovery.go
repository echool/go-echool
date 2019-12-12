package discovery

// Register provides an interface for service discovery
// and an abstraction over varying implementations
// {consul, etcd, kubernetes}
type Register interface {
	Register() error
	Deregister() error
	GetServiceName() string
	GetServiceAddress() string
}
