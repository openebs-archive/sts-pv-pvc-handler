# Specify the kubeconfig path to a Kubernetes cluster 
# to run Hostpath integration tests
ifeq (${KUBECONFIG}, )
  KUBECONFIG=${HOME}/.kube/config
  export KUBECONFIG
endif

# Requires KUBECONFIG env
.PHONY: integration-test
integration-test:
	@cd tests && go test .
