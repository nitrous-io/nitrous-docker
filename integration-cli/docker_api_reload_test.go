package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strings"

	"github.com/docker/docker/daemon"
	"github.com/docker/docker/runconfig"
	"github.com/go-check/check"
)

func (s *DockerSuite) TestContainerApiReload(c *check.C) {
	runCmd := exec.Command(dockerBinary, "run", "-d", "busybox", "top")
	out, _, err := runCommandWithOutput(runCmd)
	if err != nil {
		c.Fatalf("Error on container creation: %v, output: %q", err, out)
	}
	containerID := strings.TrimSpace(out)

	// stop the container
	out, _, err = runCommandWithOutput(exec.Command(dockerBinary, "stop", containerID))
	if err != nil {
		c.Fatalf("Error on container stop: %v, output: %q", err, out)
	}

	status, body, err := sockRequest("GET", "/info", nil)
	c.Assert(status, check.Equals, http.StatusOK)
	c.Assert(err, check.IsNil)
	var infoResponse struct {
		DockerRootDir string
	}
	if err = json.Unmarshal(body, &infoResponse); err != nil {
		c.Fatalf("unable to unmarshal response body: %v", err)
	}

	// manually modifies config.json
	configFilePath := fmt.Sprintf("%s/containers/%s/config.json", infoResponse.DockerRootDir, containerID)
	configJson, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		c.Fatalf("Can't read config file: %s. %v", configFilePath, err)
	}
	var container daemon.Container
	err = json.Unmarshal(configJson, &container)
	c.Assert(err, check.IsNil)
	container.Volumes = map[string]string{
		"/testVolume": "/testVolume",
	}
	container.VolumesRW = map[string]bool{
		"/testVolume": true,
	}
	data, err := json.Marshal(container)
	c.Assert(err, check.IsNil)
	ioutil.WriteFile(configFilePath, data, 0666)

	// manually modifies hostconfig.json
	hostConfigFilePath := fmt.Sprintf("%s/containers/%s/hostconfig.json", infoResponse.DockerRootDir, containerID)
	hostConfigJson, err := ioutil.ReadFile(hostConfigFilePath)
	if err != nil {
		c.Fatalf("Can't read config file: %s. %v", hostConfigFilePath, err)
	}
	var hostconfig runconfig.HostConfig
	err = json.Unmarshal(hostConfigJson, &hostconfig)
	c.Assert(err, check.IsNil)
	hostconfig.Binds = []string{"/testVolume:/testVolume"}
	data, err = json.Marshal(hostconfig)
	c.Assert(err, check.IsNil)
	ioutil.WriteFile(hostConfigFilePath, data, 0666)

	out, _, err = runCommandWithOutput(exec.Command(dockerBinary, "inspect", "-f", "{{.Volumes}}", containerID))
	if err != nil {
		c.Fatalf("failed to inspect %s. Output(%q), err(%v)", containerID, out, err)
	}
	if strings.Contains(out, "/testVolume") {
		c.Fatal("The volume '/testVolume' expected not to be mounted")
	}

	status, _, err = sockRequest("POST", fmt.Sprintf("/containers/%s/reload", containerID), nil)
	c.Assert(err, check.IsNil)
	c.Assert(status, check.Equals, http.StatusNoContent)

	_, err = runCommand(exec.Command(dockerBinary, "start", containerID))
	c.Assert(err, check.IsNil)

	out, _, err = runCommandWithOutput(exec.Command(dockerBinary, "inspect", "-f", "{{.Volumes}}", containerID))
	if err != nil {
		c.Fatalf("failed to inspect %s. Output(%q), err(%v)", containerID, out, err)
	}
	if !strings.Contains(out, "/testVolume") {
		c.Fatal("The volume '/testVolume' expected to be mounted")
	}
}
