package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"oci-collector/config"
	"oci-collector/util"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Check config file and print basic info",
	Long:  `Display configuration details like regions, compartments, ADs.`,
	Run: func(cmd *cobra.Command, args []string) {
		verbose, _ := cmd.Flags().GetBool("verbose")
		cfg, err := config.Getconfig()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		_, client, tenancyID, err := config.Prep(cfg)
		if err != nil {
			util.FatalIfError(err)
		}
		regions, compartments, ads, _ := config.CommonSetup(client, tenancyID)
		if regions == nil {
			fmt.Println("regions is nil")
		} else {
			fmt.Printf("subscribed regions: %v\n", len(regions))
			if verbose {
				for _, region := range regions {
					fmt.Printf("\tRegionKey: %v, name: %v \n", strings.ToLower(*region.RegionKey), *region.RegionName)
				}
			}
		}
		if ads == nil {
			fmt.Println("ads is nil")
		} else {
			fmt.Printf("ads: %v\n", len(ads))
			if verbose {
				for _, ad := range ads {
					fmt.Printf("\tAD: %v\n", *ad.Name)
				}
			}
		}
		if compartments == nil {
			fmt.Println("compartments is nil")
		} else {

			fmt.Printf("compartments: %v\n", len(compartments))
			if verbose {
				for _, comp := range compartments {
					fmt.Printf("\tCompartment Name: %v \n", *comp.Name)

				}
			}
			//fmt.Printf("compartments: %v\n", compartments)
		}
	},
}

func init() {
	configCmd.Flags().BoolP("verbose", "v", false, "get more details")
}