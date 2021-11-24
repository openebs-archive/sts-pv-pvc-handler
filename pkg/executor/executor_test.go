package executor

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/ksraj123/lister-sa/pkg/constants"
	"github.com/ksraj123/lister-sa/tests/generators"
	AppsV1 "k8s.io/api/apps/v1"
	CoreV1 "k8s.io/api/core/v1"
	StorageV1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	envtest "sigs.k8s.io/controller-runtime/pkg/envtest"
)

func startCluster() (*kubernetes.Clientset, *envtest.Environment) {
	testenv := &envtest.Environment{}
	cfg, err := testenv.Start()
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		panic(err.Error())
	}
	return clientset, testenv
}

func stopCluster(testenv *envtest.Environment) {
	testenv.Stop()
}

func TestGetPVCDanlingStatusMap(t *testing.T) {
	ctx := context.Background()
	clientSet, clusterTestEnv := startCluster()
	defer stopCluster(clusterTestEnv)

	unboundPVC := generators.GeneratePersistentVolumeClaim(fmt.Sprintf("test-pvc-unbound-%v", rand.Int()), constants.TEST_NAMESPACE, "test-storage-class")
	statefulsetReplicas := 1
	statefulset := generators.GenerateStatefulSet(fmt.Sprintf("test-sts-%v", rand.Int()), constants.TEST_NAMESPACE, int32(statefulsetReplicas), map[string]string{"role": "test"}, "standard")

	tests := map[string]struct {
		initFunc func(*kubernetes.Clientset)
		expected *CoreV1.PersistentVolumeClaim
	}{
		"Testing Dangling State of PVCs": {
			initFunc: func(clientset *kubernetes.Clientset) {
				_, err := clientset.CoreV1().PersistentVolumeClaims(constants.TEST_NAMESPACE).Create(ctx, unboundPVC, metav1.CreateOptions{})
				if err != nil {
					panic(err.Error())
				}
				_, err = clientset.AppsV1().StatefulSets(constants.TEST_NAMESPACE).Create(ctx, statefulset, metav1.CreateOptions{})
				if err != nil {
					panic(err.Error())
				}
			},
			expected: unboundPVC,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			test.initFunc(clientSet)
			persistentvolumeClaims, _ := clientSet.CoreV1().PersistentVolumeClaims(constants.TEST_NAMESPACE).List(ctx, metav1.ListOptions{})
			danglingStatusMap := GetPVCDanlingStatusMap(clientSet, ctx, constants.TEST_NAMESPACE, persistentvolumeClaims.Items)
			testFailed := false
			if danglingStatusMap[test.expected.Name] != true {
				testFailed = true
			}
			for pvcName, _ := range danglingStatusMap {
				if strings.Contains(pvcName, statefulset.Name) && danglingStatusMap[pvcName] == true {
					testFailed = true
					break
				}
			}
			if testFailed {
				t.Fatalf("Dangling PVCs determined incorrectly, %v", danglingStatusMap)
			}
		})
	}

	err := clientSet.AppsV1().StatefulSets(constants.TEST_NAMESPACE).Delete(ctx, statefulset.Name, metav1.DeleteOptions{})
	if err != nil {
		panic(err.Error())
	}
	for i := 0; i < statefulsetReplicas; i++ {
		err := clientSet.CoreV1().PersistentVolumeClaims(constants.TEST_NAMESPACE).Delete(ctx, fmt.Sprintf("pvc-%v-%v", statefulset.Name, i), metav1.DeleteOptions{})
		if err != nil {
			panic(err.Error())
		}
	}

	err = clientSet.CoreV1().PersistentVolumeClaims(constants.TEST_NAMESPACE).Delete(ctx, unboundPVC.Name, metav1.DeleteOptions{})
	if err != nil {
		panic(err.Error())
	}
}

