package resourcesearch

import (
	"context"
	"fmt"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/example/helpers"
	"github.com/oracle/oci-go-sdk/v65/resourcesearch"
)

func Search(provider common.ConfigurationProvider, tenancyID string, searchString string) {
	client, err := resourcesearch.NewResourceSearchClientWithConfigurationProvider(common.DefaultConfigProvider())
	helpers.FatalIfError(err)

	fmt.Printf("Searching for resources with the string: %s\n", searchString)

	//	query := fmt.Sprintf("createdBy LIKE '%s%%' AND lifecycleState = 'ACTIVE'", searchString)

	req := resourcesearch.SearchResourcesRequest{
		TenantId: common.String(tenancyID),
		Limit:    common.Int(514),

		SearchDetails: resourcesearch.FreeTextSearchDetails{MatchingContextType: resourcesearch.SearchDetailsMatchingContextTypeHighlights,
			Text: common.String(searchString)}}

	// Send the request using the service client
	resp, err := client.SearchResources(context.Background(), req)
	helpers.FatalIfError(err)

	// Retrieve value from the response.
	for _, item := range resp.Items {
		fmt.Println(item)
	}
}
