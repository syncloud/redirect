package main

import (
	"github.com/syncloud/redirect/cmd"
)

func main() {
	main := cmd.NewMain()
	if main == nil {
		return
	}
	main.StartWww()
}
