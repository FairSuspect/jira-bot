package slack_int

import "github.com/slack-go/slack"

// Поиск пользователя Slack по имени пользователя
func FindSlackUser(UserName *string, slackUsers *[]slack.User) *slack.User {
	var slackUser slack.User
	for _, User := range *slackUsers {
		if User.Name == *UserName {
			slackUser = User
		}

	}
	return &slackUser
}
