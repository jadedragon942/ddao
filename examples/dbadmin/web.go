package main

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jadedragon942/ddao/object"
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
	http.HandleFunc("/connect", ws.handleConnect)
	http.HandleFunc("/disconnect", ws.handleDisconnect)
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
	if !ws.adminServer.connected {
		ws.handleConnect(w, r)
		return
	}

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
                <a href="/disconnect" class="disconnect-btn">Disconnect</a>
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
                        <p><strong>Status:</strong> <span class="status-connected">Connected</span></p>
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

func (ws *WebServer) handleConnect(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		ws.showConnectForm(w, r)
	} else if r.Method == "POST" {
		ws.processConnect(w, r)
	}
}

func (ws *WebServer) showConnectForm(w http.ResponseWriter, r *http.Request) {
	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Connect to Database - Database Administration Tool</title>
    <link rel="stylesheet" href="/static/style.css">
    <script>
        function updateConnectionString() {
            const storageType = document.getElementById('storageType').value;
            const connStringInput = document.getElementById('connString');
            const exampleDiv = document.getElementById('connectionExample');

            const examples = {
                'sqlite': {
                    placeholder: 'file:database.db',
                    example: 'Examples: file:mydb.db, :memory:'
                },
                'postgres': {
                    placeholder: 'postgres://user:password@localhost/dbname?sslmode=disable',
                    example: 'Example: postgres://user:pass@localhost:5432/mydb?sslmode=disable'
                },
                'sqlserver': {
                    placeholder: 'sqlserver://user:password@localhost:1433?database=dbname',
                    example: 'Example: sqlserver://sa:MyPass123@localhost:1433?database=mydb'
                },
                'oracle': {
                    placeholder: 'oracle://user:password@localhost:1521/XE',
                    example: 'Example: oracle://system:oracle@localhost:1521/XE'
                },
                'cockroach': {
                    placeholder: 'postgres://user:password@localhost:26257/dbname?sslmode=disable',
                    example: 'Example: postgres://root@localhost:26257/defaultdb?sslmode=disable'
                },
                'yugabyte': {
                    placeholder: 'postgres://user:password@localhost:5433/dbname?sslmode=disable',
                    example: 'Example: postgres://yugabyte@localhost:5433/yugabyte?sslmode=disable'
                },
                'tidb': {
                    placeholder: 'user:password@tcp(localhost:4000)/dbname',
                    example: 'Example: root:@tcp(localhost:4000)/test'
                },
                'scylla': {
                    placeholder: 'localhost:9042/keyspace',
                    example: 'Example: 127.0.0.1:9042/mykeyspace'
                },
                's3': {
                    placeholder: 'bucket-name',
                    example: 'Example: my-s3-bucket'
                }
            };

            if (examples[storageType]) {
                connStringInput.placeholder = examples[storageType].placeholder;
                exampleDiv.textContent = examples[storageType].example;
            }
        }

        window.onload = function() {
            updateConnectionString();
        }
    </script>
</head>
<body>
    <div class="container">
        <header>
            <h1>Database Administration Tool</h1>
            <nav>
                <a href="/connect" class="active">Connect</a>
            </nav>
        </header>

        <main>
            <div class="section">
                <div class="section-header">
                    <h2>Connect to Database</h2>
                </div>

                {{if .Message}}
                <div class="{{if .Success}}success{{else}}error{{end}}">
                    <p>{{.Message}}</p>
                </div>
                {{end}}

                <form method="POST" class="form">
                    <div class="form-group">
                        <label for="storageType">Storage Type:</label>
                        <select id="storageType" name="storageType" onchange="updateConnectionString()" required>
                            <option value="sqlite">SQLite</option>
                            <option value="postgres">PostgreSQL</option>
                            <option value="sqlserver">SQL Server</option>
                            <option value="oracle">Oracle</option>
                            <option value="cockroach">CockroachDB</option>
                            <option value="yugabyte">YugabyteDB</option>
                            <option value="tidb">TiDB</option>
                            <option value="scylla">ScyllaDB/Cassandra</option>
                            <option value="s3">Amazon S3</option>
                        </select>
                    </div>

                    <div class="form-group">
                        <label for="connString">Connection String:</label>
                        <input type="text" id="connString" name="connString" required>
                        <div id="connectionExample" class="connection-example"></div>
                    </div>

                    <div class="form-actions">
                        <button type="submit" class="btn btn-primary">Connect</button>
                    </div>
                </form>

                <div class="connection-help">
                    <h3>Connection String Examples</h3>
                    <ul>
                        <li><strong>SQLite:</strong> file:database.db or :memory:</li>
                        <li><strong>PostgreSQL:</strong> postgres://user:password@host:port/database?sslmode=disable</li>
                        <li><strong>SQL Server:</strong> sqlserver://user:password@host:port?database=dbname</li>
                        <li><strong>Oracle:</strong> oracle://user:password@host:port/service</li>
                        <li><strong>CockroachDB:</strong> postgres://user:password@host:port/database?sslmode=disable</li>
                        <li><strong>YugabyteDB:</strong> postgres://user:password@host:port/database?sslmode=disable</li>
                        <li><strong>TiDB:</strong> user:password@tcp(host:port)/database</li>
                        <li><strong>ScyllaDB:</strong> host:port/keyspace</li>
                        <li><strong>Amazon S3:</strong> bucket-name</li>
                    </ul>
                </div>
            </div>
        </main>
    </div>
</body>
</html>`

	t, err := template.New("connect").Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Message string
		Success bool
	}{}

	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, data)
}

func (ws *WebServer) processConnect(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	storageType := r.FormValue("storageType")
	connString := r.FormValue("connString")

	if storageType == "" || connString == "" {
		ws.showConnectFormWithMessage(w, "Storage type and connection string are required", false)
		return
	}

	err := ws.adminServer.Connect(storageType, connString)
	if err != nil {
		ws.showConnectFormWithMessage(w, fmt.Sprintf("Failed to connect: %v", err), false)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (ws *WebServer) showConnectFormWithMessage(w http.ResponseWriter, message string, success bool) {
	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Connect to Database - Database Administration Tool</title>
    <link rel="stylesheet" href="/static/style.css">
    <script>
        function updateConnectionString() {
            const storageType = document.getElementById('storageType').value;
            const connStringInput = document.getElementById('connString');
            const exampleDiv = document.getElementById('connectionExample');

            const examples = {
                'sqlite': {
                    placeholder: 'file:database.db',
                    example: 'Examples: file:mydb.db, :memory:'
                },
                'postgres': {
                    placeholder: 'postgres://user:password@localhost/dbname?sslmode=disable',
                    example: 'Example: postgres://user:pass@localhost:5432/mydb?sslmode=disable'
                },
                'sqlserver': {
                    placeholder: 'sqlserver://user:password@localhost:1433?database=dbname',
                    example: 'Example: sqlserver://sa:MyPass123@localhost:1433?database=mydb'
                },
                'oracle': {
                    placeholder: 'oracle://user:password@localhost:1521/XE',
                    example: 'Example: oracle://system:oracle@localhost:1521/XE'
                },
                'cockroach': {
                    placeholder: 'postgres://user:password@localhost:26257/dbname?sslmode=disable',
                    example: 'Example: postgres://root@localhost:26257/defaultdb?sslmode=disable'
                },
                'yugabyte': {
                    placeholder: 'postgres://user:password@localhost:5433/dbname?sslmode=disable',
                    example: 'Example: postgres://yugabyte@localhost:5433/yugabyte?sslmode=disable'
                },
                'tidb': {
                    placeholder: 'user:password@tcp(localhost:4000)/dbname',
                    example: 'Example: root:@tcp(localhost:4000)/test'
                },
                'scylla': {
                    placeholder: 'localhost:9042/keyspace',
                    example: 'Example: 127.0.0.1:9042/mykeyspace'
                },
                's3': {
                    placeholder: 'bucket-name',
                    example: 'Example: my-s3-bucket'
                }
            };

            if (examples[storageType]) {
                connStringInput.placeholder = examples[storageType].placeholder;
                exampleDiv.textContent = examples[storageType].example;
            }
        }

        window.onload = function() {
            updateConnectionString();
        }
    </script>
</head>
<body>
    <div class="container">
        <header>
            <h1>Database Administration Tool</h1>
            <nav>
                <a href="/connect" class="active">Connect</a>
            </nav>
        </header>

        <main>
            <div class="section">
                <div class="section-header">
                    <h2>Connect to Database</h2>
                </div>

                <div class="{{if .Success}}success{{else}}error{{end}}">
                    <p>{{.Message}}</p>
                </div>

                <form method="POST" class="form">
                    <div class="form-group">
                        <label for="storageType">Storage Type:</label>
                        <select id="storageType" name="storageType" onchange="updateConnectionString()" required>
                            <option value="sqlite">SQLite</option>
                            <option value="postgres">PostgreSQL</option>
                            <option value="sqlserver">SQL Server</option>
                            <option value="oracle">Oracle</option>
                            <option value="cockroach">CockroachDB</option>
                            <option value="yugabyte">YugabyteDB</option>
                            <option value="tidb">TiDB</option>
                            <option value="scylla">ScyllaDB/Cassandra</option>
                            <option value="s3">Amazon S3</option>
                        </select>
                    </div>

                    <div class="form-group">
                        <label for="connString">Connection String:</label>
                        <input type="text" id="connString" name="connString" required>
                        <div id="connectionExample" class="connection-example"></div>
                    </div>

                    <div class="form-actions">
                        <button type="submit" class="btn btn-primary">Connect</button>
                    </div>
                </form>

                <div class="connection-help">
                    <h3>Connection String Examples</h3>
                    <ul>
                        <li><strong>SQLite:</strong> file:database.db or :memory:</li>
                        <li><strong>PostgreSQL:</strong> postgres://user:password@host:port/database?sslmode=disable</li>
                        <li><strong>SQL Server:</strong> sqlserver://user:password@host:port?database=dbname</li>
                        <li><strong>Oracle:</strong> oracle://user:password@host:port/service</li>
                        <li><strong>CockroachDB:</strong> postgres://user:password@host:port/database?sslmode=disable</li>
                        <li><strong>YugabyteDB:</strong> postgres://user:password@host:port/database?sslmode=disable</li>
                        <li><strong>TiDB:</strong> user:password@tcp(host:port)/database</li>
                        <li><strong>ScyllaDB:</strong> host:port/keyspace</li>
                        <li><strong>Amazon S3:</strong> bucket-name</li>
                    </ul>
                </div>
            </div>
        </main>
    </div>
</body>
</html>`

	t, err := template.New("connectWithMessage").Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Message string
		Success bool
	}{
		Message: message,
		Success: success,
	}

	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, data)
}

