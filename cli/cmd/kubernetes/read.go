// : Copyright Verizon Media
// : Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.
package kubernetes

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/simplexiengage/kubectl-flame/cli/cmd/kubernetes/job"
	"time"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/simplexiengage/kubectl-flame/cli/cmd/data"
)

type DataHandler interface {
	Handle(events chan string, done chan bool, ctx context.Context)
}

func GetPodDetails(podName, namespace string, ctx context.Context) (*apiv1.Pod, error) {
	podObject, err := clientSet.
		CoreV1().
		Pods(namespace).
		Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return podObject, nil
}

func WaitForPodStart(cfg *data.FlameConfig, ctx context.Context) (*apiv1.Pod, error) {
	var pod *apiv1.Pod
	err := wait.Poll(1*time.Second, 5*time.Minute,
		func() (bool, error) {
			podList, err := clientSet.
				CoreV1().
				Pods(cfg.JobConfig.Namespace).
				List(ctx, metav1.ListOptions{
					LabelSelector: fmt.Sprintf("kubectl-flame/id=%s", cfg.TargetConfig.Id),
				})

			if err != nil {
				return false, err
			}

			if len(podList.Items) == 0 {
				return false, nil
			}

			pod = &podList.Items[0]
			switch pod.Status.Phase {
			case apiv1.PodFailed:
				return false, fmt.Errorf("pod failed")
			case apiv1.PodSucceeded:
				fallthrough
			case apiv1.PodRunning:
				return true, nil
			default:
				return false, nil
			}
		})

	if err != nil {
		return nil, err
	}

	return pod, nil
}

func GetLogsFromPod(pod *apiv1.Pod, handler DataHandler, ctx context.Context) (chan bool, error) {
	done := make(chan bool)
	req := clientSet.CoreV1().
		Pods(pod.Namespace).
		GetLogs(pod.Name, &apiv1.PodLogOptions{
			Follow:    true,
			Container: job.ContainerName,
		})

	readCloser, err := req.Stream(ctx)
	if err != nil {
		return nil, err
	}

	eventsChan := make(chan string)
	go handler.Handle(eventsChan, done, ctx)
	go func() {
		defer readCloser.Close()
		r := bufio.NewReader(readCloser)
		for {
			bytes, err := r.ReadBytes('\n')

			if err != nil {
				return
			}

			eventsChan <- string(bytes)
		}
	}()

	return done, nil
}

func GetContainerId(containerName string, pod *apiv1.Pod) (string, error) {
	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.Name == containerName {
			return containerStatus.ContainerID, nil
		}
	}

	return "", errors.New("Could not find container id for " + containerName)
}
