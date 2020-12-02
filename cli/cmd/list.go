package cmd

import (
	"fmt"

	kpack "github.com/pivotal/kpack/pkg/client/clientset/versioned"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
)

func list(cmd *cobra.Command, args []string) {
	config := parseConfig(cmd)
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", config.KubeconfigPath)
	ExitfIfError(err, "an unexpected error occurred")

	client, err := kpack.NewForConfig(kubeConfig)
	ExitfIfError(err, "Failed to create a kpack client")

	images, err := client.KpackV1alpha1().Images(config.Namespace).List(metav1.ListOptions{})
	ExitfIfError(err, "Couldn't list apps")

	if len(images.Items) == 0 {
		fmt.Println("No apps")
		return
	}

	for _, image := range images.Items {
		fmt.Printf("%-32s%s\n", "Name", "Url")
		fmt.Printf("--------------------------------------------------------------------------\n")
		fmt.Printf("%-32s%s.%s.vcap.me\n", image.Name, image.Name, image.Namespace)
	}
}
