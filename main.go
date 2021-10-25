package main

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

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
		fmt.Printf("error %s, getting PVCs\n", err.Error())
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
	for _, openebsPVC := range openebsPvcs {
		fmt.Println(openebsPVC.ObjectMeta.Name)
		fmt.Println(openebsPVC.ObjectMeta.Labels)
	}

}
