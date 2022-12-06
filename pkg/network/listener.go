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
	"net"
	"sync/atomic"

	"github.com/wereliang/govoy/pkg/api"
)

func NewListener(addr net.Addr) api.Listener {
	l := &listener{addr: addr}
	return l
}

type listener struct {
	addr net.Addr
	cb   atomic.Value
}

func (nl *listener) Listen() error {
	l, err := net.Listen(nl.addr.Network(), nl.addr.String())
	if err != nil {
		return err
	}
	// log.Debug("network listen. %s", nl.addr.String())
	for {
		conn, err := l.Accept()
		if err != nil {
			panic(err)
		}

		go func() {
			cb := nl.cb.Load().(api.ListenerCallback)
			cb.OnAccept(newConnSize(conn, 4096))
		}()
	}
}

func (nl *listener) SetCallback(cb api.ListenerCallback) {
	nl.cb.Store(cb)
}

func (nl *listener) Addr() net.Addr {
	return nl.addr
}
