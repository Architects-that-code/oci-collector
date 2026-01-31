package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"oci-collector/capability"
	"oci-collector/config"
	"oci-collector/util"
)

var capabilityCmd = &cobra.Command{
	Use:   "capability",
	Short: "What types of 'things' are available for a shape type",
	Long:  `Check capabilities for specified shape type.`,
	Run: func(cmd *cobra.Command, args []string) {
		run, _ := cmd.Flags().GetBool("run")
		shapeType, _ := cmd.Flags().GetString("type")
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
		capability.OSSupport(provider, regions, tenancyID, compartments, run, shapeType)
	},
}

func init() {
	capabilityCmd.Flags().BoolP("run", "r", false, "fetch capability")
	capabilityCmd.Flags().StringP("type", "t", "E4", "use shape type E3, E4, E5, X9, A1")
}