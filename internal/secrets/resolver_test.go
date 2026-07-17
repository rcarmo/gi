package secrets

import "testing"

func TestEnvNameAndResolver(t *testing.T) {
	if got := EnvName("connect/basic-password.v1"); got != "CONNECT_BASIC_PASSWORD_V1" {
		t.Fatalf("EnvName = %q", got)
	}
	t.Setenv("CONNECT_BASIC_PASSWORD_V1", "secret")
	got, err := (EnvResolver{}).Resolve(t.Context(), "connect/basic-password.v1")
	if err != nil || got != "secret" {
		t.Fatalf("Resolve = %q, %v", got, err)
	}
}
