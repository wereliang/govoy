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

package lb

import (
	"sync"

	"github.com/wereliang/govoy/pkg/api"
)

type rbItem struct {
	host api.Host
	step int32
}

type SmoothRoundRobin struct {
	sync.RWMutex
	total int32
	items []*rbItem
}

func (lb *SmoothRoundRobin) Select(api.LoadBalancerContext) api.Host {
	lb.RLock()
	defer lb.RUnlock()

	maxIndex := 0
	for i := 0; i < len(lb.items); i++ {
		item := lb.items[i]
		item.step += int32(item.host.Weight())
		if item.step > lb.items[maxIndex].step {
			maxIndex = i
		}
		// log.Trace("host:%s weight:%d", item.host.Address().String(), item.step)
	}
	lb.items[maxIndex].step -= lb.total
	return lb.items[maxIndex].host
}

func init() {
	registLoadBalancer(api.Round_Robin, func(hosts api.HostSet) api.LoadBalancer {
		rb := &SmoothRoundRobin{}
		for _, h := range hosts {
			rb.total += int32(h.Weight())
			rb.items = append(rb.items, &rbItem{host: h})
		}
		return rb
	})
}
