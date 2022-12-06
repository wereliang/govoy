#!/bin/bash

set -ex

echo "do iptables"

# outbound
iptables -t nat -N ISTIO_REDIRECT
iptables -t nat -A ISTIO_REDIRECT -p tcp -j REDIRECT --to-port 15001

iptables -t nat -N ISTIO_OUTPUT
iptables -t nat -A OUTPUT -p tcp -j ISTIO_OUTPUT

iptables -t nat -A ISTIO_OUTPUT -o lo -s 127.0.0.6/32 -j RETURN 
#iptables -t nat -A ISTIO_OUTPUT -o lo ! -d 127.0.0.1/32 -j ISTIO_IN_REDIRECT 
iptables -t nat -A ISTIO_OUTPUT -m owner --uid-owner 1337 -j RETURN 
iptables -t nat -A ISTIO_OUTPUT -m owner --gid-owner 1337 -j RETURN 
iptables -t nat -A ISTIO_OUTPUT -d 127.0.0.1/32 -j RETURN   
iptables -t nat -A ISTIO_OUTPUT -j ISTIO_REDIRECT 

# inbound
iptables -t nat -N ISTIO_IN_REDIRECT
iptables -t nat -A ISTIO_IN_REDIRECT -p tcp -j REDIRECT --to-port 15006

iptables -t nat -N ISTIO_INBOUND
iptables -t nat -A PREROUTING -p tcp -j ISTIO_INBOUND

iptables -t nat -A ISTIO_INBOUND -p tcp --dport 22 -j RETURN
iptables -t nat -A ISTIO_INBOUND -p tcp --dport 15008 -j RETURN
iptables -t nat -A ISTIO_INBOUND -p tcp --dport 15090 -j RETURN
iptables -t nat -A ISTIO_INBOUND -p tcp --dport 15021 -j RETURN
iptables -t nat -A ISTIO_INBOUND -p tcp --dport 15020 -j RETURN
iptables -t nat -A ISTIO_INBOUND -p tcp -j ISTIO_IN_REDIRECT

iptables -t nat -A ISTIO_OUTPUT -m owner --uid-owner 0 -j RETURN 