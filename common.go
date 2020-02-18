package main

import (
	"fmt"
	"log"
	"net"
	"sync"
)

var m = sync.Mutex{}

func validateHostname(host string) (string, bool) {
	if ok := ipCheck(host); !ok {
		// get IPv4 from DNS Name
		ipv4, err := dnsNameResolve(host)
		if err != nil {
			// Not ipv4 and Not DNS Record
			return host, false
		}
		// Chenge DNS to IPv4
		return ipv4, true
	}
	// This is IPv4
	return host, true
}

func dnsNameResolve(host string) (string, error) {
	addr, err := net.ResolveIPAddr("ip4", host)
	if err != nil {
		return "", err
	}
	return addr.String(), nil
}

func ipCheck(hostName string) bool {
	v := net.ParseIP(hostName)
	if v.To4() == nil {
		return false
	}
	return true
}

func masterErrInfo(buf string) {
	if Debug {
		fmt.Println(buf)
	}
	log.Println(buf)
}
