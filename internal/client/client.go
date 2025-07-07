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

// Branch represents a Grafbase branch
type Branch struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	Graph     Graph     `json:"graph"`
}

// Subgraph represents a Grafbase subgraph
type Subgraph struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"createdAt"`
	Branch    Branch    `json:"branch"`
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

// CreateBranchInput represents the input for creating a branch
type CreateBranchInput struct {
	GraphID    string `json:"graphId"`
	BranchName string `json:"branchName"`
}

// CreateSubgraphInput represents the input for creating a subgraph
type CreateSubgraphInput struct {
	BranchID     string `json:"branchId"`
	SubgraphName string `json:"subgraphName"`
	URL          string `json:"url"`
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

// CreateBranch creates a new branch
func (c *Client) CreateBranch(ctx context.Context, input CreateBranchInput) (*Branch, error) {
	query := `
		mutation CreateBranch($input: BranchCreateInput!) {
			branchCreate(input: $input) {
				... on BranchCreateSuccess {
					branch {
						id
						name
						createdAt
						graph {
							id
							slug
							account {
								id
								slug
								name
							}
						}
					}
				}
				... on GraphDoesNotExistError {
					__typename
				}
				... on BranchAlreadyExistsError {
					__typename
				}
				... on BranchNameInvalidError {
					__typename
				}
			}
		}
	`

	variables := map[string]interface{}{
		"input": input,
	}

	resp, err := c.ExecuteQuery(ctx, query, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to create branch: %w", err)
	}

	var result struct {
		BranchCreate json.RawMessage `json:"branchCreate"`
	}

	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal create branch response: %w", err)
	}

	// Try to parse as success response
	var successResp struct {
		Branch Branch `json:"branch"`
	}
	if err := json.Unmarshal(result.BranchCreate, &successResp); err == nil && successResp.Branch.ID != "" {
		return &successResp.Branch, nil
	}

	// If not a success response, it's an error
	var errorResp map[string]interface{}
	if err := json.Unmarshal(result.BranchCreate, &errorResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return nil, fmt.Errorf("branch creation failed: %v", errorResp)
}

// GetBranch retrieves a branch by graph ID and branch name
func (c *Client) GetBranch(ctx context.Context, graphID, branchName string) (*Branch, error) {
	query := `
		query GetBranch($graphId: ID!, $branchName: String!) {
			branchByGraphId(graphId: $graphId, branchName: $branchName) {
				id
				name
				createdAt
				graph {
					id
					slug
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
		"graphId":    graphID,
		"branchName": branchName,
	}

	resp, err := c.ExecuteQuery(ctx, query, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to get branch: %w", err)
	}

	var result struct {
		BranchByGraphId *Branch `json:"branchByGraphId"`
	}

	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal get branch response: %w", err)
	}

	if result.BranchByGraphId == nil {
		return nil, fmt.Errorf("branch not found")
	}

	return result.BranchByGraphId, nil
}

// DeleteBranch deletes a branch
func (c *Client) DeleteBranch(ctx context.Context, id string) error {
	query := `
		mutation DeleteBranch($input: BranchDeleteInput!) {
			branchDelete(input: $input) {
				... on BranchDeleteSuccess {
					deletedId
				}
				... on BranchDoesNotExistError {
					query
				}
			}
		}
	`

	variables := map[string]interface{}{
		"input": map[string]interface{}{
			"id": id,
		},
	}

	resp, err := c.ExecuteQuery(ctx, query, variables)
	if err != nil {
		return fmt.Errorf("failed to delete branch: %w", err)
	}

	var result struct {
		BranchDelete json.RawMessage `json:"branchDelete"`
	}

	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return fmt.Errorf("failed to unmarshal delete branch response: %w", err)
	}

	// Try to parse as success response
	var successResp struct {
		DeletedID string `json:"deletedId"`
	}
	if err := json.Unmarshal(result.BranchDelete, &successResp); err == nil && successResp.DeletedID != "" {
		return nil
	}

	// If not a success response, it's an error
	var errorResp map[string]interface{}
	if err := json.Unmarshal(result.BranchDelete, &errorResp); err != nil {
		return fmt.Errorf("failed to parse delete branch response: %w", err)
	}

	return fmt.Errorf("branch deletion failed: %v", errorResp)
}

