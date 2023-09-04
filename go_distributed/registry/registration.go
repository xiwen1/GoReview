package registry

// this file is to claim and management services except registry
// so called service discovery

type Registration struct {
	ServiceName ServiceName
	ServiceURL  string
	RequiredServices []ServiceName
	ServiceUpdateURL string
}

type ServiceName string

type patchEntry struct {
	Name ServiceName
	URL string
}

type patch struct {
	Added []patchEntry
	Removed []patchEntry
}

const (
	LogService = ServiceName("LogService")
)