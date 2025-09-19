package main

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/jadedragon942/ddao/schema"
)

type WebServer struct {
	adminServer *AdminServer
	port        int
}

type TableInfo struct {
	Name   string
	Fields []FieldInfo
}

type FieldInfo struct {
	Name     string
	DataType string
	Nullable bool
}

type ObjectData struct {
	ID     string
	Fields map[string]interface{}
}

func NewWebServer(adminServer *AdminServer, port int) *WebServer {
	return &WebServer{
		adminServer: adminServer,
		port:        port,
	}
}

func (ws *WebServer) Start() error {
	http.HandleFunc("/", ws.handleHome)
	http.HandleFunc("/tables", ws.handleTables)
	http.HandleFunc("/table/", ws.handleTable)
	http.HandleFunc("/alter", ws.handleAlterTable)
	http.HandleFunc("/insert", ws.handleInsert)
	http.HandleFunc("/delete", ws.handleDelete)
	http.HandleFunc("/static/", ws.handleStatic)

	fmt.Printf("Open http://localhost:%d in your browser\n", ws.port)
	return http.ListenAndServe(fmt.Sprintf(":%d", ws.port), nil)
}

func (ws *WebServer) handleHome(w http.ResponseWriter, r *http.Request) {
	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Database Administration Tool</title>
    <link rel="stylesheet" href="/static/style.css">
</head>
<body>
    <div class="container">
        <header>
            <h1>Database Administration Tool</h1>
            <nav>
                <a href="/" class="active">Dashboard</a>
                <a href="/tables">Tables</a>
                <a href="/alter">Alter Table</a>
            </nav>
        </header>

        <main>
            <div class="dashboard">
                <h2>Dashboard</h2>

                <div class="stats">
                    <div class="stat-card">
                        <h3>Storage Information</h3>
                        <p><strong>Type:</strong> {{.StorageType}}</p>
                        <p><strong>Connection:</strong> {{.ConnString}}</p>
                        <p><strong>Port:</strong> {{.Port}}</p>
                    </div>

                    <div class="stat-card">
                        <h3>Quick Actions</h3>
                        <ul>
                            <li><a href="/tables">View Tables</a></li>
                            <li><a href="/alter">Alter Table</a></li>
                        </ul>
                    </div>
                </div>

                <div class="recent-activity">
                    <h3>Available Tables</h3>
                    {{if .Tables}}
                        <ul>
                            {{range .Tables}}
                            <li><a href="/table/{{.Name}}">{{.Name}}</a></li>
                            {{end}}
                        </ul>
                    {{else}}
                        <p>No tables available</p>
                    {{end}}
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

	tables := ws.getTableList()

	data := struct {
		StorageType string
		ConnString  string
		Port        int
		Tables      []TableInfo
	}{
		StorageType: ws.adminServer.config.StorageType,
		ConnString:  ws.adminServer.config.ConnString,
		Port:        ws.adminServer.config.Port,
		Tables:      tables,
	}

	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, data)
}

