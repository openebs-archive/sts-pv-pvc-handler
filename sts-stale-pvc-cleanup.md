# Stateful Set Stale PVC Cleanup

## Summary

OpenEBS Dynamic Local PV provisioner allows automatic and dynamic provisioning of Kubernetes Local Volumes using different kinds of storage available on the nodes. It dynamically creates persistent volumes (PV) and persistent volume claims (PVC) needed by a Pod. In a StatefulSet when the entire StatefulSet is deleted or it is scaled down then some or all Pods are destroyed but the PVCs they were bound to and the PVs which were created dynamically on the creation of that pod sticks around. The aim of this project is to automate the removal of such stale or dangling PVCs and PVs based on custom annotations provided in their storage class.

## Motivation

The motivation of this project is to ensure optimal utilisation of resources and reduce manual administration of the cluster required from the DevOps teams.

## Goals

- To have a frequently running process to remove stale or dangling PVCs and PVs.

- Provide means to specify default behaviour (to delete or not) for stale or dangling PVCs for openEbs storage classes.

## Non-Goals

- The functionality could be extended to non openEbs storage classes by providing a list of provisioner names via an environment variable if required but the default behaviour will include all openEbs storage classes in the cluster.

- The project is restricted to only pods under a statefulset, and will not handle dangling PVCs created by deletion of individual pods not under a statefulset.

## Proposal

This project proposes the creation of a Cron Job that runs as per a specified configurable cron schedule and identifies the stale or dangling PVCs correctly and deletes them or leaves them as they are based upon confirmation specified by cluster administrators. This project will only deal with openEbs components, i.e. storage classes that use an openEbs provisioner and PVCs associated with those storage classes. Cluster administrators will be able to specify the default behaviour for dangling PVCs through annotations in the storage classes. PVCs that are dynamically created by the OpenEBS Dynamic Local PV provisioner have the same labels as the stateful sets and one of pre-determined storage classes provided by OpenEBS. Using these we would identify PVCs that are dynamically created by the OpenEBS Dynamic Local PV provisioner. Among these PVCs we will look for the PVCs that are not bound and these would be dangling. Based upon annotations provided in the storage classes, dangling PVCs will be deleted by the Go binary running via the Cron Job.

## User Stories

We have mainly two types of users
- Cluster or System Administrator - creates storage classes and specifies the default behavior
- Developers - creates statefulsets and specified behaviour specific to the application based on its business logic.

PVCs associated to a statefulset might go into dangling state only in two ways
- Downscaling of the statefulset
- Deletion of the entire statefulset

### User Story 1

If the stateful set is deleted then all the volumes that were created dynamically for that stateful set should be deleted. On the deletion of a statefulset it has to be ensured that all the pods have terminated before deleting the volumes associated with that statefulset as per the configured behaviour via annotations and labels.

### User Story 2

Let there be a storage class in whose annotations the default behaviour is specified not to delete dangling PVs but lets say that there is a developer testing out things with that storage class and provisioning a statefulset for which he is certain that the PVCs are of no use after deletion of the Pods they were bound to, then there should be a way enable the developer to specify that in the statefulset definition by using labels. This functionality would be a stretch goal for the mentorship.

### User Story 3

In a cluster initially there is one statefulset with one pod in it associated with one volume, when we scale up the statefulset then the number of pods increase and new volumes associated with them are created dynamically by dynamic-local-pv provisioner. When we scale down the statefulset then the number of pods decreases and the pods will be deleted one by one, leaving the PVCs associated with them in a dangling state. As we know that the identity of pods persist in a stateful set, when the statefulset scales back up again the same pods will be associated with the same PVCs again. There could be situations where we want the PVCs and PVs to persist when the statefulset is scaled down so that they are available to be bound when its scaled back up again and the data in them is not lost but we might want all the PVCs and PVs to be deleted when the entire statefulset is deleted. Hence having different properties to specify behaviour in these two different situations makes sense. 

Fate of dangling PVCs created by scaling down of the statefulset should be determined by a label in the statefulset itself as this is closely related with the business logic of the application and could be different for different stateful sets hence it should be configurable by the developer. However if the same is not specified then default behaviour specified through the annotations by the cluster administrators in the storage class that the PVC uses will take effect. This would be a stretch goal for the mentorship.

