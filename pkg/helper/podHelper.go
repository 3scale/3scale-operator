package helper

import (
	"bytes"
	"context"
	"time"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	apimachinerymetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	kube "k8s.io/client-go/kubernetes"

	"k8s.io/client-go/kubernetes/scheme"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	redisDefaultImage      = "quay.io/fedora/redis-6"
	mySqlDefaultImage      = "quay.io/sclorg/mysql-80-c8s"
	postgreSqlDefaultImage = "quay.io/sclorg/postgresql-10-c8s"
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

func CreateRedisThrowAwayPod(k8sclient client.Client, namespace string) (*v1.Pod, error) {
	// Create throwaway redis deployment
	systemRedisPod := throwAwayRedis(namespace)

	err := k8sclient.Create(context.TODO(), systemRedisPod)
	if err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			// if it's other error than "AlreadyExists"
			return systemRedisPod, err
		}
	}

	// Wait for deployment to become ready
	err = wait.Poll(time.Second*5, time.Minute*3, func() (done bool, err error) {
		err = k8sclient.Get(context.TODO(), client.ObjectKey{Name: systemRedisPod.Name, Namespace: systemRedisPod.Namespace}, systemRedisPod)
		if err != nil {
			// Failed getting the deployment, trying again
			return false, nil
		}

		for _, condition := range systemRedisPod.Status.Conditions {
			if condition.Type == v1.PodReady {
				if condition.Status == v1.ConditionTrue {
					return true, nil
				} else {
					return false, nil
				}
			}
		}

		return true, nil
	})
	if err != nil {
		// Failed to boot up Redis throwaway pod
		return systemRedisPod, err
	}

	return systemRedisPod, nil
}

func DeletePod(k8sclient client.Client, pod *v1.Pod) error {
	err := k8sclient.Delete(context.TODO(), pod)
	if err != nil {
		return err
	}

	return nil
}

func CreateDatabaseThrowAwayPod(k8sclient client.Client, namespace, dbType string) (*v1.Pod, error) {
	// Create throwaway redis deployment
	dbPod := &v1.Pod{}
	if dbType == "mysql" {
		dbPod = throwAwayMysql(namespace)
	}
	if dbType == "postgres" {
		dbPod = throwAwayPostgres(namespace)
	}

	err := k8sclient.Create(context.TODO(), dbPod)
	if err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			// if it's other error than "AlreadyExists"
			return dbPod, err
		}
	}

	// Wait for deployment to become ready
	err = wait.Poll(time.Second*5, time.Minute*3, func() (done bool, err error) {
		err = k8sclient.Get(context.TODO(), client.ObjectKey{Name: dbPod.Name, Namespace: dbPod.Namespace}, dbPod)
		if err != nil {
			// Failed getting the deployment, trying again
			return false, nil
		}

		for _, condition := range dbPod.Status.Conditions {
			if condition.Type == v1.PodReady {
				if condition.Status == v1.ConditionTrue {
					return true, nil
				} else {
					return false, nil
				}
			}
		}

		return true, nil
	})
	if err != nil {
		return dbPod, err
	}

	return dbPod, nil
}

func throwAwayPostgres(namespace string) *v1.Pod {
	systemPostgresImage := GetEnvVar("RELATED_IMAGE_SYSTEM_POSTGRESQL", postgreSqlDefaultImage)
	return &v1.Pod{
		ObjectMeta: apimachinerymetav1.ObjectMeta{
			Name:      "throwaway-postgres",
			Namespace: namespace,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  "throwaway-postgres",
					Image: systemPostgresImage,
					Env: []v1.EnvVar{
						{Name: "POSTGRESQL_USER", Value: "throwaway"},
						{Name: "POSTGRESQL_PASSWORD", Value: "throwaway"},
						{Name: "POSTGRESQL_DATABASE", Value: "throwaway"},
					},
				},
			},
		},
	}
}

func throwAwayMysql(namespace string) *v1.Pod {
	systemMysqlImage := GetEnvVar("RELATED_IMAGE_SYSTEM_MYSQL", mySqlDefaultImage)
	return &v1.Pod{
		ObjectMeta: apimachinerymetav1.ObjectMeta{
			Name:      "throwaway-mysql",
			Namespace: namespace,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  "throwaway-mysql",
					Image: systemMysqlImage,
					Env: []v1.EnvVar{
						{Name: "MYSQL_USER", Value: "throwaway"},
						{Name: "MYSQL_PASSWORD", Value: "throwaway"},
						{Name: "MYSQL_DATABASE", Value: "throwaway"},
					},
				},
			},
		},
	}
}

func throwAwayRedis(namespace string) *v1.Pod {
	systemRedisImage := GetEnvVar("RELATED_IMAGE_SYSTEM_REDIS", redisDefaultImage)
	return &v1.Pod{
		ObjectMeta: apimachinerymetav1.ObjectMeta{
			Name:      "throwaway-redis",
			Namespace: namespace,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  "throwaway-redis",
					Image: systemRedisImage,
				},
			},
		},
	}
}
