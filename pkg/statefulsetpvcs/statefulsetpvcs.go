package statefulsetpvcs

import (
	"context"

	"github.com/ksraj123/lister-sa/pkg/constants"
	v1 "k8s.io/api/core/v1"
	StorageV1 "k8s.io/api/storage/v1"
	"k8s.io/client-go/kubernetes"
)

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
