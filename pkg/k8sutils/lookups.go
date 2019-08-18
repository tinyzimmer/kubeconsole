package k8sutils

import (
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k *kubeFactory) ListNamespaces() (namespaces []string, err error) {
	res, err := k.clientset.CoreV1().Namespaces().List(v1.ListOptions{})
	if err != nil {
		return
	}
	for _, ns := range res.Items {
		namespaces = append(namespaces, ns.Name)
	}
	return
}

func (k *kubeFactory) ListPods(ns string) (pods []string, err error) {
	res, err := k.clientset.CoreV1().Pods(ns).List(v1.ListOptions{})
	if err != nil {
		return
	}
	for _, po := range res.Items {
		pods = append(pods, po.Name)
	}
	return
}

func (k *kubeFactory) GetPod(ns string, pod string) (podMeta *corev1.Pod, err error) {
	res, err := k.clientset.CoreV1().Pods(ns).Get(pod, v1.GetOptions{})
	if err != nil {
		return
	}
	podMeta = res
	return
}
