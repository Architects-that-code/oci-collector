package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"oci-collector/compute"
	"oci-collector/config"
	"oci-collector/util"
)

var computeCmd = &cobra.Command{
	Use:   "compute",
	Short: "Fetch compute active instances in all YOUR regions",
	Long:  `Fetch active compute instances across subscribed regions and compartments. Optionally collect utilization metrics for the last day, week, and month.`,
	Run: func(cmd *cobra.Command, args []string) {
		run, _ := cmd.Flags().GetBool("run")
		metrics, _ := cmd.Flags().GetBool("metrics")
		enableMetrics, _ := cmd.Flags().GetBool("enable-metrics")
		discover, _ := cmd.Flags().GetBool("metrics-discover")
		discoverWindow, _ := cmd.Flags().GetString("discover-window")
		discoverInstance, _ := cmd.Flags().GetString("discover-instance")
		// metricsTypes, _ := cmd.Flags().GetStringSlice("metrics-types") // TODO: implement multiple metrics types
		format, _ := cmd.Flags().GetString("format")
		output, _ := cmd.Flags().GetString("out")

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

		if enableMetrics {
			err := compute.EnableMetrics(provider, regions, compartments)
			if err != nil {
				util.FatalIfError(err)
			}
			fmt.Println("Oracle Cloud Agent metric collection enabled on applicable instances")
			return
		}

		if discover {
			// parse lookback
			lb := time.Hour
			if discoverWindow != "" {
				if d, err := time.ParseDuration(discoverWindow); err == nil {
					lb = d
				}
			}

			summaries, err := compute.DiscoverMetrics(provider, regions, compartments, discoverInstance, lb)
			if err != nil {
				util.FatalIfError(err)
			}
			// print concise report to stdout
			for _, s := range summaries {
				if s.Found {
					metricsList := strings.Join(s.AvailableMetrics, ", ")
					fmt.Printf("FOUND %s | %s | %s | ns=%s metric=%s dim=%s rg=%s samples=%d | available: %s\n", s.Region, s.Compartment, s.Name, s.Namespace, s.Metric, s.DimensionKey, s.ResourceGroup, s.Samples, metricsList)
				} else {
					fmt.Printf("NONE  %s | %s | %s | %s\n", s.Region, s.Compartment, s.Name, s.Note)
				}
			}
			return
		}

		var instances []compute.InstanceInventory
		if run || metrics {
			instances = compute.GatherInstances(provider, regions, compartments, run)
		}

		if metrics {
			summaries, err := compute.CollectMetrics(provider, regions, compartments, instances)
			if err != nil {
				util.FatalIfError(err)
			}
			// default to json
			switch strings.ToLower(format) {
			case "csv":
				data := compute.ToCSV(summaries)
				if output != "" {
					if err := os.WriteFile(output, []byte(data), 0644); err != nil {
						util.FatalIfError(err)
					}
					fmt.Printf("metrics written to %s\n", output)
				} else {
					fmt.Print(data)
				}
			default:
				data, err := compute.ToJSON(summaries)
				util.FatalIfError(err)
				if output != "" {
					if err := os.WriteFile(output, []byte(data), 0644); err != nil {
						util.FatalIfError(err)
					}
					fmt.Printf("metrics written to %s\n", output)
				} else {
					fmt.Println(data)
				}
			}
		}

		if run {
			// Only change output if user provided -f/-o; otherwise keep default text output
			runFormat := "text"
			runOut := ""
			if cmd.Flags().Changed("format") {
				runFormat = format
			}
			if cmd.Flags().Changed("out") {
				runOut = output
				// If user asked to write output but didn't choose a format, default to json
				if !cmd.Flags().Changed("format") {
					runFormat = "json"
				}
			}
			if err := compute.RunCompute(provider, regions, tenancyID, compartments, runFormat, runOut); err != nil {
				util.FatalIfError(err)
			}
		}

		if !run && !metrics {
			fmt.Println("add -run to list instances, or use --metrics to collect utilization metrics")
		}
	},
}

func init() {
	computeCmd.Flags().BoolP("run", "r", false, "fetch compute active instances in all regions")
	computeCmd.Flags().BoolP("metrics", "m", false, "collect CPU utilization metrics (day/week/month) for running instances")
	computeCmd.Flags().Bool("enable-metrics", false, "enable Oracle Cloud Agent metric collection on instances")
	computeCmd.Flags().Bool("metrics-discover", false, "probe Monitoring to find the namespace/dimension that returns datapoints per instance")
	computeCmd.Flags().String("discover-window", "1h", "lookback duration for discovery (e.g., 1h, 24h, 7d)")
	computeCmd.Flags().String("discover-instance", "", "optional instance OCID to limit discovery")
	computeCmd.Flags().StringP("format", "f", "json", "output format: for metrics (default json), for run specify json or csv; omit for text output")
	computeCmd.Flags().StringP("out", "o", "", "optional file path to write output (metrics or run)")
	computeCmd.Flags().StringSlice("metrics-types", []string{"CpuUtilization"}, "metrics to collect (cpu,memory,disk,network); default cpu")
}
