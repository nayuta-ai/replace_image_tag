package replace

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/google/go-github/v49/github"
	"golang.org/x/oauth2"
	yaml "gopkg.in/yaml.v2"
)

type GitHubInterface interface {
	FetchBranch(string, string) (*github.Reference, error)
	NewBranch(string, string) (*github.Reference, error)
	UpdateConfigFile() (*github.RepositoryContent, []byte, error)
	NewPullRequest([]byte, *string, string) error
}

type Client struct {
	Model GitHubInterface
}

type Config struct {
	client *github.Client
}

const (
	owner      = "nayuta-ai"
	repo       = "k8s-argo"
	mainBranch = "main"
	filePath   = "dev/deployment.yaml"
)

var imageName  = "yuta42173/ubuntu:"

func ReplaceTags(tag, token *string) int {
	imageName += *tag
	if *token == "" {
		fmt.Println("input error: a personal access token does not exist")
		return 1
	}
	client := NewAuthenticatedGitHubClient(token)
	c := Client{
		Model: &Config{client: client},
	}
	// Specify the repository, new branch name and the SHA of the commit you want the branch to point to
	newBranchName := "update_tag"

	_, err := c.Model.FetchBranch(mainBranch, newBranchName)
	if err != nil {
		fmt.Println(err)
		return 1
	}

	fileContent, newContent, err := c.Model.UpdateConfigFile()
	if err != nil {
		fmt.Println(err)
		return 1
	}
	err = c.Model.NewPullRequest(newContent, fileContent.SHA, newBranchName)
	if err != nil {
		fmt.Println(err)
		return 1
	}
	return 0
}

func NewAuthenticatedGitHubClient(token *string) *github.Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: *token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return github.NewClient(tc)
}

func (c *Config) FetchBranch(mainBranchName string, newBranchName string) (*github.Reference, error) {
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

func (c *Config) NewBranch(mainBranchName string, newBranchName string) (*github.Reference, error) {
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

func (c *Config) UpdateConfigFile() (*github.RepositoryContent, []byte, error) {
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
	yamlObj.(map[interface{}]interface{})["spec"].(map[interface{}]interface{})["template"].(map[interface{}]interface{})["spec"].(map[interface{}]interface{})["containers"].([]interface{})[0].(map[interface{}]interface{})["image"] = imageName
	//Marshal struct to yaml
	newContent, err := yaml.Marshal(&yamlObj)
	if err != nil {
		return nil, nil, err
	}
	return fileContent, newContent, nil
}

func (c *Config) NewPullRequest(newContent []byte, sha *string, newBranch string) error {
	commitMessage := "Update deployment configuration"
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
