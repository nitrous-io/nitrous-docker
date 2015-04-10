package daemon

import (
	"fmt"

	"github.com/Sirupsen/logrus"
)

func (daemon *Daemon) ContainerReload(name string) error {
	container, err := daemon.Get(name)
	if err != nil {
		return err
	}

	if container.IsRunning() {
		return fmt.Errorf("Container %s is running. Cannot reconfigure from disk.\n", name)
	}

	id := container.ID
	container, err = daemon.load(id)
	if err != nil {
		return fmt.Errorf("Failed to load container %v: %v\n", id, err)
	}

	// Ignore the container if it does not support the current driver being used by the graph
	currentDriver := daemon.driver.String()
	if (container.Driver == "" && currentDriver == "aufs") || container.Driver == currentDriver {
		logrus.Debugf("Loaded container %v", container.ID)

		if entities := daemon.containerGraph.List("/", -1); entities != nil {
			for _, p := range entities.Paths() {
				e := entities[p]

				if container.ID == e.ID() {
					// Delete the container data from the daemon and set new container data to it.
					daemon.containers.Delete(container.ID)
					if err := daemon.register(container, false); err != nil {
						logrus.Debugf("Failed to register container %s: %s", container.ID, err)
					}
				}
			}
		}
	} else {
		logrus.Debugf("Cannot load container %s because it was created with another graph driver.", container.ID)
	}

	container.LogEvent("reload")
	return nil
}
