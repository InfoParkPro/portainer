package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	portainer "github.com/portainer/portainer/api"
)

const apiKeyHeader = "X-API-KEY"

type Client struct {
	baseURL    string
	apiToken   string
	httpClient *http.Client
}

type StatusResponse struct {
	Version string `json:"Version"`
}

type stackFileResponse struct {
	StackFileContent string `json:"StackFileContent"`
}

type UpdateStackPayload struct {
	StackFileContent       string           `json:"StackFileContent"`
	Env                    []portainer.Pair `json:"Env"`
	Prune                  bool             `json:"Prune"`
	RepullImageAndRedeploy bool             `json:"RepullImageAndRedeploy"`
}

func New(baseURL string, apiToken string, tlsSkipVerify bool) (*Client, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return nil, fmt.Errorf("remote Portainer URL is required")
	}

	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid remote Portainer URL: %w", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("invalid remote Portainer URL: missing scheme or host")
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	if tlsSkipVerify {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec
	}

	return &Client{
		baseURL:  baseURL,
		apiToken: apiToken,
		httpClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		},
	}, nil
}

func (c *Client) Status(ctx context.Context) (*StatusResponse, error) {
	var status StatusResponse
	if err := c.doJSON(ctx, http.MethodGet, "/api/status", nil, &status); err != nil {
		return nil, err
	}

	return &status, nil
}

func (c *Client) Stacks(ctx context.Context) ([]portainer.Stack, error) {
	var stacks []portainer.Stack
	if err := c.doJSON(ctx, http.MethodGet, "/api/stacks", nil, &stacks); err != nil {
		return nil, err
	}

	return stacks, nil
}

func (c *Client) Stack(ctx context.Context, id portainer.StackID) (*portainer.Stack, error) {
	var stack portainer.Stack
	if err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("/api/stacks/%d", id), nil, &stack); err != nil {
		return nil, err
	}

	return &stack, nil
}

func (c *Client) StackFile(ctx context.Context, id portainer.StackID) (string, error) {
	var file stackFileResponse
	if err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("/api/stacks/%d/file", id), nil, &file); err != nil {
		return "", err
	}

	return file.StackFileContent, nil
}

func (c *Client) UpdateStack(ctx context.Context, stack *portainer.Stack, payload UpdateStackPayload) (*portainer.Stack, error) {
	if stack == nil {
		return nil, fmt.Errorf("remote stack is required")
	}
	if stack.EndpointID == 0 {
		return nil, fmt.Errorf("remote stack endpoint ID is required")
	}

	path := fmt.Sprintf("/api/stacks/%d?endpointId=%d", stack.ID, stack.EndpointID)

	var updated portainer.Stack
	if err := c.doJSON(ctx, http.MethodPut, path, payload, &updated); err != nil {
		return nil, err
	}

	return &updated, nil
}

func (c *Client) doJSON(ctx context.Context, method string, path string, payload any, target any) error {
	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to encode remote Portainer request: %w", err)
		}
		body = bytes.NewReader(encoded)
	}

	request, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return fmt.Errorf("failed to create remote Portainer request: %w", err)
	}

	request.Header.Set(apiKeyHeader, c.apiToken)
	if payload != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("remote Portainer request failed: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return remoteError(response)
	}

	if target == nil {
		return nil
	}

	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		return fmt.Errorf("failed to decode remote Portainer response: %w", err)
	}

	return nil
}

func remoteError(response *http.Response) error {
	body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
	message := strings.TrimSpace(string(body))
	if message == "" {
		message = http.StatusText(response.StatusCode)
	}

	return fmt.Errorf("remote Portainer returned %d: %s", response.StatusCode, message)
}
