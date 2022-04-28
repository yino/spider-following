package spider

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/antchfx/htmlquery"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/debug"
)

type Exec struct {
	urlQueue   []string
	goNum      int
	indexQueue int // 当前执行中的 queue
	queueChan  []chan string
	mutex      sync.Mutex
	wg         sync.WaitGroup
}

func (e *Exec) AddQueue(url string) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	if e.indexQueue >= e.goNum-1 {
		e.indexQueue = 0
	} else {
		e.indexQueue++
	}
	e.queueChan[e.indexQueue] <- url
	return nil
}

func (e *Exec) Run() {
	for i := 0; i < e.goNum; i++ {
		e.wg.Add(1)
		go func(i int) {
			c := colly.NewCollector(
				colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.163 Safari/537.36"),
				colly.Debugger(&debug.LogDebugger{}),
				colly.MaxDepth(1),
				colly.Async(true),
			)
			f, err := os.OpenFile(fmt.Sprintf("data/%d.txt", i), os.O_WRONLY|os.O_CREATE, 0666)
			if err != nil {
				panic(err)
			}
			defer f.Close()
			for {
				// 获取url地址
				url := <-e.queueChan[i]
				log.Println("goroutine no", i, "url", url)
				c.Visit(url)
				c.OnResponse(func(response *colly.Response) {
					users, nextHref := ParseHtml(string(response.Body))
					if len(nextHref) > 0 {
						e.AddQueue(nextHref)
					}
					log.Println("i", i, "nextHref", nextHref)
					jsonUser, err := json.Marshal(users)
					if err != nil {
						log.Println(fmt.Sprintf("error url %s err %v", url, err))
					} else {
						f.Write([]byte(string(jsonUser) + "\n"))
					}
				})
			}
			e.wg.Done()
		}(i)
	}
	e.wg.Wait()
}

// ParseHtml .
// @param content string
// @return 用户id []string  下一页 string
func ParseHtml(content string) ([]string, string) {
	var (
		nextHref string
		users    []string
		userHref string
	)
	doc, err := htmlquery.Parse(strings.NewReader(content))
	if err != nil {
		log.Println("compile pagination error", err)
		return []string{}, ""
	}
	list := htmlquery.Find(doc, `//div[@class="pagination"]/a`)
	for _, n := range list {
		if htmlquery.InnerText(n) == "Next" {
			nextHref = htmlquery.SelectAttr(n, "href")
			log.Println("nextHref", nextHref)
		}
	}
	userList := htmlquery.Find(htmlquery.Find(htmlquery.Find(doc, `//div[@class="Layout-main"]`)[1], `//div[@class="position-relative"]`)[0], `//a[@class="d-inline-block no-underline mb-1"]`)
	for _, n := range userList {
		userHref = htmlquery.SelectAttr(n, "href")
		users = append(users, string([]rune(userHref)[1:]))
	}
	return users, nextHref
}

// NewSpider .
func NewSpider(goNum int) Spider {
	spider := &Exec{
		goNum: goNum,
		// queueChan: make([]chan string, goNum),
	}
	queueChan := make([]chan string, spider.goNum)
	for i := 0; i < spider.goNum; i++ {
		queueChan[i] = make(chan string, 1)
	}
	spider.queueChan = queueChan
	return spider
}
