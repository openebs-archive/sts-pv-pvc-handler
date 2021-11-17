package tests

import (
	"flag"
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	kubeConfigPath                  string
	openebsNamespace                string
	clientset                       *kubernetes.Clientset
	LocalPVProvisionerLabelSelector = "openebs.io/component-name=openebs-localpv-provisioner"
)

func init() {
	flag.StringVar(&kubeConfigPath, "kubeconfig", os.Getenv("KUBECONFIG"), "path to kubeconfig to invoke kubernetes API calls")
	flag.StringVar(&openebsNamespace, "openebs-namespace", "openebs", "kubernetes namespace where the OpenEBS components are present")
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		panic(err.Error())
	}
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
}

func TestSource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test application deployment")
}

var _ = BeforeSuite(func() {
	By("waiting for openebs-localpv-provisioner pod to come into running state")
	provPodCount := GetPodRunningCountEventually(
		openebsNamespace,
		LocalPVProvisionerLabelSelector,
		1,
		clientset,
	)
	Expect(provPodCount).To(Equal(1))

	// run our job
})

var _ = Describe("Shopping cart", func() {})
