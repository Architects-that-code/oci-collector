package supportresources

import (
	"testing"

	"github.com/oracle/oci-go-sdk/v65/common"
)

func TestCloseTicketNoPanic(t *testing.T) {
	CloseTicket()
}

func TestSupportFunctionSignatures(t *testing.T) {
	var listFn func(common.ConfigurationProvider, string, string, string)
	listFn = ListTickets
	if listFn == nil {
		t.Fatal("ListTickets function must be assignable and non-nil")
	}

	var createFn func(common.ConfigurationProvider, string, string, string)
	createFn = CreateTicket
	if createFn == nil {
		t.Fatal("CreateTicket function must be assignable and non-nil")
	}
}
