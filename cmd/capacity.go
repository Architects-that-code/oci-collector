package cmd

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	capcheck "oci-collector/capCheck"
	"oci-collector/config"
	"oci-collector/util"
)

var capacityCmd = &cobra.Command{
	Use:   "capacity",
	Short: "Check capacity in all YOUR regions",
	Long:  `Check capacity for specified OCPUs, memory, and shape type.`,
	Run: func(cmd *cobra.Command, args []string) {
		run, _ := cmd.Flags().GetBool("run")
		ocpus, _ := cmd.Flags().GetInt("ocpus")
		memory, _ := cmd.Flags().GetInt("memory")
		shapeType, _ := cmd.Flags().GetString("type")
		// flags defined but not used currently: ad, fd
		if run || ocpus > 0 || memory > 0 {
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
			chipsetSlice := []string{"AMD", "INTEL", "ARM"}
			if ocpus > 0 || memory > 0 {
				if !slices.Contains(chipsetSlice, strings.ToUpper(shapeType)) {
					capcheck.Check(provider, regions, tenancyID, compartments, run, ocpus, memory, shapeType)
				} else {
					capcheck.CheckFAMILY(provider, regions, tenancyID, compartments, run, ocpus, memory, shapeType)
				}
			} else {
				fmt.Println("add -ocpus and -memory -type  to run - NOTE: for -type you can use shape type 'E3', 'E4', 'E5', 'E6', 'X9', 'A1', 'A2' --- OR the ChipSet family 'AMD', 'Intel', 'ARM'")
			}
		} else {
			fmt.Println("add -ocpus and -memory -type  to run - NOTE: for -type you can use shape type 'E3', 'E4', 'E5', 'E6', 'X9', 'A1', 'A2' --- OR the ChipSet family 'AMD', 'Intel', 'ARM'")
		}
	},
}

func init() {
	capacityCmd.Flags().BoolP("run", "r", false, "fetch capacity")
	capacityCmd.Flags().IntP("ocpus", "o", 0, "number of ocpus")
	capacityCmd.Flags().IntP("memory", "m", 0, "amount of memory")
	capacityCmd.Flags().StringP("type", "t", "E4", "use shape type E3, E4, E5, X9, A1, A2 --- OR the ChipSet family AMD, Intel, ARM")
	capacityCmd.Flags().StringP("ad", "a", "", "availability domain")
	capacityCmd.Flags().StringP("fd", "f", "", "fault domain")
}