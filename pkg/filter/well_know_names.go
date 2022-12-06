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

package filter

const (
	Listener_TlsInspector  = "envoy.filters.listener.tls_inspector"
	Listener_OriginalDst   = "envoy.filters.listener.original_dst"
	Listener_HttpInspector = "envoy.filters.listener.http_inspector"

	Network_Echo                  = "envoy.filters.network.echo"
	Network_HttpConnectionManager = "envoy.filters.network.http_connection_manager"

	HTTP_Router = "envoy.filters.http.router"
)

var well_know_names = map[string]struct{}{
	Listener_TlsInspector:         {},
	Listener_OriginalDst:          {},
	Listener_HttpInspector:        {},
	Network_HttpConnectionManager: {},
	HTTP_Router:                   {},
}

func IsWellknowName(name string) bool {
	_, ok := well_know_names[name]
	return ok
}