func (ws *WebServer) handleTables(w http.ResponseWriter, r *http.Request) {
	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Tables - Database Administration Tool</title>
    <link rel="stylesheet" href="/static/style.css">
</head>
<body>
    <div class="container">
        <header>
            <h1>Database Administration Tool</h1>
            <nav>
                <a href="/">Dashboard</a>
                <a href="/tables" class="active">Tables</a>
                <a href="/alter">Alter Table</a>
            </nav>
        </header>

        <main>
            <div class="section">
                <div class="section-header">
                    <h2>Database Tables</h2>
                </div>

                <div class="table-list">
                    {{if .Tables}}
                        {{range .Tables}}
                        <div class="table-card">
                            <div class="table-header">
                                <h3><a href="/table/{{.Name}}">{{.Name}}</a></h3>
                            </div>
                            <div class="table-details">
                                <h4>Fields:</h4>
                                <ul>
                                    {{range .Fields}}
                                    <li><strong>{{.Name}}</strong> ({{.DataType}}{{if .Nullable}}, NULL{{else}}, NOT NULL{{end}})</li>
                                    {{end}}
                                </ul>
                            </div>
                        </div>
                        {{end}}
                    {{else}}
                        <div class="empty-state">
                            <p>No tables found in the database.</p>
                        </div>
                    {{end}}
                </div>
            </div>
        </main>
    </div>
</body>
</html>`

	t, err := template.New("tables").Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tables := ws.getTableList()

	data := struct {
		Tables []TableInfo
	}{
		Tables: tables,
	}

	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, data)
}

func (ws *WebServer) handleTable(w http.ResponseWriter, r *http.Request) {
	tableName := strings.TrimPrefix(r.URL.Path, "/table/")
	if tableName == "" {
		http.Error(w, "Table name required", http.StatusBadRequest)
		return
	}

	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.TableName}} - Database Administration Tool</title>
    <link rel="stylesheet" href="/static/style.css">
</head>
<body>
    <div class="container">
        <header>
            <h1>Database Administration Tool</h1>
            <nav>
                <a href="/">Dashboard</a>
                <a href="/tables">Tables</a>
                <a href="/alter">Alter Table</a>
            </nav>
        </header>

        <main>
            <div class="section">
                <div class="section-header">
                    <h2>Table: {{.TableName}}</h2>
                    <a href="/insert?table={{.TableName}}" class="btn btn-primary">Add Record</a>
                </div>

                <div class="table-schema">
                    <h3>Schema</h3>
                    <table class="data-table">
                        <thead>
                            <tr>
                                <th>Field Name</th>
                                <th>Data Type</th>
                                <th>Nullable</th>
                            </tr>
                        </thead>
                        <tbody>
                            {{range .Fields}}
                            <tr>
                                <td>{{.Name}}</td>
                                <td>{{.DataType}}</td>
                                <td>{{if .Nullable}}Yes{{else}}No{{end}}</td>
                            </tr>
                            {{end}}
                        </tbody>
                    </table>
                </div>

                {{if .Error}}
                <div class="error">
                    <p>Error loading data: {{.Error}}</p>
                </div>
                {{else if .Data}}
                <div class="table-data">
                    <h3>Data (Sample)</h3>
                    <table class="data-table">
                        <thead>
                            <tr>
                                <th>ID</th>
                                {{range .Fields}}
                                <th>{{.Name}}</th>
                                {{end}}
                                <th>Actions</th>
                            </tr>
                        </thead>
                        <tbody>
                            {{range .Data}}
                            <tr>
                                <td>{{.ID}}</td>
                                {{range $.Fields}}
                                <td>{{index $.FieldData .Name}}</td>
                                {{end}}
                                <td>
                                    <form method="POST" action="/delete" style="display: inline;">
                                        <input type="hidden" name="table" value="{{$.TableName}}">
                                        <input type="hidden" name="id" value="{{.ID}}">
                                        <button type="submit" class="btn btn-danger btn-small" onclick="return confirm('Are you sure?')">Delete</button>
                                    </form>
                                </td>
                            </tr>
                            {{end}}
                        </tbody>
                    </table>
                </div>
                {{else}}
                <div class="empty-state">
                    <p>No data found in this table. <a href="/insert?table={{.TableName}}">Add the first record</a></p>
                </div>
                {{end}}
            </div>
        </main>
    </div>
</body>
</html>`

	t, err := template.New("table").Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tableInfo := ws.getTableInfo(tableName)
	if tableInfo == nil {
		http.Error(w, "Table not found", http.StatusNotFound)
		return
	}

	data := struct {
		TableName string
		Fields    []FieldInfo
		Data      []ObjectData
		Error     string
	}{
		TableName: tableName,
		Fields:    tableInfo.Fields,
	}

	// Try to get some sample data, but don't fail if it doesn't work
	// since some storage types might not support this operation
	// For now, we'll just show the schema

	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, data)
}

func (ws *WebServer) handleAlterTable(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		ws.showAlterTableForm(w, r)
	} else if r.Method == "POST" {
		ws.processAlterTable(w, r)
	}
}

