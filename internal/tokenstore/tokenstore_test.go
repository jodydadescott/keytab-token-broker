package tokenstore

import (
	"reflect"
	"testing"
)

const validExpiredTokenStr = "eyJhbGciOiJFUzI1NiIsImtpZCI6IjVlM2RjNWNhYjE1NjhjMDAwMWU3YzJjMSIsInR5cCI6IkpXVCJ9.eyJzZXJ2aWNlIjp7IkBjbG91ZDphd3M6YW1pLWlkIjoiYW1pLTBhNTRhZWY0ZWYzYjVmODgxIiwiaWxvdmUiOiJ0aGU4MHMiLCJrZXl0YWIiOiJwdXJwbGUifSwiYXVkIjoibmljbyIsImV4cCI6MTU5OTA3MjUxNCwiaWF0IjoxNTk5MDY4OTE0LCJpc3MiOiJodHRwczovL2FwaS5jb25zb2xlLmFwb3JldG8uY29tL3YvMS9uYW1lc3BhY2VzLzVkZGMzOTZiOWZhY2VjMDAwMWQzYzg4Ni9vYXV0aGluZm8iLCJzdWIiOiI1ZjRmZGE5ZGI2ZWFlMDAwMDEwZjc3MzUifQ.B_knd8t8nvn6wNHTDcud6KpPSNOeUyifPwKQ8tEEXNkRLuDcf1igeDQ9HnUNURBE15yoDL_lgbCrzvrIIIl90g"

const invalidTokenStr = "eyJhbGciOiJFUzI1NiIsImtpZCI6IjVlM2RjNWNhYjE1NjhjMDAwMWU3YzJjMSIsInR5cCI6IkpXVCJ9.eyJzZXJ2aWNlIjp7IkBjbG91ZDphd3M6YW1pLWlkIjoiYW1pLTBhNTRhZWY0ZWYzYjVmODgxIiwiQGNsb3VkOmF3czpyZWdpb24iOiJ1cy1lYXN0LTIiLCJjbG91ZDphd3M6cm9sZSI6ImFybjphd3M6aWFtOjo0Mzk5MTcwNTUzNDA6aW5zdGFuY2UtcHJvZmlsZS9hdXRvcmVnIiwiaWxvdmUiOiJ0aGU4MHMifSwiYXVkIjoibmljbyIsImV4cCI6MTU5ODk4MzMyMCwiaWF0IjoxNTk4OTc5NzIwLCJpc3MiOiJodHRwczovL2FwaS5jb25zb2xlLmFwb3JldG8uY29tL3YvMS9uYW1lc3BhY2VzLzVkZGMzOTZiOWZhY2VjMDAwMWQzYzg4Ni9vYXV0aGluZm8iLCJzdWIiOiI1ZjRlN2UyNGI2ZWFlMDAwMDEwZjc1MzAifQ.e8Rz_A2ziaXtusi6PPqYqjaRU406uQT-mQwiE-x307Kx824q-tEF2F4iPpVXhIgW1XK6UMlBH2fOjLpWYVEEqA"

func Test1(t *testing.T) {

	s := NewConfig().Build()
	defer s.Shutdown()

	validExpiredToken, err := s.GetToken(validExpiredTokenStr)

	if err != nil {
		t.Fatalf("Error not expected: %s", err)
	}

	assertEqual(t, validExpiredToken.Iss, "https://api.console.aporeto.com/v/1/namespaces/5ddc396b9facec0001d3c886/oauthinfo")

	_, err = s.GetToken(invalidTokenStr)

	if err == nil {
		t.Fatalf("Expected error for invalid token")
	}

}

func assertEqual(t *testing.T, a interface{}, b interface{}) {
	if a == b {
		return
	}
	// debug.PrintStack()
	t.Errorf("Received %v (type %v), expected %v (type %v)", a, reflect.TypeOf(a), b, reflect.TypeOf(b))
}
