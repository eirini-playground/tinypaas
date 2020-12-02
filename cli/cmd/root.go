// Package cmd is the entrance to the wonderful world of tinypaas
package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type Config struct {
	KubeconfigPath           string `yaml:"kubeconfig_path"`
	Namespace                string `yaml:"namespace"`
	GitSecretName            string `yaml:"git_secret_name"`
	DockerRegistrySecretName string `yaml:"docker_registry_secret_name"`
	BuilderName              string `yaml:"builder_name"`
	KpackServiceAccount      string `yaml:"kpack_service_account"`
}

func Execute() {
	rootCmd := &cobra.Command{
		Use:   "tinypaas",
		Short: "A tiny PaaS based on kpack and Eirini",
		Long:  `The fastest (and simplest) way to get your application deployed on Kubernetes.`,
	}

	rootCmd.PersistentFlags().StringP("config", "c", "", "The path to tinypaas config file")

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
		Long:  `Returns a git remote where the application should be pushed for deployment`,
		Run:   create,
	}

	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Deletes the specified application",
		Long:  "Deletes the specified application and cleans up the cluster from all traces of this app",
		Run:   deleteapp,
	}

	gitPublicKeyCmd := &cobra.Command{
		Use:   "git-public-key",
		Short: "Prints the git public key",
		Long:  "Prints the public key which needs to be authorized to fetch your code.",
		Run:   getGitPublicKey,
	}

	createCmd.Flags().StringP("name", "n", "", "The name of the application")
	createCmd.Flags().StringP("git-url", "u", "", "The git remote (the SSH url)")
	createCmd.Flags().StringP("git-branch", "b", "tiny-paas", "The git branch to track (default tiny-paas)")
	createCmd.Flags().StringP("docker-repo", "d", "", "Name of the docker repository")

	deleteCmd.Flags().StringP("name", "n", "", "The name of the application to delete")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(gitPublicKeyCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func parseConfig(rootCmd *cobra.Command) (config Config) {
	configPath, err := rootCmd.Flags().GetString("config")
	ExitfIfError(err, "Couldn't parse the config argument")

	configContents, err := ioutil.ReadFile(configPath)
	ExitfIfError(err, "Couldn't read the config file")

	err = yaml.Unmarshal(configContents, &config)
	ExitfIfError(err, "Couldn't parse the config file")

	// TODO: validate struct contents
	// e.g. mandatory fields: builder_name, kubeconfig_path, ...

	return
}
