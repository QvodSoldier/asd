package controllers

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getPod(name, namespace, image, nodeName string) *corev1.Pod {
	debugPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    map[string]string{"DebugApp": name, "app": "asd"},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:            "asd",
					Image:           image,
					ImagePullPolicy: corev1.PullAlways,
					Args:            []string{"-c", "./mnt/agent/asd_agent"},
					Command:         []string{"/bin/sh"},
					SecurityContext: &corev1.SecurityContext{
						Capabilities: &corev1.Capabilities{
							Add: []corev1.Capability{"SYS_ADMIN", "SYS_CHROOT"},
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						corev1.VolumeMount{
							Name:      "asd-data",
							MountPath: "/mnt",
						},
					},
				},
			},
			NodeName:      nodeName,
			HostPID:       true,
			RestartPolicy: corev1.RestartPolicyAlways,
			Volumes: []corev1.Volume{
				corev1.Volume{
					Name: "asd-data",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: "asd-data",
						},
					},
				},
			},
		},
	}
	return debugPod
}
