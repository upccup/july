package config

import (
	"path/filepath"
)

const (
	ContainerIPStorePrefix    = "/jdjr/containers"
	ContainerDomainsStorePath = "/jdjr/container-domains"
	HostAssignedIPStorePath   = "/jdjr/hosts/assigned"
	HostIPPoolStorePath       = "/jdjr/hosts/pool"
	HostIPConfigStorePath     = "/jdjr/hosts/config"
)

func GetHostIPConfigStorePath(ip string) string {
	return filepath.Join("/jdjr/hosts/config", ip)
}

func ContainerIPPoolSotrePath(ipNet string) string {
	return filepath.Join(ContainerIPStorePrefix, ipNet, "pool")
}

func ContainerAssignedIPSotrePath(ipNet string) string {
	return filepath.Join(ContainerIPStorePrefix, ipNet, "assigned")
}

func ContainerIPConfigSotrePath(ipNet string) string {
	return filepath.Join(ContainerIPStorePrefix, ipNet, "config")
}
