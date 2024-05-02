package peopleresource

import (
	"context"
	"fmt"

	"github.com/oracle/oci-go-sdk/example/helpers"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/identity"
)

func GetAllPeople(provider common.ConfigurationProvider, client identity.IdentityClient, tenancyID string, showUsers bool) []identity.User {

	var allUsers []identity.User

	req := identity.ListUsersRequest{
		CompartmentId: &tenancyID,
		Limit:         common.Int(10000),
	}

	for {
		resp, err := client.ListUsers(context.Background(), req)
		helpers.FatalIfError(err)

		allUsers = append(allUsers, resp.Items...)

		if resp.OpcNextPage != nil {
			req.Page = resp.OpcNextPage
		} else {
			break

		}

		if showUsers {
			for _, user := range resp.Items {
				fmt.Printf("User: %s\n", *user.Name)
			}
		}
	}
	fmt.Printf("users returned %v\n", len(allUsers))
	return allUsers
}
