package main

import (
	"github.com/goplus/llgo/internal/runtime/c"
)

func concat(args ...string) (ret string) {
	for _, v := range args {
		ret += v
	}
	return
}

func info(s string) string {
	return "" + s + "..."
}

func main() {
	result := concat("Hello", " ", "World")
	c.Fprintf(c.Stderr, c.Str("Hi, %s\n"), c.AllocaCStr(result))
}
