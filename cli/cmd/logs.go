package cmd

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"

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
	wg := &sync.WaitGroup{}
	allContainers := []corev1.Container{}
	allContainers = append(allContainers, pods[0].Spec.Containers...)
	allContainers = append(allContainers, pods[0].Spec.InitContainers...)
	for _, container := range allContainers {
		wg.Add(1)
		// TODO: we should collect all logs through a channel to be thread safe: https://stackoverflow.com/a/14694630
		// TODO: Only start streaming a container when it passes the "Waiting" state (otherwise streaming will exit early).
		// TODO: Exit if a container fails with error ? (e.g. if an init container fails, the next one will stay in Waiting state forever).
		go tailContainerLogs(wg, client, namespace, pods[0].Name, container.Name)
	}
	defer wg.Wait()
}

func tailContainerLogs(wg *sync.WaitGroup, client kubernetes.Interface, namespace, pod, container string) error {
	defer wg.Done()

	req := client.CoreV1().RESTClient().Get().
		Namespace(namespace).
		Name(pod).
		Resource("pods").
		SubResource("log").
		Param("follow", strconv.FormatBool(true)).
		Param("container", container).
		Param("previous", strconv.FormatBool(false)).
		Param("timestamps", strconv.FormatBool(false))
	stream, err := req.Stream()
	if err != nil {
		return err
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
