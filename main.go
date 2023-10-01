package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func streamLogs(ctx context.Context, clientset *kubernetes.Clientset, namespace, podName string) {
	// Exclude the log-streamer pod itself to prevent a log loop
	if podName == "log-streamer" {
		return
	}

	podLogOpts := v1.PodLogOptions{
		Follow: true,
	}
	req := clientset.CoreV1().Pods(namespace).GetLogs(podName, &podLogOpts)
	podLogs, err := req.Stream(ctx)
	if err != nil {
		log.Println(err)
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
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Println("Error reading config:", err)
		return nil, nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Println("Error creating clientset:", err)
		return nil, nil, err
	}

	podList, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Println("Error fetching pods:", err)
		return nil, nil, err
	}
	pods := append([]v1.Pod{}, podList.Items...)
	return pods, clientset, err
}

func main() {
	namespace := "ide"
	ctx := context.Background()

	for {
		pods, clientset, err := fetchPods(ctx, namespace)
		if err != nil {
			log.Println("Error fetching pods:", err)
			time.Sleep(10 * time.Second)
			continue
		}

		var wg sync.WaitGroup
		for _, pod := range pods {
			wg.Add(1)
			go func(pod v1.Pod) {
				defer wg.Done()
				streamLogs(ctx, clientset, namespace, pod.Name)
			}(pod)
		}

		wg.Wait()
		log.Println("All log streaming goroutines completed, restarting in 10 seconds...")
		time.Sleep(10 * time.Second)
	}
}
