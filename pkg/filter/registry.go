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

import (
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/wereliang/govoy/pkg/api"
	"github.com/wereliang/govoy/pkg/log"
)

var (
	ListenerFilterFactory = newRegistFactory()
	NetworkFilterFactory  = newRegistFactory()
	HTTPFilterFactory     = newRegistFactory()
)

func GetListenerFactory(a *any.Any, name string) (api.ListernerFactory, proto.Message) {
	factory, pb := getFactory(a, name, ListenerFilterFactory)
	if factory != nil {
		return factory.(api.ListernerFactory), pb
	}
	return nil, pb
}

func GetNetworkFactory(a *any.Any, name string) (api.NetworkFactory, proto.Message) {
	factory, pb := getFactory(a, name, NetworkFilterFactory)
	if factory != nil {
		return factory.(api.NetworkFactory), pb
	}
	return nil, pb
}

func GetHTTPFactory(a *any.Any, name string) (api.HTTPFactory, proto.Message) {
	factory, pb := getFactory(a, name, HTTPFilterFactory)
	if factory != nil {
		return factory.(api.HTTPFactory), pb
	}
	return nil, pb
}

func getFactory(a *any.Any, name string, r *registFactory) (api.Factory, proto.Message) {
	var (
		factory api.Factory
		pb      proto.Message
	)
	if a == nil {
		factory = r.GetFactoryByName(name)
	} else {
		if factory = r.GetFactoryByType(a.TypeUrl); factory == nil {
			return nil, nil
		}
		pb = factory.CreateEmptyConfigProto()
		err := ptypes.UnmarshalAny(a, pb)
		if err != nil {
			panic(err)
		}
	}
	return factory, pb
}

type registFactory struct {
	namedFactorys map[string]api.Factory
	typedFactorys map[string]api.Factory
}

func newRegistFactory() *registFactory {
	return &registFactory{
		namedFactorys: make(map[string]api.Factory),
		typedFactorys: make(map[string]api.Factory),
	}
}

func (f *registFactory) Regist(factory api.Factory) {
	name := factory.Name()
	if name == "" {
		panic("name can't be empty")
	}
	if _, ok := f.namedFactorys[name]; ok {
		panic("factory name exist")
	}
	// fmt.Println("regist factory:", name)
	log.Debug("regist factory: %s", name)
	f.namedFactorys[name] = factory

	if pb := factory.CreateEmptyConfigProto(); pb != nil {
		any, err := ptypes.MarshalAny(pb)
		if err != nil {
			panic(err)
		}
		typed := any.GetTypeUrl()
		if _, ok := f.typedFactorys[typed]; ok {
			panic("factory typed exist")
		}
		f.typedFactorys[typed] = factory
	}
}

func (f *registFactory) GetFactoryByName(name string) api.Factory {
	if f, ok := f.namedFactorys[name]; ok {
		return f
	}
	return nil
}

func (f *registFactory) GetFactoryByType(typed string) api.Factory {
	if f, ok := f.typedFactorys[typed]; ok {
		return f
	}
	return nil
}

func (f *registFactory) RegisteredNames() []string {
	var names []string
	for n := range f.namedFactorys {
		names = append(names, n)
	}
	return names
}

func (f *registFactory) RegisteredTypes() []string {
	var types []string
	for n := range f.typedFactorys {
		types = append(types, n)
	}
	return types
}
