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

package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	_ "net/http/pprof"

	_ "github.com/wereliang/govoy/pkg/filter/http/myrouter"
	_ "github.com/wereliang/govoy/pkg/filter/http/router"
	_ "github.com/wereliang/govoy/pkg/filter/listener/http_inspector"
	_ "github.com/wereliang/govoy/pkg/filter/listener/original_dst"
	_ "github.com/wereliang/govoy/pkg/filter/listener/tls_inspector"
	_ "github.com/wereliang/govoy/pkg/filter/network/echo"
	_ "github.com/wereliang/govoy/pkg/filter/network/http_connection_manager"
	"github.com/wereliang/govoy/pkg/log"
	"github.com/wereliang/govoy/pkg/server"
)

var (
	cfg string
)

func init() {
	flag.StringVar(&cfg, "c", "", "please input config path")
	flag.Parse()
}

func main() {
	log.DefaultLog = log.NewSimpleLogger(log.TraceLevel, true)

	svc, err := server.NewGovoy(cfg)
	if err != nil {
		panic(err)
	}
	err = svc.Start()
	if err != nil {
		panic(err)
	}
	defer svc.Stop()

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c
	log.Debug("stopping...")
}
