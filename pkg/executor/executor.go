package executor

import (
	"context"
	"fmt"

	"github.com/ksraj123/lister-sa/pkg/constants"
	"github.com/ksraj123/lister-sa/pkg/listers"
	v1 "k8s.io/api/core/v1"
	StorageV1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

// Takes in Statefulset PVCs of deletion allowed storage classes as argument and returns a map containing dangling status of given PVCs.
func GetPVCDanlingStatusMap(clientset *kubernetes.Clientset, ctx context.Context, namespace string, statefulsetPvcs []v1.PersistentVolumeClaim) map[string]bool {

	allStatefulsets := listers.ListAllStatefulSets(clientset, ctx, namespace)
	pvcDanglingStatusList := make(map[string]bool)

	// initally mark all openebs statefulset pvcs as dangling
	for _, openebsPvc := range statefulsetPvcs {
		pvcDanglingStatusList[openebsPvc.ObjectMeta.Name] = true
	}

	// iterate over all pods in all statefulsets and mark the pvc they are bound to as not dagling
	for _, statefulset := range allStatefulsets {
		statefulsetLabels := statefulset.Spec.Selector.MatchLabels
		labelSelectorString := labels.SelectorFromSet(statefulsetLabels).String()

		// ToDo: Maybe this should be moved into listers
		pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: labelSelectorString})
		if err != nil {
			fmt.Printf("Could not get Pods of label %v in namespace %v, Error = %v\n", labelSelectorString, namespace, err.Error())
		}

		for _, pod := range pods.Items {
			podVolumes := pod.Spec.Volumes
			for _, volume := range podVolumes {
				if volume.PersistentVolumeClaim != nil {
					pvcDanglingStatusList[volume.PersistentVolumeClaim.ClaimName] = false
				}
			}
		}
	}
	return pvcDanglingStatusList
}

// Kubernetes copies Statefulset selector as labels on statefulset PVCs, this property helps determine if the PVC is a statefulset PVC
// there being no other way to do so once the statefulset itself by virtue of which the PVCs were created gets deleted
// an extra selector needs to be put on the sts whose name can would be the value of "sts-pvc-selector" parameter of storage class and value could be true
func GetStatefulSetPVCs(clientset *kubernetes.Clientset, ctx context.Context, pvcs []v1.PersistentVolumeClaim, openEbsStorageClassesMap map[string]*StorageV1.StorageClass) []v1.PersistentVolumeClaim {
	var statefulsetPvcs []v1.PersistentVolumeClaim
	for _, pvc := range pvcs {
		statefulsetPvcSelector := openEbsStorageClassesMap[*pvc.Spec.StorageClassName].Parameters[constants.STS_PVC_SELECTOR]
		for key, value := range pvc.Labels {
			if key == statefulsetPvcSelector && value == "true" {
				statefulsetPvcs = append(statefulsetPvcs, pvc)
			}
		}
	}
	return statefulsetPvcs
}

func DeleteDanglingPVCs(clientset *kubernetes.Clientset, ctx context.Context, namespace string, openebsPVCsStatus map[string]bool) {
	for pvcName, isDangling := range openebsPVCsStatus {
		if isDangling {
			fmt.Println(pvcName + " is dangling!")
			err := clientset.CoreV1().PersistentVolumeClaims(namespace).Delete(ctx, pvcName, metav1.DeleteOptions{})
			if err == nil {
				fmt.Printf("Dangling PVC %v in namesapce %v deleted successfully\n", pvcName, namespace)
			} else {
				panic(fmt.Sprintf("Error while deleting danling PVC %v in namespace %v", pvcName, namespace))
			}
		}
	}
}
