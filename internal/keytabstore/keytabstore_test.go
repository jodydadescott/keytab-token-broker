package keytabstore

import (
	"fmt"
	"testing"
)

func Test1(t *testing.T) {

	store := NewKeytabStoreDefault()
	defer store.Shutdown()

	store.AddPrincipal("bob@example.com")
	store.AddPrincipal("alice@example.com")

	keytab, err := store.GetKeytab("bob@example.com")

	if err == nil {
		t.Fatalf("Expected err %s", err)
	}

	fmt.Println(keytab.JSON())

}
