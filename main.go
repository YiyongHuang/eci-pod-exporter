package main

import (
	"flag"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/YiyongHuang/eci-pod-exporter/utils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
)

var (
	eciRequestCpu *prometheus.GaugeVec

	kubeClient *kubernetes.Clientset

	config = struct {
		Kubeconfig      string
		SrvAddr         string
		AdditionalLabel string
		CollectInterval int
	}{}

	baseLabels = []string{
		"pod_name",
		"pod_namespace",
	}
)

func init() {
	flagSet := flag.CommandLine
	klog.InitFlags(flagSet)

	flagSet.StringVar(&config.Kubeconfig, "kubeconfig", "", "kubeconfig path")
	flagSet.StringVar(&config.SrvAddr, "listen-addr", ":9099", "server listen address")
	flagSet.StringVar(&config.AdditionalLabel, "additional-label", "", "additional label to append on metrics")
	flagSet.IntVar(&config.CollectInterval, "collect-interval", 30, "collect hpa interval")

	flagSet.Parse(os.Args[1:])
}

func initCollectors() []prometheus.Collector {
	if config.AdditionalLabel != "" {
		baseLabels = append(baseLabels, config.AdditionalLabel)
	}

	eciRequestCpu = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "eci_pod_request_cpu",
			Help: "eci pod request cpu.",
		},
		baseLabels,
	)

	return []prometheus.Collector{
		eciRequestCpu,
	}

}

func getEciPodListV1() ([]corev1.Pod, error) {
	var err error
	if kubeClient == nil {
		kubeClient, err = utils.NewClientset(config.Kubeconfig)
		if err != nil {
			return nil, err
		}
	}

	podFieldSelector := labels.SelectorFromSet(map[string]string{"spec.nodeName": "virtual-kubelet"})
	out, err := kubeClient.CoreV1().Pods(corev1.NamespaceAll).List(metav1.ListOptions{FieldSelector: podFieldSelector.String()})
	if err != nil {
		return nil, err
	}

	return out.Items, err
}

func collectorV1(pod []corev1.Pod, additionalLabel string) {
	resetMetrics()

	for _, a := range pod {
		baseLabel := prometheus.Labels{
			"pod_name":      a.ObjectMeta.Name,
			"pod_namespace": a.ObjectMeta.Namespace,
		}

		if additionalLabel != "" {
			baseLabel[additionalLabel] = a.Labels[additionalLabel]
		}

		if a.ObjectMeta.Annotations["k8s.aliyun.com/eci-instance-cpu"] != "" {
			v, err := strconv.ParseFloat(a.ObjectMeta.Annotations["k8s.aliyun.com/eci-instance-cpu"], 32)
			if err != nil {
				klog.Error("annotation parse float err:", err)
			} else {
				eciRequestCpu.With(baseLabel).Set(v)
			}
		}
	}
}

func resetMetrics() {
	eciRequestCpu.Reset()
}

func main() {
	collectors := initCollectors()
	prometheus.MustRegister(collectors...)

	klog.Info("start eci pod exporter...")

	go func() {
		for {
			podV1, err := getEciPodListV1()
			if err != nil {
				klog.Error("list eci pod v1 err:", err)
				continue
			}

			collectorV1(podV1, config.AdditionalLabel)

			time.Sleep(time.Duration(config.CollectInterval) * time.Second)
		}
	}()

	http.Handle("/metrics", promhttp.Handler())

	klog.Fatal(http.ListenAndServe(config.SrvAddr, nil))
}