## Dependencies

- Client-go
- Docker

## Design Details

- There is a fixed number of provisioners that are supposed by OpenEBS, a list of all the provisioners will be provided as an environment variable to the Go binary.
 
- The storage class definition will have an annotation that would specify the behaviour (whether to delete or not to delete) dangling or stale PVCs and PVs. 

        apiVersion: storage.k8s.io/v1
        kind: StorageClass
        metadata:
            name: openebs-hostpath
        annotations:
            openebs.io/delete-dangling-pvc: true
            openebs.io/cas-type: local
            cas.openebs.io/config: |
                - name: StorageType
                value: "hostpath"
                - name: BasePath
                value: "/var/openebs/local/"
        provisioner: openebs.io/local
        volumeBindingMode: WaitForFirstConsumer
        reclaimPolicy: Delete

- We will list all the storage classes and iterate over them, storage classes that do not have any openEbs related provisioner or for whom the required annotation is not set will not be considered further. After this step we will have a list of storage classes for whom the dangling PVCs are to be deleted.

- We will list down all the Persistent Volume Claims from the list of all PVCs we would select only PVCs that have the storage class in the list we got from the previous step. After this we would have a list of PVCs which would have a storage class that uses an openEBS provisioner and has the deletion annotation set.

- In the next step we would list all the stateful sets, for PVCs that are dynamically provisioned by OpenEbs dynamic local PV provisioner the labels on the statefulset are carried onto the PVC, Using this we would further filter out the PVCs. The PVCs whose labels are not present on any of the statefulsets could not have been dynamically provisioned and they will not be considered further. There could be some corner cases here which would be taken care of. After this step we would have a list of PVCs that were dynamically provisioned and uses an openEbs storage class that has the deletion annotation set, so if any of the PVCs in this list is found to be dangling then it could be deleted.

- We would create a dangling status map for the list of PVCs we got from the last step, we will mark all the PVCs are dangling initially and iterate over all the pods across all the statefulsets, we would use the selector of a statefulset to get all the pods under it, we would get the PVC that the pod is bound to and mark it as not dangling in the map.

- The selectors on a Statefulset are applied as labels on statefulset PVCs. If a Statefulset PVC is not mounted by any pod then it is set to be dandling and based on the annotation in its storage class it could either be deleted or not. However, if the statefulset itself is deleted then there is no way to tell if the PVCs that were associated with it were statefulset PVCs or just standalone PCVs created with some labels applied. To handle this, an extra selector would need to be given by the developer in the statefulset which would clearly determine that a PVC is a statefulset PVC. The developer would get the name and value for this from the storage class parameters of the storage class he wants to use in the VolumeClaimTemplate. Ultimately we would iterate over the keys in the map and delete all the PVCs that are still marked as dangling.

## Test Plan

Ideally the Go binary should be able to interact with the Kubernetes API and identify and delete dangling PVCs.  Lets go over some tests to understand the expected behaviour more clearly;

### Test 1

- Step 1 - Create a Storage Class using an OpenEbs provisioner and set the delete-dangling-pvc annotation to true.
- Step 2 - Create a StatefulSet with 3 replicas and specify the storage class created in Step 1 under the volumeClaimTemplates
- Step 3 - List the PVCs in the cluster and observe the PVCs associated with the statefulset.
- Step 4 - Downscale the statefulset to 1 or 2 replicas.
- Step 5 - Run the go binary or wait for some time for the cron job to kick in.
- Step 6 - List the PVCs in the cluster again and observe the output. 
- Step 7 - Delete the entire stateful list.
- Step 8 - Run the go binary or wait for some time.
- Step 9 - List the PVCs in the cluster again and observe the output.

For this test to be positive the dangling PVCs should automatically get deleted when the statefulset is downscaled or is deleted.

### Test 2

- Step 1 - Create a Storage Class using an OpenEbs provisioner but do not set the delete-dangling-pvc annotation.
- Repeat Steps 2 to 9 from Test 1.

For this test to be positive, the dangling PVCs should stick around when the statefulset is deleted or downscaled which is according to default kubernetes behaviour as the annotation was not set in the storage class.
