package main

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"strings"
)

// WebServer handles HTTP requests for the LDAP web interface
type WebServer struct {
	ldapServer *LDAPServer
	port       int
}

// NewWebServer creates a new web server instance
func NewWebServer(ldapServer *LDAPServer, port int) *WebServer {
	return &WebServer{
		ldapServer: ldapServer,
		port:       port,
	}
}

// Start starts the web server
func (ws *WebServer) Start() error {
	http.HandleFunc("/", ws.handleHome)
	http.HandleFunc("/users", ws.handleUsers)
	http.HandleFunc("/users/add", ws.handleAddUser)
	http.HandleFunc("/users/delete", ws.handleDeleteUser)
	http.HandleFunc("/entries", ws.handleEntries)
	http.HandleFunc("/entries/add", ws.handleAddEntry)
	http.HandleFunc("/entries/delete", ws.handleDeleteEntry)
	http.HandleFunc("/search", ws.handleSearch)
	http.HandleFunc("/auth", ws.handleAuth)
	http.HandleFunc("/static/", ws.handleStatic)

	fmt.Printf("Starting web server on port %d\n", ws.port)
	fmt.Printf("Open http://localhost:%d in your browser\n", ws.port)

	return http.ListenAndServe(fmt.Sprintf(":%d", ws.port), nil)
}

// handleHome serves the main dashboard
func (ws *WebServer) handleHome(w http.ResponseWriter, r *http.Request) {
	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>LDAP Server Web Interface</title>
    <link rel="stylesheet" href="/static/style.css">
</head>
<body>
    <div class="container">
        <header>
            <h1>LDAP Server Web Interface</h1>
            <nav>
                <a href="/">Dashboard</a>
                <a href="/users">Users</a>
                <a href="/entries">Entries</a>
                <a href="/search">Search</a>
                <a href="/auth">Authentication</a>
            </nav>
        </header>

        <main>
            <div class="dashboard">
                <h2>Dashboard</h2>

                <div class="stats">
                    <div class="stat-card">
                        <h3>Quick Actions</h3>
                        <ul>
                            <li><a href="/users/add">Add New User</a></li>
                            <li><a href="/entries/add">Add New Entry</a></li>
                            <li><a href="/search">Search Entries</a></li>
                            <li><a href="/auth">Test Authentication</a></li>
                        </ul>
                    </div>

                    <div class="stat-card">
                        <h3>Server Information</h3>
                        <p><strong>Base DN:</strong> {{.BaseDN}}</p>
                        <p><strong>Admin DN:</strong> {{.AdminDN}}</p>
                        <p><strong>LDAP Port:</strong> {{.LDAPPort}}</p>
                        <p><strong>Web Port:</strong> {{.WebPort}}</p>
                    </div>
                </div>

                <div class="recent-activity">
                    <h3>Recent Activity</h3>
                    <p>No recent activity data available</p>
                </div>
            </div>
        </main>
    </div>
</body>
</html>`

	t, err := template.New("home").Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		BaseDN   string
		AdminDN  string
		LDAPPort int
		WebPort  int
	}{
		BaseDN:   ws.ldapServer.baseDN,
		AdminDN:  ws.ldapServer.bindDN,
		LDAPPort: ws.ldapServer.port,
		WebPort:  ws.port,
	}

	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, data)
}

// handleUsers manages user operations
func (ws *WebServer) handleUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		ws.showUsers(w, r)
	}
}

func (ws *WebServer) showUsers(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Get all user entries (simplified search)
	users := []LDAPEntry{}

	// Search for person entries
	baseDN := ws.ldapServer.baseDN
	usersOU := fmt.Sprintf("ou=users,%s", baseDN)

	// This is a simplified approach - in a real implementation you'd implement proper hierarchical search
	entry, err := ws.ldapServer.orm.FindByID(ctx, "entries", usersOU)
	if err == nil && entry != nil {
		// For demo purposes, we'll just show the users OU entry
		ldapEntry := LDAPEntry{
			DN:          entry.ID,
			ObjectClass: entry.Fields["object_class"].(string),
			Attributes:  entry.Fields["attributes"].(string),
			CreatedAt:   entry.Fields["created_at"].(string),
		}
		users = append(users, ldapEntry)
	}

	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Users - LDAP Server</title>
    <link rel="stylesheet" href="/static/style.css">
</head>
<body>
    <div class="container">
        <header>
            <h1>LDAP Server Web Interface</h1>
            <nav>
                <a href="/">Dashboard</a>
                <a href="/users" class="active">Users</a>
                <a href="/entries">Entries</a>
                <a href="/search">Search</a>
                <a href="/auth">Authentication</a>
            </nav>
        </header>

        <main>
            <div class="section">
                <div class="section-header">
                    <h2>Users</h2>
                    <a href="/users/add" class="btn btn-primary">Add User</a>
                </div>

                <div class="user-list">
                    {{if .Users}}
                        {{range .Users}}
                        <div class="entry-card">
                            <div class="entry-header">
                                <h3>{{.DN}}</h3>
                                <div class="entry-actions">
                                    <form method="POST" action="/users/delete" style="display: inline;">
                                        <input type="hidden" name="dn" value="{{.DN}}">
                                        <button type="submit" class="btn btn-danger btn-small" onclick="return confirm('Are you sure you want to delete this user?')">Delete</button>
                                    </form>
                                </div>
                            </div>
                            <div class="entry-details">
                                <p><strong>Object Class:</strong> {{.ObjectClass}}</p>
                                <p><strong>Attributes:</strong> {{.Attributes}}</p>
                                <p><strong>Created:</strong> {{.CreatedAt}}</p>
                                {{if .UpdatedAt}}
                                <p><strong>Updated:</strong> {{.UpdatedAt}}</p>
                                {{end}}
                            </div>
                        </div>
                        {{end}}
                    {{else}}
                        <div class="empty-state">
                            <p>No users found. <a href="/users/add">Create your first user</a></p>
                        </div>
                    {{end}}
                </div>
            </div>
        </main>
    </div>
</body>
</html>`

	t, err := template.New("users").Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Users []LDAPEntry
	}{
		Users: users,
	}

	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, data)
}

