# Windows Usage Examples

## Command Prompt (cmd.exe)
onemcp --help
onemcp status
onemcp add filesystem @modelcontextprotocol/server-filesystem

## PowerShell
.\onemcp.exe --help
.\onemcp.exe status
.\onemcp.exe add github @modelcontextprotocol/server-github

## Setting API Keys
onemcp set-key github GITHUB_PERSONAL_ACCESS_TOKEN ghp_xxxxxxxxxx
onemcp set-key brave-search BRAVE_API_KEY your-api-key

## Starting Gateway
onemcp start

## Web Interface
# Opens at http://localhost:8080
onemcp web

