package children

import (
	"context"
	"fmt"

	"github.com/oracle/oci-go-sdk/example/helpers"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/tenantmanagercontrolplane"
)

var org_id = "ocid1.organizationsentity.oc1.iad.amaaaaaa7zxxlwiah7tpyjgkemo437g5q5oiwabylaac2pzzj7hstjws66dq"

func Children(provider common.ConfigurationProvider, tenancyID string, childFetch bool, homeregion string) {
	fmt.Println("checking child tenancies")

	client, err := tenantmanagercontrolplane.NewOrganizationClientWithConfigurationProvider(provider)
	req := tenantmanagercontrolplane.GetOrganizationRequest{

		OrganizationId: common.String(org_id),
	}

	// Send the request using the service client
	resp, err := client.GetOrganization(context.Background(), req)
	helpers.FatalIfError(err)

	// Retrieve value from the response.
	Organization := resp.Organization
	fmt.Printf("Organization: %v\n", Organization)

	// does this tenancy have children
	GetChildTenancies(provider, tenancyID)
}

func GetChildTenancies(provider common.ConfigurationProvider, tenancyID string) {
	client, err := tenantmanagercontrolplane.NewOrganizationClientWithConfigurationProvider(provider)
	helpers.FatalIfError(err)
	req := tenantmanagercontrolplane.ListOrganizationTenanciesRequest{Limit: common.Int(199),
		OrganizationId: common.String(org_id),
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
