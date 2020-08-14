# eci-pod-exporter

### Features
eci pod 的 resource 配置信息是添加在 annotation 里的，无法通过 Prometheus 直接查询。通过 client-go 定时抓取 eci pod 的 quota 信息转换成 metric，从而可以用 Prometheus 实时监控。



### Metrics
| Metric name | Metric type | Labels/tags |
|-------------|-------------|-------------|
|eci_pod_request_cpu|Gauge|`pod_name`;`pod_namespace`|

### Getting start
```shell script
$ git clone https://git.qutoutiao.net/paas-k8s/eci-pod-exporter.git
```
```shell script
$ make image
$ kubectl apply -f deploy/deploy.yaml
```
