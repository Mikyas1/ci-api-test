# syntax=docker/dockerfile:1

FROM golang:1.23.4-alpine AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /api ./main.go

FROM gcr.io/distroless/static-debian12 AS build-release-stage

WORKDIR /

COPY --from=build-stage /api /api

# Your Go application is trying to bind to port 80, but in ECS Fargate, non-root 
# users cannot bind to ports below 1024. This is because of the USER nonroot:nonroot directive in your Dockerfile.
USER nonroot:nonroot 

ENTRYPOINT ["/api"]