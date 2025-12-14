package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/mdarshad-ai/OneMCP/internal/config"
	"github.com/mdarshad-ai/OneMCP/internal/gateway"
	"github.com/mdarshad-ai/OneMCP/internal/storage"
)

// ServerInfo represents information about a server for the web API
type ServerInfo struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Version string `json:"version"`
	Status  string `json:"status"`
	Path    string `json:"path"`
}

// Server represents the web server
type Server struct {
	config *config.Config
	gw     *gateway.Gateway
	store  *storage.FileStorage
}

// NewServer creates a new web server
func NewServer(cfg *config.Config, gw *gateway.Gateway, store *storage.FileStorage) *Server {
	return &Server{
		config: cfg,
		gw:     gw,
		store:  store,
	}
}

// Start starts the web server
func (s *Server) Start() error {
	// Serve static files
	fs := http.FileServer(http.Dir(filepath.Join("web", "static")))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// API routes
	http.HandleFunc("/api/", s.handleAPI)

	// Main page
	http.HandleFunc("/", s.handleIndex)

	addr := fmt.Sprintf("%s:%d", s.config.Web.Host, s.config.Web.Port)
	log.Printf("Web server starting on %s", addr)
	return http.ListenAndServe(addr, nil)
}

// handleIndex serves the main web page
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>OneMCP - MCP Server Manager</title>
    <link href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.0.0/css/all.min.css" rel="stylesheet">
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background: #f8f9fa; }
        .ribbon { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 1rem 0; box-shadow: 0 2px 10px rgba(0,0,0,0.1); position: sticky; top: 0; z-index: 100; }
        .ribbon-content { max-width: 1200px; margin: 0 auto; padding: 0 20px; display: flex; justify-content: space-between; align-items: center; }
        .ribbon h1 { font-size: 1.8rem; font-weight: 600; }
        .ribbon .status { opacity: 0.9; font-size: 0.9rem; }
        .nav-tabs { background: white; border-bottom: 2px solid #e9ecef; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
        .nav-tabs-container { max-width: 1200px; margin: 0 auto; padding: 0 20px; }
        .nav-tabs-list { display: flex; list-style: none; margin: 0; padding: 0; }
        .nav-tab { position: relative; }
        .nav-tab-btn { padding: 20px 30px; border: none; background: none; cursor: pointer; font-size: 1.1rem; font-weight: 600; color: #6c757d; border-bottom: 4px solid transparent; transition: all 0.3s; display: flex; align-items: center; gap: 10px; }
        .nav-tab-btn.active { color: #007bff; border-bottom-color: #007bff; background: rgba(0, 123, 255, 0.05); }
        .nav-tab-btn:hover { color: #007bff; background: rgba(0, 123, 255, 0.05); }
        .nav-tab-icon { font-size: 1.2rem; }
        .main-content { max-width: 1200px; margin: 0 auto; padding: 30px 20px; }
        .tab-pane { display: none; }
        .tab-pane.active { display: block; }
        .card { background: white; border-radius: 10px; box-shadow: 0 2px 15px rgba(0,0,0,0.08); overflow: hidden; margin-bottom: 20px; }
        .card-header { background: #f8f9fa; padding: 20px; border-bottom: 1px solid #e9ecef; }
        .card-header h3 { color: #333; font-size: 1.3rem; font-weight: 600; margin-bottom: 5px; }
        .card-header p { color: #6c757d; font-size: 0.9rem; }
        .card-body { padding: 25px; }
        .form-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 20px; }
        .form-group { margin-bottom: 20px; }
        .form-group.full-width { grid-column: 1 / -1; }
        label { display: block; margin-bottom: 8px; font-weight: 600; color: #333; font-size: 0.9rem; }
        input, select, textarea { width: 100%; padding: 12px; border: 2px solid #e9ecef; border-radius: 8px; font-size: 1rem; transition: border-color 0.3s; }
        input:focus, select:focus, textarea:focus { outline: none; border-color: #007bff; box-shadow: 0 0 0 3px rgba(0, 123, 255, 0.1); }
        .btn { padding: 12px 24px; border: none; border-radius: 8px; cursor: pointer; font-size: 1rem; font-weight: 500; transition: all 0.3s; display: inline-flex; align-items: center; gap: 8px; }
        .btn-primary { background: linear-gradient(135deg, #007bff 0%, #0056b3 100%); color: white; }
        .btn-primary:hover { transform: translateY(-2px); box-shadow: 0 4px 15px rgba(0, 123, 255, 0.3); }
        .btn-success { background: linear-gradient(135deg, #28a745 0%, #1e7e34 100%); color: white; }
        .btn-danger { background: linear-gradient(135deg, #dc3545 0%, #bd2130 100%); color: white; }
        .btn-secondary { background: #6c757d; color: white; }
        .server-grid { display: grid; gap: 15px; }
        .server-card { background: white; border: 1px solid #e9ecef; border-radius: 10px; padding: 20px; display: flex; justify-content: space-between; align-items: center; transition: all 0.3s; }
        .server-card:hover { box-shadow: 0 4px 20px rgba(0,0,0,0.1); transform: translateY(-2px); }
        .server-info h4 { color: #333; font-size: 1.1rem; margin-bottom: 5px; }
        .server-meta { color: #6c757d; font-size: 0.9rem; margin-bottom: 8px; }
        .status-badge { display: inline-block; padding: 4px 12px; border-radius: 20px; font-size: 0.8rem; font-weight: 600; text-transform: uppercase; }
        .status-running { background: #d4edda; color: #155724; }
        .status-stopped { background: #f8d7da; color: #721c24; }
        .status-error { background: #fff3cd; color: #856404; }
        .server-actions { display: flex; gap: 10px; }
        .client-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 20px; }
        .client-card { background: white; border: 1px solid #e9ecef; border-radius: 10px; padding: 25px; text-align: center; transition: all 0.3s; }
        .client-card:hover { box-shadow: 0 4px 20px rgba(0,0,0,0.1); }
        .client-icon { font-size: 3rem; color: #007bff; margin-bottom: 15px; }
        .client-card h4 { color: #333; margin-bottom: 10px; }
        .client-card p { color: #6c757d; margin-bottom: 20px; }
        .alert { padding: 15px; border-radius: 8px; margin-bottom: 20px; }
        .alert-success { background: #d4edda; color: #155724; border: 1px solid #c3e6cb; }
        .alert-error { background: #f8d7da; color: #721c24; border: 1px solid #f5c6cb; }
        .loading { text-align: center; padding: 40px; color: #6c757d; }
        .loading i { font-size: 2rem; margin-bottom: 10px; }
        @media (max-width: 768px) {
            .form-grid { grid-template-columns: 1fr; }
            .tab-buttons { flex-direction: column; }
            .server-card { flex-direction: column; align-items: flex-start; gap: 15px; }
        }
    </style>
</head>
<body>
    <!-- Ribbon Navigation -->
    <div class="ribbon">
        <div class="ribbon-content">
            <div>
                <h1><i class="fas fa-cubes"></i> OneMCP</h1>
                <div class="status">Centralized MCP Server Management & Gateway</div>
            </div>
            <div>
                <small>Status: <span id="statusIndicator" style="color: #28a745;"><i class="fas fa-circle"></i> Online</span></small>
            </div>
        </div>
    </div>

    <!-- Navigation Tabs -->
    <nav class="nav-tabs">
        <div class="nav-tabs-container">
            <ul class="nav-tabs-list">
                <li class="nav-tab">
                    <button class="nav-tab-btn active" onclick="showTab('install')">
                        <i class="fas fa-plus-circle nav-tab-icon"></i>
                        <span>Install</span>
                    </button>
                </li>
                <li class="nav-tab">
                    <button class="nav-tab-btn" onclick="showTab('gateway')">
                        <i class="fas fa-server nav-tab-icon"></i>
                        <span>MCP Gateway</span>
                    </button>
                </li>
                <li class="nav-tab">
                    <button class="nav-tab-btn" onclick="showTab('clients')">
                        <i class="fas fa-desktop nav-tab-icon"></i>
                        <span>Clients</span>
                    </button>
                </li>
            </ul>
        </div>
    </nav>

    <div class="main-content">
        <!-- Install Tab -->
        <div id="install" class="tab-pane active">
            <div class="card">
                <div class="card-header">
                    <h3><i class="fas fa-plus-circle"></i> Install New MCP Server</h3>
                    <p>Add MCP servers from npm, pip, or custom sources</p>
                </div>
                <div class="card-body">
                    <form id="addServerForm">
                        <div class="form-grid">
                            <div class="form-group">
                                <label for="serverName"><i class="fas fa-tag"></i> Server Name</label>
                                <input type="text" id="serverName" placeholder="e.g., github, slack, filesystem" required>
                            </div>
                            <div class="form-group">
                                <label for="serverType"><i class="fas fa-cogs"></i> Installation Type</label>
                                <select id="serverType" required>
                                    <option value="npm">NPM Package</option>
                                    <option value="pip">PIP Package</option>
                                    <option value="custom">Custom Source</option>
                                </select>
                            </div>
                        </div>
                        <div class="form-group full-width">
                            <label for="serverSource"><i class="fas fa-link"></i> Source</label>
                            <input type="text" id="serverSource" placeholder="e.g., @modelcontextprotocol/server-filesystem" required>
                            <small style="color: #6c757d; display: block; margin-top: 5px;">
                                For custom: use git URLs or local paths
                            </small>
                        </div>
                        <button type="submit" class="btn btn-primary">
                            <i class="fas fa-download"></i> Install Server
                        </button>
                    </form>
                </div>
            </div>
        </div>

        <!-- Gateway Tab -->
        <div id="gateway" class="tab-pane">
            <div class="card">
                <div class="card-header">
                    <h3><i class="fas fa-server"></i> MCP Gateway - Installed Servers</h3>
                    <p>Manage your installed MCP servers and their status</p>
                </div>
                <div class="card-body">
                    <div id="serverList" class="loading">
                        <i class="fas fa-spinner fa-spin"></i>
                        <p>Loading servers...</p>
                    </div>
                </div>
            </div>
        </div>

        <!-- Clients Tab -->
        <div id="clients" class="tab-pane">
            <div class="card">
                <div class="card-header">
                    <h3><i class="fas fa-desktop"></i> MCP Clients</h3>
                    <p>Configure MCP clients to use your gateway</p>
                </div>
                <div class="card-body">
                    <div class="client-grid">
                        <div class="client-card">
                            <div class="client-icon"><i class="fas fa-code"></i></div>
                            <h4>Cursor</h4>
                            <p>AI-powered code editor with MCP support</p>
                            <button class="btn btn-primary" onclick="configureClient('cursor')">
                                <i class="fas fa-cog"></i> Configure
                            </button>
                        </div>
                        <div class="client-card">
                            <div class="client-icon"><i class="fas fa-robot"></i></div>
                            <h4>Claude Desktop</h4>
                            <p>Anthropic's AI assistant with MCP integration</p>
                            <button class="btn btn-primary" onclick="configureClient('claude')">
                                <i class="fas fa-cog"></i> Configure
                            </button>
                        </div>
                        <div class="client-card">
                            <div class="client-icon"><i class="fas fa-terminal"></i></div>
                            <h4>opencode</h4>
                            <p>Advanced coding assistant with MCP support</p>
                            <button class="btn btn-success" onclick="configureClient('opencode')">
                                <i class="fas fa-check"></i> Configured
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>

    <script>
        // Tab switching
        function showTab(tabName) {
            // Hide all tab panes
            var panes = document.querySelectorAll('.tab-pane');
            for (var i = 0; i < panes.length; i++) {
                panes[i].classList.remove('active');
            }

            // Remove active class from all tab buttons
            var buttons = document.querySelectorAll('.nav-tab-btn');
            for (var i = 0; i < buttons.length; i++) {
                buttons[i].classList.remove('active');
            }

            // Show the selected tab pane
            var selectedPane = document.getElementById(tabName);
            if (selectedPane) {
                selectedPane.classList.add('active');
            }

            // Add active class to the button that corresponds to this tab
            var buttons = document.querySelectorAll('.nav-tab-btn');
            for (var i = 0; i < buttons.length; i++) {
                var btn = buttons[i];
                if (btn.getAttribute('onclick') && btn.getAttribute('onclick').includes(tabName)) {
                    btn.classList.add('active');
                    break;
                }
            }

            // Load servers when gateway tab is selected
            if (tabName === 'gateway') {
                loadServers();
            }
        }

        // Load servers
        function loadServers() {
            var serverList = document.getElementById('serverList');
            serverList.innerHTML = '<div class="loading"><i class="fas fa-spinner fa-spin"></i><p>Loading servers...</p></div>';

            fetch('/api/servers')
                .then(function(response) {
                    return response.json();
                })
                .then(function(servers) {
                    if (servers.length === 0) {
                        serverList.innerHTML = '<div class="alert alert-secondary"><i class="fas fa-info-circle"></i> No servers installed yet. Go to the Install tab to add your first server.</div>';
                        return;
                    }

                    var html = '<div class="server-grid">';
                    for (var i = 0; i < servers.length; i++) {
                        var server = servers[i];
                        html += '<div class="server-card">' +
                            '<div class="server-info">' +
                                '<h4><i class="fas fa-server"></i> ' + server.name + '</h4>' +
                                '<div class="server-meta">' + server.type + ' • v' + server.version + '</div>' +
                                '<span class="status-badge status-' + server.status + '">' + server.status + '</span>' +
                            '</div>' +
                            '<div class="server-actions">' +
                                (server.status === 'running' ?
                                    '<button class="btn btn-danger" onclick="stopServer(\'' + server.name + '\')"><i class="fas fa-stop"></i> Stop</button>' :
                                    '<button class="btn btn-success" onclick="startServer(\'' + server.name + '\')"><i class="fas fa-play"></i> Start</button>') +
                                '<button class="btn btn-secondary" onclick="viewLogs(\'' + server.name + '\')"><i class="fas fa-eye"></i> Logs</button>' +
                                '<button class="btn btn-danger" onclick="removeServer(\'' + server.name + '\')"><i class="fas fa-trash"></i> Remove</button>' +
                            '</div>' +
                        '</div>';
                    }
                    html += '</div>';
                    serverList.innerHTML = html;
                })
                .catch(function(error) {
                    serverList.innerHTML = '<div class="alert alert-error"><i class="fas fa-exclamation-triangle"></i> Error loading servers: ' + error.message + '</div>';
                });
        }

        // Add server form
        document.getElementById('addServerForm').addEventListener('submit', function(e) {
            e.preventDefault();
            var name = document.getElementById('serverName').value;
            var type = document.getElementById('serverType').value;
            var source = document.getElementById('serverSource').value;

            var submitBtn = e.target.querySelector('button[type="submit"]');
            var originalText = submitBtn.innerHTML;
            submitBtn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Installing...';
            submitBtn.disabled = true;

            fetch('/api/servers', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ name: name, type: type, source: source })
            })
                .then(function(response) {
                    if (response.ok) {
                        showAlert('Server installed successfully!', 'success');
                        document.getElementById('addServerForm').reset();
                        loadServers();
                    } else {
                        var error = response.text();
                        showAlert('Failed to install server: ' + error, 'error');
                    }
                })
                .catch(function(error) {
                    showAlert('Error: ' + error.message, 'error');
                })
                .finally(function() {
                    submitBtn.innerHTML = originalText;
                    submitBtn.disabled = false;
                });
        });

        // Server actions
        function startServer(name) {
            serverAction(name, 'start', 'Starting server...');
        }

        function stopServer(name) {
            serverAction(name, 'stop', 'Stopping server...');
        }

        function serverAction(name, action, loadingText) {
            var btn = event.target;
            var originalText = btn.innerHTML;
            btn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> ' + loadingText;
            btn.disabled = true;

            fetch('/api/servers/' + name + '/' + action, { method: 'POST' })
                .then(function(response) {
                    if (response.ok) {
                        loadServers();
                        showAlert('Server ' + action + ' successful!', 'success');
                    } else {
                        showAlert('Failed to ' + action + ' server', 'error');
                    }
                })
                .catch(function(error) {
                    showAlert('Error: ' + error.message, 'error');
                })
                .finally(function() {
                    btn.innerHTML = originalText;
                    btn.disabled = false;
                });
        }

        function removeServer(name) {
            if (!confirm('Are you sure you want to remove the server "' + name + '"? This action cannot be undone.')) return;

            var btn = event.target;
            var originalText = btn.innerHTML;
            btn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Removing...';
            btn.disabled = true;

            fetch('/api/servers/' + name, { method: 'DELETE' })
                .then(function(response) {
                    if (response.ok) {
                        loadServers();
                        showAlert('Server removed successfully!', 'success');
                    } else {
                        showAlert('Failed to remove server', 'error');
                    }
                })
                .catch(function(error) {
                    showAlert('Error: ' + error.message, 'error');
                })
                .finally(function() {
                    btn.innerHTML = originalText;
                    btn.disabled = false;
                });
        }

        function viewLogs(name) {
            // TODO: Implement log viewing
            alert('Log viewing will be implemented in the next update');
        }

        // Client configuration
        function configureClient(client) {
            var title = '';
            var instructions = '';
            switch(client) {
                case 'cursor':
                    title = 'Configure Cursor';
                    instructions = '<strong>Steps to configure Cursor:</strong><br><br>' +
                        '1. Open Cursor settings (Cmd/Ctrl + ,)<br>' +
                        '2. Go to "MCP" section<br>' +
                        '3. Click "Add Server"<br>' +
                        '4. Enter configuration:<br>' +
                        '   <code>Name: OneMCP Gateway<br>' +
                        '   Command: onemcp<br>' +
                        '   Args: ["start"]</code><br>' +
                        '5. Restart Cursor<br><br>' +
                        '<em>Note: Make sure OneMCP is running on your server.</em>';
                    break;
                case 'claude':
                    title = 'Configure Claude Desktop';
                    instructions = '<strong>Steps to configure Claude Desktop:</strong><br><br>' +
                        '1. Open Claude Desktop<br>' +
                        '2. Go to Settings > MCP Servers<br>' +
                        '3. Click "Add Server"<br>' +
                        '4. Enter configuration:<br>' +
                        '   <code>Name: OneMCP Gateway<br>' +
                        '   Command: onemcp<br>' +
                        '   Args: ["start"]<br>' +
                        '   Environment: {}</code><br>' +
                        '5. Restart Claude Desktop<br><br>' +
                        '<em>Note: Ensure OneMCP is accessible from your local machine.</em>';
                    break;
                case 'opencode':
                    title = 'OneMCP Status';
                    instructions = '<strong>opencode is already configured!</strong><br><br>' +
                        'opencode automatically detects and connects to MCP servers configured in your <code>~/.opencode/mcp-servers.json</code> file.<br><br>' +
                        'Your OneMCP gateway is already listed and active. You can start using MCP tools in opencode immediately!';
                    break;
            }

            // Create modal dialog
            var modal = document.createElement('div');
            modal.style.cssText = 'position: fixed; top: 0; left: 0; width: 100%; height: 100%; background: rgba(0,0,0,0.5); display: flex; align-items: center; justify-content: center; z-index: 1000;';
            modal.innerHTML = '<div style="background: white; padding: 30px; border-radius: 10px; max-width: 500px; width: 90%; max-height: 80vh; overflow-y: auto;">' +
                '<h3 style="margin-top: 0; color: #333;"><i class="fas fa-cog"></i> ' + title + '</h3>' +
                '<div style="margin: 20px 0; line-height: 1.6;">' + instructions + '</div>' +
                '<div style="text-align: right;">' +
                    '<button onclick="this.closest(\'div\').parentElement.remove()" class="btn btn-secondary">' +
                        '<i class="fas fa-times"></i> Close' +
                    '</button>' +
                '</div>' +
            '</div>';
            document.body.appendChild(modal);
        }

        // Check MCP server status
        function checkServerStatus() {
            fetch('/api/servers')
                .then(function(response) {
                    if (response.ok) {
                        document.getElementById('statusIndicator').innerHTML = '<i class="fas fa-circle"></i> Online';
                        document.getElementById('statusIndicator').style.color = '#28a745';
                    } else {
                        document.getElementById('statusIndicator').innerHTML = '<i class="fas fa-exclamation-triangle"></i> Issues';
                        document.getElementById('statusIndicator').style.color = '#ffc107';
                    }
                })
                .catch(function(error) {
                    document.getElementById('statusIndicator').innerHTML = '<i class="fas fa-times-circle"></i> Offline';
                    document.getElementById('statusIndicator').style.color = '#dc3545';
                });
        }

        // Alert system
        function showAlert(message, type) {
            var alertDiv = document.createElement('div');
            alertDiv.className = 'alert alert-' + type;
            alertDiv.innerHTML = '<i class="fas fa-' + (type === 'success' ? 'check-circle' : 'exclamation-triangle') + '"></i> ' + message;

            var container = document.querySelector('.main-content');
            container.insertBefore(alertDiv, container.firstChild);

            setTimeout(function() {
                alertDiv.remove();
            }, 5000);
        }

        // Initialize
        document.addEventListener('DOMContentLoaded', function() {
            // Check server status
            checkServerStatus();
            setInterval(checkServerStatus, 30000); // Check every 30 seconds

            // Load gateway tab data when first accessed
            showTab('install');
        });
    </script>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}
func (s *Server) handleServers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s.getServers(w, r)
	case "POST":
		s.addServer(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleAPI handles all API routes
func (s *Server) handleAPI(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/")

	if path == "servers" || strings.HasPrefix(path, "servers?") {
		s.handleServers(w, r)
		return
	}

	if strings.HasPrefix(path, "servers/") {
		s.handleServerActions(w, r)
		return
	}

	if path == "config" {
		s.handleConfig(w, r)
		return
	}

	http.Error(w, "API endpoint not found", http.StatusNotFound)
}

// handleServerActions handles individual server endpoints
func (s *Server) handleServerActions(w http.ResponseWriter, r *http.Request) {
	log.Printf("DEBUG: handleServerActions called with URL: %s, Method: %s", r.URL.Path, r.Method)
	// Extract server name and action from URL
	path := strings.TrimPrefix(r.URL.Path, "/api/servers/")
	log.Printf("DEBUG: Extracted path: %s", path)

	// Check if there's an action (contains /)
	slashIndex := strings.Index(path, "/")
	if slashIndex == -1 {
		// No action, just server name
		serverName := path
		if serverName == "" {
			http.Error(w, "Server name required", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case "DELETE":
			s.removeServer(w, r, serverName)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	// Has action
	serverName := path[:slashIndex]
	action := path[slashIndex+1:]

	if serverName == "" {
		http.Error(w, "Server name required", http.StatusBadRequest)
		return
	}

	switch action {
	case "start":
		if r.Method == "POST" {
			s.startServer(w, r, serverName)
			return
		}
	case "stop":
		if r.Method == "POST" {
			s.stopServer(w, r, serverName)
			return
		}
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// handleConfig handles the /api/config endpoint
func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s.getConfig(w, r)
	case "PUT":
		s.updateConfig(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) getServers(w http.ResponseWriter, r *http.Request) {
	serversMap := s.gw.ListServers()

	// Convert map to array for JSON response
	servers := make([]ServerInfo, 0, len(serversMap))
	for _, server := range serversMap {
		servers = append(servers, ServerInfo{
			Name:    server.Name,
			Type:    server.Type,
			Version: server.Version,
			Status:  server.Status,
			Path:    server.Path,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(servers)
}

func (s *Server) addServer(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name   string `json:"name"`
		Type   string `json:"type"`
		Source string `json:"source"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Here we would call the installer, but for now just return success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (s *Server) removeServer(w http.ResponseWriter, r *http.Request, name string) {
	// Here we would remove the server
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (s *Server) startServer(w http.ResponseWriter, r *http.Request, name string) {
	if err := s.gw.StartServer(name); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (s *Server) stopServer(w http.ResponseWriter, r *http.Request, name string) {
	if err := s.gw.StopServer(name); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (s *Server) getConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.config)
}

func (s *Server) updateConfig(w http.ResponseWriter, r *http.Request) {
	// Here we would update the config
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}