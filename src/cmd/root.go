package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"limits"
	"compute"
	"config"
	"peeps"
	"policies"
	"groups"
	"support"
	"network"
	"capacity"
	"capability"
	"children"
	"object"
	"billing"
	"schedule"
	"search"
)

// rootCmd represents the base command when called without any subcommands
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
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(limits.Cmd)
	rootCmd.AddCommand(compute.Cmd)
	rootCmd.AddCommand(config.Cmd)
	rootCmd.AddCommand(peeps.Cmd)
	rootCmd.AddCommand(policies.Cmd)
	rootCmd.AddCommand(groups.Cmd)
	rootCmd.AddCommand(support.Cmd)
	rootCmd.AddCommand(network.Cmd)
	rootCmd.AddCommand(capacity.Cmd)
	rootCmd.AddCommand(capability.Cmd)
	rootCmd.AddCommand(children.Cmd)
	rootCmd.AddCommand(object.Cmd)
	rootCmd.AddCommand(billing.Cmd)
	rootCmd.AddCommand(schedule.Cmd)
	rootCmd.AddCommand(search.Cmd)

	// Here you will define your flags and configuration settings.
	// Cobra supports Persistent Flags which, if defined here,
	// will be global for your application.

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}