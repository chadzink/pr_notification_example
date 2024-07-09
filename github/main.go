package github

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
)

type Organization struct {
	Name         string       `json:"name"`
	Repositories []Repository `json:"repositories"`
	TeamMembers  []TeamMember `json:"teamMembers"`
}

type Repository struct {
	Name         string        `json:"name"`
	PullRequests []PullRequest `json:"pullRequests"`
}

type PullRequest struct {
	Repository     string   `json:"repository"`
	Number         int      `json:"number"`
	Title          string   `json:"title"`
	Permalink      string   `json:"permalink"`
	UpdatedAt      string   `json:"updatedAt"`
	State          string   `json:"state"`
	Additions      int      `json:"additions"`
	Deletions      int      `json:"deletions"`
	ChangedFiles   int      `json:"changedFiles"`
	AuthorUserName string   `json:"authorUserName"`
	ReviewDecision string   `json:"reviewDecision"`
	Reviews        []Review `json:"reviews"`
}

type Review struct {
	ReviewerUserName string `json:"reviewerUserName"`
	PublishedAt      string `json:"publishedAt"`
	State            string `json:"state"`
}

type TeamMember struct {
	GithubUserName string        `json:"userName"`
	EMail          string        `json:"email"`
	DisplayName    string        `json:"displayName"`
	PullRequests   []PullRequest `json:"pullRequests"`
}

type RepoSummaryEntry struct {
	Repository   string                    `json:"repository"`
	TotalOpenPRs int                       `json:"totalPRs"`
	OpenPRs      []PullRequestSummaryEntry `json:"openPRs"`
}

type PullRequestSummaryEntry struct {
	Title           string `json:"title"`
	Permalink       string `json:"permalink"`
	Author          string `json:"authorUserName"`
	UpdatedAt       string `json:"updatedAt"`
	Approvals       int    `json:"approvals"`
	LastApprovalAt  string `json:"lastApprovalAt"`
	Rejections      int    `json:"rejections"`
	LastRejectionAt string `json:"lastRejectionAt"`
	Reviews         int    `json:"reviews"`
	LastReviewAt    string `json:"lastReviewAt"`
}

func (org *Organization) LoadOpenPullRequests(token string) error {
	// repoSummary := make(map[string]RepoSummaryEntry)

	// Get the OPEN pull requests for the repositories in the organization
	for index, repository := range org.Repositories {
		// Get the pull requests for the repository
		pullRequests, err := getRepoPullRequests(org.Name, repository.Name, token)
		if err != nil {
			log.Fatal(err)
		}

		// Add the pull requests to the repository
		org.Repositories[index].PullRequests = pullRequests
	}

	// Get the OPEN pull requests for the team members
	for index, teamMember := range org.TeamMembers {

		// Get the pull requests for the team member
		pullRequests, err := getPullRequestsForUser(teamMember.GithubUserName, token)

		if err != nil {
			log.Fatal(err)
		}

		// Add the pull requests to the team member
		org.TeamMembers[index].PullRequests = pullRequests
	}

	return nil
}

func PostTo(url string, payload []byte, githubToken string) string {
	// Create a new client to make a request to the github API
	client := &http.Client{}

	// Create a new request
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(payload))
	if err != nil {
		log.Fatal(err)
	}

	// Set the headers for the request
	req.Header.Set("Authorization", "Bearer "+githubToken)
	req.Header.Set("Content-Type", "application/json")

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	// Read the response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	return string(respBody)
}

// Get the repositorys for an organization on github
func getRepoPullRequests(ownerName, repoName, githubToken string) ([]PullRequest, error) {
	query := `
		query ($owner: String!, $name: String!) {
			repo: repository(owner: $owner, name: $name) {
				name
				pullRequests(states: [OPEN], last: 50) {
					nodes {
						number
						title
						permalink
						updatedAt
						state
						additions
						deletions
						changedFiles
						PrAuthor: author {
						username: login
						}
						reviewDecision
						reviews(states: [COMMENTED, APPROVED, CHANGES_REQUESTED], last: 10) {
						nodes {
							reviewer: author {
							username: login
							}
								publishedAt
								state
							}
						}
					}
				}
			}
		}`

	type PullRequestGraphQLPayload struct {
		Query     string `json:"query"`
		Variables struct {
			Owner string `json:"owner"`
			Name  string `json:"name"`
		} `json:"variables"`
	}

	payload := PullRequestGraphQLPayload{Query: query}
	payload.Variables.Owner = ownerName
	payload.Variables.Name = repoName

	// Convert the payload to a byte array
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	data := PostTo("https://api.github.com/graphql", payloadBytes, githubToken)

	type PullRequestsQuery struct {
		Data struct {
			Repo struct {
				Name         string `json:"name"`
				PullRequests struct {
					Nodes []struct {
						Number       int    `json:"number"`
						Title        string `json:"title"`
						Permalink    string `json:"permalink"`
						UpdatedAt    string `json:"updatedAt"`
						State        string `json:"state"`
						Additions    int    `json:"additions"`
						Deletions    int    `json:"deletions"`
						ChangedFiles int    `json:"changedFiles"`
						PrAuthor     struct {
							Username string `json:"username"`
						} `json:"PrAuthor"`
						ReviewDecision string `json:"reviewDecision"`
						Reviews        struct {
							Nodes []struct {
								Reviewer struct {
									Username string `json:"username"`
								} `json:"reviewer"`
								PublishedAt string `json:"publishedAt"`
								State       string `json:"state"`
							} `json:"nodes"`
						} `json:"reviews"`
					} `json:"nodes"`
				} `json:"pullRequests"`
			} `json:"repo"`
		} `json:"data"`
	}

	// Unmarshal the response into the struct
	repos := &PullRequestsQuery{}
	derr := json.Unmarshal([]byte(data), repos)
	if derr != nil {
		return nil, derr
	}

	// Map the pull requests to the PullRequest struct
	var pullRequests []PullRequest
	for _, pr := range repos.Data.Repo.PullRequests.Nodes {
		var reviews []Review
		for _, review := range pr.Reviews.Nodes {
			reviews = append(reviews, Review{
				ReviewerUserName: review.Reviewer.Username,
				PublishedAt:      review.PublishedAt,
				State:            review.State,
			})
		}

		pullRequests = append(pullRequests, PullRequest{
			Repository:     repos.Data.Repo.Name,
			Number:         pr.Number,
			Title:          pr.Title,
			Permalink:      pr.Permalink,
			UpdatedAt:      pr.UpdatedAt,
			State:          pr.State,
			Additions:      pr.Additions,
			Deletions:      pr.Deletions,
			ChangedFiles:   pr.ChangedFiles,
			AuthorUserName: pr.PrAuthor.Username,
			ReviewDecision: pr.ReviewDecision,
			Reviews:        reviews,
		})
	}

	return pullRequests, nil
}

