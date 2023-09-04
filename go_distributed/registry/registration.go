package registry

// this file is to claim and management services except registry
// so called service discovery

type Registration struct {
	ServiceName ServiceName
	ServiceURL  string
}

type ServiceName string

const (
	LogService = ServiceName("LogService")
)