package executor

import (
	"context"
	"fmt"
	"reflect"

	"github.com/ksraj123/lister-sa/pkg/listers"
	v1 "k8s.io/api/core/v1"
	StorageV1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type OpenebsPvcStatus struct {
	IsDangling bool
	Labels     map[string]string
}

func GetAllUnboundedPVCs(clientset *kubernetes.Clientset, ctx context.Context, statefulsetPvcs []v1.PersistentVolumeClaim) map[string]OpenebsPvcStatus {

	allStatefulsets := listers.ListAllStatefulSets(clientset, ctx)
	openebsPVCsStatus := make(map[string]OpenebsPvcStatus)

	for _, openebsPvc := range statefulsetPvcs {
		openebsPVCsStatus[openebsPvc.ObjectMeta.Name] = OpenebsPvcStatus{IsDangling: true, Labels: openebsPvc.ObjectMeta.Labels}
	}

	// iterate over all pods in all statefulsets and mark the pvc they are bound to
	for _, statefulset := range allStatefulsets {
		labels := statefulset.Spec.Selector.MatchLabels
		labelsKeys := reflect.ValueOf(labels).MapKeys()
		key := labelsKeys[0].Interface().(string)
		selector := key + "=" + labels[key]

		pods, err := clientset.CoreV1().Pods("default").List(ctx, metav1.ListOptions{LabelSelector: selector})

		if err != nil {
			fmt.Printf("error %s, getting PVCs\n", err.Error())
		}

		for _, pod := range pods.Items {
			podVolumes := pod.Spec.Volumes

			for _, volume := range podVolumes {
				if volume.PersistentVolumeClaim != nil {
					entry, found := openebsPVCsStatus[volume.PersistentVolumeClaim.ClaimName]
					if found {
						entry.IsDangling = false
						openebsPVCsStatus[volume.PersistentVolumeClaim.ClaimName] = entry
					}
				}
			}
		}
	}
	return openebsPVCsStatus
}

// Gets statefulset PVCs from list of OpenEBS PVCs with annotation provided
func GetStatefulSetPVCs(clientset *kubernetes.Clientset, ctx context.Context, pvcs []v1.PersistentVolumeClaim, openEbsStorageClassesMap map[string]StorageV1.StorageClass) []v1.PersistentVolumeClaim {
	var statefulsetPvcs []v1.PersistentVolumeClaim
	for _, pvc := range pvcs {
		statefulsetPvcSelector := openEbsStorageClassesMap[*pvc.Spec.StorageClassName].Parameters["sts-pvc-selector"]
		for key, value := range pvc.Labels {
			if key == "sts-pvc-selector" && value == statefulsetPvcSelector {
				statefulsetPvcs = append(statefulsetPvcs, pvc)
			}
		}
	}
	return statefulsetPvcs
}

func DeleteDanglingPVCs(openebsPVCsStatus map[string]OpenebsPvcStatus) {
	for pvc, status := range openebsPVCsStatus {
		if status.IsDangling {
			fmt.Println(pvc + " is dangling!")
			/*
				err := clientset.CoreV1().PersistentVolumeClaims("default").Delete(ctx, pvc, metav1.DeleteOptions{})

				if err == nil {
					fmt.Printf("Dangling PVC %s deleted successfully\n", pvc)
				}
			*/
		}
	}
}
