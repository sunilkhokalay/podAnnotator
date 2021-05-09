package main

import (
	"context"
	"flag"
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
)

func main() {
	// This program set the annotation annotatedBy=podAnnotator
	// for all the pods created in default namespace

	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		var kubeconfig *string
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
		} else {
			kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
		}
		flag.Parse()

		// use the current context in kubeconfig
		config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
		if err != nil {
			panic(err.Error())
		}
	}

	// creates the clientSet
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	wInf, err := clientSet.CoreV1().Pods("default").Watch(context.TODO(), metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{},
		Watch:    true,
	})

	if err != nil {
		panic(err)
	}

	eventChan := wInf.ResultChan()
	println("Starting pod watcher...")
	for {
		event := <-eventChan
		pod := event.Object.DeepCopyObject().(*v1.Pod)
		println(fmt.Sprintf("The pod \"%s\" in \"%s\" namespace was %s", pod.Name, pod.Namespace, event.Type))
		if event.Type == watch.Added {
			println("Annotating pod...")
			err := annotate(clientSet,pod)
			if err != nil {
				println(fmt.Sprintf("failed to annotate the pod \"%s\" in \"%s\" namespace", pod.Name, pod.Namespace))
			}
		}

	}
}

func annotate(c *kubernetes.Clientset,pod *v1.Pod) error {
	pod.SetAnnotations(map[string]string{"annotatedBy":"podAnnotator"})
	updatedPod, err := c.CoreV1().Pods(pod.Namespace).Update(context.TODO(),pod,metav1.UpdateOptions{})
	if err != nil{
		return err
	}
	println(fmt.Sprintf("Annotations of the pod after update: %v",updatedPod.GetAnnotations()))
	return nil
}
