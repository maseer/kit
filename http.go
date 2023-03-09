package kit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"golang.org/x/net/proxy"
)

type NetState int

const (
	NetStateUnknow NetState = 0
	NetStateOK     NetState = 1
	NetStateFaild  NetState = 2
)

type ConfigPing struct {
	Timeout time.Duration
	URLPing string
}

type HttpClient struct {
	Headers     map[string]string
	Cookies     map[string]string
	ProxySocks5 string //like "127.0.0.1:7891",or set env "socks5_proxy"
	RetryTimes  int
	*ConfigPing

	client       *http.Client
	netState     NetState
	lockCheckNet sync.Mutex
}

func DefaultClient() *HttpClient {
	client := &HttpClient{
		client: &http.Client{
			Timeout: time.Second * 2,
		},
		Headers:      map[string]string{},
		Cookies:      map[string]string{},
		lockCheckNet: sync.Mutex{},
	}
	if v := os.Getenv("socks_proxy"); v != "" {
		client.SetSocks5Proxy(v)
	}
	return client
}

func (c *HttpClient) WaitPing() error {
	c.lockCheckNet.Lock()
	defer c.lockCheckNet.Unlock()
	if c.netState == NetStateFaild {
		return fmt.Errorf("network error[0]")
	} else if c.netState == NetStateOK {
		return nil
	} else if c.netState == NetStateUnknow && c.ConfigPing == nil {
		return nil
	}
	req, err := http.NewRequest(http.MethodGet, c.ConfigPing.URLPing, nil)
	if err != nil {
		return nil
	}
	timeStart := time.Now()
	for {
		resp, err := c.client.Do(req)
		if err != nil {
			if time.Now().After(timeStart.Add(c.ConfigPing.Timeout)) {
				c.netState = NetStateFaild
				return fmt.Errorf("network error[1]")
			}
			continue
		}
		if resp.StatusCode == http.StatusOK {
			c.netState = NetStateOK
			return nil
		}
		<-time.After(time.Second * 10)
	}
}
func (c *HttpClient) SetSocks5Proxy(socks5 string) {
	dialSocksProxy, err := proxy.SOCKS5("tcp", socks5, nil, proxy.Direct)
	if err != nil {
		fmt.Println("Error connecting to proxy:", err)
	}
	tr := &http.Transport{Dial: dialSocksProxy.Dial}
	c.client.Transport = tr
}

func (c *HttpClient) do(req *http.Request, retryTimes int) (*http.Response, error) {
	if err := c.WaitPing(); err != nil {
		return nil, err
	}
	for k, v := range c.Headers {
		req.Header.Set(k, v)
	}
	for k, v := range c.Cookies {
		req.AddCookie(&http.Cookie{
			Name:  k,
			Value: v,
		})
	}
	resp, err := c.client.Do(req)
	if err != nil {
		if retryTimes > 0 {
			return c.do(req, retryTimes-1)
		}
		return nil, err
	}
	return resp, err
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
	resp, err := client.do(req, client.RetryTimes)
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
