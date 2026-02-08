package compute

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"time"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/identity"
	"github.com/oracle/oci-go-sdk/v65/monitoring"
)

// InstanceMetrics holds aggregated utilization metrics for a time window.
type InstanceMetrics struct {
	Avg         float64 `json:"avg"`
	P95         float64 `json:"p95"`
	P99         float64 `json:"p99"`
	Max         float64 `json:"max"`
	SampleCount int     `json:"samples"`
	Note        string  `json:"note,omitempty"`
}

// InstanceMetricsSummary is a digestible summary for an instance with various metrics.
type InstanceMetricsSummary struct {
	Region         string          `json:"region"`
	Compartment    string          `json:"compartment"`
	InstanceID     string          `json:"instanceId"`
	Name           string          `json:"displayName"`
	Shape          string          `json:"shape"`
	AgentInstalled bool            `json:"agentInstalled"`
	CpuDay         InstanceMetrics `json:"cpuDay"`
	CpuWeek        InstanceMetrics `json:"cpuWeek"`
	CpuMonth       InstanceMetrics `json:"cpuMonth"`
	MemoryDay      InstanceMetrics `json:"memoryDay,omitempty"`
	MemoryWeek     InstanceMetrics `json:"memoryWeek,omitempty"`
	MemoryMonth    InstanceMetrics `json:"memoryMonth,omitempty"`
	DiskDay        InstanceMetrics `json:"diskDay,omitempty"`
	DiskWeek       InstanceMetrics `json:"diskWeek,omitempty"`
	DiskMonth      InstanceMetrics `json:"diskMonth,omitempty"`
	NetworkDay     InstanceMetrics `json:"networkDay,omitempty"`
	NetworkWeek    InstanceMetrics `json:"networkWeek,omitempty"`
	NetworkMonth   InstanceMetrics `json:"networkMonth,omitempty"`
	Note           string          `json:"note,omitempty"`
}

// CollectMetrics enumerates running instances and collects utilization metrics for each over the last day, week, and month.
func CollectMetrics(provider common.ConfigurationProvider, regions []identity.RegionSubscription, compartments []identity.Compartment) ([]InstanceMetricsSummary, error) {
	var results []InstanceMetricsSummary

	client, err := monitoring.NewMonitoringClientWithConfigurationProvider(provider)
	if err != nil {
		return nil, err
	}

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

				agentInstalled := isAgentInstalled(inst)

				// Collect metrics for different time windows
				now := time.Now().UTC()
				dayStart := now.Add(-24 * time.Hour)
				weekStart := now.Add(-7 * 24 * time.Hour)
				monthStart := now.Add(-30 * 24 * time.Hour) // approx month

				// Always try CPU
				cpuDay, _ := getMetric(client, compID, id, "CpuUtilization", dayStart, now)
				cpuWeek, _ := getMetric(client, compID, id, "CpuUtilization", weekStart, now)
				cpuMonth, _ := getMetric(client, compID, id, "CpuUtilization", monthStart, now)

				summary := InstanceMetricsSummary{
					Region:         regionName,
					Compartment:    compName,
					InstanceID:     id,
					Name:           name,
					Shape:          shape,
					AgentInstalled: agentInstalled,
					CpuDay:         cpuDay,
					CpuWeek:        cpuWeek,
					CpuMonth:       cpuMonth,
				}

				// If agent is installed, collect additional metrics
				if agentInstalled {
					memDay, _ := getMetric(client, compID, id, "MemoryUtilization", dayStart, now)
					memWeek, _ := getMetric(client, compID, id, "MemoryUtilization", weekStart, now)
					memMonth, _ := getMetric(client, compID, id, "MemoryUtilization", monthStart, now)
					summary.MemoryDay = memDay
					summary.MemoryWeek = memWeek
					summary.MemoryMonth = memMonth

					diskDay, _ := getMetric(client, compID, id, "DiskUtilization", dayStart, now)
					diskWeek, _ := getMetric(client, compID, id, "DiskUtilization", weekStart, now)
					diskMonth, _ := getMetric(client, compID, id, "DiskUtilization", monthStart, now)
					summary.DiskDay = diskDay
					summary.DiskWeek = diskWeek
					summary.DiskMonth = diskMonth

					netDay, _ := getMetric(client, compID, id, "NetworkBytesIn", dayStart, now)
					netWeek, _ := getMetric(client, compID, id, "NetworkBytesIn", weekStart, now)
					netMonth, _ := getMetric(client, compID, id, "NetworkBytesIn", monthStart, now)
					summary.NetworkDay = netDay
					summary.NetworkWeek = netWeek
					summary.NetworkMonth = netMonth
				}

				results = append(results, summary)
			}
		}
	}

	return results, nil
}

