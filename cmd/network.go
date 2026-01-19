package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	network "oci-collector/networks"
	"oci-collector/config"
	"oci-collector/util"
)

var networkCmd = &cobra.Command{
	Use:   "network",
	Short: "Network related info",
	Long:  `Fetch VCN and network information across regions.`,
	Run: func(cmd *cobra.Command, args []string) {
		run, _ := cmd.Flags().GetBool("run")
		cidr, _ := cmd.Flags().GetBool("cidr")
		ip, _ := cmd.Flags().GetBool("ip")
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
			network.GetAllVcn(provider, regions, tenancyID, compartments, run, cidr, ip)
		} else {
			fmt.Println("add -run to run")
		}
	},
}

func init() {
	networkCmd.Flags().BoolP("run", "r", false, "fetch all vcn in all regions")
	networkCmd.Flags().BoolP("cidr", "c", false, "also fetch CIDR blocks")
	networkCmd.Flags().BoolP("ip", "i", false, "fetch IP inventory")
}