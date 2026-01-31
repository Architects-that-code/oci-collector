package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"oci-collector/limits"
	"oci-collector/config"
	"oci-collector/util"
)

var limitsCmd = &cobra.Command{
	Use:   "limits",
	Short: "Fetch limits in all YOUR regions",
	Long:  `Fetch service limits across subscribed regions.`,
	Run: func(cmd *cobra.Command, args []string) {
		run, _ := cmd.Flags().GetBool("run")
		write, _ := cmd.Flags().GetBool("write")
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
			regions, _, _, _ := config.CommonSetup(client, tenancyID)
			limits.RunLimits(provider, regions, tenancyID, write)
		} else {
			fmt.Println("add -run to run")
		}
	},
}

func init() {
	limitsCmd.Flags().BoolP("run", "r", false, "fetch limits in all regions")
	limitsCmd.Flags().BoolP("write", "w", false, "write limits to file")
}