package main

import (
	replace "app/src"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
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
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	tmp := os.Getenv("GITHUB_TOKEN")
	token = &tmp
	os.Exit(replace.ReplaceTags(tag, token))
}
