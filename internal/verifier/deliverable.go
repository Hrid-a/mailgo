package verifier

import (
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/smtp"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/idna"
)

// SMTP holds the results of an SMTP deliverability check
type SMTP struct {
	HostExists  bool `json:"host_exists"` // did we successfully connect to the mail server?
	FullInbox   bool `json:"full_inbox"`  // is the mailbox full?
	CatchAll    bool `json:"catch_all"`   // does the domain accept mail for any address?
	Deliverable bool `json:"deliverable"` // does the specific address exist?
	Disabled    bool `json:"disabled"`    // is the address blocked or disabled?
}

// CheckSMTP performs an SMTP check for the given domain and username.
// Returns nil if smtpCheckEnabled is false.
func (v *Verifier) CheckSMTP(domain, username string) (*SMTP, error) {

	var ret SMTP

	conn, client, mx, err := newSMTPClient(domain, v.connectTimeout)
	if err != nil {
		return &ret, ParseSMTPError(err)
	}
	defer client.Close()

	resetDeadline := func() error {
		return conn.SetDeadline(time.Now().Add(v.operationTimeout))
	}

	_ = mx // available for api-based verifiers in future

	// Sets the HELO/EHLO hostname
	if err = resetDeadline(); err != nil {
		return &ret, ParseSMTPError(err)
	}
	if err = client.Hello(v.helloName); err != nil {
		return &ret, ParseSMTPError(err)
	}

	if err = client.Mail(v.fromEmail); err != nil {
		return &ret, ParseSMTPError(err)
	}

	ret.HostExists = true

	// Default: assume catch-all until proven otherwise
	ret.CatchAll = true

	// catch-all probe: only used to determine if the server accepts all addresses
	if err = client.Rcpt(GenerateRandomEmail(domain)); err != nil {
		if e := ParseSMTPError(err); e != nil {
			switch e.Message {
			case ErrMailboxNotFound, ErrTempUnavailable:
				ret.CatchAll = false
			}
		}
	}

	if username == "" {
		return &ret, nil
	}

	if err = resetDeadline(); err != nil {
		return &ret, ParseSMTPError(err)
	}
	email := fmt.Sprintf("%s@%s", username, domain)
	if err = client.Rcpt(email); err != nil {
		if e := ParseSMTPError(err); e != nil {
			switch e.Message {
			case ErrFullInbox:
				ret.FullInbox = true
			case ErrNotAllowed:
				ret.Disabled = true
			}
		}
	} else {
		ret.Deliverable = true
	}

	return &ret, nil
}

// newSMTPClient dials the first available MX server for the domain and returns
// the connected client along with the MX record that accepted the connection
type smtpResult struct {
	conn   net.Conn
	client *smtp.Client
	mx     *net.MX
}

func newSMTPClient(domain string, connectTimeout time.Duration) (net.Conn, *smtp.Client, *net.MX, error) {
	domain = domainToASCII(domain)

	mxRecords, err := net.LookupMX(domain)
	if err != nil {
		return nil, nil, nil, err
	}
	if len(mxRecords) == 0 {
		return nil, nil, nil, errors.New("no MX records found")
	}

	ch := make(chan any, len(mxRecords))

	var done bool
	var mu sync.Mutex

	for i, r := range mxRecords {
		addr := r.Host + smtpPort
		index := i
		go func() {
			conn, c, err := dialSMTP(addr, connectTimeout)
			if err != nil {
				mu.Lock()
				if !done {
					ch <- err
				}
				mu.Unlock()
				return
			}

			mu.Lock()
			switch {
			case !done:
				done = true
				ch <- smtpResult{conn: conn, client: c, mx: mxRecords[index]}
			default:
				c.Close()
			}
			mu.Unlock()
		}()
	}

	var errs []error
	for {
		res := <-ch
		switch r := res.(type) {
		case smtpResult:
			return r.conn, r.client, r.mx, nil
		case error:
			errs = append(errs, r)
			if len(errs) == len(mxRecords) {
				return nil, nil, nil, errs[0]
			}
		default:
			return nil, nil, nil, errors.New("unexpected response dialing SMTP server")
		}
	}
}

// dialSMTP connects to an SMTP address and returns the raw conn and client.
// The caller is responsible for setting per-operation deadlines on the conn.
func dialSMTP(addr string, connectTimeout time.Duration) (net.Conn, *smtp.Client, error) {
	conn, err := net.DialTimeout("tcp", addr, connectTimeout)
	if err != nil {
		return nil, nil, err
	}

	host, _, _ := net.SplitHostPort(addr)
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		conn.Close()
		return nil, nil, err
	}

	return conn, client, nil
}

// checkDNS populates MX, SPF, and DMARC fields on the Result
func checkDNS(domain string) DNS {
	var dns DNS

	mxRecords, err := net.LookupMX(domain)
	if err == nil && len(mxRecords) > 0 {
		dns.HasMX = true
	}

	txtRecords, _ := net.LookupTXT(domain)
	for _, record := range txtRecords {
		if strings.HasPrefix(record, "v=spf1") {
			dns.HasSPF = true
			dns.SPFRecord = record
			break
		}
	}

	dmarcRecords, _ := net.LookupTXT("_dmarc." + domain)
	for _, record := range dmarcRecords {
		if strings.HasPrefix(record, "v=DMARC1") {
			dns.HasDMARC = true
			dns.DMARCRecord = record
			break
		}
	}

	return dns
}

// domainToASCII converts an internationalized domain name to its ASCII form
func domainToASCII(domain string) string {
	ascii, err := idna.ToASCII(domain)
	if err != nil {
		return domain
	}
	return ascii
}

// GenerateRandomEmail generates a random email address for catch-all detection
func GenerateRandomEmail(domain string) string {
	r := make([]byte, 32)
	for i := 0; i < 32; i++ {
		r[i] = alphanumeric[rand.Intn(len(alphanumeric))]
	}
	return fmt.Sprintf("%s@%s", string(r), domain)
}