func (ws *WebServer) handleDisconnect(w http.ResponseWriter, r *http.Request) {
	ws.adminServer.Disconnect()
	http.Redirect(w, r, "/connect", http.StatusSeeOther)
}

func (ws *WebServer) handleTables(w http.ResponseWriter, r *http.Request) {
	if !ws.adminServer.connected {
		http.Redirect(w, r, "/connect", http.StatusSeeOther)
		return
	}
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
                <a href="/disconnect" class="disconnect-btn">Disconnect</a>
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
	if !ws.adminServer.connected {
		http.Redirect(w, r, "/connect", http.StatusSeeOther)
		return
	}

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
                <a href="/disconnect" class="disconnect-btn">Disconnect</a>
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
	if !ws.adminServer.connected {
		http.Redirect(w, r, "/connect", http.StatusSeeOther)
		return
	}

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
                <a href="/disconnect" class="disconnect-btn">Disconnect</a>
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
                <a href="/disconnect" class="disconnect-btn">Disconnect</a>
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
	if !ws.adminServer.connected {
		http.Redirect(w, r, "/connect", http.StatusSeeOther)
		return
	}

	if r.Method == "GET" {
		ws.showInsertForm(w, r)
	} else if r.Method == "POST" {
		ws.processInsert(w, r)
	}
}

