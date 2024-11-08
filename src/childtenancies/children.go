package children

import (
	config "check-limits/config"
	"context"
	"fmt"

	"github.com/oracle/oci-go-sdk/example/helpers"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/governancerulescontrolplane"
	"github.com/oracle/oci-go-sdk/v65/tenantmanagercontrolplane"
)

func Children(provider common.ConfigurationProvider, tenancyID string, childFetch bool, homeregion string, config config.Config) {
	fmt.Println("checking child tenancies Children")

	client, err := tenantmanagercontrolplane.NewOrganizationClientWithConfigurationProvider(provider)
	req := tenantmanagercontrolplane.GetOrganizationRequest{

		OrganizationId: common.String(config.ORG_ID),
	}
	helpers.FatalIfError(err)
	// Send the request using the service client
	resp, err := client.GetOrganization(context.Background(), req)
	helpers.FatalIfError(err)

	// Retrieve value from the response.
	Organization := resp.Organization
	fmt.Printf("Organization: %v\n", Organization)

	// does this tenancy have children
	GetChildTenancies(provider, tenancyID, config)
}

func GetChildTenancies(provider common.ConfigurationProvider, tenancyID string, config config.Config) {
	fmt.Println("checking child tenancies GetChildTenancies")
	client, err := tenantmanagercontrolplane.NewOrganizationClientWithConfigurationProvider(provider)
	helpers.FatalIfError(err)
	req := tenantmanagercontrolplane.ListOrganizationTenanciesRequest{
		Limit:          common.Int(1000),
		OrganizationId: common.String(config.ORG_ID),
	}

	// Send the request using the service client
	resp, err := client.ListOrganizationTenancies(context.Background(), req)
	helpers.FatalIfError(err)

	// Retrieve value from the response.
	fmt.Printf("count tenancies: %v\n", len(resp.Items))
	for _, tenancy := range resp.Items {
		fmt.Printf("Tenancy: %v\n", tenancy)
	}
}
func Deets(provider common.ConfigurationProvider, tenancyID string, homeregion string, config config.Config) {
	fmt.Println("checking child tenancies Deets")
	client, err := tenantmanagercontrolplane.NewOrganizationClientWithConfigurationProvider(provider)
	helpers.FatalIfError(err)
	req := tenantmanagercontrolplane.GetOrganizationTenancyRequest{
		OrganizationId: common.String(config.ORG_ID),
		TenancyId:      common.String(tenancyID),
	}

	resp, err := client.GetOrganizationTenancy(context.Background(), req)
	helpers.FatalIfError(err)
	OrganizationTenancy := resp.OrganizationTenancy

	gclient, err := governancerulescontrolplane.NewGovernanceRuleClientWithConfigurationProvider(provider)
	helpers.FatalIfError(err)

	listReq := governancerulescontrolplane.ListGovernanceRulesRequest{
		CompartmentId: common.String(tenancyID),
	}

	listResp, err := gclient.ListGovernanceRules(context.Background(), listReq)

	helpers.FatalIfError(err)
	//fmt.Printf("list response Governance: %v\n", listResp)
	for _, rule := range listResp.Items {
		fmt.Printf("rule: %v\n", rule)
		/*
			delReq := governancerulescontrolplane.DeleteGovernanceRuleRequest{
				GovernanceRuleId: rule.Id,
			}
			delResp, err := gclient.DeleteGovernanceRule(context.Background(), delReq)
			helpers.FatalIfError(err)
			fmt.Printf("delete response Governance: %v\n", delResp)
		*/

	}

	fmt.Printf("Organization Tenancy: %v\n", OrganizationTenancy)

}
