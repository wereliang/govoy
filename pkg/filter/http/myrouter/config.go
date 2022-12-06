package myrouter

import (
	"github.com/golang/protobuf/proto"
	"github.com/wereliang/govoy/pkg/api"
	"github.com/wereliang/govoy/pkg/filter"
)

func init() {
	filter.HTTPFilterFactory.Regist(new(Factory))
}

type MyRouterFilter struct {
	cb      api.DecoderFilterCallbacks
	context api.FactoryContext
}

func (r *MyRouterFilter) SetDecoderFilterCallbacks(cb api.DecoderFilterCallbacks) {
	r.cb = cb
	r.cb.SetRoute(defaultRouter)
}

func (r *MyRouterFilter) Decode(ctx api.StreamContext) api.FilterStatus {
	return api.Continue
}

type Factory struct {
}

func (f *Factory) Name() string {
	return "envoy.filters.http.myrouter"
}

func (f *Factory) CreateEmptyConfigProto() proto.Message {
	return nil
}

func (f *Factory) CreateFilterFactory(pb proto.Message, context api.FactoryContext) api.HTTPFilterCreator {
	return func(cb api.HTTPFilterManager) {
		router := &MyRouterFilter{context: context}
		cb.AddDecodeFilter(router)
	}
}
