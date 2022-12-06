package cluster

import (
	"net"
	"strconv"
	"strings"
	"time"

	envoy_config_cluster_v3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	"github.com/wereliang/govoy/pkg/api"
	"github.com/wereliang/govoy/pkg/log"
)

type strictDNSCluster struct {
	*simpleCluster
	stopCh chan struct{}
}

func (c *strictDNSCluster) Close() {
	close(c.stopCh)
}

func newStrictDNSCluster(c *envoy_config_cluster_v3.Cluster) (api.Cluster, error) {
	dnsCluster := &strictDNSCluster{newSimpleCluster(c), make(chan struct{})}
	// save cluster config
	dnsCluster.UpdateHosts(nil)
	dnsCluster.resolveDNS()
	return dnsCluster, nil
}

func (c *strictDNSCluster) resolveDNS() {
	hosts, err := c.getConfigHosts()
	if err != nil {
		log.Error("get config hosts error: %s", err)
		return
	}
	for _, h := range hosts {
		log.Debug("hosts: %#v", h.Address())
	}

	// update dns
	refreshRate := time.Second * 5
	if rate := c.info.Config().GetDnsRefreshRate(); rate != nil {
		refreshRate = time.Duration(rate.GetSeconds()) * time.Second
	}

	go func() {
		c.resolveAndUpdate(hosts)
		t := time.NewTicker(refreshRate)
	LOOP:
		for {
			select {
			case <-c.stopCh:
				t.Stop()
				break LOOP
			case <-t.C:
				c.resolveAndUpdate(hosts)
			}
		}
		log.Info("cluster %s close.", c.info.Config().Name)
	}()
}

func (c *strictDNSCluster) resolveAndUpdate(hosts api.HostSet) {
	var destHosts api.HostSet
	for _, h := range hosts {
		var (
			name string
			port int
		)
		addrs := strings.Split(h.Address().String(), ":")
		if len(addrs) == 2 {
			name = addrs[0]
			port, _ = strconv.Atoi(addrs[1])
		} else {
			log.Error("invalid host [%s]", h)
			return
		}

		ips, err := net.LookupIP(name)
		if err != nil {
			log.Error("Could not get IPs: [%s]", name)
			return
		}
		for _, ip := range ips {
			tcphost := api.NewHost(&net.TCPAddr{IP: ip, Port: port})
			tcphost.SetWeight(h.Weight())
			destHosts = append(destHosts, tcphost)
			// log.Trace("resolve dns:%s ip:%s", name, ip)
		}
	}
	if destHosts != nil {
		c.UpdateHosts(destHosts)
	}
}

func init() {
	registCluster(api.Cluster_Strict_DNS,
		func(c *envoy_config_cluster_v3.Cluster) (api.Cluster, error) {
			return newStrictDNSCluster(c)
		})
}
