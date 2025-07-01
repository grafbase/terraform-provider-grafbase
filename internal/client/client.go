package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	DefaultAPIURL = "https://api.grafbase.com/graphql"
)

// Client represents a Grafbase API client
type Client struct {
	httpClient *http.Client
	apiURL     string
	apiKey     string
}

// NewClient creates a new Grafbase API client
func NewClient(apiKey string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiURL: DefaultAPIURL,
		apiKey: apiKey,
	}
}

// GraphQLRequest represents a GraphQL request
type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// GraphQLResponse represents a GraphQL response
type GraphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []GraphQLError  `json:"errors,omitempty"`
}

// GraphQLError represents a GraphQL error
type GraphQLError struct {
	Message    string                 `json:"message"`
	Path       []interface{}          `json:"path,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// Error implements the error interface
func (e GraphQLError) Error() string {
	return e.Message
}

// ExecuteQuery executes a GraphQL query
func (c *Client) ExecuteQuery(ctx context.Context, query string, variables map[string]interface{}) (*GraphQLResponse, error) {
	request := GraphQLRequest{
		Query:     query,
		Variables: variables,
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.apiURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var graphqlResp GraphQLResponse
	if err := json.Unmarshal(body, &graphqlResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(graphqlResp.Errors) > 0 {
		return &graphqlResp, fmt.Errorf("GraphQL errors: %v", graphqlResp.Errors)
	}

	return &graphqlResp, nil
}

// Graph represents a Grafbase graph
type Graph struct {
	ID        string    `json:"id"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"createdAt"`
	Account   Account   `json:"account"`
}

// Account represents a Grafbase account
type Account struct {
	ID   string `json:"id"`
	Slug string `json:"slug"`
	Name string `json:"name"`
}

// CreateGraphInput represents the input for creating a graph
type CreateGraphInput struct {
	AccountID string `json:"accountId"`
	GraphSlug string `json:"graphSlug"`
}

// CreateGraphResponse represents the successful response from graph creation
type CreateGraphResponse struct {
	GraphCreateSuccess struct {
		Graph Graph `json:"graph"`
	} `json:"GraphCreateSuccess"`
}

// GetAccountInput represents input for getting account by slug
type GetAccountInput struct {
	Slug string `json:"slug"`
}

// GetAccountBySlug retrieves an account by slug
func (c *Client) GetAccountBySlug(ctx context.Context, slug string) (*Account, error) {
	query := `
		query GetAccount($slug: String!) {
			accountBySlug(slug: $slug) {
				id
				slug
				name
			}
		}
	`

	variables := map[string]interface{}{
		"slug": slug,
	}

	resp, err := c.ExecuteQuery(ctx, query, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	var result struct {
		AccountBySlug *Account `json:"accountBySlug"`
	}

	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal account response: %w", err)
	}

	if result.AccountBySlug == nil {
		return nil, fmt.Errorf("account not found")
	}

	return result.AccountBySlug, nil
}

// DeleteGraphInput represents the input for deleting a graph
type DeleteGraphInput struct {
	ID string `json:"id"`
}

// CreateGraph creates a new graph
func (c *Client) CreateGraph(ctx context.Context, input CreateGraphInput) (*Graph, error) {
	query := `
		mutation CreateGraph($input: GraphCreateInput!) {
			graphCreate(input: $input) {
				... on GraphCreateSuccess {
					graph {
						id
						slug
						createdAt
						account {
							id
							slug
							name
						}
					}
				}
				... on AccountDoesNotExistError {
					__typename
				}
				... on DisabledAccountError {
					__typename
				}
				... on SlugAlreadyExistsError {
					__typename
				}
				... on SlugInvalidError {
					__typename
				}
				... on SlugTooLongError {
			    	__typename
			        maxLength
				}
			}
		}
	`

	variables := map[string]interface{}{
		"input": input,
	}

	resp, err := c.ExecuteQuery(ctx, query, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to create graph: %w", err)
	}

	var result struct {
		GraphCreate json.RawMessage `json:"graphCreate"`
	}

	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal create response: %w", err)
	}

	// Try to parse as success response
	var successResp struct {
		Graph Graph `json:"graph"`
	}
	if err := json.Unmarshal(result.GraphCreate, &successResp); err == nil && successResp.Graph.ID != "" {
		return &successResp.Graph, nil
	}

	// If not a success response, it's an error
	var errorResp map[string]interface{}
	if err := json.Unmarshal(result.GraphCreate, &errorResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return nil, fmt.Errorf("graph creation failed: %v", errorResp)
}

// GetGraph retrieves a graph by account slug and graph slug
func (c *Client) GetGraph(ctx context.Context, accountSlug, graphSlug string) (*Graph, error) {
	query := `
		query GetGraph($accountSlug: String!, $graphSlug: String!) {
			graphByAccountSlug(accountSlug: $accountSlug, graphSlug: $graphSlug) {
				id
				slug
				createdAt
				account {
					id
					slug
					name
				}
			}
		}
	`

	variables := map[string]interface{}{
		"accountSlug": accountSlug,
		"graphSlug":   graphSlug,
	}

	resp, err := c.ExecuteQuery(ctx, query, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to get graph: %w", err)
	}

	var result struct {
		GraphByAccountSlug *Graph `json:"graphByAccountSlug"`
	}

	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal get response: %w", err)
	}

	if result.GraphByAccountSlug == nil {
		return nil, fmt.Errorf("graph not found")
	}

	return result.GraphByAccountSlug, nil
}

// GetGraphByID retrieves a graph by ID using the node query
func (c *Client) GetGraphByID(ctx context.Context, id string) (*Graph, error) {
	query := `
		query GetGraphByID($id: ID!) {
			node(id: $id) {
				... on Graph {
					id
					slug
					createdAt
					account {
						id
						slug
						name
					}
				}
			}
		}
	`

	variables := map[string]interface{}{
		"id": id,
	}

	resp, err := c.ExecuteQuery(ctx, query, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to get graph by ID: %w", err)
	}

	var result struct {
		Node *Graph `json:"node"`
	}

	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal get by ID response: %w", err)
	}

	if result.Node == nil {
		return nil, fmt.Errorf("graph not found")
	}

	return result.Node, nil
}

// DeleteGraph deletes a graph
func (c *Client) DeleteGraph(ctx context.Context, id string) error {
	query := `
		mutation DeleteGraph($input: GraphDeleteInput!) {
			graphDelete(input: $input) {
				... on GraphDeleteSuccess {
					deletedId
				}
				... on GraphDoesNotExistError {
					query
				}
			}
		}
	`

	variables := map[string]interface{}{
		"input": DeleteGraphInput{
			ID: id,
		},
	}

	resp, err := c.ExecuteQuery(ctx, query, variables)
	if err != nil {
		return fmt.Errorf("failed to delete graph: %w", err)
	}

	var result struct {
		GraphDelete json.RawMessage `json:"graphDelete"`
	}

	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return fmt.Errorf("failed to unmarshal delete response: %w", err)
	}

	// Try to parse as success response
	var successResp struct {
		DeletedID string `json:"deletedId"`
	}
	if err := json.Unmarshal(result.GraphDelete, &successResp); err == nil && successResp.DeletedID != "" {
		return nil
	}

	// If not a success response, it's an error
	var errorResp map[string]interface{}
	if err := json.Unmarshal(result.GraphDelete, &errorResp); err != nil {
		return fmt.Errorf("failed to parse delete response: %w", err)
	}

	return fmt.Errorf("graph deletion failed: %v", errorResp)
}
