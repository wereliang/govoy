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

package tls_inspector

import (
	http_inspector_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/listener/http_inspector/v3"
	"github.com/golang/protobuf/proto"
	"github.com/wereliang/govoy/pkg/api"
	"github.com/wereliang/govoy/pkg/filter"
	"github.com/wereliang/govoy/pkg/log"
)

func init() {
	filter.ListenerFilterFactory.Regist(new(HttpInspectorFactory))
}

const (
	HTTP11 = "http/1.1"
)

var (
	minMethodLengh = len("GET")
	maxMethodLengh = len("CONNECT")
	httpMethod     = map[string]struct{}{
		"OPTIONS": {},
		"GET":     {},
		"HEAD":    {},
		"POST":    {},
		"PUT":     {},
		"DELETE":  {},
		"TRACE":   {},
		"CONNECT": {},
	}
)

type HttpInspectorFilter struct {
}

func (f *HttpInspectorFilter) OnAccept(cb api.ListenerFilterCallbacks) api.FilterStatus {
	c := cb.Connection()
	data, err := c.Peek(maxMethodLengh)
	if err != nil {
		log.Error("peer error. %s", err)
		return api.Continue
	}

	size := len(data)
	if size < minMethodLengh {
		log.Error("peer error. %s", err)
		return api.Continue
	}

	if size > maxMethodLengh {
		size = maxMethodLengh
	}

	// check http1, 这里不是很严谨
	for i := minMethodLengh; i <= size; i++ {
		if _, ok := httpMethod[string(data[:i])]; ok {
			cb.Connection().Context().SetApplicationProtocol(HTTP11)
			log.Debug("check http1.1 by method:%s", string(data[:i]))
			return api.Continue
		}
	}

	log.Error("check http fail")
	return api.Continue
}

type HttpInspectorFactory struct {
}

func (f *HttpInspectorFactory) Name() string {
	return filter.Listener_TlsInspector
}

func (f *HttpInspectorFactory) CreateEmptyConfigProto() proto.Message {
	return &http_inspector_v3.HttpInspector{}
}

func (f *HttpInspectorFactory) CreateFilterFactory(pb proto.Message, context api.FactoryContext) api.ListenerFilterCreator {
	return func(cb api.ListenerFilterManager) {
		cb.AddAcceptFilter(&HttpInspectorFilter{})
	}
}
