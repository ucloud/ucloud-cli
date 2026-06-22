package cmd

import "testing"

// TestProfileCompletionRegistered guards against the regression where --profile
// completion was registered on cmd.Flags() instead of cmd.PersistentFlags().
// Because "profile" is a persistent flag, it only lives on PersistentFlags() at
// registration time; cmd.Flags() silent no-ops on it, silently dropping tab
// completion. See fix: command.SetPersistentCompletion.
func TestProfileCompletionRegistered(t *testing.T) {
	root := NewCmdRoot()
	if root.PersistentFlags().GetFlagValuesFunc("profile") == nil {
		t.Fatal("profile completion not registered on persistent flags")
	}
}