// handleAddUser shows the add user form and processes user creation
func (ws *WebServer) handleAddUser(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		ws.showAddUserForm(w, r)
	} else if r.Method == "POST" {
		ws.processAddUser(w, r)
	}
}

func (ws *WebServer) showAddUserForm(w http.ResponseWriter, r *http.Request) {
	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Add User - LDAP Server</title>
    <link rel="stylesheet" href="/static/style.css">
</head>
<body>
    <div class="container">
        <header>
            <h1>LDAP Server Web Interface</h1>
            <nav>
                <a href="/">Dashboard</a>
                <a href="/users" class="active">Users</a>
                <a href="/entries">Entries</a>
                <a href="/search">Search</a>
                <a href="/auth">Authentication</a>
            </nav>
        </header>

        <main>
            <div class="section">
                <div class="section-header">
                    <h2>Add New User</h2>
                    <a href="/users" class="btn btn-secondary">Back to Users</a>
                </div>

                <form method="POST" class="form">
                    <div class="form-group">
                        <label for="cn">Common Name (CN):</label>
                        <input type="text" id="cn" name="cn" required placeholder="john.doe">
                    </div>

                    <div class="form-group">
                        <label for="givenName">First Name:</label>
                        <input type="text" id="givenName" name="givenName" required placeholder="John">
                    </div>

                    <div class="form-group">
                        <label for="sn">Last Name:</label>
                        <input type="text" id="sn" name="sn" required placeholder="Doe">
                    </div>

                    <div class="form-group">
                        <label for="mail">Email:</label>
                        <input type="email" id="mail" name="mail" placeholder="john.doe@example.com">
                    </div>

                    <div class="form-group">
                        <label for="userPassword">Password:</label>
                        <input type="password" id="userPassword" name="userPassword" required>
                    </div>

                    <div class="form-group">
                        <label for="objectClass">Object Class:</label>
                        <select id="objectClass" name="objectClass">
                            <option value="person">person</option>
                            <option value="inetOrgPerson">inetOrgPerson</option>
                        </select>
                    </div>

                    <div class="form-actions">
                        <button type="submit" class="btn btn-primary">Create User</button>
                        <a href="/users" class="btn btn-secondary">Cancel</a>
                    </div>
                </form>
            </div>
        </main>
    </div>
</body>
</html>`

	t, err := template.New("addUser").Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, nil)
}

func (ws *WebServer) processAddUser(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	cn := r.FormValue("cn")
	givenName := r.FormValue("givenName")
	sn := r.FormValue("sn")
	mail := r.FormValue("mail")
	password := r.FormValue("userPassword")
	objectClass := r.FormValue("objectClass")

	if cn == "" || givenName == "" || sn == "" || password == "" {
		http.Error(w, "Required fields missing", http.StatusBadRequest)
		return
	}

	// Create DN
	dn := fmt.Sprintf("cn=%s,ou=users,%s", cn, ws.ldapServer.baseDN)

	// Build attributes string
	attributes := fmt.Sprintf("objectClass=%s givenName=%s sn=%s", objectClass, givenName, sn)
	if mail != "" {
		attributes += fmt.Sprintf(" mail=%s", mail)
	}
	attributes += fmt.Sprintf(" userPassword=%s", password)

	// Add the entry
	err := ws.ldapServer.addEntry(dn, attributes)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create user: %v", err), http.StatusInternalServerError)
		return
	}

	// Redirect to users list
	http.Redirect(w, r, "/users", http.StatusSeeOther)
}

// handleDeleteUser processes user deletion
func (ws *WebServer) handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.ParseForm()
	dn := r.FormValue("dn")

	if dn == "" {
		http.Error(w, "DN required", http.StatusBadRequest)
		return
	}

	err := ws.ldapServer.deleteEntry(dn)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete user: %v", err), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/users", http.StatusSeeOther)
}

// handleEntries manages all LDAP entries
func (ws *WebServer) handleEntries(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Get all entries (simplified)
	entries := []LDAPEntry{}

	// For demo, try to get the base DN and a few common entries
	commonDNs := []string{
		ws.ldapServer.baseDN,
		fmt.Sprintf("ou=users,%s", ws.ldapServer.baseDN),
		fmt.Sprintf("ou=groups,%s", ws.ldapServer.baseDN),
		ws.ldapServer.bindDN,
	}

	for _, dn := range commonDNs {
		entry, err := ws.ldapServer.orm.FindByID(ctx, "entries", dn)
		if err == nil && entry != nil {
			ldapEntry := LDAPEntry{
				DN:          entry.ID,
				ObjectClass: entry.Fields["object_class"].(string),
				Attributes:  entry.Fields["attributes"].(string),
				CreatedAt:   entry.Fields["created_at"].(string),
			}
			if parentDN, exists := entry.Fields["parent_dn"]; exists && parentDN != nil {
				ldapEntry.ParentDN = parentDN.(string)
			}
			if updatedAt, exists := entry.Fields["updated_at"]; exists && updatedAt != nil {
				ldapEntry.UpdatedAt = updatedAt.(string)
			}
			entries = append(entries, ldapEntry)
		}
	}

	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Entries - LDAP Server</title>
    <link rel="stylesheet" href="/static/style.css">
</head>
<body>
    <div class="container">
        <header>
            <h1>LDAP Server Web Interface</h1>
            <nav>
                <a href="/">Dashboard</a>
                <a href="/users">Users</a>
                <a href="/entries" class="active">Entries</a>
                <a href="/search">Search</a>
                <a href="/auth">Authentication</a>
            </nav>
        </header>

        <main>
            <div class="section">
                <div class="section-header">
                    <h2>LDAP Entries</h2>
                    <a href="/entries/add" class="btn btn-primary">Add Entry</a>
                </div>

                <div class="entry-list">
                    {{if .Entries}}
                        {{range .Entries}}
                        <div class="entry-card">
                            <div class="entry-header">
                                <h3>{{.DN}}</h3>
                                <div class="entry-actions">
                                    <form method="POST" action="/entries/delete" style="display: inline;">
                                        <input type="hidden" name="dn" value="{{.DN}}">
                                        <button type="submit" class="btn btn-danger btn-small" onclick="return confirm('Are you sure you want to delete this entry?')">Delete</button>
                                    </form>
                                </div>
                            </div>
                            <div class="entry-details">
                                {{if .ParentDN}}
                                <p><strong>Parent DN:</strong> {{.ParentDN}}</p>
                                {{end}}
                                <p><strong>Object Class:</strong> {{.ObjectClass}}</p>
                                <p><strong>Attributes:</strong> {{.Attributes}}</p>
                                <p><strong>Created:</strong> {{.CreatedAt}}</p>
                                {{if .UpdatedAt}}
                                <p><strong>Updated:</strong> {{.UpdatedAt}}</p>
                                {{end}}
                            </div>
                        </div>
                        {{end}}
                    {{else}}
                        <div class="empty-state">
                            <p>No entries found. <a href="/entries/add">Create your first entry</a></p>
                        </div>
                    {{end}}
                </div>
            </div>
        </main>
    </div>
</body>
</html>`

	t, err := template.New("entries").Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Entries []LDAPEntry
	}{
		Entries: entries,
	}

	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, data)
}