func (ws *WebServer) showInsertForm(w http.ResponseWriter, r *http.Request) {
	tableName := r.URL.Query().Get("table")
	if tableName == "" {
		http.Error(w, "Table name required", http.StatusBadRequest)
		return
	}

	tableInfo := ws.getTableInfo(tableName)
	if tableInfo == nil {
		http.Error(w, "Table not found", http.StatusNotFound)
		return
	}

	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Insert Record - {{.TableName}} - Database Administration Tool</title>
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
                <a href="/disconnect" class="disconnect-btn">Disconnect</a>
            </nav>
        </header>

        <main>
            <div class="section">
                <div class="section-header">
                    <h2>Insert Record - {{.TableName}}</h2>
                    <a href="/table/{{.TableName}}" class="btn btn-secondary">Back to Table</a>
                </div>

                {{if .Message}}
                <div class="{{if .Success}}success{{else}}error{{end}}">
                    <p>{{.Message}}</p>
                </div>
                {{end}}

                <form method="POST" class="form">
                    <input type="hidden" name="table" value="{{.TableName}}">

                    {{range .Fields}}
                    <div class="form-group">
                        <label for="{{.Name}}">{{.Name}} ({{.DataType}}){{if not .Nullable}} *{{end}}:</label>
                        {{if eq .DataType "TEXT"}}
                            <input type="text" id="{{.Name}}" name="{{.Name}}" {{if not .Nullable}}required{{end}}>
                        {{else if eq .DataType "INTEGER"}}
                            <input type="number" id="{{.Name}}" name="{{.Name}}" {{if not .Nullable}}required{{end}}>
                        {{else if eq .DataType "REAL"}}
                            <input type="number" step="any" id="{{.Name}}" name="{{.Name}}" {{if not .Nullable}}required{{end}}>
                        {{else if eq .DataType "BOOLEAN"}}
                            <select id="{{.Name}}" name="{{.Name}}" {{if not .Nullable}}required{{end}}>
                                {{if .Nullable}}<option value="">-- Select --</option>{{end}}
                                <option value="true">True</option>
                                <option value="false">False</option>
                            </select>
                        {{else if eq .DataType "DATETIME"}}
                            <input type="datetime-local" id="{{.Name}}" name="{{.Name}}" {{if not .Nullable}}required{{end}}>
                        {{else if eq .DataType "JSON"}}
                            <textarea id="{{.Name}}" name="{{.Name}}" rows="4" placeholder="{}" {{if not .Nullable}}required{{end}}></textarea>
                        {{else}}
                            <input type="text" id="{{.Name}}" name="{{.Name}}" {{if not .Nullable}}required{{end}}>
                        {{end}}
                        {{if .Nullable}}<small class="field-note">Optional field</small>{{end}}
                    </div>
                    {{end}}

                    <div class="form-actions">
                        <button type="submit" class="btn btn-primary">Insert Record</button>
                        <a href="/table/{{.TableName}}" class="btn btn-secondary">Cancel</a>
                    </div>
                </form>
            </div>
        </main>
    </div>
</body>
</html>`

	t, err := template.New("insertForm").Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		TableName string
		Fields    []FieldInfo
		Message   string
		Success   bool
	}{
		TableName: tableName,
		Fields:    tableInfo.Fields,
	}

	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, data)
}

func (ws *WebServer) processInsert(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	tableName := r.FormValue("table")
	if tableName == "" {
		http.Error(w, "Table name required", http.StatusBadRequest)
		return
	}

	tableInfo := ws.getTableInfo(tableName)
	if tableInfo == nil {
		http.Error(w, "Table not found", http.StatusNotFound)
		return
	}

	// Create new object
	obj := object.New()
	obj.SetTableName(tableName)

	// Process each field
	for _, field := range tableInfo.Fields {
		value := r.FormValue(field.Name)

		// Skip empty values for nullable fields
		if value == "" && field.Nullable {
			continue
		}

		// Convert value based on data type
		switch field.DataType {
		case "TEXT":
			obj.SetField(field.Name, value)
		case "INTEGER":
			if value != "" {
				if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
					obj.SetField(field.Name, intVal)
				} else {
					ws.showInsertFormWithMessage(w, r, fmt.Sprintf("Invalid integer value for field '%s': %s", field.Name, value), false)
					return
				}
			}
		case "REAL":
			if value != "" {
				if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
					obj.SetField(field.Name, floatVal)
				} else {
					ws.showInsertFormWithMessage(w, r, fmt.Sprintf("Invalid float value for field '%s': %s", field.Name, value), false)
					return
				}
			}
		case "BOOLEAN":
			if value != "" {
				if boolVal, err := strconv.ParseBool(value); err == nil {
					obj.SetField(field.Name, boolVal)
				} else {
					ws.showInsertFormWithMessage(w, r, fmt.Sprintf("Invalid boolean value for field '%s': %s", field.Name, value), false)
					return
				}
			}
		case "DATETIME":
			if value != "" {
				if timeVal, err := time.Parse("2006-01-02T15:04", value); err == nil {
					obj.SetField(field.Name, timeVal.Format("2006-01-02 15:04:05"))
				} else {
					ws.showInsertFormWithMessage(w, r, fmt.Sprintf("Invalid datetime value for field '%s': %s", field.Name, value), false)
					return
				}
			}
		default:
			obj.SetField(field.Name, value)
		}
	}

	// Insert the object
	ctx := context.Background()
	id, created, err := ws.adminServer.storage.Insert(ctx, obj)
	if err != nil {
		ws.showInsertFormWithMessage(w, r, fmt.Sprintf("Failed to insert record: %v", err), false)
		return
	}

	if created {
		// Redirect to the table view with success message
		http.Redirect(w, r, fmt.Sprintf("/table/%s?inserted=true&id=%s", tableName, string(id)), http.StatusSeeOther)
	} else {
		ws.showInsertFormWithMessage(w, r, "Record was not created", false)
	}
}

func (ws *WebServer) showInsertFormWithMessage(w http.ResponseWriter, r *http.Request, message string, success bool) {
	tableName := r.FormValue("table")
	if tableName == "" {
		http.Error(w, "Table name required", http.StatusBadRequest)
		return
	}

	tableInfo := ws.getTableInfo(tableName)
	if tableInfo == nil {
		http.Error(w, "Table not found", http.StatusNotFound)
		return
	}

	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Insert Record - {{.TableName}} - Database Administration Tool</title>
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
                <a href="/disconnect" class="disconnect-btn">Disconnect</a>
            </nav>
        </header>

        <main>
            <div class="section">
                <div class="section-header">
                    <h2>Insert Record - {{.TableName}}</h2>
                    <a href="/table/{{.TableName}}" class="btn btn-secondary">Back to Table</a>
                </div>

                <div class="{{if .Success}}success{{else}}error{{end}}">
                    <p>{{.Message}}</p>
                </div>

                <form method="POST" class="form">
                    <input type="hidden" name="table" value="{{.TableName}}">

                    {{range .Fields}}
                    <div class="form-group">
                        <label for="{{.Name}}">{{.Name}} ({{.DataType}}){{if not .Nullable}} *{{end}}:</label>
                        {{if eq .DataType "TEXT"}}
                            <input type="text" id="{{.Name}}" name="{{.Name}}" {{if not .Nullable}}required{{end}}>
                        {{else if eq .DataType "INTEGER"}}
                            <input type="number" id="{{.Name}}" name="{{.Name}}" {{if not .Nullable}}required{{end}}>
                        {{else if eq .DataType "REAL"}}
                            <input type="number" step="any" id="{{.Name}}" name="{{.Name}}" {{if not .Nullable}}required{{end}}>
                        {{else if eq .DataType "BOOLEAN"}}
                            <select id="{{.Name}}" name="{{.Name}}" {{if not .Nullable}}required{{end}}>
                                {{if .Nullable}}<option value="">-- Select --</option>{{end}}
                                <option value="true">True</option>
                                <option value="false">False</option>
                            </select>
                        {{else if eq .DataType "DATETIME"}}
                            <input type="datetime-local" id="{{.Name}}" name="{{.Name}}" {{if not .Nullable}}required{{end}}>
                        {{else if eq .DataType "JSON"}}
                            <textarea id="{{.Name}}" name="{{.Name}}" rows="4" placeholder="{}" {{if not .Nullable}}required{{end}}></textarea>
                        {{else}}
                            <input type="text" id="{{.Name}}" name="{{.Name}}" {{if not .Nullable}}required{{end}}>
                        {{end}}
                        {{if .Nullable}}<small class="field-note">Optional field</small>{{end}}
                    </div>
                    {{end}}

                    <div class="form-actions">
                        <button type="submit" class="btn btn-primary">Insert Record</button>
                        <a href="/table/{{.TableName}}" class="btn btn-secondary">Cancel</a>
                    </div>
                </form>
            </div>
        </main>
    </div>
</body>
</html>`

	t, err := template.New("insertFormWithMessage").Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		TableName string
		Fields    []FieldInfo
		Message   string
		Success   bool
	}{
		TableName: tableName,
		Fields:    tableInfo.Fields,
		Message:   message,
		Success:   success,
	}

	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, data)
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

	if !ws.adminServer.connected || ws.adminServer.storage == nil {
		return tables
	}

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
    align-items: center;
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

/* Connection styles */
.disconnect-btn {
    background-color: #e74c3c !important;
    margin-left: auto;
}

.disconnect-btn:hover {
    background-color: #c0392b !important;
}

.status-connected {
    color: #27ae60;
    font-weight: bold;
}

.connection-example {
    font-size: 0.8rem;
    color: #7f8c8d;
    margin-top: 0.25rem;
    font-style: italic;
}

.connection-help {
    background: #f8f9fa;
    padding: 1.5rem;
    border-radius: 8px;
    border: 1px solid #e9ecef;
    margin-top: 2rem;
}

.connection-help h3 {
    color: #2c3e50;
    margin-bottom: 1rem;
    font-size: 1.1rem;
}

.connection-help ul {
    list-style-type: none;
    margin: 0;
    padding: 0;
}

.connection-help li {
    margin-bottom: 0.75rem;
    padding: 0.5rem;
    background: white;
    border-radius: 4px;
    border: 1px solid #e0e0e0;
}

.connection-help strong {
    color: #2c3e50;
    display: inline-block;
    min-width: 120px;
}

.field-note {
    display: block;
    font-size: 0.8rem;
    color: #7f8c8d;
    margin-top: 0.25rem;
    font-style: italic;
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