package autonomous

import (
    "context"
    "fmt"

    "github.com/oracle/oci-go-sdk/v65/common"
    "github.com/oracle/oci-go-sdk/v65/database"
    "github.com/oracle/oci-go-sdk/v65/identity"
    "oci-collector/util"
)

// ADBItem represents an Autonomous Database found in a specific region and compartment
type ADBItem struct {
    Region      string
    Compartment string
    Summary     database.AutonomousDatabaseSummary
}

// ListAutonomousDatabases iterates all subscribed regions and all compartments to find ADBs
func ListAutonomousDatabases(provider common.ConfigurationProvider, regions []identity.RegionSubscription, compartments []identity.Compartment) []ADBItem {
    var results []ADBItem

    // Loop regions then compartments; simple and reliable. Parallelism can be added later if needed.
    for _, r := range regions {
        region := *r.RegionName
        client, err := database.NewDatabaseClientWithConfigurationProvider(provider)
        util.FatalIfError(err)
        client.SetRegion(region)

        for _, c := range compartments {
            req := database.ListAutonomousDatabasesRequest{
                CompartmentId: c.Id,
                Limit:         common.Int(1000),
            }
            for {
                resp, err := client.ListAutonomousDatabases(context.Background(), req)
                util.FatalIfError(err)
                for _, it := range resp.Items {
                    results = append(results, ADBItem{Region: region, Compartment: *c.Name, Summary: it})
                }
                if resp.OpcNextPage != nil {
                    req.Page = resp.OpcNextPage
                } else {
                    break
                }
            }
        }
    }

    return results
}

// PrintAutonomousDatabases prints a concise list of ADBs found across regions/compartments
func PrintAutonomousDatabases(items []ADBItem) {
    if len(items) == 0 {
        fmt.Println("No Autonomous Databases found across subscribed regions/compartments.")
        return
    }
    for _, i := range items {
        name := safeStr(i.Summary.DisplayName)
        workload := ""
        if i.Summary.DbWorkload != "" {
            workload = string(i.Summary.DbWorkload)
        }
        state := ""
        if i.Summary.LifecycleState != "" {
            state = string(i.Summary.LifecycleState)
        }
        ocid := safeStr(i.Summary.Id)
        cpu := ""
        if i.Summary.CpuCoreCount != nil {
            cpu = fmt.Sprintf("cpu=%d", *i.Summary.CpuCoreCount)
        }
        storage := ""
        if i.Summary.DataStorageSizeInTBs != nil {
            storage = fmt.Sprintf("storageTB=%d", *i.Summary.DataStorageSizeInTBs)
        }
        fmt.Printf("%s | %s | %s | workload=%s %s %s | %s\n", i.Region, i.Compartment, name, workload, cpu, storage, state)
        // Optionally include OCID for traceability
        _ = ocid
    }
}

func safeStr(p *string) string {
    if p == nil {
        return ""
    }
    return *p
}
