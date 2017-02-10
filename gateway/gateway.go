package gateway

import (
	"net"

	"github.com/mostlygeek/arp"
	"github.com/pkg/errors"
)

func getDefaultIPNet() (*net.IPNet, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	lAddr, _, err := net.SplitHostPort(conn.LocalAddr().String())
	if nil != err {
		return nil, err
	}
	iff, err := net.InterfaceAddrs()
	if nil != err {
		return nil, err
	}
	ip := net.ParseIP(lAddr)
	if ip == nil {
		return nil, errors.Errorf("failed parsing address: %s", lAddr)
	}
	for i := 0; i < len(iff); i++ {
		n, ok := iff[i].(*net.IPNet)
		if !ok || n.IP.IsLoopback() {
			continue
		}
		if n.Contains(ip) {
			return n, nil
		}
	}
	return nil, errors.New("could not determine ip")

}

// GetDefault returns the default gateway if it is possible to determine it
func GetDefault() (string, error) {
	ipNet, err := getDefaultIPNet()
	if nil != err {
		return "", err
	}
	for ip := range arp.Table() {
		nIp := net.ParseIP(ip)
		if nIp == nil {
			continue
		}
		if ipNet.Contains(nIp) {
			return ip, nil
		}
	}
	return "", errors.New("could not determine gateway")
}
