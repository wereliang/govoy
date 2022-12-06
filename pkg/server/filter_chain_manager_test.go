package server

import (
	"fmt"
	"net"
	"testing"

	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_config_listener_v3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	"github.com/wereliang/govoy/pkg/api"
	"github.com/wereliang/govoy/pkg/log"
	"github.com/wereliang/govoy/pkg/network"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var chains = []*envoy_config_listener_v3.FilterChain{
	{
		FilterChainMatch: &envoy_config_listener_v3.FilterChainMatch{
			DestinationPort: wrapperspb.UInt32(15006),
		},
		Name: "001",
	},
	{
		FilterChainMatch: &envoy_config_listener_v3.FilterChainMatch{
			PrefixRanges: []*envoy_config_core_v3.CidrRange{
				{
					AddressPrefix: "0.0.0.0",
					PrefixLen:     wrapperspb.UInt32(0),
				},
			},
			TransportProtocol:    "tls",
			ApplicationProtocols: []string{"istio-http/1.0", "istio-http/1.1", "istio-h2"},
		},
		Name: "002",
	},
	{
		FilterChainMatch: &envoy_config_listener_v3.FilterChainMatch{
			PrefixRanges: []*envoy_config_core_v3.CidrRange{
				{
					AddressPrefix: "0.0.0.0",
					PrefixLen:     wrapperspb.UInt32(0),
				},
			},
			TransportProtocol:    "raw_buffer",
			ApplicationProtocols: []string{"http/1.1", "h2c"},
		},
		Name: "003",
	},
	{
		FilterChainMatch: &envoy_config_listener_v3.FilterChainMatch{
			PrefixRanges: []*envoy_config_core_v3.CidrRange{
				{
					AddressPrefix: "0.0.0.0",
					PrefixLen:     wrapperspb.UInt32(0),
				},
			},
			TransportProtocol:    "tls",
			ApplicationProtocols: []string{"istio-peer-exchange", "istio"},
		},
		Name: "004",
	},
	{
		FilterChainMatch: &envoy_config_listener_v3.FilterChainMatch{
			PrefixRanges: []*envoy_config_core_v3.CidrRange{
				{
					AddressPrefix: "0.0.0.0",
					PrefixLen:     wrapperspb.UInt32(0),
				},
			},
			TransportProtocol: "raw_buffer",
		},
		Name: "005",
	},
	{
		FilterChainMatch: &envoy_config_listener_v3.FilterChainMatch{
			PrefixRanges: []*envoy_config_core_v3.CidrRange{
				{
					AddressPrefix: "0.0.0.0",
					PrefixLen:     wrapperspb.UInt32(0),
				},
			},
			TransportProtocol: "tls",
		},
		Name: "006",
	},
	{
		FilterChainMatch: &envoy_config_listener_v3.FilterChainMatch{
			DestinationPort:      wrapperspb.UInt32(9080),
			TransportProtocol:    "tls",
			ApplicationProtocols: []string{"istio", "istio-peer-exchange", "istio-http/1.0", "istio-http/1.1", "istio-h2"},
		},
		Name: "007",
	},
	{
		FilterChainMatch: &envoy_config_listener_v3.FilterChainMatch{
			DestinationPort:   wrapperspb.UInt32(9080),
			TransportProtocol: "raw_buffer",
		},
		Name: "008",
	},
	{
		FilterChainMatch: &envoy_config_listener_v3.FilterChainMatch{
			PrefixRanges: []*envoy_config_core_v3.CidrRange{
				{
					AddressPrefix: "192.168.0.0",
					PrefixLen:     wrapperspb.UInt32(24),
				},
			},
			TransportProtocol: "raw_buffer",
		},
		Name: "009",
	},
	{
		FilterChainMatch: &envoy_config_listener_v3.FilterChainMatch{
			PrefixRanges: []*envoy_config_core_v3.CidrRange{
				{
					AddressPrefix: "10.0.20.0",
					PrefixLen:     wrapperspb.UInt32(24),
				},
			},
			TransportProtocol: "tls",
			ServerNames:       []string{"www.qq.com"},
		},
		Name: "010",
	},
	{
		FilterChainMatch: &envoy_config_listener_v3.FilterChainMatch{
			SourcePorts:       []uint32{9080, 10000},
			TransportProtocol: "raw_buffer",
		},
		Name: "011",
	},
}

func loop(k, v, arg interface{}) {
	if filterChain, ok := v.(*envoy_config_listener_v3.FilterChain); ok {
		// fmt.Printf("filterChain:%#v", filterChain)
		fmt.Printf("%v%-20v%-20s\n", arg, k, filterChain.Name)
		return
	}
	arg = fmt.Sprintf("%v%-20v", arg, k)
	t := v.(MatchTable)
	t.Loop(loop, arg)
}

func TestTable(t *testing.T) {
	filterChainManager := newFilterChainManager(chains, nil)
	filterChainManagerImpl := filterChainManager.(*FilterChainManagerImpl)
	format := "%-20s%-20s%-20s%-20s%-20s%-20s%-20s%-20s%-20s%-20s\n"
	fmt.Printf(format, "DestPort", "DestIP", "ServerName", "Transport", "Application", "DirectlyIP", "SourceType", "SourceIP", "SourcePort", "Chain")
	filterChainManagerImpl.destPortsMap.Loop(loop, "")
}

func TestMatchFilter(t *testing.T) {

	filterChainManager := newFilterChainManager(chains, nil)

	tests := []struct {
		name    string
		cs      api.ConnectionContext
		wantNil bool
		want    string
	}{
		{
			"test destinationPort",
			&network.ConnectionContextImpl{
				DestinationPort: 9080,
			},
			false,
			"008",
		},
		{
			"test destinationPort and application protocol",
			&network.ConnectionContextImpl{
				DestinationPort:     9080,
				ApplicationProtocol: "http/1.1",
			},
			false,
			"008",
		},
		{
			"test destination ip",
			&network.ConnectionContextImpl{
				DestinationIP: net.ParseIP("192.168.0.2"),
			},
			false,
			"009",
		},
		{
			"test server name fail",
			&network.ConnectionContextImpl{
				DestinationIP: net.ParseIP("10.0.20.2"),
			},
			true,
			"",
		},
		{
			"test server name fail (transport protocol)",
			&network.ConnectionContextImpl{
				DestinationIP: net.ParseIP("10.0.20.2"),
				ServerName:    "www.qq.com",
			},
			true,
			"",
		},
		{
			"test server name",
			&network.ConnectionContextImpl{
				DestinationIP:     net.ParseIP("10.0.20.2"),
				ServerName:        "www.qq.com",
				TransportProtocol: "tls",
			},
			false,
			"010",
		},
		{
			"test transport/application protocol",
			&network.ConnectionContextImpl{
				ApplicationProtocol: "istio-peer-exchange",
				TransportProtocol:   "tls",
			},
			false,
			"004",
		},
		{
			"test transport/application protocol fail",
			&network.ConnectionContextImpl{
				ApplicationProtocol: "xxxxx",
				TransportProtocol:   "tls",
			},
			false,
			"006",
		},
		{
			"test source port zero",
			&network.ConnectionContextImpl{
				SourcePort: 8000,
			},
			false,
			"005",
		},
		{
			"test source port",
			&network.ConnectionContextImpl{
				SourcePort: 10000,
			},
			false,
			"011",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filterChain := filterChainManager.FindFilterChains(tt.cs)
			if (filterChain == nil) != tt.wantNil {
				t.Errorf("FindFilterChains nil error. want: %t", tt.wantNil)
				return
			}
			if filterChain == nil {
				return
			}
			if filterChain.Name != tt.want {
				t.Errorf("FindFilterChains() = %v, want %v", filterChain.Name, tt.want)
			}
		})
	}
}

func init() {
	log.DefaultLog = log.NewSimpleLogger(log.TraceLevel, true)
}
