package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd is the root Cobra command for the OCI Collector CLI.
var rootCmd = &cobra.Command{
	Use:   "oci-collector",
	Short: "A utility belt for OCI tenancy management",
	Long:  `A loose collection of tools to help manage and monitor your OCI tenancy. Use the commands below to perform specific actions.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

// Execute runs the root command.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// init adds all subcommands to rootCmd and sets up flags.
func init() {
	rootCmd.AddCommand(limitsCmd)
	rootCmd.AddCommand(computeCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(peepsCmd)
	rootCmd.AddCommand(policiesCmd)
	rootCmd.AddCommand(groupsCmd)
	rootCmd.AddCommand(supportCmd)
	rootCmd.AddCommand(capacityCmd)
	rootCmd.AddCommand(capabilityCmd)
	rootCmd.AddCommand(childrenCmd)
	rootCmd.AddCommand(objectCmd)
	rootCmd.AddCommand(billingCmd)
	rootCmd.AddCommand(networkCmd)
	rootCmd.AddCommand(scheduleCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(cloudAdvisorCmd)
	rootCmd.AddCommand(autonomousCmd)

}
