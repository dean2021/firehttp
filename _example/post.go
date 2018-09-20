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
