package kit

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"golang.org/x/net/proxy"
)

var httpClient *http.Client

var header map[string]string

func HttpSetHeader(header1 map[string]string) {
	header = header1
}

func initClient() {
	if httpClient != nil {
		return
	}
	httpClient = &http.Client{}

	proxyStr := os.Getenv("http_proxy")
	if proxyStr != "" {
		dialSocksProxy, err := proxy.SOCKS5("tcp", proxyStr, nil, proxy.Direct)
		if err != nil {
			fmt.Println("Error connecting to proxy:", err)
		}
		tr := &http.Transport{Dial: dialSocksProxy.Dial}
		httpClient.Transport = tr
	}
}

func HttpGet(u string) (*http.Response, error) {
	initClient()
	defer func() {
		header = nil
	}()
	request, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range header {
		request.Header.Set(k, v)
	}
	resp, err := httpClient.Do(request)
	if resp.StatusCode != http.StatusOK {
		bs, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%s", bs)
	}
	return resp, err
}

func HttpGetBytes(u string) ([]byte, error) {
	resp, err := HttpGet(u)
	if err != nil {
		return nil, err
	}
	return io.ReadAll(resp.Body)
}

func HttpJson(u string, target any) error {
	bs, err := HttpGetBytes(u)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(bs, target); err != nil {
		return err
	}
	return nil
}
