package children

import (
	config "check-limits/config"
	"context"
	"fmt"

	"github.com/oracle/oci-go-sdk/example/helpers"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/governancerulescontrolplane"
	"github.com/oracle/oci-go-sdk/v65/identity"
	"github.com/oracle/oci-go-sdk/v65/tenantmanagercontrolplane"
)

func Children(provider common.ConfigurationProvider, passThruClient identity.IdentityClient, tenancyID string, childFetch bool, homeregion string, config config.Config) {
	fmt.Println("checking child tenancies Children")
	fmt.Println("OrgId: ", config.ORG_ID)

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
	GetChildTenancies(provider, passThruClient, tenancyID, config)
}

func GetChildTenancies(provider common.ConfigurationProvider, passThruClient identity.IdentityClient, tenancyID string, config config.Config) {
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
	fmt.Printf("count child OrganizationTenancies: %v\n", len(resp.Items))
	for _, tenancy := range resp.Items {
		fmt.Printf("Tenancy: %v\n", tenancy)
		getAllPeople(provider, passThruClient, *tenancy.TenancyId, true)
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

func getAllPeople(provider common.ConfigurationProvider, passThruClient identity.IdentityClient, tenancyID string, showUsers bool) []identity.User {
	fmt.Printf("Showusers: %v \n", showUsers)

	var allUsers []identity.User

	req := identity.ListUsersRequest{
		CompartmentId: &tenancyID,
		Limit:         common.Int(10000),
	}

	for {
		resp, err := passThruClient.ListUsers(context.Background(), req)
		if err != nil {
			fmt.Printf("error %v\n", tenancyID)
			break
		} else {
			fmt.Printf("success %v\n", tenancyID)
			allUsers = append(allUsers, resp.Items...)
			if resp.OpcNextPage != nil {
				req.Page = resp.OpcNextPage
			} else {
				break
			}

		}
		fmt.Printf("users returned %v\n", len(allUsers))
		if showUsers {
			for _, user := range allUsers {
				n := *user.Name
				tc := *user.TimeCreated

				fmt.Printf("User: %s\t Created: %s \t \n", n, tc)
			}
		}
	}
	fmt.Printf("Showusers: end \n")
	return allUsers
}
