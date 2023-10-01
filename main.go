package main

import (
	"log"
	"os"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

func main() {
	stopCh := make(chan struct{})
	defer close(stopCh)

	// Open the file for logging
	file, err := os.OpenFile("out.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer file.Close()

	// Set the log output to the file
	log.SetOutput(file)

	clientset := createClientSet()
	controller := createController(clientset, stopCh)

	go controller.Run(stopCh)
	select {} // Block forever
}

func createClientSet() *kubernetes.Clientset {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Error creating config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating clientset: %v", err)
	}

	return clientset
}

func createController(clientset *kubernetes.Clientset, stopCh chan struct{}) cache.Controller {
	watchlist := cache.NewListWatchFromClient(
		clientset.CoreV1().RESTClient(),
		"pods",
		v1.NamespaceAll,
		fields.Everything(),
	)

	_, controller := cache.NewInformer(
		watchlist,
		&v1.Pod{},
		0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				log.Println("Pod added")
			},
			DeleteFunc: func(obj interface{}) {
				log.Println("Pod deleted")
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				pod := newObj.(*v1.Pod)
				log.Printf("Pod updated: %s", pod.Name)
				for _, containerStatus := range pod.Status.ContainerStatuses {
					if containerStatus.LastTerminationState.Terminated != nil && containerStatus.LastTerminationState.Terminated.Reason == "OOMKilled" {
						log.Printf("Pod %s in namespace %s was OOMKilled", pod.Name, pod.Namespace)
					}
				}
			},
		},
	)

	return controller
}
