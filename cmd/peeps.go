package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	peopleresource "oci-collector/iam"
	"oci-collector/config"
	"oci-collector/util"
)

var peepsCmd = &cobra.Command{
	Use:   "peeps",
	Short: "Fetch user counts",
	Long:  `Fetch user information in the tenancy.`,
	Run: func(cmd *cobra.Command, args []string) {
		run, _ := cmd.Flags().GetBool("run")
		cfg, err := config.Getconfig()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		provider, client, tenancyID, err := config.Prep(cfg)
		if err != nil {
			util.FatalIfError(err)
		}
		peopleresource.GetAllPeople(provider, client, tenancyID, run)
	},
}

func init() {
	peepsCmd.Flags().BoolP("run", "r", false, "fetch users")
}