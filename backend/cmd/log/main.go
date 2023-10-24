package main

import (
	"fmt"
	_ "github.com/lib/pq"
	"path/filepath"
)

func main() {

	//logger := log.Default()

	db := &Database{}
	db.Connect()

	root := "/home/boris/route53/6e2aecce-5e91-4b44-a75c-23e97ccc7442"
	err := filepath.WalkDir(root, db.ProcessFile)
	if err != nil {
		fmt.Printf("error walking the path %v: %v\n", root, err)
	}
}
