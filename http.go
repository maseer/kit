package kit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"golang.org/x/net/proxy"
)

var httpClient *http.Client
var header map[string]string
var cookies map[string]string
var proxyStr = ""

func HttpSetHeaders(header1 map[string]string) {
	header = header1
}

func HttpSetCookies(cookies map[string]string) {
	if cookies == nil {
		cookies = map[string]string{}
	}
	for k, v := range cookies {
		cookies[k] = v
	}
}

func HttpSetCookie(k, v string) {
	if cookies == nil {
		cookies = map[string]string{}
	}
	cookies[k] = v
}

// HttpSetProxy 127.0.0.1:7891 => socks5://127.0..1:7891
// or http_proxy env
func HttpSetProxy(proxy string) {
	proxyStr = proxy
}

func initClient() {
	if httpClient != nil {
		return
	}
	httpClient = &http.Client{}
	if proxyStr == "" {
		proxyStr = os.Getenv("http_proxy")
	}
	if proxyStr != "" {
		dialSocksProxy, err := proxy.SOCKS5("tcp", proxyStr, nil, proxy.Direct)
		if err != nil {
			fmt.Println("Error connecting to proxy:", err)
		}
		tr := &http.Transport{Dial: dialSocksProxy.Dial}
		httpClient.Transport = tr
	}
}

func setCookieAndHeader(req *http.Request) {
	defer func() {
		cookies = map[string]string{}
		header = map[string]string{}
	}()

	for k, v := range header {
		req.Header.Set(k, v)
	}
	for k, v := range cookies {
		req.AddCookie(&http.Cookie{
			Name:  k,
			Value: v,
		})
	}
}

func Http(u string, method string, jsonObject any) (*http.Response, error) {
	initClient()
	var dataReader io.Reader
	if jsonObject != nil {
		if header == nil {
			header = map[string]string{}
		}
		header["Content-Type"] = "application/json"
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
	setCookieAndHeader(req)
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	//debug
	// bs3, _ := io.ReadAll(resp.Body)
	// fmt.Printf("%s\n", bs3)

	if resp.StatusCode != http.StatusOK {
		bs, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("StatusCode:'%d'\nbody:'%s'", resp.StatusCode, bs)
	}
	return resp, err
}

func HttpPostBytes(u string, jsonObject any) ([]byte, error) {
	resp, err := Http(u, http.MethodPost, jsonObject)
	if err != nil {
		return nil, err
	}
	return io.ReadAll(resp.Body)
}

func HttpPostTarget(u string, jsonObject any, target any) error {
	resp, err := Http(u, http.MethodPost, jsonObject)
	if err != nil {
		return err
	}
	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(bs, target)
}

func HttpGetBytes(u string) ([]byte, error) {
	resp, err := Http(u, http.MethodGet, nil)
	if err != nil {
		return nil, err
	}
	return io.ReadAll(resp.Body)
}

func HttpGetTarget(u string, target any) error {
	bs, err := HttpGetBytes(u)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(bs, target); err != nil {
		return err
	}
	return nil
}
