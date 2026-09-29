package feishu

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const defaultBaseURL = "https://open.feishu.cn"

// Doer 抽象 HTTP 传输层，便于使用 httptest.Server 进行离线测试。
type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

// Client 是 Feishu Open API 客户端。它只负责请求构造和响应校验，不保存业务状态。
type Client struct {
	BaseURL string
	Token   string
	HTTP    Doer
}

// NewClient 创建使用 Feishu Open API 的客户端。
func NewClient(baseURL, token string, httpClient Doer) *Client {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = defaultBaseURL
	}
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{BaseURL: strings.TrimRight(baseURL, "/"), Token: token, HTTP: httpClient}
}

// APIError 表示 Feishu HTTP 或业务错误。
type APIError struct {
	HTTPStatus int
	Code       int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("feishu api error: http=%d code=%d message=%s", e.HTTPStatus, e.Code, e.Message)
}

func (c *Client) do(ctx context.Context, method, path string, query url.Values, payload any, out any) error {
	endpoint := c.BaseURL + path
	if len(query) != 0 {
		endpoint += "?" + query.Encode()
	}
	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewReader(encoded)
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var envelope struct {
		Code int             `json:"code"`
		Msg  string          `json:"msg"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return fmt.Errorf("decode feishu response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 || envelope.Code != 0 {
		return &APIError{HTTPStatus: resp.StatusCode, Code: envelope.Code, Message: envelope.Msg}
	}
	if out != nil && len(envelope.Data) > 0 && string(envelope.Data) != "null" {
		if err := json.Unmarshal(envelope.Data, out); err != nil {
			return fmt.Errorf("decode feishu data: %w", err)
		}
	}
	return nil
}

func (c *Client) get(ctx context.Context, path string, query url.Values, out any) error {
	return c.do(ctx, http.MethodGet, path, query, nil, out)
}
func (c *Client) post(ctx context.Context, path string, payload any, out any) error {
	return c.do(ctx, http.MethodPost, path, nil, payload, out)
}
func (c *Client) patch(ctx context.Context, path string, payload any, out any) error {
	return c.do(ctx, http.MethodPatch, path, nil, payload, out)
}

func pageQuery(pageSize int, pageToken string) url.Values {
	q := url.Values{}
	if pageSize > 0 {
		q.Set("page_size", strconv.Itoa(pageSize))
	}
	if pageToken != "" {
		q.Set("page_token", pageToken)
	}
	return q
}

// Bitable 方法组。
func (c *Client) CreateApp(ctx context.Context, name string) (map[string]any, error) {
	var out map[string]any
	err := c.post(ctx, "/open-apis/bitable/v1/apps", map[string]any{"name": name}, &out)
	return out, err
}
func (c *Client) CreateField(ctx context.Context, appToken, tableID, fieldName string, fieldType int, property map[string]any) (map[string]any, error) {
	var out map[string]any
	err := c.post(ctx, fmt.Sprintf("/open-apis/bitable/v1/apps/%s/tables/%s/fields", appToken, tableID), map[string]any{"field_name": fieldName, "type": fieldType, "property": property}, &out)
	return out, err
}
func (c *Client) CreateRecord(ctx context.Context, appToken, tableID string, fields map[string]any) (map[string]any, error) {
	var out map[string]any
	err := c.post(ctx, fmt.Sprintf("/open-apis/bitable/v1/apps/%s/tables/%s/records", appToken, tableID), map[string]any{"fields": fields}, &out)
	return out, err
}
func (c *Client) GetRecord(ctx context.Context, appToken, tableID, recordID string) (map[string]any, error) {
	var out map[string]any
	err := c.get(ctx, fmt.Sprintf("/open-apis/bitable/v1/apps/%s/tables/%s/records/%s", appToken, tableID, recordID), nil, &out)
	return out, err
}
func (c *Client) ListRecords(ctx context.Context, appToken, tableID string, pageSize int, pageToken string) (map[string]any, error) {
	var out map[string]any
	err := c.get(ctx, fmt.Sprintf("/open-apis/bitable/v1/apps/%s/tables/%s/records", appToken, tableID), pageQuery(pageSize, pageToken), &out)
	return out, err
}
func (c *Client) UpdateRecord(ctx context.Context, appToken, tableID, recordID string, fields map[string]any) (map[string]any, error) {
	var out map[string]any
	err := c.put(ctx, fmt.Sprintf("/open-apis/bitable/v1/apps/%s/tables/%s/records/%s", appToken, tableID, recordID), map[string]any{"fields": fields}, &out)
	return out, err
}

func (c *Client) put(ctx context.Context, path string, payload any, out any) error {
	return c.do(ctx, http.MethodPut, path, nil, payload, out)
}

// Doc 方法组。
func (c *Client) Read(ctx context.Context, documentID string) (map[string]any, error) {
	var out map[string]any
	err := c.get(ctx, fmt.Sprintf("/open-apis/docx/v1/documents/%s/raw_content", documentID), nil, &out)
	return out, err
}
func (c *Client) Write(ctx context.Context, documentID string, blocks any) (map[string]any, error) {
	var out map[string]any
	err := c.post(ctx, fmt.Sprintf("/open-apis/docx/v1/documents/%s/blocks", documentID), map[string]any{"children": blocks}, &out)
	return out, err
}
func (c *Client) Append(ctx context.Context, documentID string, blocks any) (map[string]any, error) {
	return c.Write(ctx, documentID, blocks)
}
func (c *Client) CreateTable(ctx context.Context, documentID string, rowSize, columnSize int) (map[string]any, error) {
	var out map[string]any
	err := c.post(ctx, fmt.Sprintf("/open-apis/docx/v1/documents/%s/tables", documentID), map[string]any{"row_size": rowSize, "column_size": columnSize}, &out)
	return out, err
}
func (c *Client) WriteTableCells(ctx context.Context, documentID, tableID string, values [][]string) (map[string]any, error) {
	var out map[string]any
	err := c.put(ctx, fmt.Sprintf("/open-apis/docx/v1/documents/%s/tables/%s/cells", documentID, tableID), map[string]any{"values": values}, &out)
	return out, err
}

// Drive 方法组。
func (c *Client) List(ctx context.Context, folderToken string, pageSize int, pageToken string) (map[string]any, error) {
	q := pageQuery(pageSize, pageToken)
	if folderToken != "" {
		q.Set("folder_token", folderToken)
	}
	var out map[string]any
	err := c.get(ctx, "/open-apis/drive/v1/files", q, &out)
	return out, err
}
func (c *Client) Info(ctx context.Context, fileToken string) (map[string]any, error) {
	var out map[string]any
	err := c.get(ctx, fmt.Sprintf("/open-apis/drive/v1/metas/batch_query"), url.Values{"request_docs": []string{fileToken}}, &out)
	return out, err
}
func (c *Client) CreateFolder(ctx context.Context, name, parentToken string) (map[string]any, error) {
	var out map[string]any
	err := c.post(ctx, "/open-apis/drive/v1/files/create_folder", map[string]any{"name": name, "folder_token": parentToken}, &out)
	return out, err
}

// Wiki 方法组。
func (c *Client) Spaces(ctx context.Context, pageSize int, pageToken string) (map[string]any, error) {
	var out map[string]any
	err := c.get(ctx, "/open-apis/wiki/v2/spaces", pageQuery(pageSize, pageToken), &out)
	return out, err
}
func (c *Client) Nodes(ctx context.Context, spaceID, parentToken string, pageSize int, pageToken string) (map[string]any, error) {
	q := pageQuery(pageSize, pageToken)
	if parentToken != "" {
		q.Set("parent_node_token", parentToken)
	}
	var out map[string]any
	err := c.get(ctx, fmt.Sprintf("/open-apis/wiki/v2/spaces/%s/nodes", spaceID), q, &out)
	return out, err
}
func (c *Client) Get(ctx context.Context, token string) (map[string]any, error) {
	var out map[string]any
	err := c.get(ctx, fmt.Sprintf("/open-apis/wiki/v2/nodes/%s", token), nil, &out)
	return out, err
}
func (c *Client) Create(ctx context.Context, spaceID, title, parentToken string) (map[string]any, error) {
	payload := map[string]any{"space_id": spaceID, "title": title}
	if parentToken != "" {
		payload["parent_node_token"] = parentToken
	}
	var out map[string]any
	err := c.post(ctx, "/open-apis/wiki/v2/spaces", payload, &out)
	return out, err
}
