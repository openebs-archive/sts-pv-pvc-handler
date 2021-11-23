package listers

import (
	"context"
	"fmt"

	AppsV1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	StorageV1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func ListAllStatefulSets(clientset *kubernetes.Clientset, ctx context.Context, namespace string) []AppsV1.StatefulSet {
	allStatefulsets, errAllSts := clientset.AppsV1().StatefulSets(namespace).List(ctx, metav1.ListOptions{})
	if errAllSts != nil {
		fmt.Printf("error %s, getting PVCs\n", errAllSts.Error())
	}
	return allStatefulsets.Items
}

func ListAllStorageClasses(clientset *kubernetes.Clientset, ctx context.Context) []StorageV1.StorageClass {
	allSc, errSc := clientset.StorageV1().StorageClasses().List(ctx, metav1.ListOptions{})
	if errSc != nil {
		fmt.Printf("error %s, getting Storage Classes\n", errSc.Error())
	}
	return allSc.Items
}

func ListAllPersistentVolumeClaims(clientset *kubernetes.Clientset, ctx context.Context, namespace string) []v1.PersistentVolumeClaim {
	allPvcs, errPVC := clientset.CoreV1().PersistentVolumeClaims(namespace).List(ctx, metav1.ListOptions{})
	if errPVC != nil {
		fmt.Printf("error %s, getting PVCs\n", errPVC.Error())
	}
	return allPvcs.Items
}

func ListPVCsOfStorageClass(clientset *kubernetes.Clientset, ctx context.Context, namespace string, storageclasses []*StorageV1.StorageClass) []v1.PersistentVolumeClaim {
	allPvcs := ListAllPersistentVolumeClaims(clientset, ctx, namespace)
	var openebsPvcs []v1.PersistentVolumeClaim
	for _, pvc := range allPvcs {
		pvcStorageClassName := *pvc.Spec.StorageClassName
		for _, openEbsStorageClass := range storageclasses {
			if pvcStorageClassName == openEbsStorageClass.Name {
				openebsPvcs = append(openebsPvcs, pvc)
			}

		}
	}
	return openebsPvcs
}

// retuns list of storage classes that have an provisioner among the provided provisioners and have the annotation set
func ListProvisionerStorageClassesWithAnnotation(clientset *kubernetes.Clientset, ctx context.Context, provisioners []string, annotation string) []*StorageV1.StorageClass {
	allSc := ListAllStorageClasses(clientset, ctx)
	var openEbsStorageClasses []*StorageV1.StorageClass
	for _, storageclass := range allSc {
		for _, openEbsProvisioner := range provisioners {
			if storageclass.Provisioner == openEbsProvisioner && storageclass.Annotations[annotation] == "true" {
				openEbsStorageClasses = append(openEbsStorageClasses, &storageclass)
			}
		}
	}
	return openEbsStorageClasses
}
