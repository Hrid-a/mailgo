package verifier

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseAddress(t *testing.T) {

	tests := map[string]struct {
		input string
		want  Address
	}{
		"valid email": {
			input: "email_username@domain.com",
			want: Address{
				Valid:    true,
				Domain:   "domain.com",
				Username: "email_username",
			},
		},
		"invalid email": {
			input: "email_invalid@",
			want: Address{
				Valid:    false,
				Username: "",
				Domain:   "",
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := ParseAddress(tc.input)
			diff := cmp.Diff(tc.want, got)

			if diff != "" {
				t.Fatalf("diff %v", diff)
			}
		})
	}
}
