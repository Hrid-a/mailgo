package verifier

import (
	"context"
	"io"
	"net"
	"net/http"
	"regexp"
	"strings"
)

var emailRegex = regexp.MustCompile(emailRegexString)

// Address stores all information about an email Address
type Address struct {
	Username string `json:"username"`
	Domain   string `json:"domain"`
	Valid    bool   `json:"valid"`
}

// ParseAddress attempts to parse an email address and return it in the form of an Address
func ParseAddress(email string) Address {

	isAddressValid := IsAddressValid(email)
	if !isAddressValid {
		return Address{Valid: false}
	}

	index := strings.LastIndex(email, "@")
	username := email[:index]
	domain := strings.ToLower(email[index+1:])

	return Address{
		Username: username,
		Domain:   domain,
		Valid:    isAddressValid,
	}
}

// IsAddressValid checks if email address is formatted correctly by using regex
func IsAddressValid(email string) bool {
	return emailRegex.MatchString(email)
}

func getPTR() (string, error) {
	client := &http.Client{}

	resp, err := client.Get("https://api.ipify.org")
	if err != nil {
		return "localhost", nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "localhost", nil
	}

	ip := strings.TrimSpace(string(body))

	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{}
			return d.DialContext(ctx, "udp", "8.8.8.8:53")
		},
	}

	hostnames, err := resolver.LookupAddr(context.Background(), ip)
	if err != nil || len(hostnames) == 0 {
		return "localhost", nil
	}

	return strings.TrimSuffix(hostnames[0], "."), nil
}
