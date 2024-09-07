FROM golang:1.23.0-alpine3.20

WORKDIR /src
COPY . .
RUN go build -o tcptogo cmd/main.go

EXPOSE 8080

ENTRYPOINT [ "./tcptogo" ]
