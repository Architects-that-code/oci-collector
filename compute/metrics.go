package compute

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/oracle/oci-go-sdk/core"
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

// InstanceInventory is a light-weight record of a running instance in a given region/compartment.
// This is used by metrics collection to avoid re-querying instances when the caller already
// has an inventory available.
type InstanceInventory struct {
	RegionName  string
	Compartment identity.Compartment
	Instance    core.Instance
}

// GatherInstances builds an inventory of running instances across regions/compartments.
// NOTE: metrics.go lives in the compute package, so we can reuse GetInstances from compute.go.
func GatherInstances(provider common.ConfigurationProvider, regions []identity.RegionSubscription, compartments []identity.Compartment, verbose bool) []InstanceInventory {
	return GatherInstancesWithOptions(provider, regions, compartments, verbose, false)
}

// GatherInstancesWithOptions builds an inventory of running instances with optional progress reporting.
func GatherInstancesWithOptions(provider common.ConfigurationProvider, regions []identity.RegionSubscription, compartments []identity.Compartment, verbose bool, showProgress bool) []InstanceInventory {
	type gatherTask struct {
		regionName string
		comp       identity.Compartment
	}

	totalTasks := len(regions) * len(compartments)
	if totalTasks == 0 {
		return nil
	}

	tasks := make(chan gatherTask, totalTasks)
	results := make(chan []InstanceInventory, totalTasks)

	workerCount := defaultComputeWorkerPoolSize
	if workerCount <= 0 {
		workerCount = 1
	}
	if totalTasks < workerCount {
		workerCount = totalTasks
	}

	var completedTasks atomic.Int64
	var foundInstances atomic.Int64
	var progressDone chan struct{}
	if showProgress {
		fmt.Printf("Collecting compute inventory (%d regions x %d compartments = %d tasks, %d workers)\n", len(regions), len(compartments), totalTasks, workerCount)
		progressDone = make(chan struct{})
		go func() {
			ticker := time.NewTicker(2 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					done := completedTasks.Load()
					instances := foundInstances.Load()
					pct := float64(done) * 100.0 / float64(totalTasks)
					fmt.Printf("\rcompute inventory progress: %d/%d tasks (%.1f%%), instances found: %d", done, totalTasks, pct, instances)
				case <-progressDone:
					done := completedTasks.Load()
					instances := foundInstances.Load()
					fmt.Printf("\rcompute inventory progress: %d/%d tasks (100.0%%), instances found: %d\n", done, totalTasks, instances)
					return
				}
			}
		}()
	}

	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range tasks {
				instances := GetInstances(provider, task.comp, task.regionName, verbose)
				local := make([]InstanceInventory, 0, len(instances))
				for _, inst := range instances {
					local = append(local, InstanceInventory{
						RegionName:  task.regionName,
						Compartment: task.comp,
						Instance:    inst,
					})
				}
				if showProgress {
					completedTasks.Add(1)
					foundInstances.Add(int64(len(local)))
				}
				results <- local
			}
		}()
	}

	for _, region := range regions {
		regionName := ""
		if region.RegionName != nil {
			regionName = *region.RegionName
		}
		for _, comp := range compartments {
			tasks <- gatherTask{regionName: regionName, comp: comp}
		}
	}
	close(tasks)

	go func() {
		wg.Wait()
		close(results)
	}()

	inv := make([]InstanceInventory, 0)
	for chunk := range results {
		inv = append(inv, chunk...)
	}
	if showProgress {
		close(progressDone)
	}

	return inv
}

// MetricTask represents a single metric collection task
type MetricTask struct {
	Client        monitoring.MonitoringClient
	CompartmentID string
	InstanceID    string
	MetricName    string
	Start         time.Time
	End           time.Time
	Namespaces    []string
}

// MetricResult holds the result of a metric collection task
type MetricResult struct {
	Task    MetricTask
	Metrics InstanceMetrics
	Error   error
}

