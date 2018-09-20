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

		// 预设好的header
		ParentHeader: map[string]string{
			"User-Agent": "test",
		},

		ParentHTTPTimeout: time.Second * 1,
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
