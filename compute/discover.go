package compute

import (
    "context"
    "fmt"
    "time"

    "github.com/oracle/oci-go-sdk/v65/common"
    "github.com/oracle/oci-go-sdk/v65/identity"
    "github.com/oracle/oci-go-sdk/v65/monitoring"
)

type InstanceDiscovery struct {
    Region        string `json:"region"`
    Compartment   string `json:"compartment"`
    InstanceID    string `json:"instanceId"`
    Name          string `json:"displayName"`
    Shape         string `json:"shape"`
    Found         bool   `json:"found"`
    Namespace     string `json:"namespace,omitempty"`
    Metric        string `json:"metric,omitempty"`
    DimensionKey  string `json:"dimensionKey,omitempty"`
    ResourceGroup string `json:"resourceGroup,omitempty"`
    Samples       int    `json:"samples"`
    Note          string `json:"note,omitempty"`
}

// DiscoverMetrics probes Monitoring for CPU utilization metrics for each instance, returning
// which namespace/dimension combination yields datapoints (if any) in the last hour.
func DiscoverMetrics(provider common.ConfigurationProvider, regions []identity.RegionSubscription, compartments []identity.Compartment, filterOCID string, lookback time.Duration) ([]InstanceDiscovery, error) {
    var out []InstanceDiscovery

    client, err := monitoring.NewMonitoringClientWithConfigurationProvider(provider)
    if err != nil {
        return nil, err
    }

    metricNames := []string{"CPUUtilization", "CpuUtilization"}
    namespaces := []string{"oci_computeagent", "oci_compute_infrastructure", "oci_compute"}
    dimKeys := []string{"resourceId", "instanceId"}
    filters := []string{"", ", resourceGroup = 'default'"}

    if lookback <= 0 {
        lookback = time.Hour
    }
    end := time.Now().UTC()
    start := end.Add(-1 * lookback)

    for _, region := range regions {
        regionName := ""
        if region.RegionName != nil {
            regionName = *region.RegionName
        }
        client.SetRegion(regionName)

        for _, compartment := range compartments {
            instances := GetInstances(provider, compartment, regionName, false)
            for _, inst := range instances {
                id := ""
                name := ""
                shape := ""
                compName := ""
                if inst.Id != nil { id = *inst.Id }
                if inst.DisplayName != nil { name = *inst.DisplayName }
                if inst.Shape != nil { shape = *inst.Shape }
                if compartment.Name != nil { compName = *compartment.Name }

                if filterOCID != "" && id != filterOCID {
                    continue
                }

                d := InstanceDiscovery{Region: regionName, Compartment: compName, InstanceID: id, Name: name, Shape: shape}

                // try combinations, break on first with datapoints
                found := false
                for _, m := range metricNames {
                    for _, ns := range namespaces {
                        for _, dk := range dimKeys {
                            for _, f := range filters {
                                q := fmt.Sprintf("%s[1m]{%s = '%s'%s}.mean()", m, dk, id, f)
                                req := monitoring.SummarizeMetricsDataRequest{
                                    CompartmentId: compartment.Id,
                                    SummarizeMetricsDataDetails: monitoring.SummarizeMetricsDataDetails{
                                        Namespace: common.String(ns),
                                        Query:     common.String(q),
                                        StartTime: &common.SDKTime{Time: start},
                                        EndTime:   &common.SDKTime{Time: end},
                                    },
                                }
                                resp, err := client.SummarizeMetricsData(context.Background(), req)
                                if err == nil && len(resp.Items) > 0 && len(resp.Items[0].AggregatedDatapoints) > 0 {
                                    d.Found = true
                                    d.Namespace = ns
                                    d.Metric = m
                                    d.DimensionKey = dk
                                    if f != "" { d.ResourceGroup = "default" }
                                    d.Samples = len(resp.Items[0].AggregatedDatapoints)
                                    found = true
                                    break
                                }
                            }
                            if found { break }
                        }
                        if found { break }
                    }
                    if found { break }
                }

                if !found {
                    d.Note = "no datapoints in last hour"
                }
                out = append(out, d)
            }
        }
    }

    return out, nil
}
