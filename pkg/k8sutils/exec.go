package k8sutils

import (
	"io"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

type AttachOptions struct {
	Stdin     io.Reader
	Stdout    io.Writer
	PodName   string
	Namespace string
}

func (k *kubeFactory) GetExecutor(ns, pod, container string) (exec remotecommand.Executor, err error) {
	req := k.clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(pod).
		Namespace(ns).
		SubResource("exec").
		Param("container", container)

	req.VersionedParams(&v1.PodExecOptions{
		Container: container,
		Command:   []string{"/bin/sh"},
		Stdin:     true,
		Stdout:    true,
		Stderr:    true,
		TTY:       true,
	}, scheme.ParameterCodec)

	exec, err = remotecommand.NewSPDYExecutor(k.conf, "POST", req.URL())
	if err != nil {
		return
	}

	return
}
