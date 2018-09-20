# firehttp

   
    一个专门用于开发安全工具的HTTP类库.
    

### 特点

1. header、params、body 均支持string和map定义, 太方便了.
2. 可以获取到HTTP请求/响应的原始HTTP报文.
3. 使用起来,真心简单.
4. 尽量用优雅的编程方式,让go代码读起来不那么丑陋.
5. 自己摸索吧.

### Example

发送一个普通的GET请求

```go

package main

import (
	"fmt"
	"github.com/dean2021/firehttp"
	"log"
)

func main() {

	f := firehttp.New(nil)
	resp, err := f.Get("http://www.jd.com", nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(resp.String())
}


```

发送一个复杂的GET请求

```go
package main

import (
	"github.com/dean2021/firehttp"
	"time"
	"log"
	"fmt"
)

func main() {

	f := firehttp.New(&firehttp.HTTPOptions{

		// HTTP代理地址,推荐用burpsuite进行调试
		Proxy: "http://127.0.0.1:8080",

		// DNS缓存有效时间
		DNSCacheExpire: time.Minute * 5,

		// 空闲链接
		MaxIdleConn: 1,

		// HTTP握手超时设置
		TLSHandshakeTimeout: time.Second * 5,

		// 拨号建立连接完成超时时间
		DialTimeout: time.Second * 5,

		// KeepAlive 超时时间
		DialKeepAlive: time.Second * 5,
	})

	resp, err := f.Get("https://www.jd.com/index.php", &firehttp.ReqOptions{

		// GET参数,支持map[string]string 和 string
		Params: "id=1",

		// header参数,支持map[string]string 和 string
		Header: "Cookie: xxxx\r\n",

		// 请求总超时时间
		Timeout: time.Second * 10,

		// 表示这个请求是一个ajax请求,自动追加Content-Type
		IsAjax: true,

		// 禁止跳转
		DisableRedirect: true,

		// 使用cookie会话
		UseCookieJar:true,

		// 跳过https证书验证
		InsecureSkipVerify:true,

		// 禁止gzip请求压缩
		DisableCompression: true,
	})

	if err != nil {
		log.Fatal(err)
	}


	// 输出HTTP请求报文
	fmt.Println(resp.RawHTTPRequest())

	// 输出HTTP响应报文
	fmt.Println(resp.RawHTTPResponse())

	// 输出响应状态码
	fmt.Println(resp.StatusCode())

	// 输出响应内容
	fmt.Println(resp.String())
}


```

发送一个POST请求

```go
package main

import (
	"fmt"
	"github.com/dean2021/firehttp"
	"log"
)

func main() {

	f := firehttp.New(nil)
	resp, err := f.Post("https://www.jd.com", &firehttp.ReqOptions{
		Body: "xxxx",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(resp.RawHTTPRequest())
}
```
上传一个文件

```go
package main

import (
	"fmt"
	"github.com/dean2021/firehttp"
	"log"
)

func main() {

	f := firehttp.New(nil)
	resp, err := f.Post("http://www.baidu.com/upload.php", &firehttp.ReqOptions{
		Files: []firehttp.FileUpload{
			{
				FieldName: "passwd",
				FileName:  "/etc/passwd",
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(resp.StatusCode())
}
```


代码简单易懂,建议有空通读一遍代码，相信会有新的发现。



## 使用文档


### HTTP OPTION 设置参数
```go

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
}
```

### Request Option 设置参数

```go

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

```
文档待完善...

## thanks

大量参考这个两个类库的代码, 特此感谢 levigross 和 Gay4 同学.

1. grequests / https://github.com/levigross/grequests
2. zhttp / https://github.com/Greyh4t/zhttp