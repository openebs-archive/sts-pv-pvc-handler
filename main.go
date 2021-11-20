package main

// package - pvcClearner

import (
	"context"
	"fmt"

	"github.com/ksraj123/lister-sa/pkg/executor"
	"github.com/ksraj123/lister-sa/pkg/listers"
	"github.com/ksraj123/lister-sa/pkg/utils"
	StorageV1 "k8s.io/api/storage/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	clientset *kubernetes.Clientset
	ctx       context.Context
)

const (
	PROVISIONERS_ENV_VAR     = "PROVISIONERS"
	STORAGE_CLASS_ANNOTATION = "openebs.io/delete-dangling-pvc"
)

func init() {
	config, err := rest.InClusterConfig()
	if err != nil {
		fmt.Printf("error %s, getting inclusterconfig", err.Error())
	}
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("error %s, creating clientset\n", err.Error())
	}
	ctx = context.Background()
}

func main() {
	openEbsStorageClassesMap := make(map[string]StorageV1.StorageClass)
	provisioners := utils.EnvVarSlice(PROVISIONERS_ENV_VAR)
	openEbsStorageClasses := listers.ListProvisionerStorageClassesWithAnnotation(clientset, ctx, provisioners, STORAGE_CLASS_ANNOTATION)

	if len(openEbsStorageClasses) == 0 {
		panic("No Valid Storage Classes Found")
	}

	for _, storageclass := range openEbsStorageClasses {
		openEbsStorageClassesMap[storageclass.Name] = storageclass
		fmt.Println(fmt.Sprintf("OpenEBS storage class with annotation = %v", storageclass.Name))
	}
	openebsPvcs := listers.ListPVCsOfStorageClass(clientset, ctx, openEbsStorageClasses)
	statefulsetPvcs := executor.GetStatefulSetPVCs(clientset, ctx, openebsPvcs, openEbsStorageClassesMap)
	for _, pvc := range statefulsetPvcs {
		fmt.Println(pvc.Name)
	}

	openebsPVCsStatus := executor.GetAllUnboundedPVCs(clientset, ctx, statefulsetPvcs)
	executor.DeleteDanglingPVCs(openebsPVCsStatus)
}
