package test

import (
	replace "app/src"
	"context"
	"encoding/base64"
	"net/http"
	"os"
	"testing"

	"github.com/google/go-github/v49/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
)

func TestNewAuthenticatedGitHubClient(t *testing.T) {
	validToken := os.Getenv("GITHUB_TOKEN")
	var test = []struct {
		name  string
		token string
	}{
		{
			name:  "wrong case",
			token: "abc123",
		},
		{
			name:  "correct case",
			token: validToken,
		},
	}
	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			client := replace.NewAuthenticatedGitHubClient(&tt.token)
			// Type Checking
			var i interface{} = client
			if kind, ok := i.(*github.Client); !ok {
				t.Fatalf("type error: Got %v, but expected *github.Client", kind)
			}
			// Fetch GitHub branch information using client variable.
			_, _, err := client.Git.GetRef(context.Background(), os.Getenv("GITHUB_USER"), os.Getenv("GITHUB_REPOSITORY"), "heads/"+os.Getenv("GITHUB_BRANCH"))
			switch tt.name {
			case "wrong case":
				// In this case, tt.token is invalid token, so err must be empty error.
				if err == nil {
					t.Fatal("unexpected error: Got empty error, but expected non-empty error")
				}
			case "correct case":
				// In this case, tt.token is valid token, so err must be non-empty error.
				if err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func TestNewBranch(t *testing.T) {
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetReposBranchesByOwnerByRepoByBranch,
			github.Branch{
				Name: github.String("update_tag"),
				Commit: &github.RepositoryCommit{
					SHA: github.String("abcd1234"),
				},
			},
		),
		mock.WithRequestMatch(
			mock.GetReposCommitsByOwnerByRepoByRef,
			github.Commit{
				SHA: github.String("abcd1234"),
			},
		),
		mock.WithRequestMatch(
			mock.PostReposGitRefsByOwnerByRepo,
			github.Reference{
				Object: &github.GitObject{
					SHA: github.String("abcd1234"),
				},
			},
		),
	)
	client := github.NewClient(mockedHTTPClient)
	c := replace.Client{
		Model: &TestConfig{client: client},
	}
	ref, err := c.Model.NewBranch("main", "update_tag")
	if err != nil {
		t.Fatal(err)
	}
	if *ref.Object.SHA != "abcd1234" {
		t.Fatalf("Got %v, but expected abcd1234", *ref.Object.SHA)
	}
}

func TestFetchBranch(t *testing.T) {
	clientBranchExistence := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetReposGitRefByOwnerByRepoByRef,
			github.Reference{
				Ref: github.String("update_tag"),
				Object: &github.GitObject{
					SHA: github.String("abcd1234"),
				},
			},
		),
	)
	clientBranchAbsence := mock.NewMockedHTTPClient(
		mock.WithRequestMatchHandler(
			mock.GetReposGitRefByOwnerByRepoByRef,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mock.WriteError(
					w,
					http.StatusNotFound,
					"The branch name dosen't exist.",
				)
			}),
		),
		mock.WithRequestMatch(
			mock.GetReposBranchesByOwnerByRepoByBranch,
			github.Branch{
				Name: github.String("update_tag"),
				Commit: &github.RepositoryCommit{
					SHA: github.String("abcd1234"),
				},
			},
		),
		mock.WithRequestMatch(
			mock.GetReposCommitsByOwnerByRepoByRef,
			github.Commit{
				SHA: github.String("abcd1234"),
			},
		),
		mock.WithRequestMatch(
			mock.PostReposGitRefsByOwnerByRepo,
			github.Reference{
				Object: &github.GitObject{
					SHA: github.String("abcd1234"),
				},
			},
		),
	)
	test := []struct {
		name         string
		mockedClient *http.Client
	}{
		{
			name:         "Branch Exists",
			mockedClient: clientBranchExistence,
		},
		{
			name:         "No Branch Exists",
			mockedClient: clientBranchAbsence,
		},
	}
	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			client := github.NewClient(tt.mockedClient)
			c := replace.Client{
				Model: &TestConfig{client: client},
			}
			ref, err := c.Model.FetchBranch("main", "update_tag")
			if err != nil {
				t.Fatal(err)
			}
			if *ref.Object.SHA != "abcd1234" {
				t.Fatalf("Got %v, but expected abcd1234", *ref.Object.SHA)
			}
		})
	}
}

func TestUpdateConfigFile(t *testing.T) {
	content, err := os.ReadFile("sample.yaml")
	if err != nil {
		t.Fatal(err)
	}
	baseContent := base64.StdEncoding.EncodeToString(content)
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetReposContentsByOwnerByRepoByPath,
			github.RepositoryContent{
				Content: github.String(baseContent),
			},
		),
	)
	client := github.NewClient(mockedHTTPClient)
	c := replace.Client{
		Model: &TestConfig{client: client},
	}
	fileContent, newContent, err := c.Model.UpdateConfigFile()
	if err != nil {
		t.Fatal(err)
	}
	if *fileContent.Content != baseContent {
		t.Fatalf("Got %v, but expected %v", *fileContent.Content, baseContent)
	}
	content, err = os.ReadFile("test.yaml")
	if err != nil {
		t.Fatal(err)
	}
	for i, _ := range newContent {
		if newContent[i] != content[i] {
			t.Fatalf("Phrease %v: Got %v, but expected %v", i, string(newContent[i]), string(content[i]))
		}
	}
}

func TestNewPullRequest(t *testing.T) {
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.PutReposContentsByOwnerByRepoByPath,
			github.RepositoryContentResponse{},
		),
		mock.WithRequestMatch(
			mock.PostReposPullsByOwnerByRepo,
			github.PullRequest{
				Number: github.Int(32),
			},
		),
	)
	client := github.NewClient(mockedHTTPClient)
	c := replace.Client{
		Model: &TestConfig{client: client},
	}
	err := c.Model.NewPullRequest([]byte("string"), github.String("abcd1234"), "update_tag")
	if err != nil {
		t.Fatal(err)
	}
}
