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

package http

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/valyala/fasthttp"
	"github.com/wereliang/govoy/pkg/api"
	"github.com/wereliang/govoy/pkg/log"
)

type StreamServer interface {
	Dispatch(*bytes.Buffer) error
}

type StreamClient interface {
	Call(api.StreamContext) error
}

func NewStreamServer(handler Handler, conn api.ConnectionCallbacks) StreamServer {
	s := &httpStreamServer{
		handler:   handler,
		bufChan:   make(chan *bytes.Buffer),
		endChan:   make(chan struct{}),
		closeChan: make(chan struct{}),
		conn:      conn,
	}
	s.br = bufio.NewReader(s)
	go func() {
		s.serve()
	}()
	return s
}

type httpStreamServer struct {
	handler   Handler
	bufChan   chan *bytes.Buffer
	endChan   chan struct{}
	closeChan chan struct{}
	br        *bufio.Reader
	conn      api.ConnectionCallbacks
}

func (s *httpStreamServer) Dispatch(buffer *bytes.Buffer) error {
	select {
	case <-s.closeChan:
		return fmt.Errorf("stream server close")
	default:
		s.bufChan <- buffer
		<-s.endChan
	}
	return nil
}

func (s *httpStreamServer) Read(dst []byte) (n int, err error) {
	buf, ok := <-s.bufChan
	if !ok {
		err = io.EOF
	} else {
		n, err = buf.Read(dst)
	}
	s.endChan <- struct{}{}
	return n, err
}

func (s *httpStreamServer) close() {
	close(s.closeChan)
	s.conn.Close()
}

func (s *httpStreamServer) serve() {
	for {
		request, response := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
		// blocking read using fasthttp.Request.Read
		err := request.ReadLimitBody(s.br, 1024*1024)
		if err != nil {
			if err != io.EOF {
				log.Error("ReadLimitBody error: %s. conn close", err)
				s.close()
			}
			break
		}

		ctx := NewStreamContext(context.TODO(), request, response)
		err = s.handle(ctx)
		if err != nil {
			log.Error("handle error : %s", err)
		}
		response.WriteTo(s.conn)
	}
	log.Debug("server close")
}

func (s *httpStreamServer) handle(ctx api.StreamContext) error {
	if err := s.handler.Decode(ctx); err != nil {
		return err
	}
	if err := s.handler.Encode(ctx); err != nil {
		return err
	}
	return nil
}

// 此处只区分了127.0.0.6的特殊情况
func NewStreamClient(addr net.Addr) StreamClient {
	var sc *fasthttpStreamClient
	if addr == nil {
		sc = &fasthttpStreamClient{}
	} else {
		sc = &fasthttpStreamClient{
			client: fasthttp.Client{Dial: Dial},
		}
	}
	sc.client.MaxIdleConnDuration = time.Minute
	sc.client.DisableHeaderNamesNormalizing = true
	sc.client.MaxConnsPerHost = 30000
	return sc
}

type fasthttpStreamClient struct {
	client fasthttp.Client
}

func (sc *fasthttpStreamClient) Call(ctx api.StreamContext) error {
	request := ctx.Request().Raw().(*fasthttp.Request)
	response := ctx.Response().Raw().(*fasthttp.Response)
	request.UseHostHeader = true
	return sc.client.DoTimeout(request, response, time.Second*4)
}

var dialerWithLAddr = &fasthttp.TCPDialer{
	Concurrency: 1000, LocalAddr: &net.TCPAddr{IP: net.ParseIP("127.0.0.6")}}

func Dial(addr string) (net.Conn, error) {
	return dialerWithLAddr.Dial(addr)
}

var defaultStreamClient = NewStreamClient(nil)
var streamClientWithLAddr = NewStreamClient(&net.TCPAddr{})

func Call(ctx api.StreamContext, laddr net.Addr) error {
	if laddr != nil {
		return streamClientWithLAddr.Call(ctx)
	}
	return defaultStreamClient.Call(ctx)
}
