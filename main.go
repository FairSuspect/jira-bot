package main

import (
	"context"
	"fmt"
	"log"
	"os"
	jira_int "programmingpercy/slack-bot/jira"
	slack_int "programmingpercy/slack-bot/slack"
	"strconv"
	"strings"

	"github.com/andygrunwald/go-jira"
	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

func main() {

	// Load Env variables from .dot file
	godotenv.Load(".env")

	token := os.Getenv("SLACK_AUTH_TOKEN")
	appToken := os.Getenv("SLACK_APP_TOKEN")
	channelID := os.Getenv("SLACK_CHANNEL_ID")

	// Create a new client to slack by giving token
	// Set debug to true while developing
	// Also add a ApplicationToken option to the client
	client := slack.New(token, slack.OptionDebug(true), slack.OptionAppLevelToken(appToken))
	// go-slack comes with a SocketMode package that we need to use that accepts a Slack client and outputs a Socket mode client instead
	socketClient := socketmode.New(
		client,
		socketmode.OptionDebug(true),
		// Option to set a custom logger
		socketmode.OptionLog(log.New(os.Stdout, "socketmode: ", log.Lshortfile|log.LstdFlags)),
	)

	slackUsers, err := client.GetUsers()

	if err != nil {
		log.Panicf(err.Error())
	}
	log.Printf("Slack users fetched. Number of Slack users: " + fmt.Sprint(len(slackUsers)))

	// Create a context that can be used to cancel goroutine
	ctx, cancel := context.WithCancel(context.Background())
	atlassianUsers := jira_int.GetAtlassinUsers()

	// Make this cancel called properly in a real program , graceful shutdown etc
	defer cancel()
	go func(ctx context.Context, client *slack.Client, socketClient *socketmode.Client) {
		// Create a for loop that selects either the context cancellation or the events incomming
		for {
			select {
			// inscase context cancel is called exit the goroutine
			case <-ctx.Done():
				log.Println("Shutting down socketmode listener")
				return
				// Обработка ивента
			case event := <-socketClient.Events:
				switch event.Type {
				// handle EventAPI events

				// Обработка слэш команд
				case socketmode.EventTypeSlashCommand:

					eventMessage, ok := event.Data.(slack.SlashCommand)
					if !ok {
						log.Printf("Could not type cast the event to the MessageEvent: %v\n", event)
						postMessage("Внутренняя ошибка бота", client, channelID)
						continue
					}

					Arguments := strings.Split(eventMessage.Text, " ")

					UserName := Arguments[0]
					if UserName[0] != '@' {
						postMessage("Команда работает только с упоминанием пользователя", client, channelID)
						socketClient.Ack(*event.Request)
						continue
					}
					UserName = UserName[1:]

					slackUser := slack_int.FindSlackUser(&UserName, &slackUsers)

					pageSize := 10
					if len(Arguments) > 1 {
						pageSize, err = strconv.Atoi(Arguments[1])
						if err != nil {
							postMessage("Неверный второй аргумент. Должно быть число.\n "+err.Error(), client, channelID)
						}

					}

					var jiraUser jira.User

					for _, User := range atlassianUsers {

						// log.Println(strings.ToLower(slackUser.RealName) + " : " + strings.ToLower(User.DisplayName) + " or " + UserName + " : " + strings.ToLower(User.Name))
						if strings.ToLower(slackUser.RealName) == strings.ToLower(User.DisplayName) || UserName == strings.ToLower(User.Name) {
							jiraUser = User
						}
					}

					if jiraUser.DisplayName == "" {
						postMessage("Пользователь "+slackUser.RealName+" не найден.", client, channelID)
						socketClient.Ack(*event.Request)
						continue
					}

					var issues []jira.Issue
					var message string = ""
					switch eventMessage.Command {

					case "/base":
						issues = jira_int.GetIssuesInBackByAssignee(jiraUser, pageSize)
						if len(issues) == 0 {

							postMessage("У пользователя "+slackUser.RealName+" нет задач в очереди разработки", client, channelID)
							socketClient.Ack(*event.Request)
							continue
						}
						message = "Задачи из бэклога пользователя " + slackUser.RealName + ": \n"
					case "/job":
						issues = jira_int.GetIssuesInJobByAssignee(jiraUser, pageSize)
						if len(issues) == 0 {

							postMessage("У пользователя "+slackUser.RealName+" нет взятых задач в работу", client, channelID)
							socketClient.Ack(*event.Request)
							continue
						}
						message = "Задачи, взятые в работу пользователем " + slackUser.RealName + ": \n"
					}
					var issueLinks []string
					for _, issue := range issues {
						issueLinks = append(issueLinks, "https://ecos.atlassian.net/browse/"+issue.Key)
					}
					log.Println(issueLinks)

					for _, link := range issueLinks {
						message += ("\n" + link)
					}
					postMessage(message, client, channelID)
					socketClient.Ack(*event.Request)
				}
			}
		}
	}(ctx, client, socketClient)

	socketClient.Run()
	ctx.Done()
}

func postMessage(message string, client *slack.Client, channelID string) {
	attachment := slack.Attachment{
		Pretext: message,
	}

	_, timestamp, err := client.PostMessage(
		channelID,
		slack.MsgOptionAttachments(attachment))
	if err != nil {
		panic(err)
	}
	log.Printf("Message sent at %s", timestamp)
}
