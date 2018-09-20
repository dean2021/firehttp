package firehttp

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/Greyh4t/dnscache"
	"github.com/google/go-querystring/query"
	"github.com/pkg/errors"
	"golang.org/x/net/publicsuffix"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/textproto"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const VERSION = "0.1"

type FireHttp struct {
	Setting *HTTPOptions
}

// HTTP Request 参数
type HTTPOptions struct {
	// 代理地址
	Proxy string

	// DNS缓存有效时间
	DNSCacheExpire time.Duration

	// 空闲连接数
	MaxIdleConn int

	// TLS握手超时时间
	TLSHandshakeTimeout time.Duration

	// 拨号建立连接完成超时时间
	DialTimeout time.Duration

	// KeepAlive 超时时间
	DialKeepAlive time.Duration

	// 预先设置的Header
	ParentHeader map[string]string
}

var resolver *dnscache.Resolver

// 发送请求
func (f *FireHttp) DoRequest(method string, rawUrl string, ro *ReqOptions) (*Response, error) {

	if ro == nil {
		ro = &ReqOptions{}
	}

	// 默认请求超时时间
	if ro.Timeout == 0 {
		ro.Timeout = 60 * time.Second
	}

	var httpClient *http.Client
	rawURL, err := buildQuery(rawUrl, ro.Params)
	if err != nil {
		return nil, err
	}

	req, err := f.BuildRequest(method, rawURL, ro)
	if err != nil {
		return nil, err
	}

	if ro.HTTPClient != nil {
		httpClient = ro.HTTPClient
	} else {
		httpClient = f.buildHTTPClient(ro)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	return &Response{RawResponse: resp}, nil
}

// 构建HTTP Client
func (f *FireHttp) buildHTTPClient(ro *ReqOptions) *http.Client {

	var cookieJar http.CookieJar
	if ro.UseCookieJar {
		if ro.CookieJar != nil {
			cookieJar = ro.CookieJar
		} else {
			cookieJar, _ = cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
		}
	}

	return &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if ro.DisableRedirect {
				return http.ErrUseLastResponse
			} else {
				return nil
			}
		},
		Jar:       cookieJar,
		Transport: f.buildHTTPTransport(ro),
		Timeout:   ro.Timeout,
	}
}

// 构建Transport
func (f *FireHttp) buildHTTPTransport(ro *ReqOptions) *http.Transport {
	transport := &http.Transport{
		Proxy: func(request *http.Request) (*url.URL, error) {
			if f.Setting.Proxy != "" {
				return url.Parse(f.Setting.Proxy)
			}
			return nil, nil
		},
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: ro.InsecureSkipVerify},
		DisableCompression:    ro.DisableCompression,
		MaxIdleConns:          f.Setting.MaxIdleConn,
		IdleConnTimeout:       f.Setting.DialKeepAlive,
		TLSHandshakeTimeout:   f.Setting.DialKeepAlive,
		ExpectContinueTimeout: f.Setting.DialKeepAlive,
		Dial: func(network, addr string) (net.Conn, error) {
			deadline := time.Now().Add(f.Setting.DialKeepAlive)
			if resolver != nil {
				host, port, err := net.SplitHostPort(addr)
				if err == nil {
					ip, err := resolver.FetchOneString(host)
					if err != nil {
						return nil, err
					}
					addr = net.JoinHostPort(ip, port)
				}
			}
			c, err := net.DialTimeout(network, addr, f.Setting.DialKeepAlive)
			if err != nil {
				return nil, err
			}
			c.SetDeadline(deadline)
			c.SetReadDeadline(deadline)
			c.SetWriteDeadline(deadline)
			return c, nil
		},
	}

	EnsureTransporterFinalized(transport)

	return transport
}

