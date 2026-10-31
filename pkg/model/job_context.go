package model

type JobContext struct {
	Status    string `json:"status"`
	Container struct {
		ID      string `json:"id"`
		Network string `json:"network"`
	} `json:"container"`
	Services map[string]JobServiceContext `json:"services"`
}

// JobServiceContext is job.services.<id>: the service container, its
// network and the host ports its published ports are mapped to.
type JobServiceContext struct {
	ID      string            `json:"id"`
	Network string            `json:"network"`
	Ports   map[string]string `json:"ports"`
}
