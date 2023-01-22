package test

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/google/go-github/v49/github"
	yaml "gopkg.in/yaml.v2"
)

type TestConfig struct {
	client *github.Client
}

const (
	image      = "yuta42173/ubuntu:latest"
	owner      = "nayuta-ai"
	repo       = "k8s-argo"
	mainBranch = "main"
	filePath   = "dev/deployment.yaml"
)

func (c *TestConfig) FetchBranch(mainBranchName string, newBranchName string) (*github.Reference, error) {
	ref, rsp, err := c.client.Git.GetRef(context.Background(), owner, repo, "heads/"+newBranchName)
	if err != nil {
		if rsp.Response.StatusCode == 404 {
			fmt.Printf("Branch %s does not exist\n", newBranchName)
			// Create new branch
			ref, err := c.NewBranch(mainBranchName, newBranchName)
			if err != nil {
				return nil, err
			}
			fmt.Printf("Create branch %s, SHA: %s\n", newBranchName, *ref.Object.SHA)
			return ref, nil
		}
		return nil, err
	}
	fmt.Printf("Branch %s exists, SHA: %s\n", newBranchName, *ref.Object.SHA)
	return ref, nil
}

func (c *TestConfig) NewBranch(mainBranchName string, newBranchName string) (*github.Reference, error) {
	// Get the main branch
	branchInfo, _, err := c.client.Repositories.GetBranch(context.Background(), owner, repo, mainBranchName, false)
	if err != nil {
		return nil, err
	}
	// Get the current HEAD commit in the main branch
	commit, _, err := c.client.Repositories.GetCommit(context.Background(), owner, repo, *branchInfo.Commit.SHA, &github.ListOptions{})
	if err != nil {
		return nil, err
	}

	// Create new branch structure
	newRef := &github.Reference{
		Ref: github.String("refs/heads/" + newBranchName),
		Object: &github.GitObject{
			SHA: commit.SHA,
		},
	}
	ref, _, err := c.client.Git.CreateRef(context.Background(), owner, repo, newRef)
	if err != nil {
		return nil, err
	}
	return ref, nil
}

func (c *TestConfig) UpdateConfigFile() (*github.RepositoryContent, []byte, error) {
	// Get the contents of the file
	fileContent, _, _, err := c.client.Repositories.GetContents(context.Background(), owner, repo, filePath, nil)
	if err != nil {
		return nil, nil, err
	}

	// Decode the base64 encoded file contents
	decodedContent, err := base64.StdEncoding.DecodeString(*fileContent.Content)
	if err != nil {
		return nil, nil, err
	}
	// Unmarshal the YAML contents into a Deployment struct
	var yamlObj interface{}
	err = yaml.Unmarshal(decodedContent, &yamlObj)
	if err != nil {
		return nil, nil, err
	}
	yamlObj.(map[interface{}]interface{})["spec"].(map[interface{}]interface{})["template"].(map[interface{}]interface{})["spec"].(map[interface{}]interface{})["containers"].([]interface{})[0].(map[interface{}]interface{})["image"] = image
	//Marshal struct to yaml
	newContent, err := yaml.Marshal(&yamlObj)
	if err != nil {
		return nil, nil, err
	}
	return fileContent, newContent, nil
}

func (c *TestConfig) NewPullRequest(newContent []byte, sha *string, newBranch string) error {
	commitMessage := "Update deployment TestConfiguration"
	// Create a new file commit
	newCommit := &github.RepositoryContentFileOptions{
		Message: &commitMessage,
		Content: newContent,
		SHA:     sha,
		Branch:  &newBranch,
	}
	_, _, err := c.client.Repositories.UpdateFile(context.Background(), owner, repo, filePath, newCommit)
	if err != nil {
		return err
	}
	head := owner + ":" + newBranch
	base := mainBranch
	title := "New feature: Container name update"
	body := "This PR updates the container name in the deployment.yaml file."

	// Create the pull request
	newPR := &github.NewPullRequest{
		Title: &title,
		Head:  &head,
		Base:  &base,
		Body:  &body,
	}
	pr, _, err := c.client.PullRequests.Create(context.Background(), owner, repo, newPR)
	if err != nil {
		return err
	}

	fmt.Printf("Successfully created PR %d\n", *pr.Number)
	return nil
}
