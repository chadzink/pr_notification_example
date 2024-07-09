package setup

import (
	"encoding/json"
	"github_pr_notify/github"
	"os"
)

type Setup struct {
	GitHub struct {
		GitHubToken  string              // not setup in config.json - pulled from env variable
		Organization github.Organization `json:"organization"`
	} `json:"github"`
	Slack struct {
		SlackToken string
		WebhookURL string `json:"webhookUrl"`
	} `json:"slack"`
}

// Load setup loads the configuration from the setup.json file
func (c *Setup) Load() error {
	file, err := os.Open("./setup/setup.json")
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&c)
	if err != nil {
		return err
	}

	return nil
}
