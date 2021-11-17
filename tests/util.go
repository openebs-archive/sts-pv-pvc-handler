package tests

import (
	"context"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	maxRetry = 90
)

func GetPodRunningCountEventually(namespace, lselector string, expectedPodCount int, clientset *kubernetes.Clientset) int {
	var podCount int
	for i := 0; i < maxRetry; i++ {
		pods, _ := clientset.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{
			LabelSelector: lselector,
		})
		podCount := len(pods.Items)
		if podCount == expectedPodCount {
			return podCount
		}
		time.Sleep(5 * time.Second)
	}
	return podCount
}
