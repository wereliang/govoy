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

package api

import (
	"bufio"
	"bytes"

	envoy_config_listener_v3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	"github.com/golang/protobuf/proto"
)

type FilterStatus int

const (
	Stop     FilterStatus = 0
	Continue FilterStatus = 1
)

// ListenerFilterCallbacks
type ListenerFilterCallbacks interface {
	// Connection return api.Connection
	Connection() Connection
}

// ListenerFilter
type ListenerFilter interface {
	// OnAccept
	OnAccept(ListenerFilterCallbacks) FilterStatus
}

// ReadFilter for network filter
type ReadFilter interface {
	// OnData is called everytime bytes is read from the connection
	OnData(*bytes.Buffer) FilterStatus

	// OnNewConnection is called on new connection is created
	OnNewConnection() FilterStatus
}

// WriteFilter for network filter
type WriteFilter interface {
	// OnWrite is called before data write to raw connection
	OnWrite(*bufio.Writer) FilterStatus
}

// StreamDecoderFilter for http stream filter
type StreamDecoderFilter interface {
	// Decode call before request
	Decode(StreamContext) FilterStatus

	// SetDecoderFilterCallbacks
	SetDecoderFilterCallbacks(DecoderFilterCallbacks)
}

// DecoderFilterCallbacks
type DecoderFilterCallbacks interface {
	// Connection return api.Connection
	Connection() Connection

	// Route return route config matcher
	Route() RouteConfigMatcher

	// SetRoute
	SetRoute(RouteConfigMatcher)
}

// StreamEncoderFilter for http stream filter
type StreamEncoderFilter interface {
	// Encode call after request
	Encode(StreamContext) FilterStatus
}

// Factory is basic factory
type Factory interface {
	// Name is factory name
	Name() string

	// CreateEmptyConfigProto return filter protobuf config
	CreateEmptyConfigProto() proto.Message
}

// ListenerFilterManager manager listener filter
type ListenerFilterManager interface {
	// AddAcceptFilter add listener filter
	AddAcceptFilter(ListenerFilter)
}

type ListenerFilterCreator func(ListenerFilterManager)

// ListernerFactory for listener factory
type ListernerFactory interface {
	Factory

	// CreateFilterFactory create listener filter factory
	CreateFilterFactory(proto.Message, FactoryContext) ListenerFilterCreator
}

// FilterManager manager network filter
type FilterManager interface {
	// AddReadFilter add read filter
	AddReadFilter(ReadFilter)

	// AddWriteFilter add write filter
	AddWriteFilter(WriteFilter)
}

type NetworkFilterCreator func(FilterManager, ConnectionCallbacks) error

// NetworkFactory for network factory
type NetworkFactory interface {
	Factory

	// CreateFilterFactory create network filter factory
	CreateFilterFactory(proto.Message, FactoryContext) NetworkFilterCreator
}

// HTTPFilterManager manager http filter
type HTTPFilterManager interface {
	// AddDecodeFilter add http decoder filter
	AddDecodeFilter(StreamDecoderFilter)

	// AddEncodeFilter add http encoder filter
	AddEncodeFilter(StreamEncoderFilter)
}

type HTTPFilterCreator func(HTTPFilterManager)

// HTTPFactory for http factory
type HTTPFactory interface {
	Factory

	// CreateFilterFactory create http filter factory
	CreateFilterFactory(proto.Message, FactoryContext) HTTPFilterCreator
}

// FactoryContext some context for filter
type FactoryContext interface {
	// ListenerManager
	ListenerManager() ListenerManager

	// ClusterManager
	ClusterManager() ClusterManager

	// RouteConfigManager
	RouteConfigManager() RouteConfigManager
}

// FilterChainManager filter chain manager
type FilterChainManager interface {
	// AddFilterChains add filterchain
	AddFilterChains([]*envoy_config_listener_v3.FilterChain, *envoy_config_listener_v3.FilterChain)

	// FindFilterChains match filterchain
	FindFilterChains(ConnectionContext) *envoy_config_listener_v3.FilterChain
}