func TestGetStatefulSetPVCs(t *testing.T) {
	ctx := context.Background()
	clientSet, clusterTestEnv := startCluster()
	defer stopCluster(clusterTestEnv)

	storageClass := generators.GenerateStorageClass(fmt.Sprintf("test-sc-%v", rand.Int()), nil, map[string]string{constants.STS_PVC_SELECTOR: "some-selector-in-sts"}, "test-provisioner")
	statefulsetReplicas := 1
	statefulset := generators.GenerateStatefulSet(fmt.Sprintf("test-sts-%v", rand.Int()), constants.TEST_NAMESPACE, int32(statefulsetReplicas), map[string]string{storageClass.Parameters[constants.STS_PVC_SELECTOR]: "true"}, storageClass.Name)
	tests := map[string]struct {
		initFunc func(*kubernetes.Clientset)
		expected *AppsV1.StatefulSet
	}{
		"Testing Listing of Statefulset PVCs": {
			initFunc: func(clientset *kubernetes.Clientset) {
				_, err := clientset.StorageV1().StorageClasses().Create(ctx, storageClass, metav1.CreateOptions{})
				if err != nil {
					panic(err.Error())
				}
				_, err = clientset.AppsV1().StatefulSets(constants.TEST_NAMESPACE).Create(ctx, statefulset, metav1.CreateOptions{})
				if err != nil {
					panic(err.Error())
				}
			},
			expected: statefulset,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			test.initFunc(clientSet)
			// deleting the statefulset by virtue of which the PVC was created
			err := clientSet.AppsV1().StatefulSets(constants.TEST_NAMESPACE).Delete(ctx, statefulset.Name, metav1.DeleteOptions{})
			if err != nil {
				panic(err.Error())
			}
			// listing all PVCs
			persistentvolumeClaims, err := clientSet.CoreV1().PersistentVolumeClaims(constants.TEST_NAMESPACE).List(ctx, metav1.ListOptions{})
			if err != nil {
				panic(err.Error())
			}
			// filtering all PVCs to get statefulset PVCs
			pvcs := GetStatefulSetPVCs(clientSet, ctx, persistentvolumeClaims.Items, map[string]*StorageV1.StorageClass{storageClass.Name: storageClass})
			count := 0
			for _, pvc := range pvcs {
				if strings.Contains(statefulset.Name, test.expected.Name) && pvc.Labels[storageClass.Parameters[constants.STS_PVC_SELECTOR]] == test.expected.Spec.Selector.MatchLabels[storageClass.Parameters[constants.STS_PVC_SELECTOR]] {
					count++
				}
			}
			if count != int(*test.expected.Spec.Replicas) {
				t.Fatalf("Failed to get statefulset PVCs of statefulset, %v", statefulset)
			}
		})
	}

	// cleaning up resources created for testing
	for i := 0; i < int(*statefulset.Spec.Replicas); i++ {
		err := clientSet.CoreV1().PersistentVolumeClaims(constants.TEST_NAMESPACE).Delete(ctx, fmt.Sprintf("pvc-%v-%v", statefulset.Name, i), metav1.DeleteOptions{})
		if err != nil {
			panic(err.Error())
		}
	}
	err := clientSet.StorageV1().StorageClasses().Delete(ctx, storageClass.Name, metav1.DeleteOptions{})
	if err != nil {
		panic(err.Error())
	}
}

func TestDeleteDanglingPVCs(t *testing.T) {
	ctx := context.Background()
	clientSet, clusterTestEnv := startCluster()
	defer stopCluster(clusterTestEnv)

	pvc := generators.GeneratePersistentVolumeClaim(fmt.Sprintf("test-pvc-%v", rand.Int()), constants.TEST_NAMESPACE, "test-storage-class")
	tests := map[string]struct {
		initFunc func(*kubernetes.Clientset)
		expected int
	}{
		"Testing Deletion of dangling PVCs": {
			initFunc: func(clientset *kubernetes.Clientset) {
				_, err := clientset.CoreV1().PersistentVolumeClaims(constants.TEST_NAMESPACE).Create(ctx, pvc, metav1.CreateOptions{})
				if err != nil {
					panic(err.Error())
				}
			},
			expected: 0,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			test.initFunc(clientSet)
			DeleteDanglingPVCs(clientSet, ctx, constants.TEST_NAMESPACE, map[string]bool{pvc.Name: true})

			time.Sleep(5 * time.Second)
			persistentvolumeClaims, _ := clientSet.CoreV1().PersistentVolumeClaims(constants.TEST_NAMESPACE).List(ctx, metav1.ListOptions{})
			count := 0
			for _, pvcTemp := range persistentvolumeClaims.Items {
				if pvcTemp.Name == pvc.Name {
					count++
				}
			}
			if count != test.expected {
				t.Fatalf("Dangling PVCs could not be deleted successfully, %v", persistentvolumeClaims.Items)
			}
		})
	}
}