// 摘抄grequests,解决资源泄露问题
// EnsureTransporterFinalized will ensure that when the HTTP client is GCed
// the runtime will close the idle connections (so that they won't leak)
// this function was adopted from Hashicorp's go-cleanhttp package
func EnsureTransporterFinalized(httpTransport *http.Transport) {
	runtime.SetFinalizer(&httpTransport, func(transportInt **http.Transport) {
		(*transportInt).CloseIdleConnections()
	})
}

// 构建请求
// 支持body格式string或[]byte
func (f *FireHttp) BuildRequest(method string, rawURL string, ro *ReqOptions) (*http.Request, error) {

	req, err := http.NewRequest(method, rawURL, nil)
	if err != nil {
		return nil, err
	}

	if ro.Body != nil {
		req, err = f.BuildBasicRequest(method, rawURL, ro)
		if err != nil {
			return nil, err
		}
	}

	if ro.Files != nil {
		req, err = BuildFileUploadRequest(method, rawURL, ro)
		if err != nil {
			return nil, err
		}
	}

	err = f.SetHeaders(req, ro)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// 构建基础请求
func (f *FireHttp) BuildBasicRequest(method string, rawURL string, ro *ReqOptions) (*http.Request, error) {
	var reader io.Reader
	switch ro.Body.(type) {
	case string:
		reader = strings.NewReader(ro.Body.(string))
	case []byte:
		reader = bytes.NewReader(ro.Body.([]byte))
	case map[string]string:
		urlValues := &url.Values{}
		for key, value := range ro.Body.(map[string]string) {
			urlValues.Set(key, value)
		}
		reader = strings.NewReader(urlValues.Encode())
	default:
		return nil, errors.New("body type error, only support string and map[string]string or []byte type")
	}
	req, err := http.NewRequest(method, rawURL, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req, nil
}

// 构建文件上传请求
func BuildFileUploadRequest(method string, rawURL string, ro *ReqOptions) (*http.Request, error) {

	if method == "POST" {
		return buildPostFileUploadRequest(method, rawURL, ro)
	}

	// 目前仅支持POST上传
	return nil, errors.New("Upload method is wrong, currently only supports POST method to upload files")
}

// 构建post上传文件
// 支持上传多个文件
func buildPostFileUploadRequest(method string, rawURL string, ro *ReqOptions) (*http.Request, error) {

	requestBody := &bytes.Buffer{}
	multipartWriter := multipart.NewWriter(requestBody)

	for i, f := range ro.Files {
		fd, err := os.Open(f.FileName)
		if err != nil {
			return nil, err
		}
		ro.Files[i].FileBody = fd
	}

	for i, f := range ro.Files {

		fieldName := f.FieldName

		if fieldName == "" {
			if len(ro.Files) > 1 {
				fieldName = strings.Join([]string{"file", strconv.Itoa(i + 1)}, "")
			} else {
				fieldName = "file"
			}
		}

		var writer io.Writer
		var err error

		if f.FileMime != "" {
			if f.FileName == "" {
				f.FileName = "filename"
			}
			h := make(textproto.MIMEHeader)
			h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, escapeQuotes(fieldName), escapeQuotes(f.FileName)))
			h.Set("Content-Type", f.FileMime)
			writer, err = multipartWriter.CreatePart(h)
		} else {
			writer, err = multipartWriter.CreateFormFile(fieldName, f.FileName)
		}

		if err != nil {
			return nil, err
		}

		if _, err = io.Copy(writer, f.FileBody); err != nil && err != io.EOF {
			return nil, err
		}

		if err := f.FileBody.Close(); err != nil {
			return nil, err
		}
	}

	if err := multipartWriter.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, rawURL, requestBody)

	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", multipartWriter.FormDataContentType())

	return req, err

}

