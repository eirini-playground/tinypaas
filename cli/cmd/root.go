package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func Execute() {
	rootCmd := &cobra.Command{
		Use:   "tinypaas",
		Short: "A tiny PaaS based on kpack and Eirini",
		Long:  `The fastest (and simplest) way to get your application deployed on Kubernetes.`,
	}

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number of TinyPaas",
		Long:  `All software has versions. This is Tinypaas version`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("TinyPaas v0.0 -- HEAD")
		},
	}

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new application",
		Long:  `Returns a git remote where the application should be pushed for deploymend`,
		Run:   create,
	}

	createCmd.Flags().StringP("name", "n", "", "The name of the application")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(createCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
