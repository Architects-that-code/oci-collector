package compute

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "sort"
    "time"
    "strings"
    "math/rand"

    "github.com/oracle/oci-go-sdk/v65/common"
    "github.com/oracle/oci-go-sdk/v65/identity"
    "github.com/oracle/oci-go-sdk/v65/monitoring"
)

// InstanceMetrics holds aggregated utilization metrics for a time window
type InstanceMetrics struct {
    Avg         float64 `json:"avg"`
    P95         float64 `json:"p95"`
    P99         float64 `json:"p99"`
    Max         float64 `json:"max"`
    SampleCount int     `json:"samples"`
}

// InstanceMetricsSummary is a digestible summary for an instance
type InstanceMetricsSummary struct {
    Region      string           `json:"region"`
    Compartment string           `json:"compartment"`
    InstanceID  string           `json:"instanceId"`
    Name        string           `json:"displayName"`
    Shape       string           `json:"shape"`
    Day         InstanceMetrics  `json:"day"`
    Week        InstanceMetrics  `json:"week"`
    Month       InstanceMetrics  `json:"month"`
}

// CollectMetrics enumerates running instances and collects CPU utilization metrics for each over
// the last day, week, and month. Results are returned as a slice of summaries.
func CollectMetrics(provider common.ConfigurationProvider, regions []identity.RegionSubscription, compartments []identity.Compartment) ([]InstanceMetricsSummary, error) {
    var results []InstanceMetricsSummary

    for _, region := range regions {
        regionName := ""
        if region.RegionName != nil {
            regionName = *region.RegionName
        }
        for _, compartment := range compartments {
            instances := GetInstances(provider, compartment, regionName, false)
            for _, inst := range instances {
                id := ""
                name := ""
                shape := ""
                if inst.Id != nil {
                    id = *inst.Id
                }
                if inst.DisplayName != nil {
                    name = *inst.DisplayName
                }
                if inst.Shape != nil {
                    shape = *inst.Shape
                }

                compID := ""
                compName := ""
                if compartment.Id != nil {
                    compID = *compartment.Id
                }
                if compartment.Name != nil {
                    compName = *compartment.Name
                }

                fmt.Printf("Collecting metrics: region=%s compartment=%s instance=%s (%s)\n", regionName, compName, id, name)
                day, err := summarizeCPU(provider, regionName, compID, id, "1m", time.Now().Add(-24*time.Hour), time.Now())
                if err != nil {
                    // If metrics not available (agent disabled), continue but mark as empty
                    day = InstanceMetrics{}
                }
                time.Sleep(200 * time.Millisecond)
                week, err := summarizeCPU(provider, regionName, compID, id, "5m", time.Now().Add(-7*24*time.Hour), time.Now())
                if err != nil {
                    week = InstanceMetrics{}
                }
                time.Sleep(200 * time.Millisecond)
                month, err := summarizeCPU(provider, regionName, compID, id, "1h", time.Now().Add(-30*24*time.Hour), time.Now())
                if err != nil {
                    month = InstanceMetrics{}
                }

                results = append(results, InstanceMetricsSummary{
                    Region:      regionName,
                    Compartment: compName,
                    InstanceID:  id,
                    Name:        name,
                    Shape:       shape,
                    Day:         day,
                    Week:        week,
                    Month:       month,
                })
                // small delay between instances to avoid throttling
                time.Sleep(250 * time.Millisecond)
            }
        }
    }

    return results, nil
}