func (ws *WebServer) showAlterTableForm(w http.ResponseWriter, r *http.Request) {
	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Alter Table - Database Administration Tool</title>
    <link rel="stylesheet" href="/static/style.css">
</head>
<body>
    <div class="container">
        <header>
            <h1>Database Administration Tool</h1>
            <nav>
                <a href="/">Dashboard</a>
                <a href="/tables">Tables</a>
                <a href="/alter" class="active">Alter Table</a>
            </nav>
        </header>

        <main>
            <div class="section">
                <div class="section-header">
                    <h2>Alter Table - Add Column</h2>
                    <a href="/tables" class="btn btn-secondary">Back to Tables</a>
                </div>

                {{if .Message}}
                <div class="{{if .Success}}success{{else}}error{{end}}">
                    <p>{{.Message}}</p>
                </div>
                {{end}}

                <form method="POST" class="form">
                    <div class="form-group">
                        <label for="tableName">Table Name:</label>
                        <select id="tableName" name="tableName" required>
                            <option value="">Select a table...</option>
                            {{range .Tables}}
                            <option value="{{.Name}}">{{.Name}}</option>
                            {{end}}
                        </select>
                    </div>

                    <div class="form-group">
                        <label for="columnName">New Column Name:</label>
                        <input type="text" id="columnName" name="columnName" required placeholder="new_column">
                    </div>

                    <div class="form-group">
                        <label for="dataType">Data Type:</label>
                        <select id="dataType" name="dataType" required>
                            <option value="TEXT">TEXT</option>
                            <option value="INTEGER">INTEGER</option>
                            <option value="REAL">REAL</option>
                            <option value="BOOLEAN">BOOLEAN</option>
                            <option value="DATETIME">DATETIME</option>
                            <option value="JSON">JSON</option>
                            <option value="BLOB">BLOB</option>
                        </select>
                    </div>

                    <div class="form-group">
                        <label>
                            <input type="checkbox" name="nullable" value="true">
                            Allow NULL values
                        </label>
                    </div>

                    <div class="form-actions">
                        <button type="submit" class="btn btn-primary">Add Column</button>
                        <a href="/tables" class="btn btn-secondary">Cancel</a>
                    </div>
                </form>
            </div>
        </main>
    </div>
</body>
</html>`

	t, err := template.New("alterTable").Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tables := ws.getTableList()

	data := struct {
		Tables  []TableInfo
		Message string
		Success bool
	}{
		Tables: tables,
	}

	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, data)
}

func (ws *WebServer) processAlterTable(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	tableName := r.FormValue("tableName")
	columnName := r.FormValue("columnName")
	dataType := r.FormValue("dataType")
	nullable := r.FormValue("nullable") == "true"

	if tableName == "" || columnName == "" || dataType == "" {
		ws.showAlterTableFormWithMessage(w, "All fields are required", false)
		return
	}

	ctx := context.Background()
	err := ws.adminServer.storage.AlterTable(ctx, tableName, columnName, dataType, nullable)
	if err != nil {
		ws.showAlterTableFormWithMessage(w, fmt.Sprintf("Failed to alter table: %v", err), false)
		return
	}

	ws.showAlterTableFormWithMessage(w, fmt.Sprintf("Successfully added column '%s' to table '%s'", columnName, tableName), true)
}

func (ws *WebServer) showAlterTableFormWithMessage(w http.ResponseWriter, message string, success bool) {
	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Alter Table - Database Administration Tool</title>
    <link rel="stylesheet" href="/static/style.css">
</head>
<body>
    <div class="container">
        <header>
            <h1>Database Administration Tool</h1>
            <nav>
                <a href="/">Dashboard</a>
                <a href="/tables">Tables</a>
                <a href="/alter" class="active">Alter Table</a>
            </nav>
        </header>

        <main>
            <div class="section">
                <div class="section-header">
                    <h2>Alter Table - Add Column</h2>
                    <a href="/tables" class="btn btn-secondary">Back to Tables</a>
                </div>

                <div class="{{if .Success}}success{{else}}error{{end}}">
                    <p>{{.Message}}</p>
                </div>

                <form method="POST" class="form">
                    <div class="form-group">
                        <label for="tableName">Table Name:</label>
                        <select id="tableName" name="tableName" required>
                            <option value="">Select a table...</option>
                            {{range .Tables}}
                            <option value="{{.Name}}">{{.Name}}</option>
                            {{end}}
                        </select>
                    </div>

                    <div class="form-group">
                        <label for="columnName">New Column Name:</label>
                        <input type="text" id="columnName" name="columnName" required placeholder="new_column">
                    </div>

                    <div class="form-group">
                        <label for="dataType">Data Type:</label>
                        <select id="dataType" name="dataType" required>
                            <option value="TEXT">TEXT</option>
                            <option value="INTEGER">INTEGER</option>
                            <option value="REAL">REAL</option>
                            <option value="BOOLEAN">BOOLEAN</option>
                            <option value="DATETIME">DATETIME</option>
                            <option value="JSON">JSON</option>
                            <option value="BLOB">BLOB</option>
                        </select>
                    </div>

                    <div class="form-group">
                        <label>
                            <input type="checkbox" name="nullable" value="true">
                            Allow NULL values
                        </label>
                    </div>

                    <div class="form-actions">
                        <button type="submit" class="btn btn-primary">Add Column</button>
                        <a href="/tables" class="btn btn-secondary">Cancel</a>
                    </div>
                </form>
            </div>
        </main>
    </div>
</body>
</html>`

	t, err := template.New("alterTableWithMessage").Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tables := ws.getTableList()

	data := struct {
		Tables  []TableInfo
		Message string
		Success bool
	}{
		Tables:  tables,
		Message: message,
		Success: success,
	}

	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, data)
}

