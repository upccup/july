package ipamdriver

import (
	"fmt"

	"github.com/upccup/july/util"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/go-plugins-helpers/ipam"
)

type MyIPAMHandler struct {
}

func (iph *MyIPAMHandler) GetCapabilities() (response *ipam.CapabilitiesResponse, err error) {
	log.Infof("GetCapabilities")
	return &ipam.CapabilitiesResponse{RequiresMACAddress: true}, nil
}

func (iph *MyIPAMHandler) GetDefaultAddressSpaces() (response *ipam.AddressSpacesResponse, err error) {
	log.Infof("GetDefaultAddressSpaces")
	return &ipam.AddressSpacesResponse{}, nil
}

func (iph *MyIPAMHandler) RequestPool(request *ipam.RequestPoolRequest) (response *ipam.RequestPoolResponse, err error) {
	log.Infof("RequestPool: %#v", request)
	ip_net, _ := util.GetIPNetAndMask(request.Pool)
	_, ip_cidr := util.GetIPAndCIDR(request.Pool)
	options := request.Options
	return &ipam.RequestPoolResponse{ip_net, ip_cidr, options}, nil
}

func (iph *MyIPAMHandler) ReleasePool(request *ipam.ReleasePoolRequest) (err error) {
	log.Infof("ReleasePool %#v is danger, you should do this by manual.", request)
	return nil
}

func (iph *MyIPAMHandler) RequestAddress(request *ipam.RequestAddressRequest) (response *ipam.RequestAddressResponse, err error) {
	log.Infof("function RequestAddress param request: %#v", request)
	ip_net := request.PoolID
	ip := request.Address
	config, _ := GetConfig(ip_net)

	// TODO:(upccup) check is ip in the pool and legitimacy check
	if ip != "" {
		log.Infof("request ip: %s is not empty return it", ip)
		return &ipam.RequestAddressResponse{fmt.Sprintf("%s/%s", ip, config.Mask), nil}, nil
	}

	ip, err = AllocateIP(ip_net, ip)
	return &ipam.RequestAddressResponse{fmt.Sprintf("%s/%s", ip, config.Mask), nil}, err
}

func (iph *MyIPAMHandler) ReleaseAddress(request *ipam.ReleaseAddressRequest) (err error) {
	log.Infof("function ReleaseAddress param request: %#v", request)
	return ReleaseIP(request.PoolID, request.Address)
}
