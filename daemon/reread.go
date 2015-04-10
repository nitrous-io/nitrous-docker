package daemon

import (
	"github.com/docker/docker/engine"
)

func (daemon *Daemon) ContainerReread(job *engine.Job) engine.Status {
	if len(job.Args) != 1 {
		return job.Errorf("Usage: %s CONTAINER\n", job.Name)
	}

	var name = job.Args[0]
	if container := daemon.Get(name); container != nil {
		if err := container.FromDisk(); err != nil {
			return job.Errorf("Cannot read container %s from disk: %s\n", name, err)
		}
		container.LogEvent("reread")
	} else {
		return job.Errorf("No such container: %s\n", name)
	}
	return engine.StatusOK
}
