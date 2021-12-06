# Specify the kubeconfig path to a Kubernetes cluster 
# to run Hostpath integration tests
ifeq (${KUBECONFIG}, )
  KUBECONFIG=${HOME}/.kube/config
  export KUBECONFIG
endif

ifeq (${IMAGE_ORG}, )
  IMAGE_ORG = ksraj123
endif

# If IMAGE_TAG is mentioned then TAG will be set to IMAGE_TAG
# If RELEASE_TAG is mentioned then TAG will be set to RELEAE_TAG
# If both are mentioned then TAG will be set to RELEASE_TAG
TAG=ci

ifneq (${IMAGE_TAG}, )
  TAG=${IMAGE_TAG:v%=%}
endif

ifneq (${RELEASE_TAG}, )
  TAG=${RELEASE_TAG:v%=%}
endif

# Specify the name for the binaries
STALE_STS_PVC_CLEANER=stale-sts-pvc-cleaner

# Specify the name of the image
STALE_STS_PVC_CLEANER_IMAGE?=stale-sts-pvc-cleaner

# Final variable with image org, name and tag
STALE_STS_PVC_CLEANER_IMAGE_TAG=${IMAGE_ORG}/${STALE_STS_PVC_CLEANER_IMAGE}:${TAG}

# Requires KUBECONFIG env
.PHONY: integration-test
integration-test:
	@cd tests && go test -v -timeout 60m .

.PHONY: test
test:
	go test -timeout 60m ./pkg/...

.PHONY: stale-sts-pvc-cleaner
stale-sts-pvc-cleaner:
	@echo "----------------------------"
	@echo "--> stale-sts-pvc-cleaner    "
	@echo "----------------------------"
	@PNAME=${STALE_STS_PVC_CLEANER} CTLNAME=${STALE_STS_PVC_CLEANER} sh -c "'$(PWD)/buildscripts/build.sh'"


.PHONY: stale-sts-pvc-cleaner-image
stale-sts-pvc-cleaner-image:stale-sts-pvc-cleaner
	@echo "-------------------------------"
	@echo "--> stale-sts-pvc-cleaner "
	@echo "-------------------------------"
	@PNAME=${STALE_STS_PVC_CLEANER} CTLNAME=${STALE_STS_PVC_CLEANER} STALE_STS_PVC_CLEANER_IMAGE_TAG=${STALE_STS_PVC_CLEANER_IMAGE_TAG} sh -c "'$(PWD)/buildscripts/build-image.sh'"
