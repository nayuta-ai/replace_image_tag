package main

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/google/go-github/v32/github"
	"golang.org/x/oauth2"
	yaml "gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
)

func replaceTags(tag, token *string) int {
	if *token == "" {
		fmt.Println("input error: a personal access token does not exist")
		return 1
	}
	client := newAuthenticatedGitHubClient(token)
	// Specify the repository, new branch name and the SHA of the commit you want the branch to point to
	newBranchName := "update_tag"

	err := fetchBranch(client, mainBranch, newBranchName)
	if err != nil {
		fmt.Println(err)
		return 1
	}

	fileContent, newContent, err := updateConfigFile(client)
	if err != nil {
		fmt.Println(err)
		return 1
	}
	err = newPullRequest(client, newContent, fileContent.SHA, newBranchName)
	if err != nil {
		fmt.Println(err)
		return 1
	}
	return 0
}

func newAuthenticatedGitHubClient(token *string) *github.Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: *token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return github.NewClient(tc)
}

func fetchBranch(client *github.Client, mainBranchName string, newBranchName string) error {
	ref, _, err := client.Git.GetRef(context.Background(), owner, repo, "heads/"+newBranchName)
	if err != nil {
		if err, ok := err.(*github.ErrorResponse); ok {
			if err.Response.StatusCode == 404 {
				fmt.Printf("Branch %s does not exist\n", newBranchName)
				// Create new branch
				ref, err := newBranch(client, mainBranchName, newBranchName)
				if err != nil {
					return err
				}
				fmt.Printf("Create branch %s, SHA: %s\n", newBranchName, *ref.Object.SHA)
				return nil
			}
		}
		return err
	}
	fmt.Printf("Branch %s exists, SHA: %s\n", newBranchName, *ref.Object.SHA)
	return nil
}

func newBranch(client *github.Client, mainBranchName string, newBranchName string) (*github.Reference, error) {
	// Get the main branch
	branchInfo, _, err := client.Repositories.GetBranch(context.Background(), owner, repo, mainBranchName)
	if err != nil {
		return nil, err
	}
	// Get the current HEAD commit in the main branch
	commit, _, err := client.Repositories.GetCommit(context.Background(), owner, repo, *branchInfo.Commit.SHA)
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
	ref, _, err := client.Git.CreateRef(context.Background(), owner, repo, newRef)
	if err != nil {
		return nil, err
	}
	return ref, nil
}

func updateConfigFile(client *github.Client) (*github.RepositoryContent, []byte, error) {
	// Get the contents of the file
	fileContent, _, _, err := client.Repositories.GetContents(context.Background(), owner, repo, filePath, nil)
	if err != nil {
		return nil, nil, err
	}

	// Decode the base64 encoded file contents
	decodedContent, err := base64.StdEncoding.DecodeString(*fileContent.Content)
	if err != nil {
		return nil, nil, err
	}
	// Unmarshal the YAML contents into a Deployment struct
	deployment := &appsv1.Deployment{}
	err = yaml.Unmarshal(decodedContent, deployment)
	if err != nil {
		return nil, nil, err
	}
	deployment.Spec.Template.Spec.Containers[0].Image = image
	//Marshal struct to yaml
	newContent, err := yaml.Marshal(deployment)
	if err != nil {
		return nil, nil, err
	}
	return fileContent, newContent, nil
}

func newPullRequest(client *github.Client, newContent []byte, sha *string, newBranch string) error {
	commitMessage := "Update deployment configuration"
	// Create a new file commit
	newCommit := &github.RepositoryContentFileOptions{
		Message: &commitMessage,
		Content: newContent,
		SHA:     sha,
		Branch:  &newBranch,
	}
	_, _, err := client.Repositories.UpdateFile(context.Background(), owner, repo, filePath, newCommit)
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
	pr, _, err := client.PullRequests.Create(context.Background(), owner, repo, newPR)
	if err != nil {
		return err
	}

	fmt.Printf("Successfully created PR %d\n", *pr.Number)
	return nil
}
