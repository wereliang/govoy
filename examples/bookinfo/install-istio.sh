#!/bin/bash

istioctl install --set profile=minimal \
    --set .values.global.proxy.image="hunter2019/govoy:v1" \
    --set meshConfig.defaultConfig.binaryPath="/govoy/bin/govoy" \
    --set meshConfig.enablePrometheusMerge=false \
    --set .values.pilot.image="docker.io/istio/pilot:1.13.2" \
    --set .values.global.proxy_init.image="docker.io/istio/proxyv2:1.13.2" \
    --set .values.global.proxy.statusPort="0"