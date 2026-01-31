package peopleresource

import (
	"context"
	"fmt"

	"github.com/oracle/oci-go-sdk/example/helpers"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/identity"
)

func Groups(provider common.ConfigurationProvider, client identity.IdentityClient, tenancyID string, showGroup bool) []identity.Group {

	var allGroups []identity.Group

	req := identity.ListGroupsRequest{
		CompartmentId: &tenancyID,
		Limit:         common.Int(10000),
	}

	resp, err := client.ListGroups(context.Background(), req)
	helpers.FatalIfError(err)

	for _, group := range resp.Items {
		allGroups = append(allGroups, group)
		if showGroup {
			fmt.Printf("Group: %s\n", *group.Name)
			fmt.Printf("Group ID: %s\n", *group.Id)
			fmt.Printf("Group Description: %s\n", *group.Description)
			fmt.Printf("Group Lifecycle State: %s\n", string(group.LifecycleState))
			fmt.Printf("Group Time Created: %s\n", group.TimeCreated.String())
			fmt.Printf("\n")

		}

	}

	fmt.Printf("groups returned %v\n", len(allGroups))

	return allGroups
}
