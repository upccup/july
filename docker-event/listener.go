package event

import (
	dns "github.com/upccup/july/dns-handler"
	docker "github.com/upccup/july/docker-client"

	log "github.com/Sirupsen/logrus"
)

const (
	EventTypeContainer = "container"

	EventContainerCreate  = "create"
	EventContainerStart   = "start"
	EventContainerKill    = "kill"
	EventContainerDie     = "die"
	EventContainerDestory = "destory"
)

type DockerListener struct {
	Endpoint   string
	APIVersion string
	DNSClient  *dns.DNSClient
}

func (listener *DockerListener) StartListenDockerAction(addr, version string) {
	client, err := docker.NewVersionedClient(addr, version)
	if err != nil {
		log.Fatalf("create docker client got error: %+v", err)
		return
	}

	eventsChan := make(chan *docker.APIEvents, 10)
	if err := client.AddEventListener(eventsChan); err != nil {
		log.Fatalf("create docker client got error: %+v", err)
	}

	log.Info("add docker event listener success")

	for {
		select {
		case e := <-eventsChan:
			if e != nil {
				listener.HandleDockerEvent(e)
			}
		}
	}
}

func (listener *DockerListener) HandleDockerEvent(e *docker.APIEvents) {
	if e.Type != EventTypeContainer {
		log.Debugf("got event from docker: %#v, type is not container drop it!!!", e)
		return
	}

	switch e.Action {
	case EventContainerStart:
		if err := listener.DNSClient.AddDNSRecord("yaoyuntest", "192.168.1.100"); err != nil {
			log.Errorf("add dns record failed. Error: %s", err.Error())
		}
	case EventContainerDie:
	default:
		log.Debugf("got event from docker: %#v, not be interested in it drop!!", e)
		return
	}
}
