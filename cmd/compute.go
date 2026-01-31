package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"oci-collector/compute"
	"oci-collector/config"
	"oci-collector/util"
)

var computeCmd = &cobra.Command{
	Use:   "compute",
	Short: "Fetch compute active instances in all YOUR regions",
	Long:  `Fetch active compute instances across subscribed regions and compartments.`,
	Run: func(cmd *cobra.Command, args []string) {
		run, _ := cmd.Flags().GetBool("run")
		if run {
			cfg, err := config.Getconfig()
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			provider, client, tenancyID, err := config.Prep(cfg)
			if err != nil {
				util.FatalIfError(err)
			}
			regions, compartments, _, _ := config.CommonSetup(client, tenancyID)
			compute.RunCompute(provider, regions, tenancyID, compartments)
		} else {
			fmt.Println("add -run to run")
		}
	},
}

func init() {
	computeCmd.Flags().BoolP("run", "r", false, "fetch compute active instances in all regions")
}