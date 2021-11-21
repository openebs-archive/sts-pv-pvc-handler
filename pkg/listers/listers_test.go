package listers

import (
	"context"
	"testing"

	AppsV1 "k8s.io/api/apps/v1"
	CoreV1 "k8s.io/api/core/v1"
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

func getStatefulSet(replicas int32, selector map[string]string) *AppsV1.StatefulSet {
	return &AppsV1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-statefulset",
			Namespace: "default",
		},
		Spec: AppsV1.StatefulSetSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: selector,
			},
			Template: CoreV1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: selector,
				},
				Spec: CoreV1.PodSpec{
					Containers: []CoreV1.Container{
						{
							Name:            "busybox",
							Image:           "busybox",
							ImagePullPolicy: CoreV1.PullIfNotPresent,
							Command: []string{
								"sleep",
								"infinity",
							},
						},
					},
				},
			},
		},
	}
}

func TestListAllStatefulSets(t *testing.T) {
	ctx := context.Background()
	clientSet, clusterTestEnv := startCluster()
	defer stopCluster(clusterTestEnv)

	tests := map[string]struct {
		initFunc func(*kubernetes.Clientset)
		// observedStatefulSets []AppsV1.StatefulSet
		expected int
	}{
		"Listing Stateful Sets in a cluster": {
			initFunc: func(clientset *kubernetes.Clientset) {
				_, err := clientset.AppsV1().StatefulSets("default").Create(ctx, getStatefulSet(1, map[string]string{"role": "test"}), metav1.CreateOptions{})
				if err != nil {
					panic(err.Error())
				}
			},
			expected: 1,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			test.initFunc(clientSet)
			observedStatefulSets := ListAllStatefulSets(clientSet, ctx)
			if len(observedStatefulSets) != test.expected {
				t.Fatalf("Expected %v, got %v", test.expected, len(observedStatefulSets))
			}
		})
	}
}
