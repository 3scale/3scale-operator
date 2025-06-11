package helper

import (
	"bytes"
	"context"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	kube "k8s.io/client-go/kubernetes"

	"k8s.io/client-go/kubernetes/scheme"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:generate moq -out pod_executor_moq.go . PodExecutorInterface
type PodExecutorInterface interface {
	ExecuteRemoteCommand(ns string, podName string, command []string) (string, string, error)
	ExecuteRemoteContainerCommand(ns string, podName string, container string, command []string) (string, string, error)
}

type PodExecutor struct {
	Log logr.Logger
}

var _ PodExecutorInterface = &PodExecutor{}

func NewPodExecutor(log logr.Logger) *PodExecutor {
	return &PodExecutor{
		Log: log,
	}
}

// ExecuteRemoteCommand exec command on specific pod and wait the command's output.
func (p PodExecutor) ExecuteRemoteCommand(ns string, podName string, command []string) (string, string, error) {
	kubeClient, restConfig, err := getClient()
	if err != nil {
		return "", "", errors.Errorf("Failed to get client :%s", err)
	}

	req := kubeClient.CoreV1().RESTClient().Post().Resource("pods").Name(podName).
		Namespace(ns).SubResource("exec")
	option := &v1.PodExecOptions{
		Command: command,
		Stdin:   false,
		Stdout:  true,
		Stderr:  true,
		TTY:     true,
	}
	req.VersionedParams(
		option,
		scheme.ParameterCodec,
	)
	exec, err := remotecommand.NewSPDYExecutor(restConfig, "POST", req.URL())
	if err != nil {
		return "", "", errors.Errorf("Failed executing command on throwaway pod %s", podName)
	}

	buf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}

	err = exec.StreamWithContext(context.Background(), remotecommand.StreamOptions{
		Stdout: buf,
		Stderr: errBuf,
	})
	if err != nil {
		return "", "", errors.Errorf("Failed executing command on throwaway pod %s", podName)
	}

	return buf.String(), errBuf.String(), nil
}

// ExecuteRemoteContainerCommand exec command on specific pod and wait the command's output.
func (p PodExecutor) ExecuteRemoteContainerCommand(ns string, podName string, container string, command []string) (string, string, error) {
	kubeClient, restConfig, err := getClient()
	if err != nil {
		return "", "", errors.Errorf("Failed to get client during throwaway pod execuction command: %s", err)
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
		return "", "", errors.Errorf("Failed executing command on throwaway pod %s", podName)
	}

	buf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}

	err = exec.StreamWithContext(context.Background(), remotecommand.StreamOptions{
		Stdout: buf,
		Stderr: errBuf,
	})
	if err != nil {
		return "", "", errors.Errorf("Failed executing command on throwaway pod %s", podName)
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
		return nil, nil, errors.Errorf("Failed to generate new client for throwaway pod %s", err)
	}
	return kubeClient, restCfg, nil
}

func DeletePod(k8sclient client.Client, pod *v1.Pod) error {
	err := k8sclient.Delete(context.TODO(), pod)
	if err != nil {
		return err
	}

	return nil
}
