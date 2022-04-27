package main

import (
	"spider/spider"
)

func main() {
	s := spider.NewSpider(5)
	s.AddQueue("https://github.com/hello-evawang?page=1&tab=following")
	s.Run()
}