// 设置header
// 仅支持string和map[string]string类型
func (f *FireHttp) SetHeaders(req *http.Request, ro *ReqOptions) error {

	var err error

	if f.Setting.ParentHeader != nil {
		for key, val := range f.Setting.ParentHeader {
			req.Header.Set(key, val)
		}
	}

	if ro.Header != nil {
		switch ro.Header.(type) {
		case map[string]string:
			for key, val := range ro.Header.(map[string]string) {
				req.Header.Set(key, val)
			}
		case string:
			headers := strings.Split(ro.Header.(string), "\r\n")
			for _, val := range headers {
				header := strings.Split(val, ":")

				if len(header) == 2 {
					req.Header.Set(header[0], header[1])
				}
			}
		default:
			return errors.New("header type error, only support string and map[string]string type")
		}
	}

	// 基础认证
	if ro.BasicAuthUserAndPass != nil {
		if len(ro.BasicAuthUserAndPass) == 2 {
			req.SetBasicAuth(ro.BasicAuthUserAndPass[0], ro.BasicAuthUserAndPass[1])
		}
	}

	// 启用ajax请求
	if ro.IsAjax == true {
		req.Header.Set("X-Requested-With", "XMLHttpRequest")
	}

	if ro.IsJSON == true {
		req.Header.Set("Content-Type", "application/json")
	}

	if ro.IsXML == true {
		req.Header.Set("Content-Type", "application/xml")
	}

	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", "firehttp/"+VERSION)
	}

	return err
}

// 解析params,生成带Query的url
// params 支持string类型和map[string]string或struct类型
func buildQuery(rawURL string, params interface{}) (string, error) {

	if params != nil {

		parsedURL, err := url.Parse(rawURL)
		if err != nil {
			return "", err
		}

		parsedQuery, err := url.ParseQuery(parsedURL.RawQuery)
		if err != nil {
			return "", err
		}

		switch params.(type) {
		case string:
			urlQuery := params.(string)
			queryStr, err := url.ParseQuery(urlQuery)
			if err != nil {
				return "", err
			}
			for key, value := range queryStr {
				for _, v := range value {
					parsedQuery.Add(key, v)
				}
			}
		case map[string]string:
			for key, value := range params.(map[string]string) {
				parsedQuery.Set(key, value)
			}
		case struct{}:
			queryStruct, err := query.Values(params)
			if err != nil {
				return "", err
			}
			for key, value := range queryStruct {
				for _, v := range value {
					parsedQuery.Add(key, v)
				}
			}
		default:
			return "", errors.New("params type error, only support string and map[string]string or struct type")
		}
		rawURL = strings.Join([]string{strings.Replace(parsedURL.String(), "?"+parsedURL.RawQuery, "", -1), parsedQuery.Encode()}, "?")
	}
	return rawURL, nil
}

// HTTP Request 参数
type ReqOptions struct {
	// 请求get参数
	Params interface{}

	// 设置请求header
	Header interface{}

	// 整个请求（包括拨号/请求/重定向）等待的最长时间。
	Timeout time.Duration

	// 上传文件
	Files []FileUpload

	// 禁用跳转
	DisableRedirect bool

	// 请求body
	Body interface{}

	// 基础认证账号密码
	BasicAuthUserAndPass []string

	// 设置ajax header
	IsAjax bool

	// 设置JSON header
	IsJSON bool

	// 设置XML header
	IsXML bool

	// 自定义http client
	HTTPClient *http.Client

	// 使用cookie会话
	UseCookieJar bool
	CookieJar    http.CookieJar

	// 跳过证书验证
	InsecureSkipVerify bool

	// 禁用请求gzip压缩
	DisableCompression bool
}

func New(options *HTTPOptions) *FireHttp {

	if options == nil {
		options = &HTTPOptions{}
	}

	// 设置默认超时
	if options.TLSHandshakeTimeout == 0 {
		options.TLSHandshakeTimeout = 10 * time.Second
	}

	if options.DialTimeout == 0 {
		options.DialTimeout = 30 * time.Second
	}

	if options.DialKeepAlive == 0 {
		options.DialKeepAlive = 30 * time.Second
	}

	if options.DNSCacheExpire > 0 {
		resolver = dnscache.New(options.DNSCacheExpire)
	}

	return &FireHttp{
		Setting: options,
	}

}
