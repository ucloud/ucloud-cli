package cmd

import "testing"

// TestProfileCompletionRegistered guards against the regression where --profile
// completion was dropped. "profile" is a persistent flag; upstream cobra's
// RegisterFlagCompletionFunc (via command.SetPersistentCompletion) registers the
// completion on the command and GetFlagCompletionFunc resolves it.
func TestProfileCompletionRegistered(t *testing.T) {
	root := NewCmdRoot()
	if _, ok := root.GetFlagCompletionFunc("profile"); !ok {
		t.Fatal("profile completion not registered for persistent --profile flag")
	}
}
