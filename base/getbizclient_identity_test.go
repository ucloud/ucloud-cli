package base

import "testing"

// TestGetBizClientKeepsAuthCredentialIdentity locks the invariant that a token
// refresh via GetBizClient overwrites the existing *AuthCredential in place
// instead of swapping the package pointer.
//
// Product service clients (cli.NewServiceClient) capture the AuthCredential
// pointer at command-tree registration and read AccessToken lazily per request.
// If GetBizClient replaced the pointer, those already-built clients would keep
// sending the pre-refresh (expired) Bearer on the first request and only recover
// through the reactive retry handler at the cost of a wasted round-trip.
func TestGetBizClientKeepsAuthCredentialIdentity(t *testing.T) {
	savedCred, savedCfg := AuthCredential, ClientConfig
	t.Cleanup(func() { AuthCredential, ClientConfig = savedCred, savedCfg })

	AuthCredential = &CredentialConfig{AuthMode: AuthModeOAuth, AccessToken: "stale-token"}
	captured := AuthCredential // the pointer a product client would have captured

	retries := 3
	ac := &AggConfig{
		BaseURL:       "https://api.ucloud.cn/",
		Timeout:       15,
		Region:        "cn-bj2",
		ProjectID:     "org-test",
		MaxRetryTimes: &retries,
		AuthMode:      AuthModeOAuth,
		AccessToken:   "fresh-token",
	}
	if _, err := GetBizClient(ac); err != nil {
		t.Fatalf("GetBizClient returned error: %v", err)
	}

	if AuthCredential != captured {
		t.Fatal("AuthCredential pointer was replaced; registration-time product clients would keep the stale token")
	}
	if captured.AccessToken != "fresh-token" {
		t.Fatalf("captured credential not refreshed in place: got %q, want %q", captured.AccessToken, "fresh-token")
	}
}
