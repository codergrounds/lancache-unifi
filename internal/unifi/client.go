package unifi

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// Client communicates with the UniFi Network controller's static-dns API.
type Client struct {
	baseURL    string
	site       string
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a UniFi API client.
// TLS verification is disabled by default since UniFi controllers commonly
// use self-signed certificates.
func NewClient(host, site, apiKey string) *Client {
	return &Client{
		baseURL: host,
		site:    site,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		},
	}
}

func (c *Client) endpoint() string {
	return fmt.Sprintf("%s/proxy/network/v2/api/site/%s/static-dns", c.baseURL, c.site)
}

func (c *Client) newRequest(method, url string, body interface{}) (*http.Request, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshalling request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-API-KEY", c.apiKey)

	return req, nil
}

func (c *Client) doRequest(req *http.Request) ([]byte, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		log.Fatalf("[FATAL] UniFi API returned 401 Unauthorized. Please verify your UNIFI_API_KEY is correct.")
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return bodyBytes, nil
}

// ListDNSRecords returns all static DNS records from the controller.
func (c *Client) ListDNSRecords() ([]DNSRecord, error) {
	req, err := c.newRequest(http.MethodGet, c.endpoint(), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("listing DNS records: %w", err)
	}

	var result []DNSRecord
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decoding list response: %w", err)
	}

	return result, nil
}

// CreateDNSRecord creates a new static DNS record.
func (c *Client) CreateDNSRecord(record DNSRecord) error {
	// Ensure _id is not sent on create
	record.ID = ""

	req, err := c.newRequest(http.MethodPost, c.endpoint(), record)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return fmt.Errorf("creating DNS record %q: %w", record.Key, err)
	}

	return nil
}

// UpdateDNSRecord updates an existing static DNS record by ID.
func (c *Client) UpdateDNSRecord(id string, record DNSRecord) error {
	url := fmt.Sprintf("%s/%s", c.endpoint(), id)
	record.ID = id

	req, err := c.newRequest(http.MethodPut, url, record)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return fmt.Errorf("updating DNS record %q (id=%s): %w", record.Key, id, err)
	}

	return nil
}

// DeleteDNSRecord deletes a static DNS record by ID.
func (c *Client) DeleteDNSRecord(id string) error {
	url := fmt.Sprintf("%s/%s", c.endpoint(), id)

	req, err := c.newRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return fmt.Errorf("deleting DNS record (id=%s): %w", id, err)
	}

	return nil
}
