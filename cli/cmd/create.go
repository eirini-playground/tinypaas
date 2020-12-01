package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func create(cmd *cobra.Command, args []string) {
	name, err := cmd.Flags().GetString("name")
	ExitfIfError(err, "Failed to get name flag")

	config := parseConfig(cmd)
	_ = CreateKubeClient(config.KubeconfigPath)

	fmt.Printf("Creating app %v", name)
}
