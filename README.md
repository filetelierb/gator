# Gator - RSS Feed Aggregator CLI

Gator is a command-line RSS feed aggregator built in Go that allows you to manage RSS feeds, follow feeds, and browse posts from your subscribed feeds. It features user authentication, feed management, and automatic feed scraping with configurable intervals.

## Prerequisites

Before you can run Gator, you'll need to have the following installed on your system:

### Required Software

1. **PostgreSQL** - The application uses PostgreSQL as its database backend
   - Install PostgreSQL from [postgresql.org](https://www.postgresql.org/download/)
   - Make sure PostgreSQL is running and accessible

2. **Go** - The application is written in Go
   - Install Go 1.25.1 or later from [golang.org](https://golang.org/dl/)
   - Verify installation with `go version`

## Installation

### Install the Gator CLI

You can install the Gator CLI using Go's `go install` command:

```bash
go install github.com/filetelierb/gator@latest
```

This will download and install the `gator` binary to your `$GOPATH/bin` directory (or `$GOBIN` if set). Make sure this directory is in your `$PATH` to use the `gator` command from anywhere.

### Verify Installation

After installation, verify that Gator is working:

```bash
gator --help
```

## Configuration

### Database Setup

1. **Create a PostgreSQL database** for Gator:
   ```sql
   CREATE DATABASE gator;
   ```

2. **Run the database migrations** to set up the required tables:
   ```bash
   # Navigate to your project directory
   cd /path/to/gator
   
   # Run migrations (you'll need goose or similar migration tool)
   # The schema files are located in sql/schema/
   ```

### Configuration File

Gator uses a configuration file located at `~/.gatorconfig.json` in your home directory. You'll need to create this file with your database connection details:

```json
{
  "db_url": "postgres://username:password@localhost:5432/gator?sslmode=disable",
  "current_user_name": ""
}
```

**Important:** Replace the `db_url` with your actual PostgreSQL connection string:
- `username`: Your PostgreSQL username
- `password`: Your PostgreSQL password  
- `localhost:5432`: Your PostgreSQL host and port
- `gator`: Your database name
- `sslmode=disable`: Set to `require` for production

## Usage

### Getting Started

1. **Register a new user:**
   ```bash
   gator register your_username
   ```

2. **Login with an existing user:**
   ```bash
   gator login your_username
   ```

3. **View all users:**
   ```bash
   gator users
   ```

### Feed Management

1. **Add a new RSS feed:**
   ```bash
   gator addfeed "Feed Name" "https://example.com/feed.xml"
   ```

2. **List all available feeds:**
   ```bash
   gator feeds
   ```

3. **Follow an existing feed:**
   ```bash
   gator follow "https://example.com/feed.xml"
   ```

4. **View feeds you're following:**
   ```bash
   gator following
   ```

5. **Unfollow a feed:**
   ```bash
   gator unfollow "https://example.com/feed.xml"
   ```

### Content Browsing

1. **Browse recent posts:**
   ```bash
   gator browse
   ```

2. **Browse with custom page size:**
   ```bash
   gator browse 10
   ```

### Feed Aggregation

1. **Start automatic feed scraping:**
   ```bash
   gator agg
   ```

2. **Start with custom interval:**
   ```bash
   gator agg 1h    # Scrape every hour
   gator agg 15s  # Scrape every 15 seconds
   gator agg 1s   # Scrape every second
   ```

### Administrative Commands

1. **Reset user table (clears all users):**
   ```bash
   gator reset
   ```

## Available Commands

| Command | Description | Arguments | Requires Login |
|---------|-------------|-----------|----------------|
| `register` | Create a new user account | username | No |
| `login` | Login as an existing user | username | No |
| `users` | List all users | None | No |
| `addfeed` | Add a new RSS feed | name, url | Yes |
| `feeds` | List all available feeds | None | No |
| `follow` | Follow an existing feed | feed_url | Yes |
| `following` | Show feeds you're following | None | Yes |
| `unfollow` | Unfollow a feed | feed_url | Yes |
| `browse` | Browse recent posts | [page_size] | No |
| `agg` | Start feed aggregation | [interval] | No |
| `reset` | Clear all users (admin) | None | No |

## Project Structure

```
gator/
├── main.go                 # Main application entry point
├── go.mod                  # Go module definition
├── sqlc.yaml              # SQLC configuration
├── internal/
│   ├── config/            # Configuration management
│   ├── database/          # Generated database code (SQLC)
│   └── cmd/               # Command handlers
└── sql/
    ├── schema/            # Database migration files
    └── queries/           # SQL queries for SQLC
```

## Development

### Database Code Generation

This project uses SQLC to generate type-safe Go code from SQL queries. After modifying SQL files, regenerate the code:

```bash
sqlc generate
```

### Dependencies

The project uses the following main dependencies:
- `github.com/lib/pq` - PostgreSQL driver
- `github.com/google/uuid` - UUID generation

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## License

This project is open source. Please check the license file for details.

## Troubleshooting

### Common Issues

1. **Database connection errors**: Verify your PostgreSQL is running and the connection string in `~/.gatorconfig.json` is correct.

2. **Command not found**: Make sure `$GOPATH/bin` or `$GOBIN` is in your `$PATH`.

3. **Permission errors**: Ensure you have write permissions to your home directory for the config file.

4. **Feed parsing errors**: Some RSS feeds may have malformed XML. Check the feed URL manually.

For more help, please open an issue on the project repository.
