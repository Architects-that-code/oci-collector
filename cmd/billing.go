package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"oci-collector/billing"
	"oci-collector/config"
	"oci-collector/util"
)

var billingCmd = &cobra.Command{
	Use:   "billing",
	Short: "Fetch billing info",
	Long:  `Manage billing reports. Use one or more action flags: --download, --process, --redownload-errors.`,
	Run: func(cmd *cobra.Command, args []string) {
		path, _ := cmd.Flags().GetString("path")
		download, _ := cmd.Flags().GetBool("download")
		process, _ := cmd.Flags().GetBool("process")
		redownloadErrors, _ := cmd.Flags().GetBool("redownload-errors")
		if !download && !redownloadErrors && !process {
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
		if download {
			billing.Getfiles(provider, tenancyID, homeregion, cfg, path, true)
		}
		if redownloadErrors {
			fmt.Println("Re-downloading error files...")
			err := billing.RedownloadErrorFiles(provider, tenancyID, homeregion, cfg, path)
			if err != nil {
				fmt.Printf("Error re-downloading error files: %v\n", err)
				os.Exit(1)
			} else {
				fmt.Println("Re-download of error files completed.")
			}
		}
		if process {
			fmt.Println("Processing billing files...")
			err := billing.ProcessBillingFiles(path, cfg.ProfileName)
			if err != nil {
				fmt.Printf("Error processing billing files: %v\n", err)
				os.Exit(1)
			} else {
				fmt.Println("Billing files processed successfully.")
			}
		}
	},
}

func init() {
	billingCmd.Flags().StringP("path", "p", "reports", "path to save billing files - default is ./reports")
	billingCmd.Flags().BoolP("download", "d", false, "download billing files")
	billingCmd.Flags().BoolP("process", "x", false, "process downloaded billing files")
	billingCmd.Flags().BoolP("redownload-errors", "e", false, "re-download files listed in error_files.txt")
}
