package main

import (
	"math/rand"
	"strconv"
	"strings"
	"time"
)

type Command struct {
	state         State
	pipelineAfter bool
	method        func(c *Client, line string)
}

var commands = map[string]Command{
	"HELO":             {HELO, false, commandHelo},
	"EHLO":             {HELO, false, commandEhlo},
	"MAIL":             {MAIL, true, commandMail},
	"RCPT":             {RCPT, true, commandRcpt},
	"DATA":             {DATA, false, commandData},
	"RSET":             {NONE, true, commandRset},
	"QUIT":             {NONE, false, commandQuit},
	"NOOP":             {NONE, false, commandNoop},
	"HELP":             {NONE, true, commandHelp},
	"VRFY":             {NONE, true, commandVrfy},
	"POST":             {NONE, true, commandProxy},
	"CLIENT-IP:":       {NONE, true, commandProxy},
	"X-FORWARDED-FOR:": {NONE, true, commandProxy},
	"EXPN":             {NONE, true, commandUnimplemented},
	"AUTH":             {NONE, false, commandUnimplemented},
	"STARTTLS":         {NONE, true, commandUnimplemented},
}

func clientCommand(c *Client, line string) {
	if c.state == DATA {
		// TODO: Do something with message.
		if line == "." {
			commandDataEnd(c)
		}
		return
	}

	cmd, arg := splitTwo(line, " ")

	command, ok := commands[strings.ToUpper(cmd)]
	if !ok {
		command = Command{NONE, true, commandUnimplemented}
	}

	if (!command.pipelineAfter || !c.esmtp) && c.reader.Buffered() > 0 {
		c.badPipeline = true
	}

	if checkState(c, command.state) {
		command.method(c, strings.TrimSpace(arg))
	}
}

func checkState(c *Client, state State) (ok bool) {
	if c.state+1 == state || state == NONE || (state == RCPT && c.state == RCPT) {
		return true
	}

	if c.state >= state {
		if state == HELO {
			c.Write("503 You've already said HELO (#5.5.1)")
		} else {
			c.Write("503 You've already done that, try RSET (#5.5.1)")
		}
	} else if c.state == NONE {
		c.Write("503 HELO first (#5.5.1)")
	} else if c.state == HELO {
		c.Write("503 MAIL first (#5.5.1)")
	} else {
		c.Write("503 RCPT first (#5.5.1)")
	}
	return false
}

func doHelo(c *Client, arg string) {
	c.SetState(HELO)
	c.helo = arg
	c.heloValid = validHost(arg, c.clientIp)
}

func commandHelo(c *Client, arg string) {
	doHelo(c, arg)
	c.esmtp = false
	c.Write("250 " + *serverName)
}

func commandEhlo(c *Client, arg string) {
	doHelo(c, arg)
	c.esmtp = true
	c.Write("250-" + *serverName)
	c.Write("250-PIPELINING")
	c.Write("250-8BITMIME")
	c.Write("250 SIZE")
}

func commandMail(c *Client, arg string) {
	if address, ok := parseArg(arg, "FROM:"); ok {
		c.bouncing = address == ""
		c.SetState(MAIL)
		c.Write("250 ok")
	} else {
		c.Write("501 Syntax error (#5.5.4)")
	}
}

func commandRcpt(c *Client, arg string) {
	if address, ok := parseArg(arg, "TO:"); ok {
		_, domain := splitTwo(address, "@")
		if isAllowedRcptHost(domain) {
			c.SetState(RCPT)
			c.Write("250 ok")
		} else {
			c.Write("553 sorry, that domain isn't in my list of allowed rcpthosts (#5.7.1)")
		}
	} else {
		c.Write("501 Syntax error (#5.5.4)")
	}
}

func commandData(c *Client, arg string) {
	c.SetState(DATA)
	c.Write("354 go ahead")
}

func commandDataEnd(c *Client) {
	c.SetState(SENT)
	c.Write("250 ok " + strconv.FormatInt(time.Now().Unix(), 10) + " qp " + strconv.Itoa(rand.Int()&0xffff))
}

func commandRset(c *Client, arg string) {
	if c.state > HELO {
		c.SetState(HELO)
	}
	c.Write("250 flushed")
}

func commandQuit(c *Client, arg string) {
	c.SetState(QUIT)
	c.Write("221 " + *serverName)
	c.close()
}

func commandNoop(c *Client, arg string) {
	c.Write("250 ok")
}

func commandVrfy(c *Client, arg string) {
	c.Write("252 send some mail, i'll try my best")
}

func commandHelp(c *Client, arg string) {
	c.Write("214 qmail home page: http://pobox.com/~djb/qmail.html")
}

func commandProxy(c *Client, arg string) {
	c.proxy = true
	commandUnimplemented(c, arg)
}

func commandUnimplemented(c *Client, arg string) {
	c.Write("502 unimplemented (#5.5.1)")
}
