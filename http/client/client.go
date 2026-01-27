package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Response HTTP 响应
type Response struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}

// Client HTTP 客户端
type Client struct {
	*Options
}

// NewClient 创建新的客户端实例
func NewClient(opts ...Option) *Client {
	return &Client{
		Options: NewOptions(opts...),
	}
}

// Get 发送 GET 请求
func (c *Client) Get(ctx context.Context, path string) (*Response, error) {
	return c.Do(ctx, http.MethodGet, path, nil)
}

// Post 发送 POST 请求
func (c *Client) Post(ctx context.Context, path string, body []byte) (*Response, error) {
	return c.Do(ctx, http.MethodPost, path, body)
}

// Put 发送 PUT 请求
func (c *Client) Put(ctx context.Context, path string, body []byte) (*Response, error) {
	return c.Do(ctx, http.MethodPut, path, body)
}

// Delete 发送 DELETE 请求
func (c *Client) Delete(ctx context.Context, path string) (*Response, error) {
	return c.Do(ctx, http.MethodDelete, path, nil)
}


// Do 发送通用 HTTP 请求
func (c *Client) Do(ctx context.Context, method, path string, body []byte) (*Response, error) {
	// 构建完整 URL
	url := c.BaseURL + path

	// 设置超时
	timeout := c.DefaultTimeout
	if deadline, ok := ctx.Deadline(); ok {
		timeout = time.Until(deadline)
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// 创建请求
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置默认请求头
	for key, value := range c.Headers {
		req.Header.Set(key, value)
	}

	// 发送请求
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("request timeout")
		}
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       respBody,
	}, nil
}

// DoWithHeaders 发送带自定义请求头的 HTTP 请求
func (c *Client) DoWithHeaders(ctx context.Context, method, path string, body []byte, headers map[string]string) (*Response, error) {
	// 构建完整 URL
	url := c.BaseURL + path

	// 设置超时
	timeout := c.DefaultTimeout
	if deadline, ok := ctx.Deadline(); ok {
		timeout = time.Until(deadline)
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// 创建请求
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置默认请求头
	for key, value := range c.Headers {
		req.Header.Set(key, value)
	}

	// 设置自定义请求头（覆盖默认）
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// 发送请求
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("request timeout")
		}
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       respBody,
	}, nil
}