// summarizeCPU queries OCI Monitoring for CpuUtilization and computes avg, p95, p99, max for the period.
func summarizeCPU(provider common.ConfigurationProvider, region string, compartmentID string, instanceOCID string, period string, start time.Time, end time.Time) (InstanceMetrics, error) {
    client, err := monitoring.NewMonitoringClientWithConfigurationProvider(provider)
    if err != nil {
        return InstanceMetrics{}, err
    }
    client.SetRegion(region)

    // Build Monitoring query: mean over interval; we compute percentiles locally
    if period == "" { period = "1m" }
    // we'll try multiple possible metric/filters below

    // helper to execute a request with retries
    doReq := func(ns string, query string) (monitoring.SummarizeMetricsDataResponse, error) {
        req := monitoring.SummarizeMetricsDataRequest{
            CompartmentId: &compartmentID,
            SummarizeMetricsDataDetails: monitoring.SummarizeMetricsDataDetails{
                Namespace:                 common.String(ns),
                Query:                     common.String(query),
                StartTime:                 &common.SDKTime{Time: start.UTC()},
                EndTime:                   &common.SDKTime{Time: end.UTC()},
            },
        }
        var resp monitoring.SummarizeMetricsDataResponse
        var errDo error
        for attempt := 0; attempt < 6; attempt++ { // ~1m worst-case
            resp, errDo = client.SummarizeMetricsData(context.Background(), req)
            if errDo == nil {
                return resp, nil
            }
            if se, ok := errDo.(common.ServiceError); ok {
                code := se.GetHTTPStatusCode()
                if code == 429 || code >= 500 {
                    sleep := time.Duration(1<<attempt) * time.Second
                    sleep += time.Duration(rand.Intn(250)) * time.Millisecond
                    time.Sleep(sleep)
                    continue
                }
            }
            return monitoring.SummarizeMetricsDataResponse{}, errDo
        }
        return resp, errDo
    }

    // Try compute agent first, then infra, then generic compute as fallback
    namespaces := []string{"oci_computeagent", "oci_compute_infrastructure", "oci_compute"}
    metricNames := []string{"CpuUtilization", "CPUUtilization"}
    filters := []string{"", ", resourceGroup = 'default'"}
    dimKeys := []string{"resourceId", "instanceId"}
    var resp monitoring.SummarizeMetricsDataResponse
    var lastErr error
    outer:
    for _, name := range metricNames {
        for _, ns := range namespaces {
            for _, dk := range dimKeys {
                for _, f := range filters {
                    q := fmt.Sprintf("%s[%s]{%s = '%s'%s}.mean()", name, period, dk, instanceOCID, f)
                    r, errReq := doReq(ns, q)
                    if errReq == nil {
                        resp = r
                        if len(resp.Items) > 0 && len(resp.Items[0].AggregatedDatapoints) > 0 {
                            break outer
                        }
                    } else {
                        lastErr = errReq
                    }
                }
            }
        }
    }
    if len(resp.Items) == 0 || len(resp.Items[0].AggregatedDatapoints) == 0 {
        if lastErr != nil {
            return InstanceMetrics{}, lastErr
        }
        return InstanceMetrics{}, errors.New("no datapoints")
    }

    var values []float64
    var max float64
    for _, dp := range resp.Items[0].AggregatedDatapoints {
        if dp.Value == nil { continue }
        v := *dp.Value
        values = append(values, v)
        if v > max { max = v }
    }

    if len(values) == 0 {
        return InstanceMetrics{}, errors.New("no datapoints")
    }

    avg := mean(values)
    p95 := percentile(values, 95)
    p99 := percentile(values, 99)

    return InstanceMetrics{
        Avg:         avg,
        P95:         p95,
        P99:         p99,
        Max:         max,
        SampleCount: len(values),
    }, nil
}

func mean(xs []float64) float64 {
    var sum float64
    for _, v := range xs {
        sum += v
    }
    if len(xs) == 0 {
        return 0
    }
    return sum / float64(len(xs))
}

// percentile computes the pth percentile via nearest-rank method on a copy of the slice
func percentile(xs []float64, p int) float64 {
    if len(xs) == 0 {
        return 0
    }
    ys := make([]float64, len(xs))
    copy(ys, xs)
    sort.Float64s(ys)
    if p <= 0 {
        return ys[0]
    }
    if p >= 100 {
        return ys[len(ys)-1]
    }
    // Nearest-rank
    k := int(float64(p)/100.0*float64(len(ys)))
    if k <= 0 {
        k = 1
    }
    if k > len(ys) {
        k = len(ys)
    }
    return ys[k-1]
}

// ToJSON pretty prints metrics summaries
func ToJSON(summaries []InstanceMetricsSummary) (string, error) {
    b, err := json.MarshalIndent(summaries, "", "  ")
    if err != nil {
        return "", err
    }
    return string(b), nil
}

// ToCSV converts metrics summaries to a flat, digestible CSV. Each instance has three rows (day/week/month).
func ToCSV(summaries []InstanceMetricsSummary) string {
    var b strings.Builder
    // header
    b.WriteString("region,compartment,instanceId,displayName,shape,window,avg,p95,p99,max,samples\n")
    for _, s := range summaries {
        // helper to write a row
        write := func(window string, m InstanceMetrics) {
            b.WriteString(fmt.Sprintf("%s,%s,%s,%s,%s,%s,%.2f,%.2f,%.2f,%.2f,%d\n",
                s.Region, s.Compartment, s.InstanceID, s.Name, s.Shape, window,
                m.Avg, m.P95, m.P99, m.Max, m.SampleCount,
            ))
        }
        write("day", s.Day)
        write("week", s.Week)
        write("month", s.Month)
    }
    return b.String()
}
