package jira_int

import (
	"fmt"
	"log"

	"os"

	"github.com/joho/godotenv"

	jira "github.com/andygrunwald/go-jira"
)

// Get jira users
func GetUsers() []jira.User {
	// Load the .env file
	godotenv.Load(".env")
	// Create a BasicAuth Transport object
	tp := jira.BasicAuthTransport{
		Username: os.Getenv("JIRA_USER"),
		Password: os.Getenv("JIRA_TOKEN"),
	}
	// Create a new Jira Client
	client, err := jira.NewClient(tp.Client(), os.Getenv("JIRA_URL"))
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Fetching jira users with a limit of 1000")
	jiraUsers, _, _ := client.User.Find("&maxResults=1000")
	log.Printf("Jira users fetched. Number of Jira users: " + fmt.Sprint(len(jiraUsers)))

	return jiraUsers

}

// Получение пользователей atlassioan (исключены боты и тд)
func GetAtlassinUsers() []jira.User {

	jiraUsers := GetUsers()
	var atlassianUsers []jira.User

	for _, v := range jiraUsers {

		if v.AccountType == "atlassian" {
			atlassianUsers = append(atlassianUsers, v)
		}
	}
	log.Printf("Jira users shrinked to Atlassian users. Number of Atlassian users: " + fmt.Sprint(len(atlassianUsers)))

	return atlassianUsers

}

// Возвращает задачи пользователя `assignee`
func GetIssueByAssignee(assignee jira.User, pageSize int) []jira.Issue {
	godotenv.Load(".env")
	// Create a BasicAuth Transport object
	tp := jira.BasicAuthTransport{
		Username: os.Getenv("JIRA_USER"),
		Password: os.Getenv("JIRA_TOKEN"),
	}
	// Create a new Jira Client
	client, err := jira.NewClient(tp.Client(), os.Getenv("JIRA_URL"))

	jql := "project = 'ITP' AND assignee IN (" + assignee.AccountID + ") AND statusCategory in ('To Do', 'In Progress') ORDER BY created DESC"

	issues, err := GetAllIssues(client, jql, pageSize, 0)
	if err != nil {
		log.Printf(err.Error())
	}
	return issues

}

// Возвращает задачи пользователя в статусе "Разрабатывается"
func GetIssuesInJobByAssignee(assignee jira.User, pageSize int) []jira.Issue {
	godotenv.Load(".env")
	// Create a BasicAuth Transport object
	tp := jira.BasicAuthTransport{
		Username: os.Getenv("JIRA_USER"),
		Password: os.Getenv("JIRA_TOKEN"),
	}
	// Create a new Jira Client
	client, err := jira.NewClient(tp.Client(), os.Getenv("JIRA_URL"))
	jql := "project = 'ITP' AND status IN ('Разрабатывается') AND assignee = " + assignee.AccountID + " ORDER BY created DESC"

	issues, err := GetAllIssues(client, jql, pageSize, 0)
	if err != nil {
		log.Printf(err.Error())
	}
	return issues
}

// Возвращает задачи, созданные пользователем
func GetIssuesByReporter(reporter jira.User, pageSize int) []jira.Issue {
	godotenv.Load(".env")
	// Create a BasicAuth Transport object
	tp := jira.BasicAuthTransport{
		Username: os.Getenv("JIRA_USER"),
		Password: os.Getenv("JIRA_TOKEN"),
	}
	// Create a new Jira Client
	client, err := jira.NewClient(tp.Client(), os.Getenv("JIRA_URL"))
	jql := "project = 'ITP' AND reporter = " + reporter.AccountID + " ORDER BY created DESC"

	issues, err := GetAllIssues(client, jql, pageSize, 0)
	if err != nil {
		log.Printf(err.Error())
	}
	return issues
}

// Возвращает задачи пользователя в статус "Бэклог" и "Подлежит разработке"
func GetIssuesInBackByAssignee(assignee jira.User, pageSize int) []jira.Issue {
	godotenv.Load(".env")
	// Create a BasicAuth Transport object
	tp := jira.BasicAuthTransport{
		Username: os.Getenv("JIRA_USER"),
		Password: os.Getenv("JIRA_TOKEN"),
	}
	// Create a new Jira Client
	client, err := jira.NewClient(tp.Client(), os.Getenv("JIRA_URL"))
	jql := "project = 'ITP' AND status IN ('Бэклог','Подлежит разработке') AND assignee = " + assignee.AccountID + " ORDER BY created DESC"

	issues, err := GetAllIssues(client, jql, pageSize, 0)
	if err != nil {
		log.Printf(err.Error())
	}
	return issues
}

// Возвращает все задачи по JQL запросу `seatchString`
func GetAllIssues(client *jira.Client, searchString string, MaxResults int, StartAt int) ([]jira.Issue, error) {

	var issues []jira.Issue
	for {
		opt := &jira.SearchOptions{
			MaxResults: MaxResults,
			StartAt:    StartAt,
		}

		chunk, resp, err := client.Issue.Search(searchString, opt)
		if err != nil {
			return nil, err
		}

		total := resp.Total
		if issues == nil {
			issues = make([]jira.Issue, 0, total)
		}
		issues = append(issues, chunk...)
		StartAt = resp.StartAt + len(chunk)
		if StartAt >= total {
			return issues, nil
		}
	}

}