// get the pull request for a user with OPEN, CLOSED, MERGED states over a goven date range
func getPullRequestsForUser(userName, githubToken string) ([]PullRequest, error) {
	query := `
		query ($userName: String!) {
			user: user(login: $userName) {
				login
				pullRequests(states: [OPEN], first: 50, orderBy: {field: UPDATED_AT, direction: DESC}) {
					nodes {
						repository {
							name
						}
						number
						title
						permalink
						updatedAt
						state
						additions
						deletions
						changedFiles
						PrAuthor: author {
						username: login
						}
						reviewDecision
						reviews(states: [COMMENTED, APPROVED, CHANGES_REQUESTED], last: 10) {
						nodes {
							reviewer: author {
							username: login
							}
								publishedAt
								state
							}
						}
					}
				}
			}
		}`

	type PullRequestGraphQLPayload struct {
		Query     string `json:"query"`
		Variables struct {
			UserName string `json:"userName"`
		} `json:"variables"`
	}

	payload := PullRequestGraphQLPayload{Query: query}
	payload.Variables.UserName = userName

	// Convert the payload to a byte array
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	data := PostTo("https://api.github.com/graphql", payloadBytes, githubToken)

	type PullRequestsQuery struct {
		Data struct {
			User struct {
				Login        string `json:"login"`
				PullRequests struct {
					Nodes []struct {
						Repository struct {
							Name string `json:"name"`
						} `json:"repository"`
						Number       int    `json:"number"`
						Title        string `json:"title"`
						Permalink    string `json:"permalink"`
						UpdatedAt    string `json:"updatedAt"`
						State        string `json:"state"`
						Additions    int    `json:"additions"`
						Deletions    int    `json:"deletions"`
						ChangedFiles int    `json:"changedFiles"`
						PrAuthor     struct {
							Username string `json:"username"`
						} `json:"PrAuthor"`
						ReviewDecision string `json:"reviewDecision"`
						Reviews        struct {
							Nodes []struct {
								Reviewer struct {
									Username string `json:"username"`
								} `json:"reviewer"`
								PublishedAt string `json:"publishedAt"`
								State       string `json:"state"`
							} `json:"nodes"`
						} `json:"reviews"`
					} `json:"nodes"`
				} `json:"pullRequests"`
			} `json:"user"`
		} `json:"data"`
	}

	// Unmarshal the response into the struct
	repos := &PullRequestsQuery{}
	derr := json.Unmarshal([]byte(data), repos)
	if derr != nil {
		return nil, derr
	}

	// Map the pull requests to the PullRequest struct
	var pullRequests []PullRequest
	for _, pr := range repos.Data.User.PullRequests.Nodes {
		var reviews []Review
		for _, review := range pr.Reviews.Nodes {
			reviews = append(reviews, Review{
				ReviewerUserName: review.Reviewer.Username,
				PublishedAt:      review.PublishedAt,
				State:            review.State,
			})
		}

		pullRequests = append(pullRequests, PullRequest{
			Repository:     pr.Repository.Name,
			Number:         pr.Number,
			Title:          pr.Title,
			Permalink:      pr.Permalink,
			UpdatedAt:      pr.UpdatedAt,
			State:          pr.State,
			Additions:      pr.Additions,
			Deletions:      pr.Deletions,
			ChangedFiles:   pr.ChangedFiles,
			AuthorUserName: pr.PrAuthor.Username,
			ReviewDecision: pr.ReviewDecision,
			Reviews:        reviews,
		})
	}

	return pullRequests, nil
}
