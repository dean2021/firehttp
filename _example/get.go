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
