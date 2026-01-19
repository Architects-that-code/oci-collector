package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "oci-collector",
	Short: "A utility belt for OCI tenancy management",
	Long: `This is designed to be a loose collection of tools to help manage and monitor your OCI tenancy.
		
usage: oci-collector [command] [flags]
		
Specify the action you want to take:

limits: fetch limits in all YOUR regions
compute: fetch compute active instances in all YOUR regions
config: check config file and print basic info
peeps: fetch user counts (-r to show users)
policies: fetch policy counts (-run to show policies -verbose to show statements)
groups: fetch group counts (-run to show groups )
support: fetch support tickets (-list to show tickets)
capacity: check capacity in all YOUR regions (-ocpus to specify ocpus -memory to specify memory -type to specify shape type)
capability: what types of 'things' are available for a shape type (-type to specify shape type)
children: dealing with child tenancies
object: fetch object storage info
billing: fetch billing info
network: network related info
search: search for resources created by a user
schedule: fetch schedule
`,
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