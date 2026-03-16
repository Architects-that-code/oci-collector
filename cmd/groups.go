package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"oci-collector/config"
	peopleresource "oci-collector/iam"
	"oci-collector/util"
)

var groupsCmd = &cobra.Command{
	Use:   "groups",
	Short: "Fetch group counts",
	Long:  `Fetch group information in the tenancy.`,
	Run: func(cmd *cobra.Command, args []string) {
		run, _ := cmd.Flags().GetBool("run")
		if !run {
			_ = cmd.Help()
			return
		}
		cfg, err := config.Getconfig()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		provider, client, tenancyID, err := config.Prep(cfg)
		if err != nil {
			util.FatalIfError(err)
		}
		peopleresource.Groups(provider, client, tenancyID, true)
	},
}

func init() {
	groupsCmd.Flags().BoolP("run", "r", false, "fetch groups")
}
