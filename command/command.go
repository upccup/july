package command

import (
	"fmt"
	"path/filepath"

	"github.com/upccup/july/bridge"
	"github.com/upccup/july/config"
	"github.com/upccup/july/db"
	dns "github.com/upccup/july/dns-handler"
	docker "github.com/upccup/july/docker-client"
	event "github.com/upccup/july/docker-event"
	"github.com/upccup/july/ipamdriver"
	"github.com/upccup/july/util"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

func NewServerCommand() cli.Command {
	return cli.Command{
		Name:  "server",
		Usage: "start the IPAM plugin server& add docker event listener",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "docker-endpoint",
				Value: "tcp://127.0.0.1:2376",
				Usage: "the docker daemon endpoint. [$DOCKER_ENDPOINT]",
			},
			cli.StringFlag{
				Name:  "dns-endpoint",
				Value: "http://127.0.0.1:9999",
				Usage: "the dns console server endpoint. [$DNS_ENDPOINT]",
			},
		},
		Action: startServerAction,
	}
}

func startServerAction(c *cli.Context) {
	// start ipam server
	go ipamdriver.StartServer()

	log.Debug("docker endpoint: ", c.String("docker-endpoint"))
	client, err := docker.NewVersionedClient(c.String("docker-endpoint"), "1.21")
	if err != nil {
		log.Fatalf("create docker client got error: %+v", err)
		return
	}

	if err := client.Ping(); err != nil {
		log.Fatalf("connect to docker client got error: %+v", err)
		return
	}

	log.Debug("docker endpoint: ", c.String("dns-endpoint"))
	dockerEvenListener := &event.DockerListener{
		DockerClient: client,
		DNSClient:    &dns.DNSClient{Endpoint: c.String("dns-endpoint")},
	}
	dockerEvenListener.StartListenDockerAction()
}

func NewIPRangeCommand() cli.Command {
	return cli.Command{
		Name:  "ip-range",
		Usage: "set the ip range for containers",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "ip-start", Usage: "the first IP for containers in CIDR notation"},
			cli.StringFlag{Name: "ip-end", Usage: "the last IP for containers in CIDR notation"},
		},
		Action: ipRangeAction,
	}
}

func ipRangeAction(c *cli.Context) {
	ip_start := c.String("ip-start")
	ip_end := c.String("ip-end")
	if ip_start == "" || ip_end == "" {
		fmt.Println("Invalid args")
		return
	}
	ipamdriver.AllocateIPRange(ip_start, ip_end)
}

func NewReleaseIPCommand() cli.Command {
	return cli.Command{
		Name:  "release-ip",
		Usage: "release the specified IP address",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "ip", Usage: "the IP to release in CIDR notation"},
		},
		Action: releaseIPAction,
	}
}

func releaseIPAction(c *cli.Context) {
	ip_args := c.String("ip")
	if ip_args == "" {
		fmt.Println("Invalid args")
		return
	}
	ip_net, _ := util.GetIPNetAndMask(ip_args)
	ip, _ := util.GetIPAndCIDR(ip_args)
	ipamdriver.ReleaseIP(ip_net, ip)
}

func NewReleaseHostCommand() cli.Command {
	return cli.Command{
		Name:  "release-host",
		Usage: "release the specified host",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "ip", Usage: "the IP to release in CIDR notation"},
		},
		Action: releaseHostAction,
	}
}

func releaseHostAction(c *cli.Context) {
	ip := c.String("ip")
	if ip == "" {
		fmt.Println("Invalid args")
		return
	}
	bridge.ReleaseHost(ip)
}

func NewHostRangeCommand() cli.Command {
	return cli.Command{
		Name:  "host-range",
		Usage: "set the ip range for hosts",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "ip-start", Usage: "the first IP for containers in CIDR notation"},
			cli.StringFlag{Name: "ip-end", Usage: "the last IP for containers in CIDR notation"},
			cli.StringFlag{Name: "gateway", Usage: "the default gateway for the docker container network"},
		},
		Action: hostRangeAction,
	}

}

