package tests

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ksraj123/lister-sa/tests/generators"
	"github.com/ksraj123/lister-sa/tests/setup"
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

func TestMain(t *testing.T) {
	ctx := context.Background()
	clientSet, clusterTestEnv := startCluster()
	defer stopCluster(clusterTestEnv)

	// Ensure openebs-localpv-provisioner pod running
	setup.TestOpenEbsPodRunningState(t, clientSet, ctx, 90)
	setup.CreateStorageClass(t, clientSet, ctx, 10)
	statefulsetName := "test-sts"
	setup.CreateStatefulSet(t, clientSet, ctx, statefulsetName, 10)

	// Delete the statefulset

	err := clientSet.AppsV1().StatefulSets("default").Delete(ctx, statefulsetName, metav1.DeleteOptions{})
	if err != nil {
		panic(err.Error())
	}

	time.Sleep(5 * time.Second)

	// now the statefulset PVCs are in dangling state

	serviceAccountName := "openebs-maya-operator"
	image := "ksraj123/stale-sts-pvc-cleaner:0.1"
	env := []CoreV1.EnvVar{
		{
			Name:  "PROVISIONERS",
			Value: "openebs.io/local",
		},
		{
			Name:  "NAMESPACES",
			Value: "default",
		},
	}

	job := generators.GenerateJob("test-job", map[string]string{"jobGroup": "test"}, serviceAccountName, image, env)
	_, err = clientSet.BatchV1().Jobs("default").Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		panic(err.Error())
	}

	time.Sleep(5 * time.Second)

	pvcs, err := clientSet.CoreV1().PersistentVolumeClaims("default").List(ctx, metav1.ListOptions{})
	if err != nil {
		fmt.Printf("error %s, getting PVCs\n", err.Error())
	}
	for _, pvc := range pvcs.Items {
		if strings.Contains(pvc.Name, statefulsetName) {
			t.Fatalf("Dangling PVCs %v not deleted by Job", pvc.Name)
		}
	}

	// clientSet.BatchV1().Jobs("default").Delete(ctx, "test-job", metav1.DeleteOptions{})
}
