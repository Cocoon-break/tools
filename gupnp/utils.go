package main

import (
	"errors"
	"net"
	"strings"
)

// 获取网卡上的v4Ip
func GetIpV4ByName(cardName string) (ip string, err error) {
	v4, _, err := getInterfaceByName(cardName)
	if err != nil {
		return
	}
	if len(v4) == 0 {
		return "", errors.New(cardName + " not found v4")
	}
	return v4.String(), nil
}

func getInterfaceByName(cardName string) (v4, v6 net.IP, err error) {
	var ipAddr net.IP
	var ipAddrv6 net.IP
	var ipAddrv6s []net.IP
	var addrs []net.Addr
	interfaces, err := net.InterfaceByName(cardName)
	if err != nil {
		return nil, nil, err
	}
	// get addresses
	if addrs, err = interfaces.Addrs(); err != nil {
		return ipAddr, ipAddr, err
	}
	for _, addr := range addrs {
		addrIpNet, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}
		addrIp := addrIpNet.IP
		if addrIp.To4() != nil {
			ipAddr = addrIp
			continue
		}
		if addrIp.To16() != nil {
			ipAddrv6 = addrIp
			ipAddrv6s = append(ipAddrv6s, addrIp)
		}
	}
	//多个v6ip，优先用公网v6
	if len(ipAddrv6s) > 1 {
		for _, addrv6 := range ipAddrv6s {
			if addrv6.IsGlobalUnicast() && !IsInnerIpv6(addrv6.String()) {
				ipAddrv6 = addrv6
				break
			}
		}
	}
	return ipAddr, ipAddrv6, nil
}

// 判断是否是内网V6ip
func IsInnerIpv6(ipStr string) bool {
	ss := strings.Split(ipStr, ":")
	if len(ss) < 2 {
		return false
	}
	// IPV6 内网标识
	if ss[0] == "fe80" {
		return true
	}
	return false
}
