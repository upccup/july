package dns

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
)

type DNSClient struct {
	Endpoint string
}

const (
	DNSPassword = "abcc"
)

type DNSRecord struct {
	FullDomain     string          `json:"full_domain"`
	Main           string          `json:"main"`
	AddressRecords []AddressRecord `json:"records"`
}

type AddressRecord struct {
	Address string `json:"address"`
	Type    string `json:"type"`
	Area    int    `json:"area"`
}

func (dClient *DNSClient) AddDNSRecord(domain, address string) error {
	timeStamp := strconv.FormatInt(time.Now().Unix(), 10)
	tokenStrList := []string{timeStamp, timeStamp, DNSPassword}
	tokenStr := strings.Join(tokenStrList, "")
	token := fmt.Sprintf("%x", md5.Sum([]byte(tokenStr)))

	addressRecord := AddressRecord{
		Address: address,
		Type:    "A",
		Area:    1,
	}

	dnsRecord := DNSRecord{
		FullDomain:     domain + ".cbpmgt.com.",
		Main:           "cbpmgt.com.",
		AddressRecords: []AddressRecord{addressRecord},
	}

	body, err := encodeData([]DNSRecord{dnsRecord})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", dClient.Endpoint, body)
	if err != nil {
		return err
	}

	req.Header.Set("Auth-User", "yanhong3")
	req.Header.Set("Auth-Random", timeStamp)
	req.Header.Set("Auth-TimeStamp", timeStamp)
	req.Header.Set("Auth-Token", token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp == nil {
		return fmt.Errorf("add dns record failed: empty response")
	}

	if resp.StatusCode != http.StatusOK {
		result, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		return fmt.Errorf("add dns record failed: status code %d, result %s", resp.StatusCode, string(result))
	}

	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	log.Infof("add dns record got response %s", string(result))
	return nil
}

func encodeData(data interface{}) (*bytes.Buffer, error) {
	params := bytes.NewBuffer(nil)
	if data != nil {
		if err := json.NewEncoder(params).Encode(data); err != nil {
			return nil, err
		}
	}
	return params, nil
}
