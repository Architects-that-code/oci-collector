package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	supportresources "oci-collector/support"
	"oci-collector/config"
	"oci-collector/util"
)

var supportCmd = &cobra.Command{
	Use:   "support",
	Short: "Fetch support tickets",
	Long:  `Fetch support ticket information using CSI.`,
	Run: func(cmd *cobra.Command, args []string) {
		// list flag currently unused
		cfg, err := config.Getconfig()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		provider, client, tenancyID, err := config.Prep(cfg)
		if err != nil {
			util.FatalIfError(err)
		}
		_, _, _, homeregion := config.CommonSetup(client, tenancyID)
		supportresources.ListTickets(provider, tenancyID, homeregion, cfg.CSI)
	},
}

func init() {
	supportCmd.Flags().BoolP("list", "l", false, "list tickets")
}