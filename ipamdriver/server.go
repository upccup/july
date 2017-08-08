package ipamdriver

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
	"github.com/docker/go-plugins-helpers/ipam"
)

type Config struct {
	Ipnet string
	Mask  string
}

func StartServer() {
	d := &MyIPAMHandler{}
	h := ipam.NewHandler(d)
	h.ServeUnix("root", "jdjr")
}

func AllocateIPRange(ip_start, ip_end string) []string {
	ips := util.GetIPRange(ip_start, ip_end)
	ipNet, mask := util.GetIPNetAndMask(ip_start)
	for _, ip := range ips {
		if checkIPAssigned(ipNet, ip) {
			log.Warnf("IP %s has been allocated", ip)
			continue
		}

		if err := AddContainerIP(ipNet, ip); err != nil {
			log.Errorf("add contaienr ip %s failed. Error: %s", ip, err.Error())
			continue
		}
	}

	initializeConfig(ipNet, mask)
	log.Info("Allocate Containers IP Done! Total:", len(ips))
	return ips
}

func AddContainerIP(ipNet, ip string) error {
	return db.SetKey(filepath.Join(config.ContainerIPPoolSotrePath(ipNet), ip), "")
}

func ReleaseIP(ipNet, ip string) error {
	if err := db.DeleteKey(filepath.Join(config.ContainerAssignedIPSotrePath(ipNet), ip)); err != nil {
		log.Infof("Skip Release IP %s", ip)
		return err
	}

	if err := db.SetKey(filepath.Join(config.ContainerIPPoolSotrePath(ipNet), ip), ""); err != nil {
		return err
	}

	log.Infof("Release IP %s", ip)
	return nil
}

func AllocateIP(ipNet, ip string) (string, error) {
	ip_pool, err := db.GetKeys(config.ContainerIPPoolSotrePath(ipNet))
	if err != nil {
		return ip, err
	}

	if len(ip_pool) == 0 {
		return ip, errors.New("Pool is empty")
	}

	if ip == "" {
		find_ip := strings.Split(ip_pool[0].Key, "/")
		ip = find_ip[len(find_ip)-1]
	}

	if checkIPAssigned(ipNet, ip) {
		return ip, errors.New(fmt.Sprintf("IP %s has been allocated", ip))
	}

	if err := db.DeleteKey(filepath.Join(config.ContainerIPPoolSotrePath(ipNet), ip)); err != nil {
		return ip, err
	}

	db.SetKey(filepath.Join(config.ContainerAssignedIPSotrePath(ipNet), ip), "")
	log.Infof("Allocated IP %s", ip)
	return ip, err
}

func checkIPAssigned(ipNet, ip string) bool {
	return db.IsKeyExist(filepath.Join(config.ContainerAssignedIPSotrePath(ipNet), ip))
}

func initializeConfig(ipNet, mask string) error {
	ipConfig := &Config{Ipnet: ipNet, Mask: mask}
	config_bytes, err := json.Marshal(ipConfig)
	if err != nil {
		log.Fatal(err)
	}

	if err := db.SetKey(config.ContainerIPConfigSotrePath(ipNet), string(config_bytes)); err != nil {
		log.Fatal(err)
	}

	log.Infof("Initialized Config %s for network %s", string(config_bytes), ipNet)
	return nil
}

func DeleteNetWork(ip_net string) error {
	err := db.DeleteKey(filepath.Join(config.ContainerIPStorePrefix, ip_net))
	if err == nil {
		log.Infof("DeleteNetwork %s", ip_net)
	}
	return err
}

func GetConfig(ipNet string) (*Config, error) {
	config, err := db.GetKey(config.ContainerIPConfigSotrePath(ipNet))
	if err == nil {
		log.Debugf("GetConfig %s from network %s", config, ipNet)
	}
	conf := &Config{}
	json.Unmarshal([]byte(config), conf)
	return conf, err
}
