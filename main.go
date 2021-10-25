package main

import (
	"context"
	"fmt"
	"reflect"

	v1 "k8s.io/api/core/v1"
	// "k8s.io/kubernetes/pkg/controller/volume/ephemeral"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	// "k8s.io/kubernetes/pkg/controller/volume/common"
)

type OpenebsPvcStatus struct {
	isDangling bool
	labels     map[string]string
}

func main() {

	openEbsStorageClassesNames := [3]string{"openebs-device", "openebs-hostpath"}

	config, err := rest.InClusterConfig()
	if err != nil {
		fmt.Printf("error %s, getting inclusterconfig", err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		// handle error
		fmt.Printf("error %s, creating clientset\n", err.Error())
	}

	ctx := context.Background()

	allPvcs, errPVC := clientset.CoreV1().PersistentVolumeClaims("default").List(ctx, metav1.ListOptions{})

	if errPVC != nil {
		fmt.Printf("error %s, getting PVCs\n", errPVC.Error())
	}

	var openebsPvcs []*v1.PersistentVolumeClaim

	for _, pvc := range allPvcs.Items {
		pvcdetails, err := clientset.CoreV1().PersistentVolumeClaims("default").Get(ctx, pvc.Name, metav1.GetOptions{})
		if err != nil {
			fmt.Println(err)
		} else {
			pvcStorageClassName := *pvcdetails.Spec.StorageClassName
			for _, openEbsStorageClass := range openEbsStorageClassesNames {
				if pvcStorageClassName == openEbsStorageClass {
					openebsPvcs = append(openebsPvcs, pvcdetails)
				}

			}
		}
	}

	fmt.Println("Printing Filtered PVCs")
	for _, openebsPvc := range openebsPvcs {
		fmt.Println(openebsPvc.ObjectMeta.Name)
		fmt.Println(openebsPvc.ObjectMeta.Labels)
	}

	openebsPVCsStatus := make(map[string]OpenebsPvcStatus)

	for _, openebsPvc := range openebsPvcs {
		openebsPVCsStatus[openebsPvc.ObjectMeta.Name] = OpenebsPvcStatus{isDangling: true, labels: openebsPvc.ObjectMeta.Labels}
	}

	allStatefulsets, errAllSts := clientset.AppsV1().StatefulSets("default").List(ctx, metav1.ListOptions{})
	if errAllSts != nil {
		fmt.Printf("error %s, getting PVCs\n", errAllSts.Error())
	}

	// Objective - Get the pods under a statefulset
	// Objective - Get what PVC a Pod is bound to
	fmt.Println("Printing All Stateful Sets")
	for _, statefulset := range allStatefulsets.Items {
		labels := statefulset.Spec.Selector.MatchLabels
		labelsKeys := reflect.ValueOf(labels).MapKeys()
		key := labelsKeys[0].Interface().(string)
		selector := key + "=" + labels[key]

		fmt.Println(selector)

		pods, err := clientset.CoreV1().Pods("default").List(ctx, metav1.ListOptions{LabelSelector: selector})

		if err != nil {
			fmt.Printf("error %s, getting PVCs\n", err.Error())
		}

		for _, pod := range pods.Items {
			fmt.Println("Printing Pod ")
			fmt.Println(pod.ObjectMeta.Name)
			podVolumes := pod.Spec.Volumes

			for _, volume := range podVolumes {
				if volume.PersistentVolumeClaim != nil {
					fmt.Println(volume.PersistentVolumeClaim.ClaimName)
					entry, found := openebsPVCsStatus[volume.PersistentVolumeClaim.ClaimName]
					if found {
						entry.isDangling = false
						openebsPVCsStatus[volume.PersistentVolumeClaim.ClaimName] = entry
					}
				}
			}
		}
	}

	for pvc, status := range openebsPVCsStatus {
		if status.isDangling {
			fmt.Println(pvc + " is dangling!")
			err := clientset.CoreV1().PersistentVolumeClaims("default").Delete(ctx, pvc, metav1.DeleteOptions{})

			if err == nil {
				fmt.Printf("Dangling PVC %s deleted successfully\n", pvc)
			}
		}
	}

	// fmt.Println("Printing Pod Pvc Index")
	// fmt.Println(common.PodPVCIndex)
}
