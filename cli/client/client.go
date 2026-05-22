package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

type Client struct {
	baseURL string
}

type APIResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   string          `json:"error,omitempty"`
}

func NewClient() (*Client, error) {
	port, err := getNativeHostPort()
	if err != nil {
		return nil, fmt.Errorf("failed to find native-host port: %w", err)
	}
	return &Client{
		baseURL: fmt.Sprintf("http://127.0.0.1:%d/api/v1", port),
	}, nil
}

func (c *Client) Get(path string) (*APIResponse, error) {
	return c.doRequest("GET", path, nil)
}

func (c *Client) Post(path string, body interface{}) (*APIResponse, error) {
	return c.doRequest("POST", path, body)
}

func (c *Client) Delete(path string) (*APIResponse, error) {
	return c.doRequest("DELETE", path, nil)
}

func (c *Client) Put(path string, body interface{}) (*APIResponse, error) {
	return c.doRequest("PUT", path, body)
}

func (c *Client) doRequest(method, path string, body interface{}) (*APIResponse, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(data)
	}

	url := c.baseURL + path
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}

	return &apiResp, nil
}

// getNativeHostPort 从 nativehost_port 文件读取端口
func getNativeHostPort() (int, error) {
	// 1. ~/.browser-bridge/nativehost_port
	homeDir, err := os.UserHomeDir()
	if err == nil {
		if port, err := readPortFile(filepath.Join(homeDir, ".browser-bridge", "nativehost_port")); err == nil {
			return port, nil
		}
	}

	// 2. 环境变量
	if portStr := os.Getenv("BROWSER_BRIDGE_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			return port, nil
		}
	}

	return 0, fmt.Errorf("nativehost_port file not found in ~/.browser-bridge/, is native-host running?")
}

func readPortFile(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(string(data))
}
