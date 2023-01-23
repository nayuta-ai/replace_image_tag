package main

import (
	replace "app/src"
	"flag"
	"log"
	"os"
)

var (
	tag   *string
	token *string
)

func init() {
	tag = flag.String("tag", "latest", "a tag of the docker image")
	token = flag.String("token", "", "a personal github token for the owner")
	log.Printf("tag: %v, token: %v", tag, token)
}

func main() {
	flag.Parse()
	os.Exit(replace.ReplaceTags(tag, token))
}
