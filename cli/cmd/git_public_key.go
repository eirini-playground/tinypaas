package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getGitPublicKey(cmd *cobra.Command, args []string) {
	config := parseConfig(cmd)

	client := CreateKubeClient(config.KubeconfigPath)
	secret, err := client.CoreV1().Secrets(config.Namespace).Get(config.GitSecretName, metav1.GetOptions{})
	ExitfIfError(err, "Couldn't find the git secret in your cluster")

	fmt.Printf("Authorise this SSH key to pull from your repository:\n\n%s", string(secret.Data["ssh-publickey"]))
}
