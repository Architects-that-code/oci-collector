package compute

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/oracle/oci-go-sdk/core"
	"github.com/oracle/oci-go-sdk/example/helpers"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/identity"
)

type InstanceGroups struct {
	Region      string
	Compartment string
	Instance    []core.Instance
}

// InstanceRow is a flat record suitable for JSON/CSV output
type InstanceRow struct {
	Region         string            `json:"region"`
	Compartment    string            `json:"compartment"`
	InstanceID     string            `json:"instanceId"`
	DisplayName    string            `json:"displayName"`
	Shape          string            `json:"shape"`
	Freeform       map[string]string `json:"freeformTags,omitempty"`
	AgentInstalled bool              `json:"agentInstalled"`
}

func RunCompute(provider common.ConfigurationProvider, regions []identity.RegionSubscription, tenancyID string, compartments []identity.Compartment, format string, output string) error {
	//client, err := core.NewComputeClientWithConfigurationProvider(provider)
	//helpers.FatalIfError(err)
	fmt.Println("in RunCompute")

	//loop thru regionsconst

	//		in region loop thru compartments
	//TODO: ADD turbonium to region and compartments

	var allInstances []InstanceGroups
	var wg sync.WaitGroup
	wg.Add(len(regions))
	var regionalSlices = make(chan []InstanceGroups, len(regions))

	for _, region := range regions {
		go func(region identity.RegionSubscription) {
			defer wg.Done()
			for _, compartment := range compartments {
				instances := GetInstances(provider, compartment, *region.RegionName, true)
				allInstances = append(allInstances, InstanceGroups{Region: *region.RegionName, Compartment: *compartment.Name, Instance: instances})
				//fmt.Printf("region: \t%v  \tcomp:%v: \t\t%v\n", *region.RegionName, *compartment.Name, len(instances))
			}
			regionalSlices <- allInstances
		}(region)
	}
	wg.Wait()

	//fmt.Printf("Total instances: %v\n", len(allInstances))

	//fmt.Printf("allInstances: %v\n", allInstances)
	sort.Slice(allInstances, func(i, j int) bool {
		return len(allInstances[i].Instance) > len(allInstances[j].Instance)
	})
	// If a structured format is requested, output JSON/CSV instead of text
	switch strings.ToLower(format) {
	case "json":
		rows := flattenInstances(allInstances)
		b, err := json.MarshalIndent(rows, "", "  ")
		if err != nil {
			return err
		}
		if output != "" {
			if err := os.WriteFile(output, b, 0644); err != nil {
				return err
			}
			fmt.Printf("instances written to %s\n", output)
		} else {
			fmt.Println(string(b))
		}
	case "csv":
		rows := flattenInstances(allInstances)
		data := toCSV(rows)
		if output != "" {
			if err := os.WriteFile(output, []byte(data), 0644); err != nil {
				return err
			}
			fmt.Printf("instances written to %s\n", output)
		} else {
			fmt.Print(data)
		}
	default:
		fmt.Println("ONLY PRINTING for ACTIVE instances per region/compartment")
		for _, instanceGroup := range allInstances {
			if len(instanceGroup.Instance) > 0 {
				fmt.Printf("all instances: Region: %v Compartment: %v  NumInstance: %v \n", instanceGroup.Region, instanceGroup.Compartment, len(instanceGroup.Instance))
		for _, instance := range instanceGroup.Instance {
			name := safeStr(instance.DisplayName)
			shape := safeStr(instance.Shape)
			agent := isAgentInstalled(instance)
			fmt.Printf("DisplayName: %s\t Shape: %s\t AgentInstalled: %t\t tags: freeform: %v\t defined: %v\n", name, shape, agent, instance.FreeformTags, instance.DefinedTags)
		}
			}
		}
	}

	return nil
}

// flattenInstances converts grouped instances into flat rows for export
func flattenInstances(groups []InstanceGroups) []InstanceRow {
	var rows []InstanceRow
	for _, g := range groups {
		for _, inst := range g.Instance {
			rows = append(rows, InstanceRow{
				Region:         g.Region,
				Compartment:    g.Compartment,
				InstanceID:     safeStr(inst.Id),
				DisplayName:    safeStr(inst.DisplayName),
				Shape:          safeStr(inst.Shape),
				Freeform:       inst.FreeformTags,
				AgentInstalled: isAgentInstalled(inst),
			})
		}
	}
	return rows
}

// toCSV renders instance rows to CSV with header
func toCSV(rows []InstanceRow) string {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"region", "compartment", "instanceId", "displayName", "shape", "agentInstalled"})
	for _, r := range rows {
		_ = w.Write([]string{r.Region, r.Compartment, r.InstanceID, r.DisplayName, r.Shape, fmt.Sprintf("%t", r.AgentInstalled)})
	}
	w.Flush()
	return buf.String()
}

