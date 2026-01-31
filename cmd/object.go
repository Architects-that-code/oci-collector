package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	oos "oci-collector/objectstorage"
	"oci-collector/config"
	"oci-collector/util"
)

var objectCmd = &cobra.Command{
	Use:   "object",
	Short: "Fetch object storage info",
	Long:  `Fetch object storage buckets and sizes across regions.`,
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
		regions, compartments, _, homeregion := config.CommonSetup(client, tenancyID)
		oos.ObjectStorageSize(provider, regions, tenancyID, compartments, run, homeregion)
	},
}

func init() {
	objectCmd.Flags().BoolP("run", "r", false, "fetch object storage")
}