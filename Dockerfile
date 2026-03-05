FROM golang:1.26 AS development

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# Copy all source files recursively
COPY internal/ ./internal/
COPY cmd/server/ ./cmd/server/

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /docker-gs-ping ./cmd/server/main.go

EXPOSE 8080
CMD [ "/docker-gs-ping" ]