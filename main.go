package main

import (
	"bufio"
	"context"
	"fmt"
	
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sync"
)

func streamLogs(ctx context.Context,clientset *kubernetes.Clientset, namespace, podName string, wg *sync.WaitGroup) {
	defer wg.Done()

	podLogOpts := v1.PodLogOptions{
		Follow: true,
	}
	req := clientset.CoreV1().Pods(namespace).GetLogs(podName, &podLogOpts)
	podLogs, err := req.Stream(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer podLogs.Close()

	scanner := bufio.NewScanner(podLogs)
	for scanner.Scan() {
		logLine := scanner.Text()
		fmt.Printf("log from pod %s: %s\n", podName, logLine)
	}
}

func fetchPods(ctx context.Context, namespace string) ([]v1.Pod, *kubernetes.Clientset, error) {
	var config *rest.Config
	var err error

	config, err = rest.InClusterConfig()
	if err != nil {
		fmt.Println("Error reading config:", err)
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Println("Error creating clientset:", err)
		panic(err)
	}

	podList, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Println("Error fetching pods:", err)
		panic(err)
	}
	pods := make([]v1.Pod, 0)
	for _, pod := range podList.Items {
		pods = append(pods, pod)
	}
	return pods, clientset, err
}

func main() {
	namespace := "ide-local"
	ctx := context.Background()

	pods, clientset, err := fetchPods(ctx, namespace)
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	for _, pod := range pods {
		wg.Add(1)
		go streamLogs(ctx, clientset, namespace, pod.Name, &wg)
	}

	wg.Wait()
}
