package verifier

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestVerifyWithEmailArg(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input string
		want  Result
	}{
		"invalid email - no domain": {
			input: "user@",
			want: Result{
				Email:  "user@",
				Valid:  false,
				Status: StatusUndeliverable,
			},
		},
		"invalid email - no @": {
			input: "notanemail",
			want: Result{
				Email:  "notanemail",
				Valid:  false,
				Status: StatusUndeliverable,
			},
		},
		"invalid email - empty": {
			input: "@domain.com",
			want: Result{
				Email:  "@domain.com",
				Valid:  false,
				Status: StatusUndeliverable,
			},
		},
	}

	v, err := NewVerifier(WithEmailArg("placeholder@example.com"))
	if err != nil {
		t.Fatalf("NewVerifier failed: %v", err)
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := v.Verify(tc.input)
			if err != nil {
				t.Fatalf("Verify failed: %v", err)
			}

			diff := cmp.Diff(tc.want, got)
			if diff != "" {
				t.Fatalf("diff %v", diff)
			}
		})
	}
}

func TestWriteResult(t *testing.T) {
	t.Parallel()

	result := Result{
		Email:  "test@example.com",
		Domain: "example.com",
		Valid:  true,
		Status: StatusDeliverable,
		DNS: DNS{
			HasMX:     true,
			HasSPF:    true,
			SPFRecord: "v=spf1 include:_spf.example.com ~all",
		},
		SMTP: SMTP{
			HostExists:  true,
			Deliverable: true,
		},
	}

	t.Run("text output contains key fields", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		v := &Verifier{output: &buf}

		if err := v.writeResult(result); err != nil {
			t.Fatalf("writeResult failed: %v", err)
		}

		out := buf.String()
		for _, want := range []string{
			result.Email,
			result.Domain,
			string(result.Status),
			result.DNS.SPFRecord,
		} {
			if !strings.Contains(out, want) {
				t.Errorf("output missing %q\ngot:\n%s", want, out)
			}
		}
	})

	t.Run("json output is valid and matches result", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		v := &Verifier{output: &buf, jsonOutput: true}

		if err := v.writeResult(result); err != nil {
			t.Fatalf("writeResult failed: %v", err)
		}

		var got Result
		if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
			t.Fatalf("output is not valid JSON: %v", err)
		}

		diff := cmp.Diff(result, got)
		if diff != "" {
			t.Fatalf("JSON result mismatch:\n%s", diff)
		}
	})
}
