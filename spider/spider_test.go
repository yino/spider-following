package spider

import (
	"fmt"
	"os"
	"testing"
)

func TestParseHtml(t *testing.T) {
	content, err := os.ReadFile("./test.html")
	if err != nil {
		panic(err)
	}
	userList, nextHref := ParseHtml(string(content))
	fmt.Println(userList, nextHref)
}
