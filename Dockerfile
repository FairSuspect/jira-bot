FROM golang:1.18 as builder

ENV GOOS linux
ENV CGO_ENABLED 0

WORKDIR /usr/src/app

COPY go.mod go.sum ./

RUN go mod download && go mod verify
COPY . .
RUN go build -v -o ./dist/jira-work .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY --from=builder /usr/src/app/dist/jira-work /app/jira-work
ENTRYPOINT ["/app/jira-work"]
