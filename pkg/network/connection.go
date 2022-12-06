/*
 * MIT License
 *
 * Copyright (c) 2022 wereliang
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package network

import (
	"bufio"
	"net"

	"github.com/wereliang/govoy/pkg/api"
)

func newConn(c net.Conn) api.Connection {
	return &connection{Conn: c, ctx: newConnectionContext(c), reader: bufio.NewReader(c)}
}

func newConnSize(c net.Conn, size int) api.Connection {
	return &connection{Conn: c, ctx: newConnectionContext(c), reader: bufio.NewReaderSize(c, size)}
}

type connection struct {
	net.Conn
	ctx    api.ConnectionContext
	reader *bufio.Reader
}

func (c *connection) Context() api.ConnectionContext {
	return c.ctx
}

func (c *connection) Raw() net.Conn {
	return c.Conn
}

func (c *connection) Peek(n int) ([]byte, error) {
	return c.reader.Peek(n)
}

func (c *connection) Read(p []byte) (int, error) {
	return c.reader.Read(p)
}

func newConnectionContext(c net.Conn) api.ConnectionContext {
	cs := &ConnectionContextImpl{}
	remote := c.RemoteAddr().(*net.TCPAddr)
	cs.SourceIP = remote.IP
	cs.SourcePort = uint32(remote.Port)
	local := c.LocalAddr().(*net.TCPAddr)
	cs.DestinationIP = local.IP
	cs.DestinationPort = uint32(local.Port)
	return cs
}

type ConnectionContextImpl struct {
	DestinationPort      uint32
	DestinationIP        net.IP
	ServerName           string
	TransportProtocol    string
	ApplicationProtocol  string
	DirectSourceIP       net.IP
	SourceType           int32
	SourceIP             net.IP
	SourcePort           uint32
	localAddressRestored bool
}

func (cs *ConnectionContextImpl) GetDestinationPort() uint32 {
	return cs.DestinationPort
}

func (cs *ConnectionContextImpl) GetDestinationIP() net.IP {
	return cs.DestinationIP
}

func (cs *ConnectionContextImpl) GetServerName() string {
	return cs.ServerName
}

func (cs *ConnectionContextImpl) GetTransportProtocol() string {
	return cs.TransportProtocol
}

func (cs *ConnectionContextImpl) GetApplicationProtocol() string {
	return cs.ApplicationProtocol
}

func (cs *ConnectionContextImpl) GetDirectSourceIP() net.IP {
	return cs.DirectSourceIP
}

func (cs *ConnectionContextImpl) GetSourceType() int32 {
	return cs.SourceType
}

func (cs *ConnectionContextImpl) GetSourceIP() net.IP {
	return cs.SourceIP
}

func (cs *ConnectionContextImpl) GetSourcePort() uint32 {
	return cs.SourcePort
}

func (cs *ConnectionContextImpl) SetOriginalDestination(ip net.IP, port uint32) {
	cs.DestinationIP = ip
	cs.DestinationPort = port
	cs.localAddressRestored = true
}

func (cs *ConnectionContextImpl) SetApplicationProtocol(s string) {
	cs.ApplicationProtocol = s
}

func (cs *ConnectionContextImpl) LocalAddressRestored() bool {
	return cs.localAddressRestored
}
