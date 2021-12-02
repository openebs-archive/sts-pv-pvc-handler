package setup

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ksraj123/lister-sa/pkg/constants"
	"github.com/ksraj123/lister-sa/tests/generators"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func TestOpenEbsPodRunningState(t *testing.T, clientset *kubernetes.Clientset, ctx context.Context, maxRetry int) {
	LocalPVProvisionerLabelSelector := "openebs.io/component-name=openebs-localpv-provisioner"

	openebsNamespace := constants.OPENEBS_NAMESPACe

	tests := map[string]struct {
		expected int
	}{
		"Testing openebs-localpv-provisioner pod state": {
			expected: 1,
		},
	}

	testPassed := false

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			for i := 0; i < maxRetry; i++ {
				pods, _ := clientset.CoreV1().Pods(openebsNamespace).List(ctx, metav1.ListOptions{
					LabelSelector: LocalPVProvisionerLabelSelector,
				})
				if len(pods.Items) == test.expected {
					testPassed = true
					break
				}
				time.Sleep(5 * time.Second)
			}
			if !testPassed {
				t.Fatalf("openebs-localpv-provisioner pod not running in namespace, %v", openebsNamespace)
			}
		})
	}
}

func TestDanglingPVCDeleted(t *testing.T, clientset *kubernetes.Clientset, ctx context.Context, statefulsetName string, maxRetry int) {
	isSuccessful := false
	var danglingPvcName string
	for i := 0; i < maxRetry; i++ {
		pvcs, err := clientset.CoreV1().PersistentVolumeClaims("default").List(ctx, metav1.ListOptions{})
		if err != nil {
			fmt.Printf("error %s, getting PVCs\n", err.Error())
		}
		count := 0
		for _, pvc := range pvcs.Items {
			if strings.Contains(pvc.Name, statefulsetName) {
				count++
				danglingPvcName = pvc.Name
			}
		}
		if count == 0 {
			isSuccessful = true
			break
		}
		time.Sleep(5 * time.Second)
	}
	if !isSuccessful {
		t.Fatalf("Dangling PVCs %v not deleted by Job", danglingPvcName)
	}
}

func CreateStorageClass(t *testing.T, clientset *kubernetes.Clientset, ctx context.Context, maxRetry int) {
	storageClass := generators.GenerateStorageClass(
		"test-storage-class",
		map[string]string{
			constants.STORAGE_CLASS_ANNOTATION: "true",
			"openebs.io/cas-type":              "local",
			"cas.openebs.io/config": `- name: StorageType
  value: "hostpath"
- name: BasePath
  value: "/var/openebs/local/"`,
		},
		map[string]string{
			constants.STS_PVC_SELECTOR: "sts-pvc",
		},
		"openebs.io/local")
	_, err := clientset.StorageV1().StorageClasses().Create(ctx, storageClass, metav1.CreateOptions{})
	if err != nil {
		panic(err.Error())
	}
	time.Sleep(5 * time.Second)
}

func CreateStatefulSet(t *testing.T, clientset *kubernetes.Clientset, ctx context.Context, name string, maxRetry int) {
	statefulset := generators.GenerateStatefulSet(name, "default", 1, map[string]string{"sts-pvc": "true"}, "test-storage-class")
	_, err := clientset.AppsV1().StatefulSets("default").Create(ctx, statefulset, metav1.CreateOptions{})
	if err != nil {
		panic(err.Error())
	}
	time.Sleep(5 * time.Second)
}
