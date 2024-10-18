package peopleresource

import (
	"context"
	"fmt"

	"github.com/oracle/oci-go-sdk/example/helpers"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/identity"
)

func GetAllPeople(provider common.ConfigurationProvider, client identity.IdentityClient, tenancyID string, showUsers bool) []identity.User {
	fmt.Println(showUsers)

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

	}
	fmt.Printf("users returned %v\n", len(allUsers))
	if showUsers {
		for _, user := range allUsers {
			n := *user.Name
			tc := *user.TimeCreated

			fmt.Printf("User: %s\t Created: %s \t \n", n, tc)
		}
	}
	return allUsers
}
