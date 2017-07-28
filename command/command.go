package command

import (
	"fmt"
	"os"

	"github.com/upccup/july/bridge"
	"github.com/upccup/july/db"
	dns "github.com/upccup/july/dns-handler"
	docker "github.com/upccup/july/docker-client"
	event "github.com/upccup/july/docker-event"
	"github.com/upccup/july/ipamdriver"
	"github.com/upccup/july/util"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

var (
	debug bool
)

func initialize_log() {
	log.SetOutput(os.Stderr)
	if debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
}

func NewServerCommand() cli.Command {
	return cli.Command{
		Name:   "server",
		Usage:  "start the TalkingData IPAM plugin",
		Action: startServerAction,
	}
}

func startServerAction(c *cli.Context) {
	debug = c.GlobalBool("debug")
	db.SetDBAddr(c.GlobalString("cluster-store"))
	initialize_log()
	ipamdriver.StartServer()
}

func NewDockerAgentCommand() cli.Command {
	return cli.Command{
		Name:   "agent",
		Usage:  "start an agent listen docker event",
		Action: startListenDockerAction,
	}
}

func startListenDockerAction(c *cli.Context) {
	client, err := docker.NewVersionedClient("tcp://172.25.60.39:8092", "1.21")
	if err != nil {
		log.Fatalf("create docker client got error: %+v", err)
		return
	}

	dockerEvenListener := &event.DockerListener{
		DockerClient: client,
		DNSClient:    &dns.DNSClient{Endpoint: "http://dns2-test.cbpmgt.com/api/domain_add"},
	}
	dockerEvenListener.StartListenDockerAction("tcp://172.25.60.39:8092", "1.21")
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
	db.SetDBAddr(c.GlobalString("cluster-store"))
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
	db.SetDBAddr(c.GlobalString("cluster-store"))
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
	db.SetDBAddr(c.GlobalString("cluster-store"))
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
	db.SetDBAddr(c.GlobalString("cluster-store"))
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
	db.SetDBAddr(c.GlobalString("cluster-store"))
	ip := c.String("ip")
	name := c.String("name")
	bridge.CreateNetwork(ip, name)
}
