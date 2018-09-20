package firehttp

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

// 发送GET请求
func (f *FireHttp) Get(rawUrl string, options *ReqOptions) (*Response, error) {
	return f.DoRequest("GET", rawUrl, options)
}

// 发送POST请求
func (f *FireHttp) Post(rawUrl string, options *ReqOptions) (*Response, error) {
	return f.DoRequest("POST", rawUrl, options)
}

// 发送PUT请求
func (f *FireHttp) Put(rawUrl string, options *ReqOptions) (*Response, error) {
	return f.DoRequest("PUT", rawUrl, options)
}

// 发送Del请求
func (f *FireHttp) Del(rawUrl string, options *ReqOptions) (*Response, error) {
	return f.DoRequest("DELETE", rawUrl, options)
}


// 发送Haed请求
func (f *FireHttp) Head(rawUrl string, options *ReqOptions) (*Response, error) {
	return f.DoRequest("HEAD", rawUrl, options)
}

// 发送Patch请求
func (f *FireHttp) Patch(rawUrl string, options *ReqOptions) (*Response, error) {
	return f.DoRequest("PATCH", rawUrl, options)
}

// 发送Options请求
func (f *FireHttp) Options(rawUrl string, options *ReqOptions) (*Response, error) {
	return f.DoRequest("OPTIONS", rawUrl, options)
}

// 获取原始的http请求报文
func RawHTTPRequest(req *http.Request) string {
	rawRequest := req.Method + " " + req.URL.RequestURI() + " " + req.Proto + "\r\n"
	host := req.Host
	if host == "" {
		host = req.URL.Host
	}
	rawRequest += "Host: " + host + "\r\n"
	for key, val := range req.Header {
		rawRequest += key + ": " + val[0] + "\r\n"
	}
	rawRequest += "\r\n"
	if req.GetBody != nil {
		b, err := req.GetBody()
		if err == nil {
			buf, _ := ioutil.ReadAll(b)
			rawRequest += string(buf)
		}
	}
	return rawRequest
}

// 获取原始的http请求报文
func RawHTTPResponse(resp *http.Response) string {
	httpMsg := fmt.Sprintf("%s %s \r\n", resp.Proto, resp.Status)
	for key, val := range resp.Header {
		httpMsg += fmt.Sprintf("%s:%s \r\n", key, val[0])
	}
	httpMsg += "\r\n"
	buf, _ := ioutil.ReadAll(resp.Body)
	httpMsg += string(buf)
	return httpMsg
}