// handleAddEntry shows the add entry form and processes entry creation
func (ws *WebServer) handleAddEntry(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		ws.showAddEntryForm(w, r)
	} else if r.Method == "POST" {
		ws.processAddEntry(w, r)
	}
}

func (ws *WebServer) showAddEntryForm(w http.ResponseWriter, r *http.Request) {
	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Add Entry - LDAP Server</title>
    <link rel="stylesheet" href="/static/style.css">
</head>
<body>
    <div class="container">
        <header>
            <h1>LDAP Server Web Interface</h1>
            <nav>
                <a href="/">Dashboard</a>
                <a href="/users">Users</a>
                <a href="/entries" class="active">Entries</a>
                <a href="/search">Search</a>
                <a href="/auth">Authentication</a>
            </nav>
        </header>

        <main>
            <div class="section">
                <div class="section-header">
                    <h2>Add New Entry</h2>
                    <a href="/entries" class="btn btn-secondary">Back to Entries</a>
                </div>

                <form method="POST" class="form">
                    <div class="form-group">
                        <label for="dn">Distinguished Name (DN):</label>
                        <input type="text" id="dn" name="dn" required placeholder="cn=example,ou=units,{{.BaseDN}}">
                        <small>Full DN for the new entry</small>
                    </div>

                    <div class="form-group">
                        <label for="objectClass">Object Class:</label>
                        <select id="objectClass" name="objectClass">
                            <option value="organizationalUnit">organizationalUnit</option>
                            <option value="person">person</option>
                            <option value="inetOrgPerson">inetOrgPerson</option>
                            <option value="group">group</option>
                            <option value="organization">organization</option>
                        </select>
                    </div>

                    <div class="form-group">
                        <label for="attributes">Attributes (key=value format, space separated):</label>
                        <textarea id="attributes" name="attributes" rows="4" placeholder="givenName=John sn=Doe mail=john@example.com"></textarea>
                        <small>Example: givenName=John sn=Doe mail=john@example.com</small>
                    </div>

                    <div class="form-actions">
                        <button type="submit" class="btn btn-primary">Create Entry</button>
                        <a href="/entries" class="btn btn-secondary">Cancel</a>
                    </div>
                </form>
            </div>
        </main>
    </div>
</body>
</html>`

	t, err := template.New("addEntry").Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		BaseDN string
	}{
		BaseDN: ws.ldapServer.baseDN,
	}

	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, data)
}

func (ws *WebServer) processAddEntry(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	dn := r.FormValue("dn")
	objectClass := r.FormValue("objectClass")
	attributes := r.FormValue("attributes")

	if dn == "" || objectClass == "" {
		http.Error(w, "DN and object class are required", http.StatusBadRequest)
		return
	}

	// Prepend object class to attributes
	fullAttributes := fmt.Sprintf("objectClass=%s %s", objectClass, attributes)

	err := ws.ldapServer.addEntry(dn, fullAttributes)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create entry: %v", err), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/entries", http.StatusSeeOther)
}

// handleDeleteEntry processes entry deletion
func (ws *WebServer) handleDeleteEntry(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.ParseForm()
	dn := r.FormValue("dn")

	if dn == "" {
		http.Error(w, "DN required", http.StatusBadRequest)
		return
	}

	err := ws.ldapServer.deleteEntry(dn)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete entry: %v", err), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/entries", http.StatusSeeOther)
}

// handleSearch provides search functionality
func (ws *WebServer) handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		ws.showSearchForm(w, r)
	} else if r.Method == "POST" {
		ws.processSearch(w, r)
	}
}

func (ws *WebServer) showSearchForm(w http.ResponseWriter, r *http.Request) {
	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Search - LDAP Server</title>
    <link rel="stylesheet" href="/static/style.css">
</head>
<body>
    <div class="container">
        <header>
            <h1>LDAP Server Web Interface</h1>
            <nav>
                <a href="/">Dashboard</a>
                <a href="/users">Users</a>
                <a href="/entries">Entries</a>
                <a href="/search" class="active">Search</a>
                <a href="/auth">Authentication</a>
            </nav>
        </header>

        <main>
            <div class="section">
                <h2>Search LDAP Entries</h2>

                <form method="POST" class="form">
                    <div class="form-group">
                        <label for="baseDN">Base DN:</label>
                        <input type="text" id="baseDN" name="baseDN" value="{{.BaseDN}}" required>
                    </div>

                    <div class="form-group">
                        <label for="filter">Filter (optional):</label>
                        <input type="text" id="filter" name="filter" placeholder="objectClass=person">
                        <small>Leave empty to search all entries under base DN</small>
                    </div>

                    <div class="form-actions">
                        <button type="submit" class="btn btn-primary">Search</button>
                    </div>
                </form>

                {{if .Results}}
                <div class="search-results">
                    <h3>Search Results</h3>
                    {{range .Results}}
                    <div class="entry-card">
                        <div class="entry-header">
                            <h4>{{.DN}}</h4>
                        </div>
                        <div class="entry-details">
                            {{if .ParentDN}}
                            <p><strong>Parent DN:</strong> {{.ParentDN}}</p>
                            {{end}}
                            <p><strong>Object Class:</strong> {{.ObjectClass}}</p>
                            <p><strong>Attributes:</strong> {{.Attributes}}</p>
                            <p><strong>Created:</strong> {{.CreatedAt}}</p>
                        </div>
                    </div>
                    {{end}}
                </div>
                {{end}}

                {{if .Error}}
                <div class="error">
                    <p>Error: {{.Error}}</p>
                </div>
                {{end}}
            </div>
        </main>
    </div>
</body>
</html>`

	t, err := template.New("search").Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		BaseDN  string
		Results []LDAPEntry
		Error   string
	}{
		BaseDN: ws.ldapServer.baseDN,
	}

	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, data)
}

