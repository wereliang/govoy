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
	"net"

	envoy_config_listener_v3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
)

type ListenerType int32

const (
	LISTENER_STATIC ListenerType = 0
	LISTENER_EDS    ListenerType = 3
)

// ActiveConnection is connection handler
type ActiveConnection interface {
	FilterManager
	Connection

	// OnLoop is loop by connection
	OnLoop()
}

// ActiveListener is listener handler
type ActiveListener interface {
	ListenerFilterManager

	// Start
	Start() error

	// Type is listener type
	Type() ListenerType

	// Listener return api.Listener
	Listener() Listener

	// GetUseOriginalDst return is use original-dst
	GetUseOriginalDst() bool

	// GetBindToPort return is bind to port
	GetBindToPort() bool
}

// ListenerManager
type ListenerManager interface {
	// AddOrUpdateListener is add or update listerner
	AddOrUpdateListener(ListenerType, *envoy_config_listener_v3.Listener) error

	// FindListenerByAddress find listerner by address
	FindListenerByAddress(net.Addr) ActiveListener

	// FindListenerByName find listener by name
	FindListenerByName(string) ActiveListener

	// DeleteListener delete exist listener
	DeleteListener(string) error

	// Range like sync.map.range
	Range(func(string, ObjectConfig) bool)
}
