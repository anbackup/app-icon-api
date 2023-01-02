FROM golang:latest
WORKDIR $GOPATH/src/app-icon-api
ADD . $GOPATH/src/app-icon-api
RUN go build .
EXPOSE 3000
ENTRYPOINT ["./app-icon-api"]