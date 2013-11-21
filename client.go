package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"time"
)

type State int

const (
	NONE State = iota
	HELO
	MAIL
	RCPT
	DATA
	SENT
	QUIT
	CLOSE
)

type Client struct {
	conn         *net.TCPConn
	reader       *bufio.Reader
	scanner      *bufio.Scanner
	writer       *bufio.Writer
	state        State
	connectionId uint64
	created      time.Time
	lastRead     time.Time
	clientIp     string
	rDns         string
	helo         string
	esmtp        bool
	rDnsValid    bool
	heloValid    bool
	bouncing     bool
	proxy        bool
	badPipeline  bool
}

func NewClient(conn *net.TCPConn, connectionId uint64) (c *Client) {
	addr := conn.RemoteAddr().String()
	ip, _, err := net.SplitHostPort(addr)
	if err != nil {
		log.Fatalln("Couldn't SplitHostPort " + addr)
	}

	c = &Client{
		conn:         conn,
		clientIp:     ip,
		reader:       bufio.NewReaderSize(conn, 1024),
		writer:       bufio.NewWriterSize(conn, 1024),
		state:        NONE,
		connectionId: connectionId,
		created:      time.Now(),
	}

	c.updateLastRead()

	c.conn.SetKeepAlive(true)

	return c
}

func (c *Client) updateLastRead() {
	c.lastRead = time.Now()

	to := c.lastRead.Add(time.Minute)
	if cto := c.created.Add(5 * time.Minute); cto.Before(to) {
		to = cto
	}

	c.conn.SetReadDeadline(to)
}

func (c *Client) Read() (line string, ok bool) {
	// DoS protection: ReadLine limits its response to the size of the
	// c.reader buffer.  Scanner and ReadString grow without bound.
	buf, prefix, err := c.reader.ReadLine()

	if err == nil && !prefix {
		c.updateLastRead()

		// ReadLine returns a reference to a circular buffer.  Need
		// to copy to a string to keep it.
		line = copyToString(buf)
		ok = true
	} else {
		if prefix {
			c.Write("421 out of memory (#4.3.0)")
		} else if isTimeout(err) {
			c.Write("451 timeout (#4.4.2)")
		}
		c.close()

		line = ""
		ok = false
	}
	return
}

func (c *Client) Write(s string) (ok bool) {
	c.conn.SetWriteDeadline(time.Now().Add(time.Minute))

	_, err := fmt.Fprintln(c.writer, s)
	ok = err == nil
	if !ok {
		c.close()
	}
	return
}

func (c *Client) Flush() (ok bool) {
	ok = c.writer.Flush() == nil
	if !ok {
		c.close()
	}
	return
}

func (c *Client) SetState(state State) {
	if c.state != CLOSE {
		c.state = state
	}
}

func (c *Client) close() {
	c.writer.Flush()
	c.conn.Close()
	c.state = CLOSE
}
