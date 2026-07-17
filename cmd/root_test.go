package cmd

import "testing"

// TestParseWaitTimeoutSec covers the --wait-timeout-sec startup pre-scan
// helper, which must accept both the space form (--wait-timeout-sec 1800)
// and the equals form (--wait-timeout-sec=1800). See issue #119: the equals
// form was previously silently ignored because init()'s manual os.Args scan
// only matched arg == "--wait-timeout-sec" exactly.
func TestParseWaitTimeoutSec(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantSec   int
		wantFound bool
		wantErr   bool
	}{
		{
			name:      "space form",
			args:      []string{"ucloud", "uhost", "--wait-timeout-sec", "1800"},
			wantSec:   1800,
			wantFound: true,
			wantErr:   false,
		},
		{
			name:      "equals form",
			args:      []string{"ucloud", "uhost", "--wait-timeout-sec=1800"},
			wantSec:   1800,
			wantFound: true,
			wantErr:   false,
		},
		{
			name:      "absent",
			args:      []string{"ucloud", "uhost"},
			wantSec:   0,
			wantFound: false,
			wantErr:   false,
		},
		{
			name:      "equals with bad int",
			args:      []string{"ucloud", "--wait-timeout-sec=abc"},
			wantSec:   0,
			wantFound: true,
			wantErr:   true,
		},
		{
			name:      "space with bad int",
			args:      []string{"ucloud", "--wait-timeout-sec", "abc"},
			wantSec:   0,
			wantFound: true,
			wantErr:   true,
		},
		{
			name:      "trailing flag no value",
			args:      []string{"ucloud", "--wait-timeout-sec"},
			wantSec:   0,
			wantFound: false,
			wantErr:   false,
		},
		{
			name:      "empty equals",
			args:      []string{"ucloud", "--wait-timeout-sec="},
			wantSec:   0,
			wantFound: false,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sec, found, err := parseWaitTimeoutSec(tt.args)
			if sec != tt.wantSec {
				t.Errorf("sec = %d, want %d", sec, tt.wantSec)
			}
			if found != tt.wantFound {
				t.Errorf("found = %v, want %v", found, tt.wantFound)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("err = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestScanFlagValue covers the generalized startup pre-scan helper. Same #119
// bug as parseWaitTimeoutSec (equals form silently ignored), now for the
// connection-class flags. The critical -pfoo case pins the R3 boundary:
// attached/combined shorthand is intentionally NOT recognized.
func TestScanFlagValue(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		names     []string
		wantVal   string
		wantFound bool
	}{
		{
			name:      "long space form",
			args:      []string{"ucloud", "config", "--profile", "foo"},
			names:     []string{"--profile", "-p"},
			wantVal:   "foo",
			wantFound: true,
		},
		{
			name:      "long equals form",
			args:      []string{"ucloud", "config", "--profile=foo"},
			names:     []string{"--profile", "-p"},
			wantVal:   "foo",
			wantFound: true,
		},
		{
			name:      "short space form",
			args:      []string{"ucloud", "config", "-p", "foo"},
			names:     []string{"--profile", "-p"},
			wantVal:   "foo",
			wantFound: true,
		},
		{
			name:      "short equals form",
			args:      []string{"ucloud", "config", "-p=foo"},
			names:     []string{"--profile", "-p"},
			wantVal:   "foo",
			wantFound: true,
		},
		{
			// R3 boundary: attached shorthand is out of scope, must NOT match.
			name:      "attached shorthand not recognized",
			args:      []string{"ucloud", "config", "-pfoo"},
			names:     []string{"--profile", "-p"},
			wantVal:   "",
			wantFound: false,
		},
		{
			name:      "empty equals is no hit",
			args:      []string{"ucloud", "config", "--profile="},
			names:     []string{"--profile", "-p"},
			wantVal:   "",
			wantFound: false,
		},
		{
			name:      "trailing flag no value",
			args:      []string{"ucloud", "config", "--profile"},
			names:     []string{"--profile", "-p"},
			wantVal:   "",
			wantFound: false,
		},
		{
			name:      "absent",
			args:      []string{"ucloud", "config"},
			names:     []string{"--profile", "-p"},
			wantVal:   "",
			wantFound: false,
		},
		{
			name:      "base-url equals form single name",
			args:      []string{"ucloud", "uhost", "list", "--base-url=http://x/"},
			names:     []string{"--base-url"},
			wantVal:   "http://x/",
			wantFound: true,
		},
		{
			// leftmost hit wins (pre-scan only needs one early value, no override)
			name:      "leftmost wins",
			args:      []string{"ucloud", "--profile", "a", "--profile", "b"},
			names:     []string{"--profile", "-p"},
			wantVal:   "a",
			wantFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, found := scanFlagValue(tt.args, tt.names...)
			if val != tt.wantVal {
				t.Errorf("val = %q, want %q", val, tt.wantVal)
			}
			if found != tt.wantFound {
				t.Errorf("found = %v, want %v", found, tt.wantFound)
			}
		})
	}
}
