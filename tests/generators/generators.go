package generators

import (
	AppsV1 "k8s.io/api/apps/v1"
	CoreV1 "k8s.io/api/core/v1"
	StorageV1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GenerateStatefulSet(name string, namespace string, replicas int32, selector map[string]string) *AppsV1.StatefulSet {
	return &AppsV1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
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

func GeneratePersistentVolumeClaim(name string, namespace string, storageClassName string) *CoreV1.PersistentVolumeClaim {
	storage, _ := resource.ParseQuantity("1Gi")
	accessModes := []CoreV1.PersistentVolumeAccessMode{CoreV1.ReadWriteOnce}
	return &CoreV1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: CoreV1.PersistentVolumeClaimSpec{
			StorageClassName: &storageClassName,
			Resources: CoreV1.ResourceRequirements{
				Requests: map[CoreV1.ResourceName]resource.Quantity{CoreV1.ResourceStorage: storage},
			},
			AccessModes: accessModes,
		},
	}
}

func GenerateStorageClass(name string, annotations map[string]string, parameters map[string]string, provisioner string) *StorageV1.StorageClass {
	var deletePolicy CoreV1.PersistentVolumeReclaimPolicy = "Delete"
	return &StorageV1.StorageClass{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Annotations: annotations,
		},
		Parameters:    parameters,
		Provisioner:   provisioner,
		ReclaimPolicy: &deletePolicy,
	}
}
