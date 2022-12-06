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

package router

import (
	"fmt"
	"strings"

	"github.com/wereliang/govoy/pkg/api"
)

type routerType int8

const (
	TypeHost routerType = iota
	TypePath
)

const (
	WILDCARD = "*"
)

type Router interface {
	NewRoute() Route
	Host(tpl string) Route
	Path(tpl string) Route
	PathPrefix(tpl string) Route
	Match(api.RequestHeader) (interface{}, bool)
}

type Route interface {
	Host(tpl string) Route
	Path(tpl string) Route
	PathPrefix(tpl string) Route
	Handler(handler interface{}) Route
	Match(api.RequestHeader) (interface{}, bool)
}

type router struct {
	routes []Route // 一个满足即可
}

func NewRouter() Router {
	return &router{}
}

func (r *router) NewRoute() Route {
	route := &route{}
	r.routes = append(r.routes, route)
	return route
}

func (r *router) Host(tpl string) Route {
	return r.NewRoute().Host(tpl)
}

func (r *router) Path(tpl string) Route {
	return r.NewRoute().Path(tpl)
}

func (r *router) PathPrefix(tpl string) Route {
	return r.NewRoute().PathPrefix(tpl)
}

func (r *router) Match(headers api.RequestHeader) (interface{}, bool) {
	for _, route := range r.routes {
		if h, b := route.Match(headers); b {
			return h, b
		}
	}
	return nil, false
}

type matcherWrap struct {
	matcher matcher
	rtype   routerType
}

type route struct {
	handler interface{}
	matcher []matcherWrap // 全部满足
}

func (r *route) Host(tpl string) Route {
	m := matcherWrap{rtype: TypeHost}
	idx := strings.Index(tpl, WILDCARD)
	if idx == -1 {
		m.matcher = &exactMatcher{tpl}
	} else {
		if len(tpl) == 1 {
			m.matcher = &wildcardMatcher{}
		} else if idx == 0 {
			m.matcher = &suffixMatcher{tpl[1:]}
		} else if idx == len(tpl)-1 {
			m.matcher = &prefixMatcher{tpl[:len(tpl)-1]}
		} else {
			panic(fmt.Sprintf("not support domain match:%s", tpl))
		}
	}
	r.matcher = append(r.matcher, m)
	return r
}

func (r *route) Path(tpl string) Route {
	r.matcher = append(r.matcher, matcherWrap{&exactMatcher{tpl}, TypePath})
	return r
}

func (r *route) PathPrefix(tpl string) Route {
	r.matcher = append(r.matcher, matcherWrap{&prefixMatcher{tpl}, TypePath})
	return r
}

func (r *route) Match(headers api.RequestHeader) (interface{}, bool) {
	for _, m := range r.matcher {
		switch m.rtype {
		case TypeHost:
			if !m.matcher.MatchRoute(headers.Host()) {
				return nil, false
			}
		case TypePath:
			if !m.matcher.MatchRoute(headers.Path()) {
				return nil, false
			}
		}
	}
	return r.handler, true
}

func (r *route) Handler(handler interface{}) Route {
	r.handler = handler
	return r
}

type matcher interface {
	MatchRoute(obj []byte) bool
}

type prefixMatcher struct {
	prefix string
}

func (m *prefixMatcher) MatchRoute(obj []byte) bool {
	return strings.HasPrefix(string(obj), m.prefix)
}

type suffixMatcher struct {
	suffix string
}

func (m *suffixMatcher) MatchRoute(obj []byte) bool {
	return strings.HasSuffix(string(obj), m.suffix)
}

type exactMatcher struct {
	exact string
}

func (m *exactMatcher) MatchRoute(obj []byte) bool {
	return string(obj) == m.exact
}

type wildcardMatcher struct {
}

func (m *wildcardMatcher) MatchRoute(obj []byte) bool {
	return true
}

// TODO
type regexMatcher struct {
}
