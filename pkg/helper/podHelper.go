package helper

import (
	"bytes"
	l "github.com/go-logr/logr"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	kube "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
)

//go:generate moq -out pod_executor_moq.go . PodExecutorInterface
type PodExecutorInterface interface {
	ExecuteRemoteCommand(ns string, podName string, command []string) (string, string, error)
	ExecuteRemoteContainerCommand(ns string, podName string, container string, command []string) (string, string, error)
}

type PodExecutor struct {
	Log l.Logger
}

var _ PodExecutorInterface = &PodExecutor{}

func NewPodExecutor(log l.Logger) *PodExecutor {
	return &PodExecutor{
		Log: log,
	}
}

func (p PodExecutor) ExecuteRemoteCommand(ns string, podName string, command []string) (string, string, error) {

	kubeClient, restConfig, err := getClient()
	if err != nil {
		return "", "", errors.Wrapf(err, "Failed to get client")
	}

	req := kubeClient.CoreV1().RESTClient().Post().Resource("pods").Name(podName).
		Namespace(ns).SubResource("exec")
	option := &v1.PodExecOptions{
		Command: command,
		Stdin:   false,
		Stdout:  true,
		Stderr:  true,
		TTY:     true,
		//Container: container,
	}
	req.VersionedParams(
		option,
		scheme.ParameterCodec,
	)
	exec, err := remotecommand.NewSPDYExecutor(restConfig, "POST", req.URL())
	if err != nil {
		return "", "", errors.Wrapf(err, "Failed executing command %s on %s/%s", command, ns, podName)
	}

	buf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}

	err = exec.Stream(remotecommand.StreamOptions{
		Stdout: buf,
		Stderr: errBuf,
	})
	if err != nil {
		return "", "", errors.Wrapf(err, "Failed executing command %s on %s/%s", command, ns, podName)
	}

	return buf.String(), errBuf.String(), nil
}

func (p PodExecutor) ExecuteRemoteContainerCommand(ns string, podName string, container string, command []string) (string, string, error) {
	kubeClient, restConfig, err := getClient()
	if err != nil {
		return "", "", errors.Wrapf(err, "Failed to get client")
	}

	req := kubeClient.CoreV1().RESTClient().Post().Resource("pods").Name(podName).
		Namespace(ns).SubResource("exec")
	option := &v1.PodExecOptions{
		Command:   command,
		Stdin:     false,
		Stdout:    true,
		Stderr:    true,
		TTY:       true,
		Container: container,
	}
	req.VersionedParams(
		option,
		scheme.ParameterCodec,
	)
	exec, err := remotecommand.NewSPDYExecutor(restConfig, "POST", req.URL())
	if err != nil {
		return "", "", errors.Wrapf(err, "Failed executing command %s on %s/%s", command, ns, podName)
	}

	buf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}

	streamOptions := remotecommand.StreamOptions{
		Stdout: buf,
		Stderr: errBuf,
		Tty:    option.TTY,
	}
	
	err = exec.Stream(streamOptions)
	if err != nil {
		return "", "", errors.Wrapf(err, "Failed executing command %s on %s/%s", command, ns, podName)
	}

	return buf.String(), errBuf.String(), nil
}

func getClient() (*kube.Clientset, *restclient.Config, error) {

	kubeCfg := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)
	restCfg, err := kubeCfg.ClientConfig()
	if err != nil {
		return nil, nil, err
	}

	kubeClient, err := kube.NewForConfig(restCfg)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Failed to generate new client")
	}
	return kubeClient, restCfg, nil
}
