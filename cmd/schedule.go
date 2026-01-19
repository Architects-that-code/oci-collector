package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	scheduler "oci-collector/schedule"
	"oci-collector/config"
	"oci-collector/util"
)

var scheduleCmd = &cobra.Command{
	Use:   "schedule",
	Short: "Fetch schedule",
	Long:  `Run the scheduler for OCI resources.`,
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
			regions, compartments, _, homeregion := config.CommonSetup(client, tenancyID)
			scheduler.RunSchedule(provider, regions, tenancyID, compartments, homeregion)
		} else {
			fmt.Println("add -run to run")
		}
	},
}

func init() {
	scheduleCmd.Flags().BoolP("run", "r", false, "fetch schedule")
}