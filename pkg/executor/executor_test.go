package executor

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/ksraj123/lister-sa/tests/generators"
	CoreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	envtest "sigs.k8s.io/controller-runtime/pkg/envtest"
)

const (
	TEST_NAMESPACE = "test"
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

	statefulsetReplicas := 1
	unboundPVC := generators.GeneratePersistentVolumeClaim(fmt.Sprintf("test-pvc-unbound-%v", rand.Int()), TEST_NAMESPACE, "test-storage-class")
	statefulset := generators.GenerateStatefulSet(fmt.Sprintf("test-sts-%v", rand.Int()), TEST_NAMESPACE, int32(statefulsetReplicas), map[string]string{"role": "test"}, "standard")

	tests := map[string]struct {
		initFunc func(*kubernetes.Clientset)
		expected *CoreV1.PersistentVolumeClaim
	}{
		"Testing Dangling State of PVCs": {
			initFunc: func(clientset *kubernetes.Clientset) {
				_, err := clientset.CoreV1().PersistentVolumeClaims(TEST_NAMESPACE).Create(ctx, unboundPVC, metav1.CreateOptions{})
				if err != nil {
					panic(err.Error())
				}
				_, err = clientset.AppsV1().StatefulSets(TEST_NAMESPACE).Create(ctx, statefulset, metav1.CreateOptions{})
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
			persistentvolumeClaims, _ := clientSet.CoreV1().PersistentVolumeClaims(TEST_NAMESPACE).List(ctx, metav1.ListOptions{})
			danglingStatusMap := GetPVCDanlingStatusMap(clientSet, ctx, TEST_NAMESPACE, persistentvolumeClaims.Items)
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

	err := clientSet.AppsV1().StatefulSets(TEST_NAMESPACE).Delete(ctx, statefulset.Name, metav1.DeleteOptions{})
	if err != nil {
		panic(err.Error())
	}
	for i := 0; i < statefulsetReplicas; i++ {
		err := clientSet.CoreV1().PersistentVolumeClaims(TEST_NAMESPACE).Delete(ctx, fmt.Sprintf("pvc-%v-%v", statefulset.Name, i), metav1.DeleteOptions{})
		if err != nil {
			panic(err.Error())
		}
	}

	err = clientSet.CoreV1().PersistentVolumeClaims(TEST_NAMESPACE).Delete(ctx, unboundPVC.Name, metav1.DeleteOptions{})
	if err != nil {
		panic(err.Error())
	}
}
