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

	stagingPods, err := client.CoreV1().Pods(config.Namespace).List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("image.kpack.io/image=%s", name),
	})
	ExitfIfError(err, "Couldn't get staging logs for app")

	printLogs(client, config.Namespace, stagingPods.Items)
}

func printLogs(client kubernetes.Interface, namespace string, pods []corev1.Pod) {
	if len(pods) == 0 {
		fmt.Println("No logs found")
		return
	}

	pod := selectLatestPod(pods)
	allContainers := []corev1.Container{}
	allContainers = append(allContainers, pod.Spec.InitContainers...)
	allContainers = append(allContainers, pod.Spec.Containers...)
	fmt.Printf("Showing logs for build #%s\n", pod.Labels["image.kpack.io/buildNumber"])
	for _, container := range allContainers {
		err := tailContainerLogs(client, namespace, pod.Name, container.Name)
		ExitfIfError(err, "failed to get build logs")
	}
}

func selectLatestPod(pods []corev1.Pod) corev1.Pod {
	pod := pods[0]
	lastBuildNumber := 0
	for i, p := range pods {
		currentBuild, err := strconv.Atoi(p.Labels["image.kpack.io/buildNumber"])
		if err != nil {
			panic(fmt.Sprintf("failed to parse build number: %s", p.Labels["image.kpack.io/buildNumber"]))
		}

		if currentBuild > lastBuildNumber {
			pod = pods[i]
			lastBuildNumber = currentBuild
		}
	}

	return pod
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
