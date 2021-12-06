package statefulsetpvcs

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

func TestGetStatefulSetPVCs(t *testing.T) {
	ctx := context.Background()
	clientSet, clusterTestEnv := startCluster()
	defer stopCluster(clusterTestEnv)

	storageClass := generators.GenerateStorageClass(fmt.Sprintf("test-sc-%v", rand.Int()), nil, map[string]string{constants.STS_PVC_SELECTOR: "some-selector-in-sts"}, "test-provisioner")
	statefulsetReplicas := 1
	statefulsetSelector := map[string]string{storageClass.Parameters[constants.STS_PVC_SELECTOR]: "true"}
	statefulset := generators.GenerateStatefulSet(fmt.Sprintf("test-sts-%v", rand.Int()), constants.TEST_NAMESPACE, int32(statefulsetReplicas), statefulsetSelector, storageClass.Name)
	pvcSts := generators.GeneratePersistentVolumeClaim("test-pvc-1", constants.TEST_NAMESPACE, storageClass.Name, statefulsetSelector)
	pvcOther := generators.GeneratePersistentVolumeClaim("test-pvc-2", constants.TEST_NAMESPACE, storageClass.Name, nil)

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
				_, err = clientset.CoreV1().PersistentVolumeClaims(constants.TEST_NAMESPACE).Create(ctx, pvcSts, metav1.CreateOptions{})
				if err != nil {
					panic(err.Error())
				}
				_, err = clientset.CoreV1().PersistentVolumeClaims(constants.TEST_NAMESPACE).Create(ctx, pvcOther, metav1.CreateOptions{})
				if err != nil {
					panic(err.Error())
				}
			},
			expected: statefulset,
		},
	}

	time.Sleep(5 * time.Second)

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
}
