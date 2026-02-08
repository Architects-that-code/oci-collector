package cloudadvisor

import (
	"context"
	"fmt"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/optimizer"
)

// ListAllRecommendations fetches all Cloud Advisor recommendations for the tenancy
// Cloud Advisor (Optimizer) is a home-region service, so region must be the tenancy home region
func ListAllRecommendations(provider common.ConfigurationProvider, tenancyID string, homeRegion string, includeOrganization bool, childTenancyIDs []string) ([]optimizer.RecommendationSummary, error) {
	client, err := optimizer.NewOptimizerClientWithConfigurationProvider(provider)
	if err != nil {
		return nil, fmt.Errorf("create optimizer client: %w", err)
	}
	client.SetRegion(homeRegion)

	var all []optimizer.RecommendationSummary
	seen := make(map[string]bool)
	// ensure we collect all by enumerating every known status and lifecycle state
	var statuses []optimizer.ListRecommendationsStatusEnum
	for _, s := range optimizer.GetListRecommendationsStatusEnumStringValues() {
		if v, ok := optimizer.GetMappingListRecommendationsStatusEnum(s); ok {
			statuses = append(statuses, v)
		}
	}
	var lifecycles []optimizer.ListRecommendationsLifecycleStateEnum
	for _, s := range optimizer.GetListRecommendationsLifecycleStateEnumStringValues() {
		if v, ok := optimizer.GetMappingListRecommendationsLifecycleStateEnum(s); ok {
			lifecycles = append(lifecycles, v)
		}
	}
	for _, lc := range lifecycles {
		for _, st := range statuses {
			req := optimizer.ListRecommendationsRequest{
				CompartmentId:          common.String(tenancyID),
				CompartmentIdInSubtree: common.Bool(true),
				Status:                 st,
				LifecycleState:         lc,
				Limit:                  common.Int(1000),
			}
			if includeOrganization {
				req.IncludeOrganization = common.Bool(true)
			}
			if len(childTenancyIDs) > 0 {
				req.ChildTenancyIds = childTenancyIDs
			}
			for {
				resp, err := client.ListRecommendations(context.Background(), req)
				if err != nil {
					return nil, err
				}
				if resp.RecommendationCollection.Items != nil {
					for _, it := range resp.RecommendationCollection.Items {
						if it.Id != nil && !seen[*it.Id] {
							seen[*it.Id] = true
							all = append(all, it)
						}
					}
				}
				if resp.OpcNextPage != nil && len(*resp.OpcNextPage) > 0 {
					req.Page = resp.OpcNextPage
				} else {
					break
				}
			}
		}
	}

	// Fallback catch-all request with no filters to ensure nothing is missed by SDK enums
	catchAllReq := optimizer.ListRecommendationsRequest{
		CompartmentId:          common.String(tenancyID),
		CompartmentIdInSubtree: common.Bool(true),
		Limit:                  common.Int(1000),
	}
	if includeOrganization {
		catchAllReq.IncludeOrganization = common.Bool(true)
	}
	if len(childTenancyIDs) > 0 {
		catchAllReq.ChildTenancyIds = childTenancyIDs
	}
	for {
		resp, err := client.ListRecommendations(context.Background(), catchAllReq)
		if err != nil {
			return nil, err
		}
		if resp.RecommendationCollection.Items != nil {
			for _, it := range resp.RecommendationCollection.Items {
				if it.Id != nil && !seen[*it.Id] {
					seen[*it.Id] = true
					all = append(all, it)
				}
			}
		}
		if resp.OpcNextPage != nil && len(*resp.OpcNextPage) > 0 {
			catchAllReq.Page = resp.OpcNextPage
		} else {
			break
		}
	}
	return all, nil
}

// ListAllResourceActions fetches all resource actions across all statuses, optionally filtered by recommendationId
// If recommendationID is empty, returns all actions for the tenancy; otherwise only actions for that recommendation.
func ListAllResourceActions(provider common.ConfigurationProvider, tenancyID string, homeRegion string, recommendationID string, includeOrganization bool, childTenancyIDs []string) ([]optimizer.ResourceActionSummary, error) {
	client, err := optimizer.NewOptimizerClientWithConfigurationProvider(provider)
	if err != nil {
		return nil, fmt.Errorf("create optimizer client: %w", err)
	}
	client.SetRegion(homeRegion)

	var all []optimizer.ResourceActionSummary
	seen := make(map[string]bool)
	// iterate over all possible statuses to ensure full coverage
	var statuses []optimizer.ListResourceActionsStatusEnum
	for _, s := range optimizer.GetListResourceActionsStatusEnumStringValues() {
		if v, ok := optimizer.GetMappingListResourceActionsStatusEnum(s); ok {
			statuses = append(statuses, v)
		}
	}
	var lifecycles []optimizer.ListResourceActionsLifecycleStateEnum
	for _, s := range optimizer.GetListResourceActionsLifecycleStateEnumStringValues() {
		if v, ok := optimizer.GetMappingListResourceActionsLifecycleStateEnum(s); ok {
			lifecycles = append(lifecycles, v)
		}
	}
	for _, lc := range lifecycles {
		for _, st := range statuses {
			req := optimizer.ListResourceActionsRequest{
				CompartmentId:          common.String(tenancyID),
				CompartmentIdInSubtree: common.Bool(true),
				Status:                 st,
				LifecycleState:         lc,
				Limit:                  common.Int(1000),
			}
			if includeOrganization {
				req.IncludeOrganization = common.Bool(true)
			}
			if len(childTenancyIDs) > 0 {
				req.ChildTenancyIds = childTenancyIDs
			}
			if recommendationID != "" {
				req.RecommendationId = common.String(recommendationID)
			}

			for {
				resp, err := client.ListResourceActions(context.Background(), req)
				if err != nil {
					return nil, err
				}
				if resp.ResourceActionCollection.Items != nil {
					for _, it := range resp.ResourceActionCollection.Items {
						if it.Id != nil && !seen[*it.Id] {
							seen[*it.Id] = true
							all = append(all, it)
						}
					}
				}
				if resp.OpcNextPage != nil && len(*resp.OpcNextPage) > 0 {
					req.Page = resp.OpcNextPage
				} else {
					break
				}
			}
		}
	}

	// Fallback catch-all request with no filters
	catchAllReq := optimizer.ListResourceActionsRequest{
		CompartmentId:          common.String(tenancyID),
		CompartmentIdInSubtree: common.Bool(true),
		Limit:                  common.Int(1000),
	}
	if includeOrganization {
		catchAllReq.IncludeOrganization = common.Bool(true)
	}
	if len(childTenancyIDs) > 0 {
		catchAllReq.ChildTenancyIds = childTenancyIDs
	}
	if recommendationID != "" {
		catchAllReq.RecommendationId = common.String(recommendationID)
	}
	for {
		resp, err := client.ListResourceActions(context.Background(), catchAllReq)
		if err != nil {
			return nil, err
		}
		if resp.ResourceActionCollection.Items != nil {
			for _, it := range resp.ResourceActionCollection.Items {
				if it.Id != nil && !seen[*it.Id] {
					seen[*it.Id] = true
					all = append(all, it)
				}
			}
		}
		if resp.OpcNextPage != nil && len(*resp.OpcNextPage) > 0 {
			catchAllReq.Page = resp.OpcNextPage
		} else {
			break
		}
	}
	return all, nil
}
