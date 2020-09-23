package keytabs

import (
	"testing"
)

func Test1(t *testing.T) {

	var principals []string
	principals = append(principals, "bob@example.com")
	principals = append(principals, "alice@example.com")

	config := &Config{
		Principals: principals,
	}

	store, err := config.Build()
	if err != nil {
		t.Fatalf("Unexpected err %s", err)
	}
	defer store.Shutdown()

	_, err = store.GetKeytab("bob@example.com")

	if err != nil {
		t.Fatalf("Unexpected err %s", err)
	}

	_, err = store.GetKeytab("invalid@example.com")

	if err == nil {
		t.Fatalf("Expected err %s", err)
	}

}
