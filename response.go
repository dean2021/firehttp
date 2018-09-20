package firehttp

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

type Response struct {
	RawResponse *http.Response
}

func (r *Response) StatusCode() int {
	return r.RawResponse.StatusCode
}

func (r *Response) Headers() http.Header {
	return r.RawResponse.Header
}

func (r *Response) Cookies() []*http.Cookie {
	return r.RawResponse.Cookies()
}

// 获取原始header
func (r *Response) RawHeaders() string {
	var rawHeader string
	for k, v := range r.RawResponse.Header {
		for _, value := range v {
			rawHeader += k + ": " + value + "\r\n"
		}
	}
	return strings.TrimSuffix(rawHeader, "\r\n")
}

// 获取原始的set cookie
func (r *Response) RawCookies() string {
	var rawCookie string
	for _, v := range r.RawResponse.Cookies() {
		rawCookie += fmt.Sprintf("Set-Cookie: %s\r\n", v.Raw)
	}
	return rawCookie
}

// 获取response响应内容,string格式
func (r *Response) String() string {
	body, _ := ioutil.ReadAll(r.RawResponse.Body)
	return string(body)
}

// 获取response响应内容,byte格式
func (r *Response) Byte() []byte {
	body, _ := ioutil.ReadAll(r.RawResponse.Body)
	return body
}

// 读取指定长度的body
func (r *Response) ReadN(n int64) []byte {
	body, _ := ioutil.ReadAll(io.LimitReader(r.RawResponse.Body, n))
	return body
}

// 获取原始HTTP请求报文
func (r *Response) RawHTTPRequest() string {
	return RawHTTPRequest(r.RawResponse.Request)
}

// 获取原始HTTP响应报文
func (r *Response) RawHTTPResponse() string {
	return RawHTTPResponse(r.RawResponse)
}

func (r *Response) Close() error {
	return r.RawResponse.Body.Close()
}
