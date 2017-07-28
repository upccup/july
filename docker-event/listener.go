package event

import (
	"errors"

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
	DockerClient *docker.Client
	DNSClient    *dns.DNSClient
}

type ContainerIPInfo struct {
	IP     string
	Domain string
	Labels map[string]string
}

func (listener *DockerListener) StartListenDockerAction(addr, version string) {
	eventsChan := make(chan *docker.APIEvents, 10)
	if err := listener.DockerClient.AddEventListener(eventsChan); err != nil {
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
		containerIPInfo, err := listener.GetContainerIPInfo(e.ID)
		if err != nil {
			log.Errorf("get container ip info failed. Error: %s", err.Error())
			return
		}

		if err := listener.DNSClient.AddDNSRecord(containerIPInfo.Domain, containerIPInfo.IP); err != nil {
			log.Errorf("add dns record failed. Error: %s", err.Error())
			return
		}
	case EventContainerDie:
		log.Infof("got container died event, container ID: %s", e.ID)
	default:
		log.Debugf("got event from docker: %#v, not be interested in it drop!!", e)
		return
	}
}

func (listener *DockerListener) GetContainerIPInfo(ID string) (*ContainerIPInfo, error) {
	containerInfo, err := listener.DockerClient.InspectContainer(ID)
	if err != nil {
		return nil, err
	}

	if containerInfo == nil || containerInfo.Config.Labels == nil ||
		containerInfo.NetworkSettings == nil || containerInfo.NetworkSettings.Networks == nil {
		return nil, errors.New("get container IP info failed: null response")
	}

	//TODO(upccup): read container labels get domain info
	var domain, ip string
	domain = "dockertest"
	for _, value := range containerInfo.NetworkSettings.Networks {
		ip = value.IPAddress
	}

	if ip == "" {
		return nil, errors.New("container ip is empty")
	}

	return &ContainerIPInfo{Domain: domain, IP: ip}, nil
}
