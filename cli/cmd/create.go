package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func create(cmd *cobra.Command, args []string) {
	name, err := cmd.Flags().GetString("name")
	ExitfIfError(err, "Failed to get name flag")

	kubeconfig, err := cmd.Flags().GetString("kubeconfig")
	ExitfIfError(err, "Failed to get kubeconfig flag")

	_ = CreateKubeClient(kubeconfig)

	fmt.Printf("Creating app %v", name)
}
