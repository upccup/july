package event

import (
	"encoding/json"
	"errors"
	"path/filepath"

	"github.com/upccup/july/config"
	"github.com/upccup/july/db"
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

	DomainZoneKey = "JR_DOMAIN_ZONE"
	DomainNameKey = "JR_DOMAIN_NAME"
)

type DockerListener struct {
	DockerClient *docker.Client
	DNSClient    *dns.DNSClient
}

type ContainerIPInfo struct {
	IP     string
	Domain string
	Zone   string
	Labels map[string]string
}

func (listener *DockerListener) StartListenDockerAction() {
	eventsChan := make(chan *docker.APIEvents, 10)
	if err := listener.DockerClient.AddEventListener(eventsChan); err != nil {
		log.Fatalf("create docker client got error: %+v", err)
	}

	log.Info("add docker event listener success")

	defer func() {
		if err := listener.DockerClient.RemoveEventListener(eventsChan); err != nil {
			log.Fatal(err)
		}
	}()

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
		log.Infof("got container start event, container ID: %s", e.ID)
		containerIPInfo, err := listener.GetContainerIPInfo(e.ID)
		if err != nil {
			log.Errorf("get container ip info failed. Error: %s", err.Error())
			return
		}

		ipInfoBytes, err := json.Marshal(containerIPInfo)
		if err != nil {
			log.Errorf("Marshal ContainerIPInfo: %+v failed. Err: %s", containerIPInfo, err.Error())
			return
		}

		if err := db.SetKey(filepath.Join(config.ContainerDomainsStorePath, e.ID), string(ipInfoBytes)); err != nil {
			log.Errorf("store container %s domain %s failed. Error: %s", e.ID, string(ipInfoBytes), err.Error())
			return
		}

		if err := listener.DNSClient.AddDNSRecord(containerIPInfo.Zone, containerIPInfo.Domain, containerIPInfo.IP); err != nil {
			log.Errorf("add dns record failed. Error: %s", err.Error())
			return
		}
	case EventContainerDie:
		log.Infof("got container died event, container ID: %s", e.ID)
		domainStoreKey := filepath.Join(config.ContainerDomainsStorePath, e.ID)
		domainBytes, err := db.GetKey(domainStoreKey)
		if err != nil {
			log.Errorf("get container %s domain failed. Error: %s", e.ID, err.Error())
			return
		}

		var ipInfo ContainerIPInfo
		if err := json.Unmarshal([]byte(domainBytes), &ipInfo); err != nil {
			log.Errorf("Unmarshal %s failed. Error: %s", string(domainBytes), err.Error())
			return
		}

		if err := listener.DNSClient.DeleteDNSRecord(ipInfo.Domain); err != nil {
			log.Errorf("delete dns record %s.%s failed. Error: %s", ipInfo.Zone, ipInfo.Domain, err.Error())
			return
		}

		if err := db.DeleteKey(domainStoreKey); err != nil {
			log.Errorf("delete container %s dns from db failed. Error: %s", domainStoreKey, err.Error())
			return
		}
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

	containerLabels := containerInfo.Config.Labels

	domainMain, ok := containerLabels[DomainZoneKey]
	if !ok {
		return nil, errors.New("get container domain main info failed: null response")
	}

	domainName, ok := containerLabels[DomainNameKey]
	if !ok {
		return nil, errors.New("get container domain name info failed: null response")
	}

	var ip string
	for _, value := range containerInfo.NetworkSettings.Networks {
		ip = value.IPAddress
	}

	if ip == "" {
		return nil, errors.New("container ip is empty")
	}

	return &ContainerIPInfo{Domain: domainName, Zone: domainMain, IP: ip}, nil
}
