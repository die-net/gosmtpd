gosmtpd
=======

Golang-based SMTP Listener

Copyright &copy; 2013 Aaron Hopkins tools@die.net

This is an unfinished SMTP listener.  It receives connections, speaks SMTP,
and discards any mail delivered to it.  I might finish it some day, but I'm
publishing it in the hope that someone will find bits of it interesting or
useful.

This was my first non-trivial Go app, and is a port of a high-volume evented
SMTP listener I'd written in C years ago.  The equivalent C code is 5 times
as much code, single-threaded, slightly slower, and took 10 times as long to
get correct.

Building:
--------

Install [Go](http://golang.org/doc/install) and git, then:

	git clone https://github.com/die-net/gosmtpd.git
	cd gosmtpd
	go build

And you'll end up with an "gosmtpd" binary in the current directory.

Command-line flags:
------------------

	-listen=":2525": The [IP]:port to listen for incoming connections on.
	-rcpthosts="localhost": Comma-separated list of domains to accept mail for.
	-servername="localhost": The hostname to show to clients

It defaults to dual-stack IPv4/IPv6.  If you want IPv4-only, specify an IPv4
listen address, like -listen="0.0.0.0:2525".