func (ws *WebServer) processSearch(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	baseDN := r.FormValue("baseDN")
	filter := r.FormValue("filter")

	results, err := ws.ldapServer.searchEntries(baseDN, filter)

	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Search - LDAP Server</title>
    <link rel="stylesheet" href="/static/style.css">
</head>
<body>
    <div class="container">
        <header>
            <h1>LDAP Server Web Interface</h1>
            <nav>
                <a href="/">Dashboard</a>
                <a href="/users">Users</a>
                <a href="/entries">Entries</a>
                <a href="/search" class="active">Search</a>
                <a href="/auth">Authentication</a>
            </nav>
        </header>

        <main>
            <div class="section">
                <h2>Search LDAP Entries</h2>

                <form method="POST" class="form">
                    <div class="form-group">
                        <label for="baseDN">Base DN:</label>
                        <input type="text" id="baseDN" name="baseDN" value="{{.BaseDN}}" required>
                    </div>

                    <div class="form-group">
                        <label for="filter">Filter (optional):</label>
                        <input type="text" id="filter" name="filter" value="{{.Filter}}" placeholder="objectClass=person">
                        <small>Leave empty to search all entries under base DN</small>
                    </div>

                    <div class="form-actions">
                        <button type="submit" class="btn btn-primary">Search</button>
                    </div>
                </form>

                {{if .Results}}
                <div class="search-results">
                    <h3>Search Results ({{len .Results}} found)</h3>
                    {{range .Results}}
                    <div class="entry-card">
                        <div class="entry-header">
                            <h4>{{.DN}}</h4>
                        </div>
                        <div class="entry-details">
                            {{if .ParentDN}}
                            <p><strong>Parent DN:</strong> {{.ParentDN}}</p>
                            {{end}}
                            <p><strong>Object Class:</strong> {{.ObjectClass}}</p>
                            <p><strong>Attributes:</strong> {{.Attributes}}</p>
                            <p><strong>Created:</strong> {{.CreatedAt}}</p>
                        </div>
                    </div>
                    {{end}}
                </div>
                {{else}}
                    {{if not .Error}}
                    <div class="empty-state">
                        <p>No entries found matching your search criteria.</p>
                    </div>
                    {{end}}
                {{end}}

                {{if .Error}}
                <div class="error">
                    <p>Error: {{.Error}}</p>
                </div>
                {{end}}
            </div>
        </main>
    </div>
</body>
</html>`

	t, err := template.New("searchResults").Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		BaseDN  string
		Filter  string
		Results []LDAPEntry
		Error   string
	}{
		BaseDN: baseDN,
		Filter: filter,
	}

	if err != nil {
		data.Error = err.Error()
	} else {
		data.Results = results
	}

	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, data)
}

// handleAuth provides authentication testing
func (ws *WebServer) handleAuth(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		ws.showAuthForm(w, r)
	} else if r.Method == "POST" {
		ws.processAuth(w, r)
	}
}

func (ws *WebServer) showAuthForm(w http.ResponseWriter, r *http.Request) {
	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Authentication - LDAP Server</title>
    <link rel="stylesheet" href="/static/style.css">
</head>
<body>
    <div class="container">
        <header>
            <h1>LDAP Server Web Interface</h1>
            <nav>
                <a href="/">Dashboard</a>
                <a href="/users">Users</a>
                <a href="/entries">Entries</a>
                <a href="/search">Search</a>
                <a href="/auth" class="active">Authentication</a>
            </nav>
        </header>

        <main>
            <div class="section">
                <h2>Test Authentication</h2>

                <form method="POST" class="form">
                    <div class="form-group">
                        <label for="dn">Distinguished Name (DN):</label>
                        <input type="text" id="dn" name="dn" value="{{.AdminDN}}" required placeholder="cn=admin,dc=example,dc=com">
                    </div>

                    <div class="form-group">
                        <label for="password">Password:</label>
                        <input type="password" id="password" name="password" required>
                    </div>

                    <div class="form-actions">
                        <button type="submit" class="btn btn-primary">Test Authentication</button>
                    </div>
                </form>

                {{if .Result}}
                <div class="auth-result {{if .Success}}success{{else}}error{{end}}">
                    <h3>Authentication Result</h3>
                    <p>{{.Result}}</p>
                </div>
                {{end}}
            </div>
        </main>
    </div>
</body>
</html>`

	t, err := template.New("auth").Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		AdminDN string
		Result  string
		Success bool
	}{
		AdminDN: ws.ldapServer.bindDN,
	}

	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, data)
}

