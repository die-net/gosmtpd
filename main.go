package main

import (
	"flag"
	"log"
	"math/rand"
	"net"
	"strings"
	"time"
)

var serverName = flag.String("servername", hostname(), "The hostname to show to clients")
var listenAddr = flag.String("listen", ":2525", "The [IP]:port to listen for incoming connections on.")
var rcptHosts = flag.String("rcpthosts", hostname(), "Comma-separated list of domains to accept mail for.")

var allowedRcptHosts = map[string]bool{}

func main() {
	flag.Parse()

	parseRcptHosts()

	rand.Seed(time.Now().UnixNano())

	addr, err := net.ResolveTCPAddr("tcp", *listenAddr)
	if err != nil {
		log.Fatalln("FATAL: ResolveTCPAddr: " + err.Error())
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Fatalln("FATAL: ListenTCP: " + err.Error())
	}

	var connectionid uint64 = 1

	for {
		tcpconn, err := listener.AcceptTCP()
		if err != nil {
			continue
		}

		go handleClient(NewClient(tcpconn, connectionid))

		connectionid++
	}
}

func handleClient(c *Client) {
	defer c.close()

	c.rDns, c.rDnsValid = reverseDNS(c.clientIp)

	c.Write("220 " + *serverName + " ESMTP")

	for c.state != CLOSE && c.Flush() {
		line, ok := c.Read()
		if !ok {
			break
		}

		clientCommand(c, line)
	}
}

func parseRcptHosts() {
	hosts := strings.Split(*rcptHosts, ",")
	for _, h := range hosts {
		allowedRcptHosts[h] = true
	}
}

func isAllowedRcptHost(hostname string) (ok bool) {
	// DoS protection: Limit how many subdomains we'll skip.
	for i := 0; i < 4 && hostname != ""; i++ {
		if allowedRcptHosts[hostname] {
			return true
		}

		// Skip leading dot, remove everything before next dot.
		if dot := strings.Index(hostname[1:], "."); dot >= 0 {
			hostname = hostname[dot+1:]
		} else {
			break
		}
	}
	return false
}
