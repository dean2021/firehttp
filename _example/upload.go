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
