package tokens

import (
	"reflect"
	"testing"
)

// TODO Remove these tokens and have the test pull current tokens

// func Test1(t *testing.T) {

// 	config := NewConfig()

// 	s, err := NewTokenStore(config)
// 	if err != nil {
// 		t.Fatalf("Error not expected: %s", err)
// 	}
// 	defer s.Shutdown()

// 	validExpiredToken, err := s.GetToken(validExpiredTokenStr)

// 	if err != nil {
// 		t.Fatalf("Error not expected: %s", err)
// 	}

// 	assertEqual(t, validExpiredToken.Iss, "https://api.console.aporeto.com/v/1/namespaces/5ddc396b9facec0001d3c886/oauthinfo")

// 	_, err = s.GetToken(invalidTokenStr)

// 	if err == nil {
// 		t.Fatalf("Expected error for invalid token")
// 	}

// }

func assertEqual(t *testing.T, a interface{}, b interface{}) {
	if a == b {
		return
	}
	// debug.PrintStack()
	t.Errorf("Received %v (type %v), expected %v (type %v)", a, reflect.TypeOf(a), b, reflect.TypeOf(b))
}
