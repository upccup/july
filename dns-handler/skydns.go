package dns

import (
	"encoding/json"
	"path/filepath"
	"strings"

	"github.com/upccup/july/db"
)

// From https://github.com/skynetservices/skydns/blob/master/msg/service.go#L23
// This *is* the rdata from a SRV record, but with a twist.
// Host (Target in SRV) must be a domain name, but if it looks like an IP
// address (4/6), we will treat it like an IP address.
type Service struct {
	Host     string `json:"host,omitempty"`
	Port     int    `json:"port,omitempty"`
	Priority int    `json:"priority,omitempty"`
	Weight   int    `json:"weight,omitempty"`
	Text     string `json:"text,omitempty"`
	Mail     bool   `json:"mail,omitempty"` // Be an MX record. Priority becomes Preference.
	Ttl      uint32 `json:"ttl,omitempty"`

	// When a SRV record with a "Host: IP-address" is added, we synthesize
	// a srv.Target domain name.  Normally we convert the full Key where
	// the record lives to a DNS name and use this as the srv.Target.  When
	// TargetStrip > 0 we strip the left most TargetStrip labels from the
	// DNS name.
	TargetStrip int `json:"targetstrip,omitempty"`

	// Group is used to group (or *not* to group) different services
	// together. Services with an identical Group are returned in the same
	// answer.
	Group string `json:"group,omitempty"`

	// Etcd key where we found this service and ignored from json un-/marshalling
	Key string `json:"-"`
}

func Reverse(source string) string {
	s := strings.Split(source, ".")
	reverse(s)
	return strings.Join(s, "/")
}

func reverse(ss []string) {
	last := len(ss) - 1
	for i := 0; i < len(ss)/2; i++ {
		ss[i], ss[last-i] = ss[last-i], ss[i]
	}

	return
}

func AddDNSRecord(zone, domain, ip string) error {
	value, err := json.Marshal(Service{Host: ip})
	if err != nil {
		return err
	}

	return db.SetKey(filepath.Join(zone, Reverse(domain)), string(value))
}

func RemoveDNSRecord(zone, domain string) error {
	return db.DeleteKey(filepath.Join(zone, Reverse(domain)))
}
