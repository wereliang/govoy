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

package original_dst

import (
	"net"

	envoy_filters_listener_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/listener/original_dst/v3"
	"github.com/golang/protobuf/proto"
	"github.com/wereliang/govoy/pkg/api"
	"github.com/wereliang/govoy/pkg/filter"
	"github.com/wereliang/govoy/pkg/log"
	"github.com/wereliang/govoy/pkg/utils"
)

func init() {
	filter.ListenerFilterFactory.Regist(new(OriginalDstFactory))
}

type OriginalDstFilter struct {
}

func (f *OriginalDstFilter) OnAccept(cb api.ListenerFilterCallbacks) api.FilterStatus {
	ctx := cb.Connection().Context()
	conn := cb.Connection().Raw()
	ip, port, err := utils.GetOriginalAddr(conn)
	if err != nil {
		log.Error("get original addr error. %s", err)
		return api.Continue
	}
	ctx.SetOriginalDestination(net.IP(ip), uint32(port))
	log.Debug("original ip:%s port:%d", ctx.GetDestinationIP(), ctx.GetDestinationPort())
	return api.Continue
}

type OriginalDstFactory struct {
}

func (f *OriginalDstFactory) Name() string {
	return filter.Listener_OriginalDst
}

func (f *OriginalDstFactory) CreateEmptyConfigProto() proto.Message {
	return &envoy_filters_listener_v3.OriginalDst{}
}

func (f *OriginalDstFactory) CreateFilterFactory(pb proto.Message, context api.FactoryContext) api.ListenerFilterCreator {
	return func(cb api.ListenerFilterManager) {
		cb.AddAcceptFilter(&OriginalDstFilter{})
	}
}