// CreateSubgraph creates a new subgraph
func (c *Client) CreateSubgraph(ctx context.Context, input CreateSubgraphInput) (*Subgraph, error) {
	query := `
		mutation CreateSubgraph($input: SubgraphCreateInput!) {
			subgraphCreate(input: $input) {
				... on SubgraphCreateSuccess {
					subgraph {
						id
						name
						url
						createdAt
						branch {
							id
							name
							graph {
								id
								slug
								account {
									id
									slug
									name
								}
							}
						}
					}
				}
				... on BranchDoesNotExistError {
					__typename
				}
				... on SubgraphAlreadyExistsError {
					__typename
				}
				... on SubgraphNameInvalidError {
					__typename
				}
				... on SubgraphUrlInvalidError {
					__typename
				}
			}
		}
	`

	variables := map[string]interface{}{
		"input": input,
	}

	resp, err := c.ExecuteQuery(ctx, query, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to create subgraph: %w", err)
	}

	var result struct {
		SubgraphCreate json.RawMessage `json:"subgraphCreate"`
	}

	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal create subgraph response: %w", err)
	}

	// Try to parse as success response
	var successResp struct {
		Subgraph Subgraph `json:"subgraph"`
	}
	if err := json.Unmarshal(result.SubgraphCreate, &successResp); err == nil && successResp.Subgraph.ID != "" {
		return &successResp.Subgraph, nil
	}

	// If not a success response, it's an error
	var errorResp map[string]interface{}
	if err := json.Unmarshal(result.SubgraphCreate, &errorResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return nil, fmt.Errorf("subgraph creation failed: %v", errorResp)
}

// GetSubgraph retrieves a subgraph by branch ID and subgraph name
func (c *Client) GetSubgraph(ctx context.Context, branchID, subgraphName string) (*Subgraph, error) {
	query := `
		query GetSubgraph($branchId: ID!, $subgraphName: String!) {
			subgraphByBranchId(branchId: $branchId, subgraphName: $subgraphName) {
				id
				name
				url
				createdAt
				branch {
					id
					name
					graph {
						id
						slug
						account {
							id
							slug
							name
						}
					}
				}
			}
		}
	`

	variables := map[string]interface{}{
		"branchId":     branchID,
		"subgraphName": subgraphName,
	}

	resp, err := c.ExecuteQuery(ctx, query, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to get subgraph: %w", err)
	}

	var result struct {
		SubgraphByBranchId *Subgraph `json:"subgraphByBranchId"`
	}

	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal get subgraph response: %w", err)
	}

	if result.SubgraphByBranchId == nil {
		return nil, fmt.Errorf("subgraph not found")
	}

	return result.SubgraphByBranchId, nil
}

// UpdateSubgraph updates a subgraph
func (c *Client) UpdateSubgraph(ctx context.Context, id, url string) (*Subgraph, error) {
	query := `
		mutation UpdateSubgraph($input: SubgraphUpdateInput!) {
			subgraphUpdate(input: $input) {
				... on SubgraphUpdateSuccess {
					subgraph {
						id
						name
						url
						createdAt
						branch {
							id
							name
							graph {
								id
								slug
								account {
									id
									slug
									name
								}
							}
						}
					}
				}
				... on SubgraphDoesNotExistError {
					__typename
				}
				... on SubgraphUrlInvalidError {
					__typename
				}
			}
		}
	`

	variables := map[string]interface{}{
		"input": map[string]interface{}{
			"id":  id,
			"url": url,
		},
	}

	resp, err := c.ExecuteQuery(ctx, query, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to update subgraph: %w", err)
	}

	var result struct {
		SubgraphUpdate json.RawMessage `json:"subgraphUpdate"`
	}

	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal update subgraph response: %w", err)
	}

	// Try to parse as success response
	var successResp struct {
		Subgraph Subgraph `json:"subgraph"`
	}
	if err := json.Unmarshal(result.SubgraphUpdate, &successResp); err == nil && successResp.Subgraph.ID != "" {
		return &successResp.Subgraph, nil
	}

	// If not a success response, it's an error
	var errorResp map[string]interface{}
	if err := json.Unmarshal(result.SubgraphUpdate, &errorResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return nil, fmt.Errorf("subgraph update failed: %v", errorResp)
}

// DeleteSubgraph deletes a subgraph
func (c *Client) DeleteSubgraph(ctx context.Context, id string) error {
	query := `
		mutation DeleteSubgraph($input: SubgraphDeleteInput!) {
			subgraphDelete(input: $input) {
				... on SubgraphDeleteSuccess {
					deletedId
				}
				... on SubgraphDoesNotExistError {
					query
				}
			}
		}
	`

	variables := map[string]interface{}{
		"input": map[string]interface{}{
			"id": id,
		},
	}

	resp, err := c.ExecuteQuery(ctx, query, variables)
	if err != nil {
		return fmt.Errorf("failed to delete subgraph: %w", err)
	}

	var result struct {
		SubgraphDelete json.RawMessage `json:"subgraphDelete"`
	}

	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return fmt.Errorf("failed to unmarshal delete subgraph response: %w", err)
	}

	// Try to parse as success response
	var successResp struct {
		DeletedID string `json:"deletedId"`
	}
	if err := json.Unmarshal(result.SubgraphDelete, &successResp); err == nil && successResp.DeletedID != "" {
		return nil
	}

	// If not a success response, it's an error
	var errorResp map[string]interface{}
	if err := json.Unmarshal(result.SubgraphDelete, &errorResp); err != nil {
		return fmt.Errorf("failed to parse delete subgraph response: %w", err)
	}

	return fmt.Errorf("subgraph deletion failed: %v", errorResp)
}
