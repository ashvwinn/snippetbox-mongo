# SnippetBox
SnippetBox is a simple web application for storing and sharing text snippets on the web built with Go and MongoDB.


# Installation
1. Clone the repository:
```bash
git clone https://github.com/ashvwinn/snippetbox-mongo.git
cd snippetbox
```

2. Set environment variables (look at .env.example for reference)

3. Setup database with Docker Compose:
```bash
docker compose up -d
```

4. Start the application:
With `go` or with [`air`](https://github.com/air-verse/air)
```bash
air
# or
go run ./cmd/web
# or
go build -o tmp/main ./cmd/web
./tmp/main
```
