package kit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type HttpClient struct {
	Headers     map[string]string
	Cookies     map[string]string
	ProxySocks5 string //like "127.0.0.1:7891"
	client      *http.Client
}

func DefaultClient() *HttpClient {
	return &HttpClient{
		client:  &http.Client{},
		Headers: map[string]string{},
		Cookies: map[string]string{},
	}
}

func (client *HttpClient) do(req *http.Request) (*http.Response, error) {
	for k, v := range client.Headers {
		req.Header.Set(k, v)
	}
	for k, v := range client.Cookies {
		req.AddCookie(&http.Cookie{
			Name:  k,
			Value: v,
		})
	}
	return client.client.Do(req)
}

func (client *HttpClient) Http(u string, method string, jsonObject any) (*http.Response, error) {
	var dataReader io.Reader
	if jsonObject != nil {
		client.Headers["Content-Type"] = "application/json"
		jsonStr, err := json.Marshal(jsonObject)
		if err != nil {
			return nil, err
		}
		dataReader = bytes.NewReader(jsonStr)
	}
	req, err := http.NewRequest(method, u, dataReader)
	if err != nil {
		return nil, err
	}
	resp, err := client.do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		bs, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("StatusCode:'%d'\nbody:'%s'", resp.StatusCode, bs)
	}
	return resp, err
}

func (c *HttpClient) HttpPostBytes(u string, jsonObject any) ([]byte, error) {
	resp, err := c.Http(u, http.MethodPost, jsonObject)
	if err != nil {
		return nil, err
	}
	return io.ReadAll(resp.Body)
}

func (c *HttpClient) HttpPostTarget(u string, jsonObject any, target any) error {
	resp, err := c.Http(u, http.MethodPost, jsonObject)
	if err != nil {
		return err
	}
	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(bs, target)
}

func (c *HttpClient) HttpGetBytes(u string) ([]byte, error) {
	resp, err := c.Http(u, http.MethodGet, nil)
	if err != nil {
		return nil, err
	}
	return io.ReadAll(resp.Body)
}

func (c *HttpClient) HttpGetTarget(u string, target any) error {
	bs, err := c.HttpGetBytes(u)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(bs, target); err != nil {
		return err
	}
	return nil
}
