package main

import (
	"context"
	"errors"
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

	if _, err := os.Stat(".env"); err == nil {
		godotenv.Load(".env")
	}

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
					channelID = eventMessage.ChannelID
					var UserName string
					if len(eventMessage.Text) == 0 {
						UserName = eventMessage.UserName
					} else {
						UserName, err = parseUserNameFromFirstArgument(eventMessage)
						if err != nil {
							postMessage(err.Error(), client, channelID)
							socketClient.Ack(*event.Request)
							continue
						}
						UserName = UserName[1:]
					}
					var slackUser *slack.User
					slackUser = slack_int.FindSlackUser(&UserName, &slackUsers)

					const pageSize int = 1000

					var jiraUser jira.User

					for _, User := range atlassianUsers {

						// log.Println(strings.ToLower(slackUser.RealName) + " : " + strings.ToLower(User.DisplayName) + " or " + UserName + " : " + strings.ToLower(User.Name))
						if strings.ToLower(slackUser.RealName) == strings.ToLower(User.DisplayName) || UserName == strings.ToLower(User.Name) {
							jiraUser = User
						}
					}
					// var senderJiraUser jira.User

					// for _, User := range atlassianUsers {

					// 	// log.Println(strings.ToLower(slackUser.RealName) + " : " + strings.ToLower(User.DisplayName) + " or " + UserName + " : " + strings.ToLower(User.Name))
					// 	if strings.ToLower(slackUser.RealName) == strings.ToLower(User.DisplayName) || UserName == strings.ToLower(User.Name) {
					// 		jiraUser = User
					// 	}
					// }

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

					case "/j":
						issues = jira_int.GetIssuesInJobByAssignee(jiraUser, pageSize)
						if len(issues) == 0 {

							postMessage("У пользователя "+slackUser.RealName+" нет взятых задач в работу", client, channelID)
							socketClient.Ack(*event.Request)
							continue
						}
						message = "Задачи, взятые в работу пользователем " + slackUser.RealName + ": \n"

					case "/report":
						issues = jira_int.GetIssuesByReporter(jiraUser, pageSize)
						if len(issues) == 0 {

							postMessage("Задачи, созданные пользователем "+slackUser.RealName+", не найдены", client, channelID)
							socketClient.Ack(*event.Request)
							continue
						}
						message = "Задачи, созданные пользователем " + slackUser.RealName + " (" + strconv.Itoa(len(issues)) + "): \n"
					case "/help":
						postMessage(helpMessage, client, channelID)
						socketClient.Ack(*event.Request)
						continue
					}

					var issueLinks []string
					for _, issue := range issues {
						// const limit int = 51
						var issueSummary string
						// if len(issue.Fields.Summary) > limit {
						// 	issueSummary = issue.Fields.Summary[0:limit] + "..."
						// } else {

						issueSummary = issue.Fields.Summary
						// }
						issueLinks = append(issueLinks, "\t•\t"+issueSummary+" [<https://ecos.atlassian.net/browse/"+issue.Key+"|"+issue.Key+">]\t-\t["+issue.Fields.Status.Name+"]")
					}
					log.Println(issueLinks)

					for i, link := range issueLinks {
						message += (link + "\n")
						if i%10 == 9 {
							postMessage(message, client, eventMessage.ChannelID)
							message = " "
						}
					}
					postMessage(message, client, eventMessage.ChannelID)
					socketClient.Ack(*event.Request)
				}
			}
		}
	}(ctx, client, socketClient)

	socketClient.Run()
	ctx.Done()
}

func parseUserNameFromFirstArgument(eventMessage slack.SlashCommand) (string, error) {
	Arguments := strings.Split(eventMessage.Text, " ")

	UserName := Arguments[0]
	if UserName[0] != '@' {
		return "", errors.New("Команда работает только с упоминанием пользователя")

	}
	return UserName, nil
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

const helpMessage string = `Команды
- /job @UserName - поиск задач в статусе 'В разработке' исполнителя @UserName
- /base @UserName - поиск задач в статус 'Бэклог' и 'Подлежит разработке' исполнителя @UserName
- /report @UserName - поиск задач, созданных пользователем @UserName`
