# 项目描述
govoy是一个学习项目，通过golang实现envoy最基础的功能，并且可以作为istio的数据面跑在k8s上。一期开发实现了http代理，跑通istio的bookinfo用例。
- 项目参考了envoy代码及文档，部分代码参考了mosn的实现。
- 由于是学习项目，代码尽量精简，尽量做到代码即文档。（对于有精力的同学可以直接学习envoy，mosn的源码）
- 目前代码没有做很好的异常处理和性能优化
- 由于个人精力能力有限，存在理解偏差和错漏的地方，希望交流指正。

# 功能列表
实现了以下功能，足以跑起bookinfo用例
- listener插件：original dst、http inspector
- network插件：http connection manager
- http插件：router
- admin config dump接口
- xds client与istiod进行通信，实现agg stow通信方式
- loadbalancer：smooth roundrobin

# 快速体验
istio的bookinfo用例请参考 https://istio.io/latest/zh/docs/examples/bookinfo/ , 下面介绍如何将istio的数据面替换为govoy并跑起来。

(1) 安装istioctl，参考 https://istio.io/latest/zh/docs/setup/getting-started/

(2) 安装istio，指定govoy作为数据面，另外指定isito版本为1.13.2（其他版本暂未做验证），由于iptables的规则设置依然采用istio的实现，所以proxy_init也需要做指定
```
istioctl install --set profile=minimal \
    --set .values.global.proxy.image="hunter2019/govoy:v1" \
    --set meshConfig.defaultConfig.binaryPath="/govoy/bin/govoy" \
    --set meshConfig.enablePrometheusMerge=false \
    --set .values.pilot.image="docker.io/istio/pilot:1.13.2" \
    --set .values.global.proxy_init.image="docker.io/istio/proxyv2:1.13.2" \
    --set .values.global.proxy.statusPort="0"
```

(3) 部署bookinfo demo，资源文件在examples/bookinfo/bookinfo.yaml下面
```
kubectl create namespace bookinfo
kubectl label namespaces bookinfo istio-injection=enabled
kubectl apply -f examples/bookinfo/bookinfo.yaml -n bookinfo
```

(4) 访问productpage，在上面的bookinfo.yaml里头已经定义了一个nodeport service，可以直接通过node节点进行访问
http://$(NODEIP):30001/productpage，正常可以看到页面，通过刷新页面可以看到review在三种进行切换，底层是govoy做了流量分配。

# TODO List
- 异常处理，资源回收
- 性能优化
- listener、network、http filter插件丰富
- xds delta实现
