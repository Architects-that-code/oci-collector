package resourcesearch

import (
	"testing"

	"github.com/oracle/oci-go-sdk/v65/common"
)

func TestSearchFunctionSignature(t *testing.T) {
	var fn func(common.ConfigurationProvider, string, string)
	fn = Search
	if fn == nil {
		t.Fatal("Search function must be assignable and non-nil")
	}
}