func (ws *WebServer) handleInsert(w http.ResponseWriter, r *http.Request) {
	// Placeholder for insert functionality
	http.Error(w, "Insert functionality not implemented yet", http.StatusNotImplemented)
}

func (ws *WebServer) handleDelete(w http.ResponseWriter, r *http.Request) {
	// Placeholder for delete functionality
	http.Error(w, "Delete functionality not implemented yet", http.StatusNotImplemented)
}

func (ws *WebServer) handleStatic(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/static/style.css" {
		w.Header().Set("Content-Type", "text/css")
		w.Write([]byte(getCSS()))
		return
	}
	http.NotFound(w, r)
}

func (ws *WebServer) getTableList() []TableInfo {
	var tables []TableInfo

	// Get schema from storage if available
	if baseStorage, ok := ws.adminServer.storage.(interface{ GetSchema() *schema.Schema }); ok {
		sch := baseStorage.GetSchema()
		if sch != nil {
			for _, table := range sch.Tables {
				tableInfo := TableInfo{
					Name:   table.TableName,
					Fields: make([]FieldInfo, 0, len(table.Fields)),
				}

				for _, fieldName := range table.FieldOrder {
					if field, exists := table.Fields[fieldName]; exists {
						tableInfo.Fields = append(tableInfo.Fields, FieldInfo{
							Name:     field.Name,
							DataType: field.DataType,
							Nullable: field.Nullable,
						})
					}
				}

				tables = append(tables, tableInfo)
			}
		}
	}

	return tables
}

func (ws *WebServer) getTableInfo(tableName string) *TableInfo {
	tables := ws.getTableList()
	for _, table := range tables {
		if table.Name == tableName {
			return &table
		}
	}
	return nil
}

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

/* Tables */
.table-list {
    display: grid;
    gap: 1rem;
}

.table-card {
    background: white;
    border: 1px solid #ddd;
    border-radius: 8px;
    padding: 1.5rem;
    box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

.table-header h3 {
    color: #2c3e50;
    margin-bottom: 1rem;
}

.table-header h3 a {
    color: #2c3e50;
    text-decoration: none;
}

.table-header h3 a:hover {
    color: #3498db;
}

.table-details h4 {
    color: #2c3e50;
    margin-bottom: 0.5rem;
    font-size: 1rem;
}

.table-details ul {
    list-style: none;
    margin-left: 1rem;
}

.table-details li {
    margin-bottom: 0.25rem;
    font-size: 0.9rem;
}

/* Data tables */
.data-table {
    width: 100%;
    border-collapse: collapse;
    margin-top: 1rem;
    background: white;
    border-radius: 8px;
    overflow: hidden;
    box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

.data-table th,
.data-table td {
    padding: 0.75rem;
    text-align: left;
    border-bottom: 1px solid #ddd;
}

.data-table th {
    background: #f8f9fa;
    font-weight: 600;
    color: #2c3e50;
}

.data-table tr:hover {
    background: #f8f9fa;
}

.table-schema {
    margin-bottom: 2rem;
}

.table-schema h3 {
    color: #2c3e50;
    margin-bottom: 1rem;
}

.table-data h3 {
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

.form-group input[type="checkbox"] {
    width: auto;
    margin-right: 0.5rem;
}

.form-actions {
    display: flex;
    gap: 1rem;
    margin-top: 2rem;
}

/* Messages */
.success {
    background: #d4edda;
    color: #155724;
    padding: 1rem;
    border-radius: 4px;
    border: 1px solid #c3e6cb;
    margin-bottom: 1rem;
}

.error {
    background: #f8d7da;
    color: #721c24;
    padding: 1rem;
    border-radius: 4px;
    border: 1px solid #f5c6cb;
    margin-bottom: 1rem;
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

    .data-table {
        font-size: 0.8rem;
    }

    .data-table th,
    .data-table td {
        padding: 0.5rem;
    }
}
`
}