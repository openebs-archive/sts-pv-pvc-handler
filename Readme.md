# Stale Statefulset PersistentVolumeClaims cleaner (Experimental)

Kuberenetes allows dynamic creation of persistent volume claims (PVCs) for Statefulsets based volume claim templates. These statefulset PVCs get bound the pods automatically and even after destruction of the pods either by scaling down the statefulset or deleting the statefulset entirely these PVCs stick around which leave them in a 'dangling' state. In the context of persistent identity of pods in a statefulset, this might be beneficial in certain situations as the PVCs would bind to the same pod if the statefulset scaled back up again. However, mostly these dangling PVCs take up resources and need manual intervention to be deleted. This project creates a binary which could be run as a job or a cron job to interact with the Kubernetes API and indentify and delete such dangling PVCs.

## Setup

- Create service account and cluster roles to allow the binary to interact with the Kuberenets API

  `kubectl apply -f deploy/sa.yaml`

- Build the latest image (can be skipped if pre-built image is used)
  
  `make stale-sts-pvc-cleaner-image`

- Update image and service account details in yaml and create the job

  `kubectl apply -f deploy/job.yaml`

## Build and Release

To build binary for a desired platform and architecture, run `make stale-sts-pvc-cleaner` with envrionmet variables `XC_OS` and `XC_ARCH` specifying the platform and architecture. The binaries will get created under the `bin` directory.

To build multiple binaries at once, update `.goreleaser.yml` and run `goreleaser build`, the binaries will be created under the `dist` directory. To build and release on github and/or dockerhub run `goreleaser`. Setting up goreleaser is a prequisite for this, please find details for the same [here](https://github.com/goreleaser/goreleaser).

## Testing

### Unit Tests

`envtest` is a prequisite to run unit tests. `envtest` helps write in testing controllers by setting up and starting an instance of etcd and the Kubernetes API server, without kubelet, controller-manager or other components. Find more details [here](https://book.kubebuilder.io/reference/envtest.html).

To run unit tests, run `make test`

### Integration Tests

Requires an active Kubernetes cluster.

  `make integration-test`
