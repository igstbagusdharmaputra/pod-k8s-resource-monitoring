package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/bndr/gotabulate"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

var (
	kubeConfig string
	namespace  string
)

func init() {
	flag.StringVar(&kubeConfig, "kubeConfig", "./config", "Give the file config (optional)")
	flag.StringVar(&namespace, "namespace", "", "*Give the namespace name (default namespace)")
}

func loadKubeconfig() (*rest.Config, error) {
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
func main() {
	flag.Parse()

	fmt.Println("Loading kube config")
	config, err := loadKubeconfig()
	if err != nil {
		log.Fatal(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}
	clientmetrics, err := metricsv.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}
	podsmetrics, err := clientmetrics.MetricsV1beta1().PodMetricses(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}
	currentTime := time.Now()
	timeGetInfo := currentTime.Format("2006-01-02 15:04:05")
	rows := make([][]interface{}, len(podsmetrics.Items))
	cpuUsage := make([]interface{}, len(podsmetrics.Items))
	memoryUsage := make([]interface{}, len(podsmetrics.Items))

	fmt.Printf("Timestamp : %s\n", timeGetInfo)
	for index, pod := range podsmetrics.Items {
		podContainers := pod.Containers
		for _, c := range podContainers {
			cpuUsageData := c.Usage.Cpu().MilliValue()

			memoryUsageData := c.Usage.Memory().AsDec().String()
			memoryUsageFloat, _ := strconv.ParseFloat(memoryUsageData, 64)

			cpuUsage[index] = fmt.Sprintf("%v%s", cpuUsageData, "m")
			memoryUsage[index] = fmt.Sprintf("%v%s", memoryUsageFloat/1024/1024, "Mi")

		}
	}

	for index, pod := range pods.Items {
		for _, c := range pod.Spec.Containers {
			reqCpu := c.Resources.Requests.Cpu()
			reqMem := c.Resources.Requests.Memory()
			rows[index] = []interface{}{index + 1, pod.Name, reqCpu, cpuUsage[index], reqMem, memoryUsage[index]}
		}
	}

	table := gotabulate.Create(rows)
	table.SetHeaders([]string{"No", "Pod Name", "CPU Request", "CPU Usage", "Memory Request", "Memory Usage"})
	table.SetMaxCellSize(20)
	table.SetWrapStrings(true)
	table.SetAlign("center")
	fmt.Println(table.Render("grid"))

}
