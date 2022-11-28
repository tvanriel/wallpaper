FROM golang:latest

RUN go install github.com/githubnemo/CompileDaemon@latest

CMD bash -c "CompileDaemon -build=\"${BUILD_COMMAND:-go build -o /tmp/build .}\" -command=\"${RUN_COMMAND:-/tmp/build}\" -directory=\"${WATCH_DIR:-.}\""