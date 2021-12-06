package listers

import (
	"context"
	"fmt"
	"math/rand"
	"testing"

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

func TestListAllStatefulSets(t *testing.T) {
	ctx := context.Background()
	clientSet, clusterTestEnv := startCluster()
	defer stopCluster(clusterTestEnv)

	replicas := 1
	statefulset := generators.GenerateStatefulSet(fmt.Sprintf("test-sts-%v", rand.Int()), constants.TEST_NAMESPACE, int32(replicas), map[string]string{"role": "test"}, "standard")

	tests := map[string]struct {
		initFunc func(*kubernetes.Clientset)
		expected *AppsV1.StatefulSet
	}{
		"Listing Stateful Sets in a namespace": {
			initFunc: func(clientset *kubernetes.Clientset) {
				_, err := clientset.AppsV1().StatefulSets(constants.TEST_NAMESPACE).Create(ctx, statefulset, metav1.CreateOptions{})
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
			observedStatefulSets := ListAllStatefulSets(clientSet, ctx, constants.TEST_NAMESPACE)
			expectedStatefulsetFound := false
			for _, sts := range observedStatefulSets {
				if sts.Name == test.expected.Name {
					expectedStatefulsetFound = true
					break
				}
			}
			if !expectedStatefulsetFound {
				t.Fatalf("Expected Statefulset %v, not found", test.expected)
			}
		})
	}
}

func TestListAllPersistentVolumeClaims(t *testing.T) {
	ctx := context.Background()
	clientSet, clusterTestEnv := startCluster()
	defer stopCluster(clusterTestEnv)

	persistentVolumeClaim := generators.GeneratePersistentVolumeClaim(fmt.Sprintf("test-pvc-%v", rand.Int()), constants.TEST_NAMESPACE, "test-storage-class", nil)

	tests := map[string]struct {
		initFunc func(*kubernetes.Clientset)
		expected *CoreV1.PersistentVolumeClaim
	}{
		"Listing Persistent Volume Claims in a namespace": {
			initFunc: func(clientset *kubernetes.Clientset) {
				_, err := clientset.CoreV1().PersistentVolumeClaims(constants.TEST_NAMESPACE).Create(ctx, persistentVolumeClaim, metav1.CreateOptions{})
				if err != nil {
					panic(err.Error())
				}
			},
			expected: persistentVolumeClaim,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			test.initFunc(clientSet)
			observedPVCs := ListAllPersistentVolumeClaims(clientSet, ctx, constants.TEST_NAMESPACE)
			expectedPVCFound := false
			for _, pvc := range observedPVCs {
				if pvc.Name == test.expected.Name {
					expectedPVCFound = true
					break
				}
			}
			if !expectedPVCFound {
				t.Fatalf("Expected PVC %v, not found", test.expected)
			}
		})
	}
}

func TestListAllStorageClasses(t *testing.T) {
	ctx := context.Background()
	clientSet, clusterTestEnv := startCluster()
	defer stopCluster(clusterTestEnv)

	storageClass := generators.GenerateStorageClass(fmt.Sprintf("test-sc-%v", rand.Int()), nil, nil, "test-provisioner")

	tests := map[string]struct {
		initFunc func(*kubernetes.Clientset)
		expected *StorageV1.StorageClass
	}{
		"Listing Storage Classes in a Cluster": {
			initFunc: func(clientset *kubernetes.Clientset) {
				_, err := clientset.StorageV1().StorageClasses().Create(ctx, storageClass, metav1.CreateOptions{})
				if err != nil {
					panic(err.Error())
				}
			},
			expected: storageClass,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			test.initFunc(clientSet)
			observedSCs := ListAllStorageClasses(clientSet, ctx)
			expectedSCFound := false
			for _, sc := range observedSCs {
				if sc.Name == test.expected.Name {
					expectedSCFound = true
					break
				}
			}
			if !expectedSCFound {
				t.Fatalf("Expected PVC %v, not found", test.expected)
			}
		})
	}
}

func TestListPVCsOfStorageClass(t *testing.T) {
	ctx := context.Background()
	clientSet, clusterTestEnv := startCluster()
	defer stopCluster(clusterTestEnv)

	storageClass := generators.GenerateStorageClass(fmt.Sprintf("test-sc-%v", rand.Int()), nil, nil, "test-provisioner")
	persistentVolumeClaim := generators.GeneratePersistentVolumeClaim(fmt.Sprintf("test-pvc-%v", rand.Int()), constants.TEST_NAMESPACE, storageClass.Name, nil)

	tests := map[string]struct {
		initFunc func(*kubernetes.Clientset)
		expected *StorageV1.StorageClass
	}{
		"Listing Persistent Volume Claims of a Storage Class": {
			initFunc: func(clientset *kubernetes.Clientset) {
				_, err := clientset.StorageV1().StorageClasses().Create(ctx, storageClass, metav1.CreateOptions{})
				if err != nil {
					panic(err.Error())
				}
				_, err = clientset.CoreV1().PersistentVolumeClaims(constants.TEST_NAMESPACE).Create(ctx, persistentVolumeClaim, metav1.CreateOptions{})
				if err != nil {
					panic(err.Error())
				}
			},
			expected: storageClass,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			test.initFunc(clientSet)
			observedPVCs := ListPVCsOfStorageClass(clientSet, ctx, constants.TEST_NAMESPACE, []*StorageV1.StorageClass{storageClass})
			expectedPVCFound := false
			t.Logf("%v", observedPVCs)
			for _, pvc := range observedPVCs {
				if *pvc.Spec.StorageClassName == test.expected.Name {
					expectedPVCFound = true
					break
				}
			}
			if !expectedPVCFound {
				t.Fatalf("PVC of Storage Class %v, not found", storageClass)
			}
		})
	}
}

func TestListProvisionerStorageClassesWithAnnotation(t *testing.T) {
	ctx := context.Background()
	clientSet, clusterTestEnv := startCluster()
	defer stopCluster(clusterTestEnv)
	annotation := "test-annotation"
	provisioner := "test-provisioner"

	storageClass := generators.GenerateStorageClass(fmt.Sprintf("test-sc-%v", rand.Int()), map[string]string{annotation: "true"}, nil, provisioner)

	type ProvisionerAnnotation struct {
		provisioner string
		annotation  string
	}
	tests := map[string]struct {
		initFunc func(*kubernetes.Clientset)
		expected ProvisionerAnnotation
	}{
		"Listing Storage Classes of given provisioners and annotation in a Cluster": {
			initFunc: func(clientset *kubernetes.Clientset) {
				_, err := clientset.StorageV1().StorageClasses().Create(ctx, storageClass, metav1.CreateOptions{})
				if err != nil {
					panic(err.Error())
				}
			},
			expected: ProvisionerAnnotation{
				provisioner: provisioner,
				annotation:  annotation,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			test.initFunc(clientSet)
			observedSCs := ListProvisionerStorageClassesWithAnnotation(clientSet, ctx, []string{provisioner}, annotation)
			expectedSCFound := false
			for _, sc := range observedSCs {
				if sc.Provisioner == test.expected.provisioner && sc.Annotations[test.expected.annotation] == "true" {
					expectedSCFound = true
					break
				}
			}
			if !expectedSCFound {
				t.Fatalf("Expected Storage class with the given provisioner %v, and annotation %v not found", test.expected.provisioner, test.expected.annotation)
			}
		})
	}
}
