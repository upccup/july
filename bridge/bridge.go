package bridge

import (
	"encoding/json"
	"errors"
	"path/filepath"

	"github.com/upccup/july/config"
	"github.com/upccup/july/db"

	log "github.com/Sirupsen/logrus"
)

type IPConfig struct {
	Subnet  string
	Gateway string
}

func AddHostIP(ip, subnet, gateway string) error {
	ipConfig := IPConfig{Subnet: subnet, Gateway: gateway}
	configBytes, err := json.Marshal(ipConfig)
	if err != nil {
		return err
	}

	if err := db.SetKey(config.GetHostIPConfigStorePath(ip), string(configBytes)); err != nil {
		return err
	}

	return nil
}

func getConfig(ip string) (*IPConfig, error) {
	config, err := db.GetKey(config.GetHostIPConfigStorePath(ip))
	if err != nil {
		return nil, err
	}

	log.Debugf("getConfig of ip %s success. config: %s", ip, config)
	conf := &IPConfig{}
	if err := json.Unmarshal([]byte(config), conf); err != nil {
		return nil, err
	}

	return conf, nil
}

func allocateHost(ip string) error {
	if !db.IsKeyExist(config.GetHostIPConfigStorePath(ip)) {
		return errors.New("ip config not found")
	}

	if err := db.SetKey(filepath.Join(config.HostAssignedIPStorePath, ip), ""); err != nil {
		return err
	}

	log.Infof("Allocated host %s", ip)
	return nil
}

func checkIPAssigned(ip string) bool {
	return db.IsKeyExist(filepath.Join(config.HostAssignedIPStorePath, ip))
}

func ReleaseHost(ip string) error {
	if err := db.DeleteKey(filepath.Join(config.HostAssignedIPStorePath, ip)); err != nil {
		log.Fatal(err)
	}

	log.Infof("Release host %s", ip)
	return nil
}

func CreateNetwork(ip, networkName string) {
	var config *IPConfig
	var err error

	if config, err = getConfig(ip); err != nil {
		log.Fatal(err)
	}

	if err = allocateHost(ip); err != nil {
		log.Fatal(err)
	}

	if err = createBridge(ip, config.Subnet, config.Gateway, networkName); err != nil {
		log.Fatal(err)
	}

	// TODO(upccup): now restart network maybe make network donot work, after find the reason and
	// fix it this will work again
	//if err = restart_network(); err != nil {
	//	log.Fatal(err)
	//}
	//log.Infof("Create network %s done", assigned_ip)
	log.Infof("Create network on ip:%s done, please restart network or reboot!!", ip)
}
