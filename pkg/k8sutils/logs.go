package k8sutils

import (
	"context"
	"io"

	v1 "k8s.io/api/core/v1"
)

func (k *kubeFactory) GetLogStream(ns, pod, container string, ctx context.Context) (stream io.ReadCloser, err error) {
	tail := int64(50)
	logOpts := &v1.PodLogOptions{}
	logOpts.Follow = true
	logOpts.TailLines = &tail
	if container != "" {
		logOpts.Container = container
	}
	stream, err = k.clientset.CoreV1().Pods(ns).GetLogs(pod, logOpts).Stream()
	return
}
