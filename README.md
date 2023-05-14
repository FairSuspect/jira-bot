Slack бот, получающий список задач из Jira для Slack пользователей.

Сопоставление пользователей выполняется по их username.

## Инструкция по развёртыванию
 - Создать файл .env со следующими полями:
    + SLACK_AUTH_TOKEN="приватный слак токен авторизации"
    + SLACK_APP_TOKEN="приватный слак токен приложения"
    + SLACK_CHANNEL_ID="channel id(публичный)"
    + JIRA_USER="email пользователя jira"
    + JIRA_TOKEN="приватный api токен пользователя jira"
    + JIRA_URL="jira URL"
 - go install
 - go run main.go

## Команды
 - /job @UserName - поиск задач в статусе "В разработке" исполнителя @UserName
 - /base @UserName - поиск задач в статус "Бэклог" и "Подлежит разработке" исполнителя @UserName
 - /report @UserName - поиск незавершенных задач, созданных пользователем @UserName

