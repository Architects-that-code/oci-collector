package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "oci-collector",
	Short: "A utility belt for OCI tenancy management",
    Long: `A loose collection of tools to help manage and monitor your OCI tenancy. Use the commands below to perform specific actions.`,
	RunE: func(cmd *cobra.Command, args []string) error {
	return cmd.Help()
},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

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

	// Here you will define your flags and configuration settings.
	// Cobra supports Persistent Flags which, if defined here,
	// will be global for your application.

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}