func (ws *WebServer) processAuth(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	dn := r.FormValue("dn")
	password := r.FormValue("password")

	result := ws.ldapServer.handleBind(dn, password)
	success := strings.Contains(result, "OK:")

	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Authentication - LDAP Server</title>
    <link rel="stylesheet" href="/static/style.css">
</head>
<body>
    <div class="container">
        <header>
            <h1>LDAP Server Web Interface</h1>
            <nav>
                <a href="/">Dashboard</a>
                <a href="/users">Users</a>
                <a href="/entries">Entries</a>
                <a href="/search">Search</a>
                <a href="/auth" class="active">Authentication</a>
            </nav>
        </header>

        <main>
            <div class="section">
                <h2>Test Authentication</h2>

                <form method="POST" class="form">
                    <div class="form-group">
                        <label for="dn">Distinguished Name (DN):</label>
                        <input type="text" id="dn" name="dn" value="{{.DN}}" required placeholder="cn=admin,dc=example,dc=com">
                    </div>

                    <div class="form-group">
                        <label for="password">Password:</label>
                        <input type="password" id="password" name="password" required>
                    </div>

                    <div class="form-actions">
                        <button type="submit" class="btn btn-primary">Test Authentication</button>
                    </div>
                </form>

                <div class="auth-result {{if .Success}}success{{else}}error{{end}}">
                    <h3>Authentication Result</h3>
                    <p>{{.Result}}</p>
                </div>
            </div>
        </main>
    </div>
</body>
</html>`

	t, err := template.New("authResult").Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		DN      string
		Result  string
		Success bool
	}{
		DN:      dn,
		Result:  strings.TrimSpace(result),
		Success: success,
	}

	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, data)
}

// handleStatic serves static CSS files
func (ws *WebServer) handleStatic(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/static/style.css" {
		w.Header().Set("Content-Type", "text/css")
		w.Write([]byte(getCSS()))
		return
	}
	http.NotFound(w, r)
}

// getCSS returns the CSS styling for the web interface
func getCSS() string {
	return `
/* Reset and base styles */
* {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
}

body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
    line-height: 1.6;
    color: #333;
    background-color: #f5f5f5;
}

