package listener

import "net"

type Listener interface {
	Accept() (conn net.Conn, err error)
	Addr() string
	Port() int
}
