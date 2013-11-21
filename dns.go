package main

import (
	"net"
)

// Lookup hostname and verify that at least one result matches ip.
func validHost(hostname string, ip string) (match bool) {
	if addrs, err := net.LookupHost(hostname); err == nil {
		for _, addr := range addrs {
			if addr == ip {
				return true
			}
		}
	}
	return false
}

// Lookup rDNS of ip, then lookup returned hostnames, and verify that at
// least one result matches ip.  Return hostname and if it matches.
func reverseDNS(ip string) (hostname string, match bool) {
	if names, err := net.LookupAddr(ip); err == nil && len(names) > 0 {
		for _, name := range names {
			if validHost(name, ip) {
				return name, true
			}
		}
		return names[0], false
	}
	return ip, false
}
