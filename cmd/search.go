package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	resourcesearch "oci-collector/search"
	"oci-collector/config"
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search for resources",
	Long:  `Search for resources created by a user using a search string.`,
	Run: func(cmd *cobra.Command, args []string) {
		searchString, _ := cmd.Flags().GetString("searchstring")
		if searchString == "" {
			fmt.Println("add --searchstring to run")
			return
		}
		cfg, err := config.Getconfig()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		provider, _, tenancyID, _ := config.Prep(cfg)
		resourcesearch.Search(provider, tenancyID, searchString)
	},
}

func init() {
	searchCmd.Flags().StringP("searchstring", "s", "", "search string")
}