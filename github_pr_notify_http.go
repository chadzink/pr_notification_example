package github_pr_notify

import (
	"fmt"
	"github_pr_notify/setup"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
)

func init() {
	functions.HTTP("PrNotifyHTTP", PrNotifyHTTP)
}

// HelloHTTP is an HTTP Cloud Function with a request parameter.
func PrNotifyHTTP(w http.ResponseWriter, r *http.Request) {
	var gitHubToken string = ""
	if envGitHubToken := loadGitHubTokenFromDotEnvFile(); envGitHubToken != "" {
		gitHubToken = envGitHubToken
	}

	// Load the setup using the local package for setup
	Setup := setup.Setup{}
	Setup.Load()
	Setup.GitHub.GitHubToken = gitHubToken

	// Print the github token if it was found
	if gitHubToken != "" {
		fmt.Fprintf(w, "GitHub Token: %s\n", gitHubToken)
	} else {
		fmt.Fprint(w, "GitHub Token not found\n")
	}
}

func loadGitHubTokenFromDotEnvFile() string {
	return loadValueFromDotEnvFile("GITHUB_TOKEN")
}

func loadSlackTokenFromDotEnvFile() string {
	return loadValueFromDotEnvFile("SLACK_TOKEN")
}

func loadValueFromDotEnvFile(key string) string {
	// Check the cwd for a .env file, If it exists, load the value for the key
	cmd := exec.Command("sh", "-c", "cat .env | grep "+key+" | cut -d '=' -f 2")

	stdout, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}

	// If the file does not exist, return an empty string
	if len(stdout) > 0 {
		return string(stdout)
	}

	// try to load the value from the environment
	if envValue := os.Getenv(key); envValue != "" {
		return envValue
	}

	return ""
}
