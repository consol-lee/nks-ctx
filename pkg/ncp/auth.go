package ncp

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// GenerateHMACSignature generates HMAC signature for NCP API authentication.
//
// Signature message format (per NCP docs):
//
//	{METHOD} {URL}\n{TIMESTAMP}\n{ACCESS_KEY}
//
// where URL = URI or URI?queryString if query string is present.
func GenerateHMACSignature(method, uri, queryString, timestamp, accessKey, secretKey string) string {
	url := uri
	if queryString != "" {
		url = uri + "?" + queryString
	}
	message := fmt.Sprintf("%s %s\n%s\n%s", method, url, timestamp, accessKey)
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(message))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return signature
}

// PrepareAuthHeaders prepares authentication headers for NCP API request
func (c *Client) PrepareAuthHeaders(method, uri, queryString string) (map[string]string, error) {
	timestamp := strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10)
	signature := GenerateHMACSignature(method, uri, queryString, timestamp, c.accessKey, c.secretKey)

	headers := map[string]string{
		"x-ncp-apigw-timestamp":    timestamp,
		"x-ncp-iam-access-key":     c.accessKey,
		"x-ncp-apigw-signature-v2": signature,
		"Content-Type":             "application/json",
	}

	return headers, nil
}

// ExtractURI extracts URI path from full URL
func ExtractURI(url string) string {
	// Remove protocol and domain
	if idx := strings.Index(url, "://"); idx != -1 {
		url = url[idx+3:]
	}
	if idx := strings.Index(url, "/"); idx != -1 {
		url = url[idx:]
	} else {
		url = "/"
	}
	// Remove query string for URI extraction
	if idx := strings.Index(url, "?"); idx != -1 {
		url = url[:idx]
	}
	return url
}

// ExtractQueryString extracts query string from URL
func ExtractQueryString(url string) string {
	if idx := strings.Index(url, "?"); idx != -1 {
		return url[idx+1:]
	}
	return ""
}
