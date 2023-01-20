package main

import (
	"flag"
	"os"
)

const (
	image      = "yuta42173/ubuntu"
	owner      = "nayuta-ai"
	repo       = "k8s-argo"
	mainBranch = "main"
	filePath   = "dev/deployment.yaml"
)

var (
	tag   *string
	token *string
)

func init() {
	tag = flag.String("tag", "latest", "a tag of the docker image")
	token = flag.String("token", "", "a personal github token for the owner")
}

func main() {
	flag.Parse()
	os.Exit(replaceTags(tag, token))
}