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
	"fmt"

	"github.com/wereliang/govoy/pkg/api"
)

type Handler interface {
	api.HTTPFilterManager
	api.DecoderFilterCallbacks
	Decode(api.StreamContext) error
	Encode(api.StreamContext) error
}

func NewHandler(r api.RouteConfigMatcher, c api.Connection) Handler {
	return &httpHandler{routeMatcher: r, connection: c}
}

type httpHandler struct {
	decodeFilters []api.StreamDecoderFilter
	encodeFilters []api.StreamEncoderFilter
	routeMatcher  api.RouteConfigMatcher
	connection    api.Connection
}

func (h *httpHandler) Route() api.RouteConfigMatcher {
	return h.routeMatcher
}

func (h *httpHandler) SetRoute(r api.RouteConfigMatcher) {
	h.routeMatcher = r
}

func (h *httpHandler) Connection() api.Connection {
	return h.connection
}

func (h *httpHandler) AddDecodeFilter(f api.StreamDecoderFilter) {
	f.SetDecoderFilterCallbacks(h)
	h.decodeFilters = append(h.decodeFilters, f)
}

func (h *httpHandler) AddEncodeFilter(f api.StreamEncoderFilter) {
	h.encodeFilters = append(h.encodeFilters, f)
}

func (h *httpHandler) Decode(ctx api.StreamContext) error {
	for _, f := range h.decodeFilters {
		if f.Decode(ctx) == api.Stop {
			return fmt.Errorf("decode error")
		}
	}
	return nil
}

func (h *httpHandler) Encode(ctx api.StreamContext) error {
	for _, f := range h.encodeFilters {
		if f.Encode(ctx) == api.Stop {
			return fmt.Errorf("encode error")
		}
	}
	return nil
}
