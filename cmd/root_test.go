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