// safeStr dereferences a *string safely
func safeStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func GetInstances(provider common.ConfigurationProvider, compartment identity.Compartment, region string, verbose bool) []core.Instance {
	client, err := core.NewComputeClientWithConfigurationProvider(provider)
	helpers.FatalIfError(err)
	client.SetRegion(region)
	var allCompute []core.Instance
	if verbose {
		fmt.Printf("Checking: Region: %v\t Compartment: %v\t\n", region, *compartment.Name)
	}
	req := core.ListInstancesRequest{
		CompartmentId:  compartment.Id,
		LifecycleState: core.InstanceLifecycleStateRunning,
	}
	for {
		// Send the request using the service client with backoff on throttling
		var resp core.ListInstancesResponse
		var errDo error
		for attempt := 0; attempt < 8; attempt++ { // up to ~ (1+2+4+...)=~255s worst-case, capped by sleep
			resp, errDo = client.ListInstances(context.Background(), req)
			if errDo == nil {
				break
			}
			if isTooManyRequests(errDo) {
				sleep := backoffWithJitter(attempt)
				time.Sleep(sleep)
				continue
			}
			// non-throttle errors are fatal
			helpers.FatalIfError(errDo)
		}
		// if still failing after retries, bail out
		helpers.FatalIfError(errDo)

		if verbose {
			for _, instance := range resp.Items {
				GetIPs(provider, instance, region, compartment)
			}
		}

		allCompute = append(allCompute, resp.Items...)
		if resp.OpcNextPage != nil {
			req.Page = resp.OpcNextPage
		} else {
			break
		}
	}

	// Retrieve value from the response.
	return allCompute
}

// TODO: add fetching IPs for a compute instance  - specifically so we can get the public IP
func GetIPs(provider common.ConfigurationProvider, instance core.Instance, region string, compartment identity.Compartment) {
	client, err := core.NewComputeClientWithConfigurationProvider(provider)
	helpers.FatalIfError(err)
	client.SetRegion(region)

	request := core.ListVnicAttachmentsRequest{
		CompartmentId: instance.CompartmentId,
		InstanceId:    instance.Id,
	}

	var response core.ListVnicAttachmentsResponse
	var errDo error
	for attempt := 0; attempt < 8; attempt++ {
		response, errDo = client.ListVnicAttachments(context.Background(), request)
		if errDo == nil {
			break
		}
		if isTooManyRequests(errDo) {
			time.Sleep(backoffWithJitter(attempt))
			continue
		}
		helpers.FatalIfError(errDo)
	}
	helpers.FatalIfError(errDo)

	for _, vnic := range response.Items {
		vnicID := *vnic.VnicId
		getPubIP(provider, vnicID, region, compartment)
	}
}

func getPubIP(provider common.ConfigurationProvider, vnicID string, region string, compartment identity.Compartment) {
	client, err := core.NewVirtualNetworkClientWithConfigurationProvider(provider)
	helpers.FatalIfError(err)
	client.SetRegion(region)

	req := core.GetVnicRequest{VnicId: &vnicID}

	var resp core.GetVnicResponse
	var errDo error
	for attempt := 0; attempt < 8; attempt++ {
		resp, errDo = client.GetVnic(context.Background(), req)
		if errDo == nil {
			break
		}
		if isTooManyRequests(errDo) {
			time.Sleep(backoffWithJitter(attempt))
			continue
		}
		helpers.FatalIfError(errDo)
	}
	helpers.FatalIfError(errDo)

	fmt.Printf("vnicID: %s\n", vnicID)
	fmt.Printf("RESPpriv: %s\n", *resp.Vnic.PrivateIp)
	if resp.Vnic.PublicIp != nil {
		fmt.Printf("RESPpub: %s\n", *resp.Vnic.PublicIp)
	}
}

// isTooManyRequests checks for OCI throttling errors without depending on SDK-specific error types
func isTooManyRequests(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	if strings.Contains(msg, "TooManyRequests") {
		return true
	}
	if strings.Contains(msg, "429") {
		return true
	}
	if strings.Contains(strings.ToLower(msg), "throttl") {
		return true
	}
	return false
}

// EnableMetrics enables Oracle Cloud Agent metric collection on instances that have the agent but monitoring disabled
// Note: UpdateInstanceAgentConfig API not available in current SDK version (v65)
// This is a placeholder for future implementation
func EnableMetrics(provider common.ConfigurationProvider, regions []identity.RegionSubscription, compartments []identity.Compartment) error {
	fmt.Println("EnableMetrics: API not available in current SDK version")
	fmt.Println("To enable metrics manually:")
	fmt.Println("1. SSH to instance")
	fmt.Println("2. Run: sudo /opt/oracle/cloud-agent/bin/cloud-agent-control enable monitoring")
	fmt.Println("3. Or use OCI Console: Compute > Instances > Instance Details > Oracle Cloud Agent")
	return nil
}

// isAgentInstalled checks if Oracle Cloud Agent is installed and monitoring is enabled
func isAgentInstalled(inst core.Instance) bool {
	if inst.AgentConfig == nil {
		return false
	}
	// If monitoring is explicitly disabled, agent is not considered installed for monitoring
	if inst.AgentConfig.IsMonitoringDisabled != nil && *inst.AgentConfig.IsMonitoringDisabled {
		return false
	}
	// If AgentConfig exists and monitoring not disabled, assume agent is installed
	return true
}

// backoffWithJitter returns an exponential backoff duration with jitter and a sane cap
func backoffWithJitter(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}
	base := time.Second
	d := base << attempt // 1s,2s,4s,...
	if d > 30*time.Second {
		d = 30 * time.Second
	}
	j := time.Duration(rand.Intn(300)) * time.Millisecond
	return d + j
}
