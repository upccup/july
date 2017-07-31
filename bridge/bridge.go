package bridge

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/upccup/july/config"
	"github.com/upccup/july/db"
	"github.com/upccup/july/util"

	log "github.com/Sirupsen/logrus"
)

type IPConfig struct {
	Subnet  string
	Gateway string
}

func AllocateHostRange(ip_start, ip_end, gateway string) []string {
	ips := util.GetIPRange(ip_start, ip_end)
	ip_net, mask := util.GetIPNetAndMask(ip_start)
	for _, ip := range ips {
		if checkIPAssigned(ip) {
			log.Warnf("IP %s has been allocated", ip)
			continue
		}
		db.SetKey(filepath.Join(config.HostIPPoolStorePath, ip), "")
	}
	initializeConfig(ip_net, fmt.Sprint(ip_net, "/", mask), gateway)
	fmt.Println("Allocate Hosts Done! Total:", len(ips))
	return ips
}

func initializeConfig(ip_net, subnet, gateway string) error {
	ipConfig := &IPConfig{Subnet: subnet, Gateway: gateway}
	config_bytes, _ := json.Marshal(ipConfig)
	if err := db.SetKey(config.HostIPConfigStorePath, string(config_bytes)); err != nil {
		return err
	}

	log.Infof("Initialized Config %s for network %s success", string(config_bytes), ip_net)
	return nil
}

func getConfig() (*IPConfig, error) {
	config, err := db.GetKey(config.HostIPConfigStorePath)
	if err != nil {
		return nil, err
	}

	log.Debugf("getConfig %s", config)
	conf := &IPConfig{}
	json.Unmarshal([]byte(config), conf)
	return conf, err
}

func allocateHost(ip string) error {
	if ip == "" {
		return errors.New("arg ip is lack")
	}

	if err := db.DeleteKey(filepath.Join(config.HostIPPoolStorePath, ip)); err != nil {
		return err
	}

	if err := db.SetKey(filepath.Join(config.HostAssignedIPStorePath, ip), ""); err != nil {
		return err
	}

	log.Infof("Allocated host %s", ip)
	return nil
}

func getHost(ip string) (string, error) {
	ip_pool, err := db.GetKeys(config.HostIPPoolStorePath)
	if err != nil {
		return "", err
	}

	if len(ip_pool) == 0 {
		return "", errors.New("Pool is empty")
	}

	if ip == "" {
		find_ip := strings.Split(ip_pool[0].Key, "/")
		ip = find_ip[len(find_ip)-1]
	} else if !db.IsKeyExist(filepath.Join(config.HostIPPoolStorePath, ip)) {
		return "", errors.New(fmt.Sprintf("Host %s not in pool", ip))
	}

	if checkIPAssigned(ip) {
		return "", errors.New(fmt.Sprintf("Host %s has been allocated", ip))
	}

	log.Infof("Host IP: %s", ip)
	return ip, nil
}

func checkIPAssigned(ip string) bool {
	return db.IsKeyExist(filepath.Join(config.HostAssignedIPStorePath, ip))
}

func ReleaseHost(ip string) error {
	if err := db.DeleteKey(filepath.Join(config.HostAssignedIPStorePath, ip)); err != nil {
		log.Fatal(err)
	}

	if err := db.SetKey(filepath.Join(config.HostIPPoolStorePath, ip), ""); err != nil {
		log.Fatal(err)
	}

	log.Infof("Release host %s", ip)
	return nil
}

func CreateNetwork(ip, networkName string) {
	var assigned_ip string
	var config *IPConfig
	var err error

	if config, err = getConfig(); err != nil {
		log.Fatal(err)
	}

	if assigned_ip, err = getHost(ip); err != nil {
		log.Fatal(err)
	}

	if err = allocateHost(assigned_ip); err != nil {
		log.Fatal(err)
	}

	if err = createBridge(assigned_ip, config.Subnet, config.Gateway, networkName); err != nil {
		log.Fatal(err)
	}

	// TODO(upccup): now restart network maybe make network donot work, after find the reason and
	// fix it this will work again
	//if err = restart_network(); err != nil {
	//	log.Fatal(err)
	//}
	//log.Infof("Create network %s done", assigned_ip)
	log.Infof("Create network on ip:%s done, please restart network or reboot!!", assigned_ip)
}
