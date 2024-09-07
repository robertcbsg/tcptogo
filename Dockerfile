FROM golang:1.23.0-alpine3.20 AS build

WORKDIR /src
COPY . .
RUN go build -o tcptogo cmd/main.go

# multi-stage build just because
FROM golang:1.23.0-alpine3.20

WORKDIR /src
COPY --from=build /src/tcptogo .

EXPOSE 8080

ENTRYPOINT [ "./tcptogo" ]
