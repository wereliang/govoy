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

package config

import (
	"io/ioutil"
	"path/filepath"
	"time"

	envoy_config_bootstrap_v3 "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v3"
	"github.com/ghodss/yaml"
	"github.com/wereliang/govoy/pkg/log"

	"github.com/golang/protobuf/jsonpb"
)

type GovoyConfig interface {
	Bootstrap() Bootstrap
}

type govoyConfig struct {
	bootrap Bootstrap
}

func (g *govoyConfig) Bootstrap() Bootstrap {
	return g.bootrap
}

type Bootstrap interface {
	Config() *envoy_config_bootstrap_v3.Bootstrap
	LastUpdated() time.Time
}

type bootstrap struct {
	bs *envoy_config_bootstrap_v3.Bootstrap
	ts time.Time
}

func (bs *bootstrap) Config() *envoy_config_bootstrap_v3.Bootstrap {
	return bs.bs
}

func (bs *bootstrap) LastUpdated() time.Time {
	return bs.ts
}

func LoadBootstrapConfig(path string) (GovoyConfig, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg, bootrap := &govoyConfig{}, &envoy_config_bootstrap_v3.Bootstrap{}

	if yamlFormat(path) {
		content, err = yaml.YAMLToJSON(content)
		if err != nil {
			return nil, err
		}
	}

	if err = jsonpb.UnmarshalString(string(content), bootrap); err != nil {
		log.Error("jsonpb.UnmarshalString error %s", err)
		return nil, err
	}

	if err = bootrap.Validate(); err != nil {
		return nil, err
	}
	cfg.bootrap = &bootstrap{bs: bootrap, ts: time.Now()}
	return cfg, nil
}

func yamlFormat(path string) bool {
	ext := filepath.Ext(path)
	if ext == ".yaml" || ext == ".yml" {
		return true
	}
	return false
}
