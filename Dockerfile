# syntax=docker/dockerfile:1

# Build stage

FROM golang:1.19-alpine AS BuildStage

WORKDIR /src

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . .

EXPOSE 8080

RUN go build -o /backend-hof

# Deploy Stage

FROM alpine:latest

COPY --from=BuildStage /backend-hof /backend-hof

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT [ "/backend-hof" ]