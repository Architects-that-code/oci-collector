package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"oci-collector/autonomous"
	"oci-collector/config"
	"oci-collector/util"
)

var autonomousCmd = &cobra.Command{
	Use:   "autonomous",
	Short: "List Autonomous Databases across subscribed regions and compartments",
	Long:  `Find and list all Autonomous Databases (ATP/ADW/AJD) across your subscribed regions and all compartments in the tenancy.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Getconfig()
		if err != nil {
			util.FatalIfError(err)
		}

		provider, client, tenancyID, err := config.Prep(cfg)
		if err != nil {
			util.FatalIfError(err)
		}

		regions, compartments, _, _ := config.CommonSetup(client, tenancyID)

		items := autonomous.ListAutonomousDatabases(provider, regions, compartments)
		fmt.Println("Region | Compartment | Name | Details | State")
		autonomous.PrintAutonomousDatabases(items)
	},
}

func init() {
	// no flags for now; prints a simple list
}
