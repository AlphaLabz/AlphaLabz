package smtp

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/miekg/dns"
)

func VerifyDomain(domain string) (bool, error) {
	tokenFilePath := "alphalab_DNS_token.json"
	tokenData, err := tokenGenerator(tokenFilePath, domain) // Get TokenData struct
	if err != nil {
		return false, fmt.Errorf("failed to generate token: %w", err)
	}
	log.Printf("Your verification token is: %s", tokenData.Token)
	log.Printf("Please create a TXT record with the following details:\n")
	log.Printf("Record Name: _alphalab-verification.%s\nRecord Value: %s\n", domain, tokenData.Token)
	log.Println("Waiting for verification...")

	// try evert 15sec for 10 times
	const maxRetries = 10
	const retryInterval = 15 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		log.Printf("Attempt %d of %d: Verifying DNS TXT record...", attempt, maxRetries)
		verified, err := verifyDNSTXTRecord(domain, tokenData.Token) // Pass the Token field
		if err != nil {
			log.Printf("Error during verification: %v", err)
			log.Println("Retrying...")
		} else if verified {
			log.Println("Domain verified successfully!")
			return true, nil
		} else {
			log.Println("Verification failed. Ensure the TXT record is set up correctly.")
		}

		if attempt < maxRetries {
			log.Printf("Waiting %s before retrying...\n", retryInterval)
			time.Sleep(retryInterval)
		} else {
			log.Println("Maximum retries reached. Verification failed.")
		}
	}

	return false, fmt.Errorf("domain verification failed after %d attempts", maxRetries)
}

// tokenGenerator generates a token in the format:
// "16-digit random code - domain HEX8 - timestamp HEX8" and saves it in a JSON file.
// If the file already exists, it reads and returns the existing data.
// TokenData represents the structure of the token JSON file
type TokenData struct {
	Domain string `json:"domain"`
	Token  string `json:"token"`
}

func tokenGenerator(filePath, domain string) (TokenData, error) {
	// Check if the token file already exists
	if _, err := os.Stat(filePath); err == nil {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return TokenData{}, fmt.Errorf("failed to read token file: %w", err)
		}

		var tokenData TokenData
		if err := json.Unmarshal(data, &tokenData); err != nil {
			return TokenData{}, fmt.Errorf("failed to parse token JSON: %w", err)
		}
		return tokenData, nil
	}

	// Generate a random 16-digit code
	randomBytes := make([]byte, 8) // 16 characters
	if _, err := rand.Read(randomBytes); err != nil {
		return TokenData{}, fmt.Errorf("failed to generate random bytes: %w", err)
	}
	randomCode := hex.EncodeToString(randomBytes)

	// Generate HEX8 for the domain
	domainBytes := make([]byte, 4) // 8 characters
	if _, err := rand.Read(domainBytes); err != nil {
		return TokenData{}, fmt.Errorf("failed to generate random bytes for domain: %w", err)
	}
	domainHex := hex.EncodeToString(domainBytes)

	// Generate HEX8 for the current timestamp
	timestamp := time.Now().Unix()
	timestampHex := fmt.Sprintf("%08x", timestamp)

	// Create the token
	token := fmt.Sprintf("%s-%s-%s", randomCode, domainHex, timestampHex)

	// Construct the token data
	tokenData := TokenData{
		Domain: domain,
		Token:  token,
	}

	// Save the token as JSON
	jsonData, err := json.MarshalIndent(tokenData, "", "  ")
	if err != nil {
		return TokenData{}, fmt.Errorf("failed to serialize token JSON: %w", err)
	}
	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return TokenData{}, fmt.Errorf("failed to write token to file: %w", err)
	}

	return tokenData, nil
}

// verifyDNSTXTRecord verifies if the TXT record for a given domain contains the specified token.
func verifyDNSTXTRecord(domain string, token string) (bool, error) {
	// Append `_alphalab-verification.` to the domain for verification
	queryDomain := fmt.Sprintf("_alphalab-verification.%s", domain)

	// Create a DNS client and query for TXT records
	c := new(dns.Client)
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(queryDomain), dns.TypeTXT)

	// Query the DNS server (use default resolver)
	resp, _, err := c.Exchange(m, "8.8.8.8:53") // Google DNS server
	if err != nil {
		return false, fmt.Errorf("failed to query DNS: %v", err)
	}

	// Parse the response for TXT records
	if resp.Rcode != dns.RcodeSuccess {
		return false, fmt.Errorf("DNS query failed with Rcode: %d", resp.Rcode)
	}

	for _, answer := range resp.Answer {
		if txt, ok := answer.(*dns.TXT); ok {
			for _, txtValue := range txt.Txt {
				if txtValue == token {
					return true, nil
				}
			}
		}
	}

	return false, nil
}
