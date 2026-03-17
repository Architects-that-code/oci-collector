package compute

import (
	"context"
	"sort"
	"strings"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/computeinstanceagent"
	"oci-collector/util"
)

// AgentPluginState captures an instance agent plugin and its current state.
// Status is runtime state when sourced from Compute Instance Agent API
// and desired state when sourced from instance AgentConfig fallback.
type AgentPluginState struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

func listInstanceAgentPlugins(client computeinstanceagent.PluginClient, compartmentID string, instanceID string) ([]AgentPluginState, error) {
	req := computeinstanceagent.ListInstanceAgentPluginsRequest{
		CompartmentId:   common.String(compartmentID),
		InstanceagentId: common.String(instanceID),
		Limit:           common.Int(1000),
	}

	var all []AgentPluginState
	for {
		resp, err := util.RetryWithBackoff(func() (computeinstanceagent.ListInstanceAgentPluginsResponse, error) {
			return client.ListInstanceAgentPlugins(context.Background(), req)
		})
		if err != nil {
			return nil, err
		}

		for _, item := range resp.Items {
			all = append(all, AgentPluginState{
				Name:   safeStr(item.Name),
				Status: string(item.Status),
			})
		}

		if resp.OpcNextPage == nil {
			break
		}
		req.Page = resp.OpcNextPage
	}

	sortAgentPlugins(all)
	return all, nil
}

func sortAgentPlugins(plugins []AgentPluginState) {
	sort.Slice(plugins, func(i, j int) bool {
		if plugins[i].Name == plugins[j].Name {
			return plugins[i].Status < plugins[j].Status
		}
		return plugins[i].Name < plugins[j].Name
	})
}

func isAgentInstalledWithPlugins(instConfigInstalled bool, plugins []AgentPluginState) bool {
	if len(plugins) > 0 {
		return true
	}
	return instConfigInstalled
}

func formatAgentPlugins(plugins []AgentPluginState) string {
	if len(plugins) == 0 {
		return ""
	}
	parts := make([]string, 0, len(plugins))
	for _, p := range plugins {
		name := strings.TrimSpace(p.Name)
		status := strings.TrimSpace(p.Status)
		if name == "" && status == "" {
			continue
		}
		switch {
		case name == "":
			parts = append(parts, status)
		case status == "":
			parts = append(parts, name)
		default:
			parts = append(parts, name+":"+status)
		}
	}
	return strings.Join(parts, ";")
}
