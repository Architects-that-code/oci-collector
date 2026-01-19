package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	peopleresource "oci-collector/iam"
	"oci-collector/config"
	"oci-collector/util"
)

var policiesCmd = &cobra.Command{
	Use:   "policies",
	Short: "Fetch policy counts",
	Long:  `Fetch policy information across compartments.`,
	Run: func(cmd *cobra.Command, args []string) {
		run, _ := cmd.Flags().GetBool("run")
		verbose, _ := cmd.Flags().GetBool("verbose")
		cfg, err := config.Getconfig()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		provider, client, tenancyID, err := config.Prep(cfg)
		if err != nil {
			util.FatalIfError(err)
		}
		_, compartments, _, _ := config.CommonSetup(client, tenancyID)
		peopleresource.GetAllPolicies(provider, client, tenancyID, compartments, run, verbose)
	},
}

func init() {
	policiesCmd.Flags().BoolP("run", "r", false, "fetch policy")
	policiesCmd.Flags().BoolP("verbose", "v", false, "show policies")
}