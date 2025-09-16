# Wiki Application using DDAO

A simple wiki application demonstrating how to use DDAO for building web applications with user authentication and content management.

## Features

- **User Registration & Authentication**: Users can register, login, and logout
- **Wiki Page Management**: Create, edit, and view wiki pages
- **Markdown Support**: Write pages in Markdown with live preview
- **Session Management**: Secure session-based authentication
- **Responsive Design**: Mobile-friendly interface
- **SQLite Storage**: Uses DDAO with SQLite backend

## Getting Started

### Prerequisites

- Go 1.21 or later
- Git

### Installation

1. Navigate to the wiki example directory:
```bash
cd examples/wiki
```

2. Install dependencies:
```bash
go mod tidy
```

3. Build the application:
```bash
go build
```

4. Run the wiki server:
```bash
./wiki
```

5. Open your browser and visit: http://localhost:8080

## Usage

### First Time Setup

1. Visit http://localhost:8080
2. Click "Register" to create a new account
3. Fill in username, email, and password
4. Login with your credentials

### Creating Wiki Pages

1. After logging in, enter a page title on the home page
2. Click "Create/Edit Page" to create a new page
3. Write content in Markdown format
4. Use the "Show Preview" button to see how it will look
5. Click "Create Page" to save

### Viewing Pages

1. Enter a page title and click "View Page"
2. If the page exists, you'll see the rendered content
3. If not, you'll be offered the option to create it

### Editing Pages

1. When viewing a page, click "Edit Page"
2. Modify the content in the editor
3. Use "Show Preview" to see changes
4. Click "Update Page" to save

## Project Structure

```
wiki/
├── main.go           # Application entry point
├── models.go         # Data models (User, WikiPage, Session)
├── schema.go         # Database schema definition
├── auth.go           # Authentication service
├── wiki.go           # Wiki page service
├── handlers.go       # HTTP handlers
├── templates/        # HTML templates
│   ├── base.html
│   ├── home.html
│   ├── login.html
│   ├── register.html
│   ├── view_page.html
│   ├── edit_page.html
│   └── page_not_found.html
├── static/           # CSS and static assets
│   └── style.css
├── go.mod
├── go.sum
└── README.md
```

## Technical Details

### Database Schema

The application uses three main tables:

1. **users**: User accounts with authentication
2. **wiki_pages**: Wiki page content and metadata
3. **sessions**: User session management

### DDAO Usage

This example demonstrates key DDAO features:

- **Schema Definition**: Programmatic table schema creation
- **ORM Operations**: Insert, Update, Find, Delete operations
- **Storage Backend**: SQLite with automatic table creation
- **Type Safety**: Strongly typed models with DDAO objects

### Security Features

- **Password Hashing**: Uses bcrypt for secure password storage
- **Session Management**: Time-based session expiration
- **CSRF Protection**: Basic form-based protection
- **Input Validation**: Server-side validation of user inputs

### Frontend Features

- **Markdown Rendering**: Client-side markdown parsing with marked.js
- **Syntax Highlighting**: Code highlighting with highlight.js
- **Live Preview**: Real-time markdown preview while editing
- **Responsive Design**: Mobile-friendly CSS layout

## Customization

### Changing Database Backend

To use a different database (PostgreSQL, CockroachDB, etc.):

1. Import the desired storage package
2. Replace the sqlite storage initialization in `main.go`
3. Update the connection string

Example for PostgreSQL:
```go
import "github.com/jadedragon942/ddao/storage/postgres"

// In main.go
storage := postgres.New()
err := storage.Connect(ctx, "postgres://user:password@localhost/wiki")
```

### Adding Features

The modular structure makes it easy to add features:

- **Page History**: Track page revisions
- **User Permissions**: Role-based access control
- **Search**: Full-text search across pages
- **File Uploads**: Image and document attachments
- **Categories**: Organize pages by topics

### Styling

Customize the appearance by modifying `static/style.css`. The CSS uses:

- CSS Grid and Flexbox for layout
- CSS custom properties for theming
- Responsive design patterns
- Modern typography

## Development

### Running in Development Mode

For development with auto-reload, you can use tools like `air`:

```bash
# Install air
go install github.com/cosmtrek/air@latest

# Run with hot reload
air
```

### Database File

The SQLite database file (`wiki.db`) is created automatically in the current directory. To reset the application:

```bash
rm wiki.db
./wiki
```

## License

This example is part of the DDAO project and follows the same MIT license.

## Contributing

This example demonstrates basic DDAO usage. For improvements or additional features, please contribute to the main DDAO project.