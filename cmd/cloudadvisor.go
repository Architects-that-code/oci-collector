package cmd

import (
    "encoding/json"
    "fmt"
    "os"
    "strings"

    "github.com/spf13/cobra"

    "oci-collector/cloudadvisor"
    "oci-collector/config"
    "oci-collector/util"
    "github.com/oracle/oci-go-sdk/v65/optimizer"
)

var cloudAdvisorCmd = &cobra.Command{
    Use:   "cloudadvisor",
    Short: "Fetch Cloud Advisor recommendations (home region)",
    Long:  `Collect all Cloud Advisor (Optimizer) recommendations for the tenancy. Cloud Advisor is a home-region service.`,
    Run: func(cmd *cobra.Command, args []string) {
        format, _ := cmd.Flags().GetString("format")
        output, _ := cmd.Flags().GetString("out")
        includeActions, _ := cmd.Flags().GetBool("actions")
        includeOrg, _ := cmd.Flags().GetBool("org")
        childCSV, _ := cmd.Flags().GetString("child-tenancies")
        var childTenancyIDs []string
        if strings.TrimSpace(childCSV) != "" {
            for _, p := range strings.Split(childCSV, ",") {
                v := strings.TrimSpace(p)
                if v != "" {
                    childTenancyIDs = append(childTenancyIDs, v)
                }
            }
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
        // Only need home region
        _, _, _, home := config.CommonSetup(client, tenancyID)

        recs, err := cloudadvisor.ListAllRecommendations(provider, tenancyID, home, includeOrg && len(childTenancyIDs) == 0, childTenancyIDs)
        if err != nil {
            util.FatalIfError(err)
        }

        var actions interface{} = nil
        if includeActions {
            acts, err := cloudadvisor.ListAllResourceActions(provider, tenancyID, home, "", includeOrg && len(childTenancyIDs) == 0, childTenancyIDs)
            if err != nil { util.FatalIfError(err) }
            actions = acts
        }

        switch strings.ToLower(format) {
        case "json":
            // if actions requested, wrap output to include both arrays
            if includeActions {
                payload := map[string]interface{}{
                    "homeRegion": home,
                    "recommendations": recs,
                }
                if len(childTenancyIDs) > 0 {
                    payload["childTenancies"] = childTenancyIDs
                } else if includeOrg {
                    payload["includeOrganization"] = true
                }
                if actions != nil {
                    payload["actions"] = actions
                }
                b, err := json.MarshalIndent(payload, "", "  ")
                if err != nil { util.FatalIfError(err) }
                if output != "" {
                    if err := os.WriteFile(output, b, 0644); err != nil { util.FatalIfError(err) }
                    fmt.Printf("cloudadvisor data written to %s\n", output)
                } else {
                    fmt.Println(string(b))
                }
                return
            }
            b, err := json.MarshalIndent(recs, "", "  ")
            if err != nil { util.FatalIfError(err) }
            if output != "" {
                if err := os.WriteFile(output, b, 0644); err != nil { util.FatalIfError(err) }
                fmt.Printf("cloudadvisor recommendations written to %s\n", output)
            } else {
                fmt.Println(string(b))
            }
        default:
            fmt.Printf("Home Region: %s | Recommendations: %d\n", home, len(recs))
            for _, r := range recs {
                // SDK structs implement String() via common.PointerString
                fmt.Println(r.String())
            }
            if includeActions {
                fmt.Println("---- Resource Actions ----")
                if acts, ok := actions.([]optimizer.ResourceActionSummary); ok {
                    for _, a := range acts {
                        fmt.Println(a.String())
                    }
                }
            }
        }
    },
}

func init() {
    cloudAdvisorCmd.Flags().StringP("format", "f", "json", "output format: json or text")
    cloudAdvisorCmd.Flags().StringP("out", "o", "", "optional file path to write output")
    cloudAdvisorCmd.Flags().Bool("actions", false, "include all resource actions in output")
    cloudAdvisorCmd.Flags().Bool("org", false, "include data for all child tenancies in your organization (requires parent tenancy context)")
    cloudAdvisorCmd.Flags().String("child-tenancies", "", "comma-separated list of child tenancy OCIDs to include (mutually exclusive with --org)")
}
