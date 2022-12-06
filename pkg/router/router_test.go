package router

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
	"github.com/wereliang/govoy/pkg/api"
)

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

func newRequestHeader(request *fasthttp.Request) api.RequestHeader {
	return &requestHeader{RequestHeader: &request.Header, uri: request.URI()}
}

func TestHost(t *testing.T) {
	router := NewRouter()

	r := router.Host("*")
	request := &fasthttp.Request{}
	header := &request.Header
	header.SetHost("www.baidu.com")
	_, b := r.Match(newRequestHeader(request))
	assert.True(t, b)

	r = router.Host("*qq.com")
	header.SetHost("www.baidu.com")
	_, b = r.Match(newRequestHeader(request))
	assert.False(t, b)
	header.SetHost("go.qq.com")
	_, b = r.Match(newRequestHeader(request))
	assert.True(t, b)

	r = router.Host("www.qq.*")
	header.SetHost("www.qq.org")
	_, b = r.Match(newRequestHeader(request))
	assert.True(t, b)
	header.SetHost("go.qq.com")
	_, b = r.Match(newRequestHeader(request))
	assert.False(t, b)

	assert.Panics(t, func() { router.Host("www.*.com") })
}

func TestPath(t *testing.T) {
	router := NewRouter()

	r := router.Path("/foo").Handler("foo")
	request := &fasthttp.Request{}
	header := &request.Header
	header.SetRequestURI("/foo")
	h, b := r.Match(newRequestHeader(request))
	assert.True(t, b)
	assert.EqualValues(t, h, "foo")

	r = router.PathPrefix("/").Handler("root")
	header.SetRequestURI("/foo")
	h, b = r.Match(newRequestHeader(request))
	assert.True(t, b)
	assert.EqualValues(t, h, "root")

	header.SetRequestURI("hello")
	h, b = r.Match(newRequestHeader(request))
	assert.False(t, b)
}

func TestRouter(t *testing.T) {
	router := NewRouter()
	router.Host("www.qq.com")
	router.Host("www.baidu.com")

	request := &fasthttp.Request{}
	header := &request.Header

	header.SetHost("www.alibaba.com")
	_, b := router.Match(newRequestHeader(request))
	assert.False(t, b)

	header.SetHost("www.baidu.com")
	_, b = router.Match(newRequestHeader(request))
	assert.True(t, b)

	router = NewRouter()
	router.Path("/foo").Handler("foo")
	router.Path("/bar").Handler("bar")

	header.SetRequestURI("/bar")
	h, b := router.Match(newRequestHeader(request))
	assert.True(t, b)
	assert.EqualValues(t, h, "bar")
}
