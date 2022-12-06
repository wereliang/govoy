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
	"context"

	"github.com/valyala/fasthttp"
	"github.com/wereliang/govoy/pkg/api"
)

func NewStreamContext(context context.Context,
	req *fasthttp.Request, rsp *fasthttp.Response) api.StreamContext {
	return &streamContext{
		context:  context,
		request:  newRequest(req),
		response: newResponse(rsp),
	}
}

type streamContext struct {
	context  context.Context
	request  api.Request
	response api.Response
}

func (sc *streamContext) Context() context.Context {
	return sc.context
}

func (sc *streamContext) Request() api.Request {
	return sc.request
}

func (sc *streamContext) Response() api.Response {
	return sc.response
}

func newRequest(r *fasthttp.Request) api.Request {
	return &request{
		Request: r,
		header:  newRequestHeader(r),
		body:    newRequestBody(r),
	}
}

type request struct {
	*fasthttp.Request
	header api.RequestHeader
	body   api.Body
}

func (r *request) Header() api.RequestHeader {
	return r.header
}

func (r *request) Body() api.Body {
	return r.body
}

func (r *request) Raw() interface{} {
	return r.Request
}

func newResponse(r *fasthttp.Response) api.Response {
	return &response{
		Response: r,
		header:   newResponseHeader(r),
		body:     newResponseBody(r),
	}
}

type response struct {
	*fasthttp.Response
	header api.ResponseHeader
	body   api.Body
}

func (r *response) Header() api.ResponseHeader {
	return r.header
}

func (r *response) Body() api.Body {
	return r.body
}

func (r *response) Raw() interface{} {
	return r.Response
}

func newRequestHeader(request *fasthttp.Request) api.RequestHeader {
	return &requestHeader{RequestHeader: &request.Header, uri: request.URI()}
}

func newResponseHeader(response *fasthttp.Response) api.ResponseHeader {
	return &responseHeader{&response.Header}
}

func newRequestBody(request *fasthttp.Request) api.Body {
	return &requestBody{Request: request}
}

func newResponseBody(response *fasthttp.Response) api.Body {
	return &responseBody{Response: response}
}

type requestHeader struct {
	*fasthttp.RequestHeader
	uri *fasthttp.URI
}

func (h *requestHeader) Get(key string) []byte {
	return h.RequestHeader.Peek(key)
}

func (h *requestHeader) Path() []byte {
	return h.uri.Path()
}

func (h *requestHeader) SetPath(path string) {
	h.uri.SetPath(path)
}

type responseHeader struct {
	*fasthttp.ResponseHeader
}

func (h *responseHeader) Get(key string) []byte {
	return h.ResponseHeader.Peek(key)
}

type requestBody struct {
	*fasthttp.Request
}

type responseBody struct {
	*fasthttp.Response
}
