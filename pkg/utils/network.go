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

package utils

import (
	"fmt"
	"net"
	"strings"
	"syscall"

	envoy_config_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	"github.com/wereliang/govoy/pkg/network"
)

// OriginDST, option for syscall.GetsockoptIPv6Mreq
const (
	SO_ORIGINAL_DST      = 80
	IP6T_SO_ORIGINAL_DST = 80
)

func GetOriginalAddr(conn net.Conn) ([]byte, int, error) {
	tc := conn.(*net.TCPConn)

	f, err := tc.File()
	if err != nil {
		return nil, 0, fmt.Errorf("[originaldst] get conn file error, err: %v", err)
	}
	defer f.Close()

	fd := int(f.Fd())
	addr, _ := syscall.GetsockoptIPv6Mreq(fd, syscall.IPPROTO_IP, SO_ORIGINAL_DST)

	if err := syscall.SetNonblock(fd, true); err != nil {
		return nil, 0, fmt.Errorf("setnonblock %v", err)
	}

	p0 := int(addr.Multiaddr[2])
	p1 := int(addr.Multiaddr[3])

	port := p0*256 + p1
	ip := addr.Multiaddr[4:8]
	return ip, port, nil
}

func ToNetAddr(addr *envoy_config_v3.Address) (net.Addr, error) {
	return toNetAddr(addr, func(saddr *envoy_config_v3.SocketAddress) net.Addr {
		return &net.TCPAddr{IP: net.ParseIP(saddr.GetAddress()), Port: int(saddr.GetPortValue())}
	})
}

func ToDNSAddr(addr *envoy_config_v3.Address) (net.Addr, error) {
	return toNetAddr(addr, func(saddr *envoy_config_v3.SocketAddress) net.Addr {
		return &network.DNSAddr{Addr: saddr.GetAddress(), Port: int(saddr.GetPortValue())}
	})
}

func toNetAddr(
	addr *envoy_config_v3.Address,
	fn func(*envoy_config_v3.SocketAddress) net.Addr) (net.Addr, error) {

	var naddr net.Addr
	saddr := addr.GetSocketAddress()
	switch strings.ToLower(saddr.GetProtocol().String()) {
	case "tcp":
		naddr = fn(saddr)
	default:
		return nil, fmt.Errorf("not support protocol")
	}
	return naddr, nil
}