.container {
    max-width: 1200px;
    margin: 0 auto;
    background: white;
    min-height: 100vh;
    box-shadow: 0 0 20px rgba(0,0,0,0.1);
}

/* Header and Navigation */
header {
    background: #2c3e50;
    color: white;
    padding: 1rem 2rem;
    border-bottom: 3px solid #3498db;
}

header h1 {
    margin-bottom: 1rem;
    font-size: 1.8rem;
}

nav {
    display: flex;
    gap: 2rem;
}

nav a {
    color: #ecf0f1;
    text-decoration: none;
    padding: 0.5rem 1rem;
    border-radius: 4px;
    transition: background-color 0.3s;
}

nav a:hover {
    background-color: #34495e;
}

nav a.active {
    background-color: #3498db;
    color: white;
}

/* Main content */
main {
    padding: 2rem;
}

.section {
    margin-bottom: 2rem;
}

.section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1.5rem;
    padding-bottom: 1rem;
    border-bottom: 2px solid #ecf0f1;
}

.section-header h2 {
    color: #2c3e50;
    font-size: 1.5rem;
}

/* Dashboard */
.dashboard {
    max-width: 100%;
}

.stats {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
    gap: 1.5rem;
    margin-bottom: 2rem;
}

.stat-card {
    background: white;
    padding: 1.5rem;
    border-radius: 8px;
    border: 1px solid #ddd;
    box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

.stat-card h3 {
    color: #2c3e50;
    margin-bottom: 1rem;
    font-size: 1.2rem;
}

.stat-card ul {
    list-style: none;
}

.stat-card li {
    margin-bottom: 0.5rem;
}

.stat-card a {
    color: #3498db;
    text-decoration: none;
}

.stat-card a:hover {
    text-decoration: underline;
}

.recent-activity {
    background: white;
    padding: 1.5rem;
    border-radius: 8px;
    border: 1px solid #ddd;
    box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

.recent-activity h3 {
    color: #2c3e50;
    margin-bottom: 1rem;
}

/* Buttons */
.btn {
    display: inline-block;
    padding: 0.75rem 1.5rem;
    background: #3498db;
    color: white;
    text-decoration: none;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    font-size: 0.9rem;
    transition: background-color 0.3s;
}

.btn:hover {
    background: #2980b9;
}

.btn-primary {
    background: #3498db;
}

.btn-primary:hover {
    background: #2980b9;
}

.btn-secondary {
    background: #95a5a6;
}

.btn-secondary:hover {
    background: #7f8c8d;
}

.btn-danger {
    background: #e74c3c;
}

.btn-danger:hover {
    background: #c0392b;
}

.btn-small {
    padding: 0.5rem 1rem;
    font-size: 0.8rem;
}

/* Forms */
.form {
    background: white;
    padding: 2rem;
    border-radius: 8px;
    border: 1px solid #ddd;
    box-shadow: 0 2px 4px rgba(0,0,0,0.1);
    max-width: 600px;
}

.form-group {
    margin-bottom: 1.5rem;
}

.form-group label {
    display: block;
    margin-bottom: 0.5rem;
    font-weight: 600;
    color: #2c3e50;
}

.form-group input,
.form-group select,
.form-group textarea {
    width: 100%;
    padding: 0.75rem;
    border: 1px solid #ddd;
    border-radius: 4px;
    font-size: 1rem;
}

.form-group input:focus,
.form-group select:focus,
.form-group textarea:focus {
    outline: none;
    border-color: #3498db;
    box-shadow: 0 0 0 2px rgba(52, 152, 219, 0.2);
}

.form-group small {
    display: block;
    margin-top: 0.25rem;
    color: #7f8c8d;
    font-size: 0.9rem;
}

.form-actions {
    display: flex;
    gap: 1rem;
    margin-top: 2rem;
}

/* Entry cards */
.entry-list,
.user-list {
    display: grid;
    gap: 1rem;
}

.entry-card {
    background: white;
    border: 1px solid #ddd;
    border-radius: 8px;
    padding: 1.5rem;
    box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

.entry-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 1rem;
    padding-bottom: 1rem;
    border-bottom: 1px solid #ecf0f1;
}

.entry-header h3,
.entry-header h4 {
    color: #2c3e50;
    font-size: 1.1rem;
    margin: 0;
    word-break: break-all;
}

.entry-actions {
    display: flex;
    gap: 0.5rem;
}

.entry-details p {
    margin-bottom: 0.5rem;
    font-size: 0.9rem;
}

.entry-details strong {
    color: #2c3e50;
}

/* Search results */
.search-results {
    margin-top: 2rem;
}

.search-results h3 {
    color: #2c3e50;
    margin-bottom: 1rem;
}

/* Messages */
.success {
    background: #d4edda;
    color: #155724;
    padding: 1rem;
    border-radius: 4px;
    border: 1px solid #c3e6cb;
    margin-top: 1rem;
}

.error {
    background: #f8d7da;
    color: #721c24;
    padding: 1rem;
    border-radius: 4px;
    border: 1px solid #f5c6cb;
    margin-top: 1rem;
}

.auth-result {
    margin-top: 1.5rem;
    padding: 1.5rem;
    border-radius: 8px;
    border: 1px solid;
}

.auth-result h3 {
    margin-bottom: 0.5rem;
}

.empty-state {
    text-align: center;
    padding: 3rem;
    color: #7f8c8d;
}

.empty-state a {
    color: #3498db;
    text-decoration: none;
}

.empty-state a:hover {
    text-decoration: underline;
}

/* Responsive design */
@media (max-width: 768px) {
    .container {
        margin: 0;
        box-shadow: none;
    }

    header {
        padding: 1rem;
    }

    nav {
        flex-wrap: wrap;
        gap: 1rem;
    }

    main {
        padding: 1rem;
    }

    .section-header {
        flex-direction: column;
        align-items: flex-start;
        gap: 1rem;
    }

    .stats {
        grid-template-columns: 1fr;
    }

    .form {
        padding: 1.5rem;
    }

    .form-actions {
        flex-direction: column;
    }

    .entry-header {
        flex-direction: column;
        gap: 1rem;
    }

    .entry-actions {
        align-self: flex-start;
    }
}
`
}