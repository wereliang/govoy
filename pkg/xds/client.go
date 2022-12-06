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

package xds

import (
	xds_v3 "github.com/wzshiming/xds/v3"
)

type XDSClient interface {
	SendLDS([]string) error
	SendRDS([]string) error
	SendCDS([]string) error
	SendEDS([]string) error
}

type xdsClient struct {
	cli *xds_v3.Client
}

func (x *xdsClient) Client() *xds_v3.Client {
	return x.cli
}

func (x *xdsClient) SendLDS(rsc []string) error {
	return x.send(xds_v3.ListenerType, rsc)
}

func (x *xdsClient) SendRDS(rsc []string) error {
	return x.send(xds_v3.RouteType, rsc)
}

func (x *xdsClient) SendCDS(rsc []string) error {
	return x.send(xds_v3.ClusterType, rsc)
}

func (x *xdsClient) SendEDS(rsc []string) error {
	return x.send(xds_v3.EndpointType, rsc)
}

func (x *xdsClient) send(typeurl string, rsc []string) error {
	return x.cli.SendRsc(typeurl, rsc)
}
