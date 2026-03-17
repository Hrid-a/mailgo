package verifier

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCheckSMTP(t *testing.T) {

	v, err := NewVerifier()
	if err != nil {
		t.Fatalf("failed to create verifier: %v", err)
	}

	tests := map[string]struct {
		domain   string
		username string
		want     *SMTP
	}{
		"valid domain": {
			domain:   "ahmedhrid.com",
			username: "me",
			want: &SMTP{
				HostExists:  true,
				CatchAll:    true,
				Deliverable: true,
			},
		},
		"non catch-all domain": {
			domain:   "gmail.com",
			username: "",
			want: &SMTP{
				HostExists: true,
				CatchAll:   false,
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := v.CheckSMTP(tc.domain, tc.username)
			if err != nil {
				t.Fatalf("CheckSMTP failed: %v", err)
			}

			diff := cmp.Diff(tc.want, got)
			if diff != "" {
				t.Fatalf("diff %v", diff)
			}
		})
	}
}
