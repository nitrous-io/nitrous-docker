package daemon

import (
	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/engine"
)

func (daemon *Daemon) ContainerReload(job *engine.Job) engine.Status {
	if len(job.Args) != 1 {
		return job.Errorf("Usage: %s CONTAINER\n", job.Name)
	}

	var name = job.Args[0]
	container, err := daemon.Get(name)
	if err != nil {
		return job.Errorf("Failed to get container %s: %s\n", name, err)
	}
	if container == nil {
		return job.Errorf("No such container: %s\n", name)
	}

	if container.IsRunning() {
		return job.Errorf("Container %s is running. Cannot reconfigure from disk.\n", name)
	}

	id := container.ID
	container, err = daemon.load(id)
	if err != nil {
		return job.Errorf("Failed to load container %v: %v\n", id, err)
	}

	// Ignore the container if it does not support the current driver being used by the graph
	currentDriver := daemon.driver.String()
	if (container.Driver == "" && currentDriver == "aufs") || container.Driver == currentDriver {
		log.Debugf("Loaded container %v", container.ID)

		if entities := daemon.containerGraph.List("/", -1); entities != nil {
			for _, p := range entities.Paths() {
				e := entities[p]

				if container.ID == e.ID() {
					// Delete the container data from the daemon and set new container data to it.
					daemon.containers.Delete(container.ID)
					if err := daemon.register(container, false); err != nil {
						log.Debugf("Failed to register container %s: %s", container.ID, err)
					}
				}
			}
		}
	} else {
		log.Debugf("Cannot load container %s because it was created with another graph driver.", container.ID)
	}

	container.LogEvent("reload")
	return engine.StatusOK
}
