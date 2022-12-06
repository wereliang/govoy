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

import "context"

// HeaderMap is some header action
type HeaderMap interface {
	Del(key string)
	Add(key, value string)
	Set(key, value string)
	Get(key string) []byte
}

// StreamContext http stream context
type StreamContext interface {
	Context() context.Context
	Request() Request
	Response() Response
}

// Request http request
type Request interface {
	Header() RequestHeader
	Body() Body
	SetHost(string)
	Raw() interface{}
}

// Response http response
type Response interface {
	Header() ResponseHeader
	Body() Body
	Raw() interface{}
}

// RequestHeader http request header
type RequestHeader interface {
	HeaderMap
	Method() []byte
	SetMethod(method string)
	Host() []byte
	SetHost(host string)
	RequestURI() []byte
	SetRequestURI(requestURI string)
	Path() []byte
	SetPath(path string)
}

// ResponseHeader http response header
type ResponseHeader interface {
	HeaderMap
	StatusCode() int
	SetStatusCode(statusCode int)
}

// Body http body action
type Body interface {
	AppendBody([]byte)
	SetBody([]byte)
}