// CollectMetrics enumerates running instances and collects utilization metrics for each over the last day, week, and month.
func CollectMetrics(provider common.ConfigurationProvider, regions []identity.RegionSubscription, compartments []identity.Compartment, inventory []InstanceInventory) ([]InstanceMetricsSummary, error) {
	// Create base monitoring client
	baseClient, err := monitoring.NewMonitoringClientWithConfigurationProvider(provider)
	if err != nil {
		return nil, err
	}

	if inventory == nil {
		inventory = GatherInstances(provider, regions, compartments, false)
	}
	var allInstances []struct {
		region         string
		compartment    string
		compID         string
		instance       core.Instance
		agentInstalled bool
	}
	for _, entry := range inventory {
		regionName := entry.RegionName
		compName := safeStr(entry.Compartment.Name)
		compID := safeStr(entry.Compartment.Id)
		allInstances = append(allInstances, struct {
			region         string
			compartment    string
			compID         string
			instance       core.Instance
			agentInstalled bool
		}{
			region:         regionName,
			compartment:    compName,
			compID:         compID,
			instance:       entry.Instance,
			agentInstalled: isAgentInstalled(entry.Instance),
		})
	}

	// Create tasks for all metric collections
	var tasks []MetricTask
	now := time.Now().UTC()
	dayStart := now.Add(-24 * time.Hour)
	weekStart := now.Add(-7 * 24 * time.Hour)
	monthStart := now.Add(-30 * 24 * time.Hour)

	for _, inst := range allInstances {
		id := ""
		if inst.instance.Id != nil {
			id = *inst.instance.Id
		}

		// Always collect CPU metrics
		for _, timeWindow := range []struct{ start, end time.Time }{
			{dayStart, now},
			{weekStart, now},
			{monthStart, now},
		} {
			tasks = append(tasks, MetricTask{
				Client:        baseClient,
				CompartmentID: inst.compID,
				InstanceID:    id,
				MetricName:    "CpuUtilization",
				Start:         timeWindow.start,
				End:           timeWindow.end,
				Namespaces:    []string{"oci_computeagent", "oci_compute_infrastructure"},
			})
		}

		// Collect additional metrics if agent is installed
		if inst.agentInstalled {
			for _, metric := range []string{"MemoryUtilization", "DiskUtilization", "NetworkBytesIn"} {
				for _, timeWindow := range []struct{ start, end time.Time }{
					{dayStart, now},
					{weekStart, now},
					{monthStart, now},
				} {
					tasks = append(tasks, MetricTask{
						Client:        baseClient,
						CompartmentID: inst.compID,
						InstanceID:    id,
						MetricName:    metric,
						Start:         timeWindow.start,
						End:           timeWindow.end,
						Namespaces:    []string{"oci_computeagent"},
					})
				}
			}
		}
	}

	// Execute tasks with rate limiting
	results := executeMetricTasks(tasks)

	// Group results by instance
	instanceMap := make(map[string]*InstanceMetricsSummary)
	for _, inst := range allInstances {
		id := ""
		if inst.instance.Id != nil {
			id = *inst.instance.Id
		}
		name := ""
		if inst.instance.DisplayName != nil {
			name = *inst.instance.DisplayName
		}
		shape := ""
		if inst.instance.Shape != nil {
			shape = *inst.instance.Shape
		}

		instanceMap[id] = &InstanceMetricsSummary{
			Region:         inst.region,
			Compartment:    inst.compartment,
			InstanceID:     id,
			Name:           name,
			Shape:          shape,
			AgentInstalled: inst.agentInstalled,
		}
	}

	// Populate metrics results
	for _, result := range results {
		if result.Error != nil {
			continue // Skip failed metrics
		}

		summary := instanceMap[result.Task.InstanceID]
		if summary == nil {
			continue
		}

		duration := result.Task.End.Sub(result.Task.Start)
		var timeKey string
		if duration <= 25*time.Hour {
			timeKey = "Day"
		} else if duration <= 8*24*time.Hour {
			timeKey = "Week"
		} else {
			timeKey = "Month"
		}

		switch result.Task.MetricName {
		case "CpuUtilization":
			switch timeKey {
			case "Day":
				summary.CpuDay = result.Metrics
			case "Week":
				summary.CpuWeek = result.Metrics
			case "Month":
				summary.CpuMonth = result.Metrics
			}
		case "MemoryUtilization":
			switch timeKey {
			case "Day":
				summary.MemoryDay = result.Metrics
			case "Week":
				summary.MemoryWeek = result.Metrics
			case "Month":
				summary.MemoryMonth = result.Metrics
			}
		case "DiskUtilization":
			switch timeKey {
			case "Day":
				summary.DiskDay = result.Metrics
			case "Week":
				summary.DiskWeek = result.Metrics
			case "Month":
				summary.DiskMonth = result.Metrics
			}
		case "NetworkBytesIn":
			switch timeKey {
			case "Day":
				summary.NetworkDay = result.Metrics
			case "Week":
				summary.NetworkWeek = result.Metrics
			case "Month":
				summary.NetworkMonth = result.Metrics
			}
		}
	}

	// Convert map to slice
	var finalResults []InstanceMetricsSummary
	for _, summary := range instanceMap {
		finalResults = append(finalResults, *summary)
	}

	return finalResults, nil
}

// executeMetricTasks runs metric collection tasks with rate limiting
func executeMetricTasks(tasks []MetricTask) []MetricResult {
	const maxConcurrent = 10 // Limit concurrent API calls
	semaphore := make(chan struct{}, maxConcurrent)
	results := make(chan MetricResult, len(tasks))

	// Start workers
	for _, task := range tasks {
		go func(t MetricTask) {
			semaphore <- struct{}{}        // Acquire
			defer func() { <-semaphore }() // Release

			metrics, err := getMetric(t.Client, t.CompartmentID, t.InstanceID, t.MetricName, t.Start, t.End)
			results <- MetricResult{
				Task:    t,
				Metrics: metrics,
				Error:   err,
			}
		}(task)
	}

	// Collect results
	var finalResults []MetricResult
	for i := 0; i < len(tasks); i++ {
		finalResults = append(finalResults, <-results)
	}

	return finalResults
}

// getMetric queries Monitoring for a specific metric over the given time window
func getMetric(client monitoring.MonitoringClient, compartmentID, instanceID, metricName string, start, end time.Time) (InstanceMetrics, error) {
	// Determine namespace based on metric
	var namespaces []string
	if metricName == "CpuUtilization" {
		// For CPU, try both namespaces
		namespaces = []string{"oci_computeagent", "oci_compute_infrastructure"}
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
			//fmt.Printf("Trying query: namespace=%s, query=%s\n", ns, q)
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
