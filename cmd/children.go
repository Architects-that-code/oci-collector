package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	children "oci-collector/childtenancies"
	"oci-collector/config"
	"oci-collector/util"
)

var childrenCmd = &cobra.Command{
	Use:   "children",
	Short: "Dealing with child tenancies",
	Long:  `Fetch child tenancy information.`,
	Run: func(cmd *cobra.Command, args []string) {
		run, _ := cmd.Flags().GetBool("run")
		write, _ := cmd.Flags().GetBool("write")
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
		children.Children(provider, client, tenancyID, run, homeregion, cfg, write)
	},
}

func init() {
	childrenCmd.Flags().BoolP("run", "r", false, "fetch child tenancies")
	childrenCmd.Flags().BoolP("write", "w", false, "write to file")
}