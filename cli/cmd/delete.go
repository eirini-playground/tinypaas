package cmd

import (
	"fmt"

	kpack "github.com/pivotal/kpack/pkg/client/clientset/versioned"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
)

func deleteapp(cmd *cobra.Command, args []string) {
	name, err := cmd.Flags().GetString("name")
	ExitfIfError(err, "Failed to get name flag")

	config := parseConfig(cmd)
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", config.KubeconfigPath)
	ExitfIfError(err, "an unexpected error occurred")

	client, err := kpack.NewForConfig(kubeConfig)
	ExitfIfError(err, "Failed to create a kpack client")

	propagationPolicy := metav1.DeletePropagationBackground
	err = client.KpackV1alpha1().Images(config.Namespace).Delete(name, &metav1.DeleteOptions{PropagationPolicy: &propagationPolicy})
	ExitfIfError(err, "Couldn't delete app")

	fmt.Printf("Deleted app %v", name)
}
