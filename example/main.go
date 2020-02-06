package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"spider"
)

func main() {
	sp := &spider.Spider{
		URL:                spider.URL{"https://www.zhihu.com"},
		SpiderDepth:        3,
		Goroutines:         20,
		AllowedDomain:      spider.Domain{"zhihu.com"},
		AllowedContentType: spider.ContentType{spider.HtmlType, spider.ImageAll},
		PreAction: func(req *http.Request) bool {
			if req.URL.Scheme == "http" {
				fmt.Println("block", req.URL.String())
				return true
			}
			fmt.Println("Visit:", req.URL.String())
			return false
		},
		PostAction: func(req *http.Request, resp *http.Response, body *bytes.Buffer) bool {
			fmt.Println("Content-Type", resp.Header.Get("Content-Type"), "detective:", http.DetectContentType(body.Bytes()))
			return false
		},
	}
	log.Fatalln(sp.Run())
}
