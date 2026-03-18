package verifier

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"time"
)

type Verifier struct {
	// SMTP config
	fromEmail        string
	helloName        string
	connectTimeout   time.Duration
	operationTimeout time.Duration

	// input
	emails []string
	input  io.Reader
	closer io.Closer

	// output
	output      io.Writer
	jsonOutput  bool
	concurrency int
}

type Status string

const (
	StatusDeliverable   Status = "deliverable"
	StatusUndeliverable Status = "undeliverable"
	StatusRisky         Status = "risky"
	StatusUnknown       Status = "unknown"
)

type Result struct {
	Email  string `json:"email"`
	Domain string `json:"domain"`
	Valid  bool   `json:"valid"`
	Status Status `json:"status"`
	SMTP   SMTP   `json:"smtp"`
	Error  string `json:"error,omitempty"`
}

type Option func(*Verifier) error

func WithFromEmail(email string) Option {
	return func(v *Verifier) error {
		if email == "" {
			return errors.New("fromEmail cannot be empty")
		}
		v.fromEmail = email
		return nil
	}
}

func WithConnectTimeout(d time.Duration) Option {
	return func(v *Verifier) error {
		v.connectTimeout = d
		return nil
	}
}

func WithOperationTimeout(d time.Duration) Option {
	return func(v *Verifier) error {
		v.operationTimeout = d
		return nil
	}
}

func WithEmailArg(email string) Option {
	return func(v *Verifier) error {
		if email == "" {
			return errors.New("email cannot be empty")
		}
		v.emails = []string{email}
		return nil
	}
}

func WithEmailsFromFile(path string) Option {
	return func(v *Verifier) error {
		f, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("could not open file: %w", err)
		}
		v.input = f
		v.closer = f
		return nil
	}
}

func WithOutputFile(path string) Option {
	return func(v *Verifier) error {
		f, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("could not create output file: %w", err)
		}
		v.output = f
		return nil
	}
}

func WithJSONOutput() Option {
	return func(v *Verifier) error {
		v.jsonOutput = true
		return nil
	}
}

func NewVerifier(opts ...Option) (*Verifier, error) {
	v := &Verifier{
		fromEmail:        "verify@gmail.com",
		connectTimeout:   10 * time.Second,
		operationTimeout: 15 * time.Second,
		output:           os.Stdout,
		concurrency:      5,
	}

	for _, opt := range opts {
		if err := opt(v); err != nil {
			return nil, err
		}
	}

	helloName, err := getPTR()
	if err != nil {
		return nil, err
	}
	v.helloName = helloName

	return v, nil
}

func (v *Verifier) writeResult(r Result) error {
	if v.jsonOutput {
		b, err := json.MarshalIndent(r, "", "  ")
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(v.output, "%s\n", b)
		return err
	}

	errLine := ""
	if r.Error != "" {
		errLine = "Error       : " + r.Error + "\n"
	}

	_, err := fmt.Fprintf(v.output,
		"\n\nEmail       : %s\nDomain      : %s\nValid       : %v\nStatus      : %v\n%sHost Exists : %v\nCatch-All   : %v\nDeliverable : %v\nFull Inbox  : %v\nDisabled    : %v\n\n%s\n",
		r.Email,
		r.Domain,
		r.Valid,
		r.Status,
		errLine,
		r.SMTP.HostExists,
		r.SMTP.CatchAll,
		r.SMTP.Deliverable,
		r.SMTP.FullInbox,
		r.SMTP.Disabled,
		"---",
	)
	return err
}

func (v *Verifier) Verify(email string) (Result, error) {

	address := ParseAddress(email)
	if !address.Valid {
		return Result{
			Email:  email,
			Valid:  false,
			Status: StatusUndeliverable,
		}, nil
	}

	result := Result{
		Email:  email,
		Domain: address.Domain,
	}

	// SMTP deliverability check

	smtp, err := v.CheckSMTP(address.Domain, address.Username)
	if err != nil {
		result.Error = err.Error()
	}

	if smtp == nil {
		result.Status = StatusUnknown
		result.Valid = false
		return result, nil
	}

	result.SMTP = *smtp
	switch {
	case smtp.Deliverable:
		result.Status = StatusDeliverable
	case smtp.CatchAll:
		result.Status = StatusUnknown
	case smtp.HostExists:
		// Connected fine but RCPT TO was rejected — could be server
		// anti-harvesting policy, not necessarily a missing mailbox
		result.Status = StatusRisky
	default:
		result.Status = StatusUndeliverable
	}

	result.Valid = result.Status == StatusDeliverable

	return result, nil

}

func (v *Verifier) Run() error {
	if v.closer != nil {
		defer v.closer.Close()
	}

	process := func(email string) error {
		result, err := v.Verify(email)
		if err != nil {
			result.Error = err.Error()
		}
		return v.writeResult(result)
	}

	if v.input != nil {
		scanner := bufio.NewScanner(v.input)
		for scanner.Scan() {
			if err := process(scanner.Text()); err != nil {
				return err
			}
		}
		return scanner.Err()
	}

	for _, email := range v.emails {
		if err := process(email); err != nil {
			return err
		}
	}

	return nil
}
