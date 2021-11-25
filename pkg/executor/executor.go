package executor

import (
	"fmt"

	"context"

	"github.com/ksraj123/lister-sa/pkg/constants"
	"github.com/ksraj123/lister-sa/pkg/danglingpvcs"
	"github.com/ksraj123/lister-sa/pkg/listers"
	"github.com/ksraj123/lister-sa/pkg/statefulsetpvcs"
	"github.com/ksraj123/lister-sa/pkg/utils"

	StorageV1 "k8s.io/api/storage/v1"
	"k8s.io/client-go/kubernetes"
)

// ToDo: check if error in one namespace does not stop execution for others

func Execute(clientset *kubernetes.Clientset, ctx context.Context, namespace string) {
	openEbsStorageClassesMap := make(map[string]*StorageV1.StorageClass)
	provisioners := utils.EnvVarSlice(constants.PROVISIONERS_ENV_VAR)
	openEbsStorageClasses := listers.ListProvisionerStorageClassesWithAnnotation(clientset, ctx, provisioners, constants.STORAGE_CLASS_ANNOTATION)

	if len(openEbsStorageClasses) == 0 {
		panic("No Valid Storage Classes Found")
	}

	for _, storageclass := range openEbsStorageClasses {
		openEbsStorageClassesMap[storageclass.Name] = storageclass
		fmt.Println("OpenEBS storage class with annotation = ", storageclass.Name)
	}
	openebsPvcs := listers.ListPVCsOfStorageClass(clientset, ctx, namespace, openEbsStorageClasses)
	statefulsetPvcs := statefulsetpvcs.GetStatefulSetPVCs(clientset, ctx, openebsPvcs, openEbsStorageClassesMap)
	for _, pvc := range statefulsetPvcs {
		fmt.Println(pvc.Name)
	}

	openebsPVCsStatus := danglingpvcs.GetStatusMap(clientset, ctx, namespace, statefulsetPvcs)
	danglingpvcs.Delete(clientset, ctx, namespace, openebsPVCsStatus)
}
