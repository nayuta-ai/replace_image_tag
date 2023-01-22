package test

import (
	replace "app/src"
	"context"
	"os"
	"testing"

	"github.com/google/go-github/v32/github"
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

func TestNewBranch(t *testing.T) {}