func hostRangeAction(c *cli.Context) {
	ip_start := c.String("ip-start")
	ip_end := c.String("ip-end")
	gateway := c.String("gateway")
	if ip_start == "" || ip_end == "" || gateway == "" {
		fmt.Println("Invalid args")
		return
	}
	bridge.AllocateHostRange(ip_start, ip_end, gateway)
}

func NewCreateNetworkCommand() cli.Command {
	return cli.Command{
		Name:  "create-network",
		Usage: "create docker network br0",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "ip", Usage: "the IP docker bridge use"},
			cli.StringFlag{Name: "name", Usage: "the docker network name"},
		},
		Action: createNetworkAction,
	}
}

func createNetworkAction(c *cli.Context) {
	ip := c.String("ip")
	name := c.String("name")
	bridge.CreateNetwork(ip, name)
}

func NewShowAssignedIPCommand() cli.Command {
	return cli.Command{
		Name: "ip-assigned",
		Flags: []cli.Flag{
			cli.BoolFlag{Name: "container", Usage: "show container ip pool"},
			cli.BoolFlag{Name: "host", Usage: "show host ip pool"},
		},
		Usage:  "show the which has been assigned",
		Action: showAssignedIPAction,
	}
}

func showAssignedIPAction(c *cli.Context) {
	// show all assigned host IP
	if c.Bool("host") {
		hostNodes, err := db.GetKeys(config.HostAssignedIPStorePath)
		if err != nil {
			log.Fatal("get assigned ip failed. Error: ", err)
			return
		}

		log.Info("assgined host IP:   ")
		for _, hostNode := range hostNodes {
			log.Info(hostNode.Key, "  ", hostNode.Value)
		}
	}

	if c.Bool("container") {
		log.Info("assigned container IP: ")
		containerNets, err := db.GetKeys(config.ContainerIPStorePrefix)
		if err != nil {
			log.Fatal("get contaienr nets failed. Error: ", err)
			return
		}

		for _, containerNet := range containerNets {
			assignedNodes, err := db.GetKeys(filepath.Join(containerNet.Key, "assigned"))
			if err != nil {
				log.Fatalf("get contaienr net %s assigned ips failed. Error: %s", containerNet.Key, err.Error())
				return
			}

			for _, assignedNode := range assignedNodes {
				log.Info(assignedNode.Key, "  ", assignedNode.Value)
			}
		}
	}
}

func NewShowIPPoolCommand() cli.Command {
	return cli.Command{
		Name:  "ip-pool",
		Usage: "show the ip poop",
		Flags: []cli.Flag{
			cli.BoolFlag{Name: "container", Usage: "show container ip pool"},
			cli.BoolFlag{Name: "host", Usage: "show host ip pool"},
		},
		Action: showIPPoolAction,
	}
}

func showIPPoolAction(c *cli.Context) {
	if c.Bool("host") {
		hostNodes, err := db.GetKeys(config.HostIPPoolStorePath)
		if err != nil {
			log.Fatal("get assigned ip failed. Error: ", err)
			return
		}

		log.Info("host IP pool:   ")
		for _, hostNode := range hostNodes {
			log.Info(hostNode.Key, "  ", hostNode.Value)
		}
	}

	if c.Bool("container") {
		log.Info("container IP pool: ")
		containerNets, err := db.GetKeys(config.ContainerIPStorePrefix)
		if err != nil {
			log.Fatal("get contaienr nets failed. Error: ", err)
			return
		}

		for _, containerNet := range containerNets {
			idleNodes, err := db.GetKeys(filepath.Join(containerNet.Key, "pool"))
			if err != nil {
				log.Fatalf("get contaienr net %s assigned ips failed. Error: %s", containerNet.Key, err.Error())
				return
			}

			for _, idleNode := range idleNodes {
				log.Info(idleNode.Key, "  ", idleNode.Value)
			}
		}
	}
}
