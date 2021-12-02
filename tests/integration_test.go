package tests

import (
	"context"
	"flag"
	"os"
	"testing"
	"time"

	"github.com/ksraj123/lister-sa/tests/generators"
	"github.com/ksraj123/lister-sa/tests/setup"
	CoreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	kubeConfigPath                  string
	openebsNamespace                string
	clientSet                       *kubernetes.Clientset
	LocalPVProvisionerLabelSelector = "openebs.io/component-name=openebs-localpv-provisioner"
)

func init() {
	flag.StringVar(&kubeConfigPath, "kubeconfig", os.Getenv("KUBECONFIG"), "path to kubeconfig to invoke kubernetes API calls")
	flag.StringVar(&openebsNamespace, "openebs-namespace", "openebs", "kubernetes namespace where the OpenEBS components are present")
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		panic(err.Error())
	}
	clientSet, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
}

func TestMain(t *testing.T) {
	ctx := context.Background()

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

	setup.TestDanglingPVCDeleted(t, clientSet, ctx, statefulsetName, 90)

	err = clientSet.BatchV1().Jobs("default").Delete(ctx, "test-job", metav1.DeleteOptions{})
	if err != nil {
		panic(err.Error())
	}

	err = clientSet.StorageV1().StorageClasses().Delete(ctx, "test-storage-class", metav1.DeleteOptions{})
	if err != nil {
		panic(err.Error())
	}
}
