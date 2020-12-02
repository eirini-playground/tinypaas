package cmd

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func logs(cmd *cobra.Command, args []string) {
	name, err := cmd.Flags().GetString("name")
	ExitfIfError(err, "Failed to get name flag")

	config := parseConfig(cmd)
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", config.KubeconfigPath)
	ExitfIfError(err, "an unexpected error occurred")

	client, err := kubernetes.NewForConfig(kubeConfig)
	ExitfIfError(err, "Failed to create k8s client")

	printStagingLogs(client, config.Namespace, name)
}

func printStagingLogs(client kubernetes.Interface, namespace, name string) {
	stagingPods, err := client.CoreV1().Pods(namespace).List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("image.kpack.io/image=%s", name),
	})
	ExitfIfError(err, "Couldn't get staging logs for app")

	printLogs(client, namespace, stagingPods.Items)
}

func printLogs(client kubernetes.Interface, namespace string, pods []corev1.Pod) {
	if len(pods) == 0 {
		fmt.Println("No logs found")
		return
	}
	initContainers := pods[0].Spec.InitContainers
	containers := pods[0].Spec.Containers
	for _, container := range initContainers {
		for {
			pod, err := client.CoreV1().Pods(namespace).Get(pods[0].Name, metav1.GetOptions{})
			if err != nil {
				ExitfIfError(err, "Couldn't fetch a pod")
			}
			var s corev1.ContainerStatus
			for _, s = range pod.Status.InitContainerStatuses {
				if s.Name == container.Name {
					break
				}
			}
			if s.Name == "" {
				time.Sleep(1 * time.Second)
				continue
			} else {
				if s.State.Waiting == nil {
					break
				}
			}
		}

		err := tailContainerLogs(client, namespace, pods[0].Name, container.Name)
		ExitfIfError(err, "failed to get build logs")
	}
	for _, container := range containers {
		for {
			pod, err := client.CoreV1().Pods(namespace).Get(pods[0].Name, metav1.GetOptions{})
			if err != nil {
				ExitfIfError(err, "Couldn't fetch a pod")
			}
			var s corev1.ContainerStatus
			for _, s = range pod.Status.InitContainerStatuses {
				if s.Name == container.Name {
					break
				}
			}
			if s.Name == "" {
				time.Sleep(1 * time.Second)
				continue
			} else {
				if s.State.Waiting == nil {
					break
				}
			}
		}
		err := tailContainerLogs(client, namespace, pods[0].Name, container.Name)
		ExitfIfError(err, "failed to get build logs")
	}
}

func tailContainerLogs(client kubernetes.Interface, namespace, pod, container string) error {
	var stream io.ReadCloser

	req := client.CoreV1().RESTClient().Get().
		Namespace(namespace).
		Name(pod).
		Resource("pods").
		SubResource("log").
		Param("follow", strconv.FormatBool(true)).
		Param("container", container).
		Param("previous", strconv.FormatBool(false)).
		Param("timestamps", strconv.FormatBool(false))

	for {
		var err error
		stream, err = req.Stream()
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}

	defer stream.Close()
	reader := bufio.NewReader(stream)
	for {
		line, err := reader.ReadBytes('\n')
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		_, err = fmt.Printf("[%s]: %s\n", container, strings.TrimSpace(string(line)))
		if err != nil {
			return err
		}
	}

	return nil
}
