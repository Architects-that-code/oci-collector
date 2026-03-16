package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"oci-collector/config"
	supportresources "oci-collector/support"
	"oci-collector/util"
)

var supportCmd = &cobra.Command{
	Use:   "support",
	Short: "Fetch support tickets",
	Long:  `Support actions. Use --list to list support tickets using configured CSI.`,
	Run: func(cmd *cobra.Command, args []string) {
		list, _ := cmd.Flags().GetBool("list")
		if !list {
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
		_, _, _, homeregion := config.CommonSetup(client, tenancyID)
		supportresources.ListTickets(provider, tenancyID, homeregion, cfg.CSI)
	},
}

func init() {
	supportCmd.Flags().BoolP("list", "l", false, "list tickets")
}
