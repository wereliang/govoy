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
	tls_inspectorv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/listener/tls_inspector/v3"
	"github.com/golang/protobuf/proto"
	"github.com/wereliang/govoy/pkg/api"
	"github.com/wereliang/govoy/pkg/filter"
)

func init() {
	filter.ListenerFilterFactory.Regist(new(TlsInspectorFactory))
}

type TlsInspectorFilter struct {
}

func (f *TlsInspectorFilter) OnAccept(cb api.ListenerFilterCallbacks) api.FilterStatus {
	return api.Continue
}

type TlsInspectorFactory struct {
}

func (f *TlsInspectorFactory) Name() string {
	return filter.Listener_HttpInspector
}

func (f *TlsInspectorFactory) CreateEmptyConfigProto() proto.Message {
	return &tls_inspectorv3.TlsInspector{}
}

func (f *TlsInspectorFactory) CreateFilterFactory(pb proto.Message, context api.FactoryContext) api.ListenerFilterCreator {
	return func(cb api.ListenerFilterManager) {
		cb.AddAcceptFilter(&TlsInspectorFilter{})
	}
}
