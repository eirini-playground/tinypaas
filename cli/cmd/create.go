package cmd

import (
	"fmt"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	kpack "github.com/pivotal/kpack/pkg/client/clientset/versioned"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
)

func create(cmd *cobra.Command, args []string) {
	name, err := cmd.Flags().GetString("name")
	ExitfIfError(err, "Failed to get name flag")

	repo, err := cmd.Flags().GetString("docker-repo")
	ExitfIfError(err, "Failed to get docker-repo flag")

	url, err := cmd.Flags().GetString("git-url")
	ExitfIfError(err, "Failed to get git-url flag")

	branch, err := cmd.Flags().GetString("git-branch")
	ExitfIfError(err, "Failed to get git-branch flag")

	config := parseConfig(cmd)
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", config.KubeconfigPath)
	ExitfIfError(err, "an unexpected error occurred")

	client, err := kpack.NewForConfig(kubeConfig)
	ExitfIfError(err, "Failed to create a kpack client")

	failIfImageExists(client, name, config.Namespace)

	image := v1alpha1.Image{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1alpha1.ImageSpec{
			Tag: repo,
			Builder: corev1.ObjectReference{
				Kind:      "Builder",
				Namespace: config.Namespace,
				Name:      config.BuilderName,
			},
			ServiceAccount: config.KpackServiceAccount,
			Source: v1alpha1.SourceConfig{
				Git: &v1alpha1.Git{
					URL:      url,
					Revision: branch,
				},
			},
		},
	}
	_, err = client.KpackV1alpha1().Images(config.Namespace).Create(&image)
	ExitfIfError(err, "Couldn't create a new Image for kpack")

	fmt.Printf("Created app %v", name)
}

func failIfImageExists(client *kpack.Clientset, name, namespace string) {
	images, err := client.KpackV1alpha1().Images(namespace).List(metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", name),
	})
	ExitfIfError(err, "couldn't list kpack images")

	if len(images.Items) > 0 {
		ExitfWithMessage("image with name %q already exists in namespace %s", name, namespace)
	}
}
