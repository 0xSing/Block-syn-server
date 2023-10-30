package utils

import (
	"fmt"
	"net"
)

func GetIPAddress() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		// 检查IP地址类型是否为IPv4或IPv6
		ipNet, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}
		ip := ipNet.IP
		// 排除回环地址和IPv6地址
		if ip.IsLoopback() || ip.To4() == nil {
			continue
		}
		return ip.String(), nil
	}
	return "", fmt.Errorf("not found ip address")
}