// getMetric queries Monitoring for a specific metric over the given time window
func getMetric(client monitoring.MonitoringClient, compartmentID, instanceID, metricName string, start, end time.Time) (InstanceMetrics, error) {
	// Determine namespace based on metric
	var namespaces []string
	if metricName == "CpuUtilization" && start.After(time.Now().Add(-24*time.Hour)) {
		// For recent CPU, try both
		namespaces = []string{"oci_computeagent", "oci_compute_infrastructure"}
	} else if metricName == "CpuUtilization" {
		namespaces = []string{"oci_compute_infrastructure", "oci_computeagent"}
	} else {
		// Memory, Disk, Network require agent
		namespaces = []string{"oci_computeagent"}
	}

	var avg, p95, p99, max float64
	var sampleCount int

	// Get average
	avgVal, samples := getMetricStat(client, compartmentID, instanceID, metricName, start, end, namespaces, ".mean()")
	if avgVal != nil {
		avg = *avgVal
		sampleCount = samples
	}

	// Get P95
	p95Val, _ := getMetricStat(client, compartmentID, instanceID, metricName, start, end, namespaces, ".percentile(0.95)")
	if p95Val != nil {
		p95 = *p95Val
	}

	// Get P99
	p99Val, _ := getMetricStat(client, compartmentID, instanceID, metricName, start, end, namespaces, ".percentile(0.99)")
	if p99Val != nil {
		p99 = *p99Val
	}

	// Get Max
	maxVal, _ := getMetricStat(client, compartmentID, instanceID, metricName, start, end, namespaces, ".max()")
	if maxVal != nil {
		max = *maxVal
	}

	if sampleCount == 0 {
		return InstanceMetrics{}, fmt.Errorf("no datapoints found")
	}

	return InstanceMetrics{
		Avg:         avg,
		P95:         p95,
		P99:         p99,
		Max:         max,
		SampleCount: sampleCount,
	}, nil
}

// getMetricStat gets a specific statistic for a metric
func getMetricStat(client monitoring.MonitoringClient, compartmentID, instanceID, metricName string, start, end time.Time, namespaces []string, statFunc string) (*float64, int) {
	// Determine interval based on time range
	duration := end.Sub(start)
	var interval string
	if duration > 7*24*time.Hour {
		interval = "1h"
	} else {
		interval = "1m"
	}

	for _, ns := range namespaces {
		queries := []string{
			fmt.Sprintf(`%s[%s]{resourceId = "%s"}%s`, metricName, interval, instanceID, statFunc),
			fmt.Sprintf(`%s[%s]{resourceId = "%s", resourceGroup = "default"}%s`, metricName, interval, instanceID, statFunc),
		}

		for _, q := range queries {
			fmt.Printf("Trying query: namespace=%s, query=%s\n", ns, q)
			req := monitoring.SummarizeMetricsDataRequest{
				CompartmentId: &compartmentID,
				SummarizeMetricsDataDetails: monitoring.SummarizeMetricsDataDetails{
					Namespace: &ns,
					Query:     &q,
					StartTime: &common.SDKTime{Time: start},
					EndTime:   &common.SDKTime{Time: end},
				},
			}

			resp, err := client.SummarizeMetricsData(context.Background(), req)
			if err != nil {
				continue
			}

			if len(resp.Items) > 0 && len(resp.Items[0].AggregatedDatapoints) > 0 {
				// For statistic queries, take the first/only value
				if resp.Items[0].AggregatedDatapoints[0].Value != nil {
					value := *resp.Items[0].AggregatedDatapoints[0].Value
					samples := len(resp.Items[0].AggregatedDatapoints)
					return &value, samples
				}
			}
		}
	}

	return nil, 0
}

// ToJSON serializes summaries to indented JSON
func ToJSON(summaries []InstanceMetricsSummary) (string, error) {
	b, err := json.MarshalIndent(summaries, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// ToCSV renders summaries to CSV (CPU metrics only)
func ToCSV(summaries []InstanceMetricsSummary) string {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{
		"region", "compartment", "instanceId", "displayName", "shape", "agentInstalled",
		"cpu_day_avg", "cpu_day_p95", "cpu_day_p99", "cpu_day_max", "cpu_day_samples",
		"cpu_week_avg", "cpu_week_p95", "cpu_week_p99", "cpu_week_max", "cpu_week_samples",
		"cpu_month_avg", "cpu_month_p95", "cpu_month_p99", "cpu_month_max", "cpu_month_samples",
		"note",
	})
	for _, s := range summaries {
		note := s.Note
		if s.CpuDay.Note != "" {
			note += " cpu_day:" + s.CpuDay.Note
		}
		if s.CpuWeek.Note != "" {
			note += " cpu_week:" + s.CpuWeek.Note
		}
		if s.CpuMonth.Note != "" {
			note += " cpu_month:" + s.CpuMonth.Note
		}
		_ = w.Write([]string{
			s.Region, s.Compartment, s.InstanceID, s.Name, s.Shape, fmt.Sprintf("%t", s.AgentInstalled),
			fmt.Sprintf("%.2f", s.CpuDay.Avg), fmt.Sprintf("%.2f", s.CpuDay.P95), fmt.Sprintf("%.2f", s.CpuDay.P99), fmt.Sprintf("%.2f", s.CpuDay.Max), fmt.Sprintf("%d", s.CpuDay.SampleCount),
			fmt.Sprintf("%.2f", s.CpuWeek.Avg), fmt.Sprintf("%.2f", s.CpuWeek.P95), fmt.Sprintf("%.2f", s.CpuWeek.P99), fmt.Sprintf("%.2f", s.CpuWeek.Max), fmt.Sprintf("%d", s.CpuWeek.SampleCount),
			fmt.Sprintf("%.2f", s.CpuMonth.Avg), fmt.Sprintf("%.2f", s.CpuMonth.P95), fmt.Sprintf("%.2f", s.CpuMonth.P99), fmt.Sprintf("%.2f", s.CpuMonth.Max), fmt.Sprintf("%d", s.CpuMonth.SampleCount),
			note,
		})
	}
	w.Flush()
	return buf.String()
}
