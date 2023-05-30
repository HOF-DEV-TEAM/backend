# Hof Backend - Docker Multistage Build
# Build command: DOCKER_BUILDKIT=1 docker build -t hof_backend -f Dockerfile .
# Run command: docker run -it --rm --name hof_backend hof_backend
#

ARG ALPINE_VERSION=3.15

# BUILD STAGE
FROM golang:1.19.2-alpine${ALPINE_VERSION} AS base

WORKDIR /app

RUN apk add --no-cache jq

COPY go.mod go.sum ./

RUN --mount=type=ssh go mod download

COPY . $WORKDIR

RUN --mount=type=ssh apk add git

RUN export APP_VERSION=$(jq --raw-output '.version' package.json) && \
   export GIT_BRANCH_NAME=$(git log -n1 --decorate=short --pretty="tformat:%f") && \
   export GIT_COMMIT_ID=$(git rev-parse --short HEAD) && \
   export GIT_COMMIT_TIME=$(git --no-pager log -1 --date=format:"%Y-%m-%d|%T" --format="%ad") && \
   export GIT_TAGS=$(git tag --points-at HEAD --format="%(refname:short)," | sort -V | tr -d '\n' | sed 's/%//') && \
   export APP_CREATION_TIME=$(date '+%Y-%m-%d|%H:%M:%S') && \
   go build -ldflags \
    '-X "'$GOPRIVATE'/deploy_pipelines/constants.GitBranchName='$GIT_BRANCH_NAME'" \
     -X "'$GOPRIVATE'/deploy_pipelines/constants.GitCommitID='$GIT_COMMIT_ID'" \
     -X "'$GOPRIVATE'/deploy_pipelines/constants.GitCommitTime='$GIT_COMMIT_TIME'" \
     -X "'$GOPRIVATE'/deploy_pipelines/constants.GitTags='$GIT_TAGS'" \
     -X "'$GOPRIVATE'/deploy_pipelines/constants.AppCreationTime='$APP_CREATION_TIME'" \
     -X "'$GOPRIVATE'/deploy_pipelines/constants.AppName='$APP_NAME'" \
     -X "'$GOPRIVATE'/deploy_pipelines/constants.AppVersion='$APP_VERSION'"' \
   cmd/main.go

# PRODUCTION STAGE
FROM alpine:${ALPINE_VERSION}
ENV WORKDIR /app
RUN apk --no-cache add ca-certificates
COPY --from=base $WORKDIR/main .
COPY --from=base $WORKDIR/migrations ./migrations
COPY --from=base $WORKDIR/templates ./templates
COPY --from=base $WORKDIR/go.mod .
COPY --from=base $WORKDIR/go.sum .

EXPOSE 8080 8082

ENTRYPOINT ["./main"]
