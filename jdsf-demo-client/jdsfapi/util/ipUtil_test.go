package util

import (
	"fmt"
	"testing"
)

func TestGetHostIp(t *testing.T) {
	hostIp := GetHostIp()
	fmt.Println(hostIp)
}

func TestGetHostIpUsePing(t *testing.T) {
	hostIp := GetHostIpUsePing("192.168.181.57:14268")
	fmt.Println(hostIp)
}