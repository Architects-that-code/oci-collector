package peopleresource

import (
	"context"
	"fmt"

	"github.com/oracle/oci-go-sdk/example/helpers"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/identity"
)

func GetAllPolicies(provider common.ConfigurationProvider, client identity.IdentityClient, tenancyID string, compartments []identity.Compartment, showPolicies bool, verbose bool) []identity.Policy {

	var policies []identity.Policy

	for _, compartment := range compartments {
		//fmt.Printf("compartment: %v\n", *compartment.Name)
		req := identity.ListPoliciesRequest{
			CompartmentId: compartment.Id,
		}

		// Send the request using the service client
		resp, err := client.ListPolicies(context.Background(), req)
		helpers.FatalIfError(err)

		fmt.Printf("in comp: %v policies returned %v\n", *compartment.Name, len(resp.Items))

		for _, pol := range resp.Items {
			policies = append(policies, pol)
			if showPolicies {
				fmt.Printf("\tComp: %v Policies: %s\n", *compartment.Name, *pol.Name)
				if verbose {
					fmt.Printf("\t\tstatements %v\n", len(pol.Statements))
					for _, statement := range pol.Statements {
						fmt.Printf("\t\t\tStatement: %s\n", statement)
					}
				}
			}
		}
	}

	//return resp.Items
	fmt.Printf("Total policies returned %v\n", len(policies))
	return policies
}
