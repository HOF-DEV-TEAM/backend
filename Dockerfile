# Hof Backend - Docker Multistage Build
# Build command: DOCKER_BUILDKIT=1 docker build -t hof_backend -f Dockerfile .
# Run command: docker run -it --rm --name hof_backend hof_backend
#

ARG ALPINE_VERSION=latest

# BUILD STAGE
FROM golang:latest AS build-stage

WORKDIR /app

COPY packages/server/go.mod ./
COPY packages/server/go.sum ./


RUN go mod download

COPY packages/server .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-w" -a -o /main .

# Bundle the admin app
# FROM node:14.15-alpine3.12 AS node_stage
# COPY --from=build-stage /app/packages/admin ./
# RUN npm install
# RUN npm run build

# PRODUCTION STAGE
FROM alpine:${ALPINE_VERSION}
RUN apk --no-cache add ca-certificates
COPY --from=build-stage /main ./
# COPY --from=node_stage /build ./admin
RUN chmod +x ./main
EXPOSE 80 8080 8082
CMD ./main