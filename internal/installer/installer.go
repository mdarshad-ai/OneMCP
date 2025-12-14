package installer

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/yourusername/onemcp/internal/storage"
)

// Installer handles MCP server installation from various sources
type Installer struct {
	cacheDir string
}

// NewInstaller creates a new installer instance
func NewInstaller(cacheDir string) *Installer {
	return &Installer{
		cacheDir: cacheDir,
	}
}

// InstallResult represents the result of an installation
type InstallResult struct {
	Name        string
	Type        storage.ServerType
	Package     string
	Version     string
	InstallPath string
	Success     bool
	Error       string
}

// InstallFromNPM installs an MCP server from npm
func (i *Installer) InstallFromNPM(packageName string) (*InstallResult, error) {
	// Check if Node.js is available
	if err := i.checkNodeJS(); err != nil {
		return &InstallResult{Success: false, Error: err.Error()}, err
	}

	// Create installation directory
	installDir := filepath.Join(i.cacheDir, "npm", strings.ReplaceAll(packageName, "/", "_"))
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return &InstallResult{Success: false, Error: fmt.Sprintf("failed to create install directory: %v", err)}, err
	}

	// Install the package globally in the cache directory
	cmd := exec.Command("npm", "install", "-g", "--prefix", installDir, packageName)
	cmd.Env = append(os.Environ(), "npm_config_global=true")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return &InstallResult{
			Success: false,
			Error:   fmt.Sprintf("npm install failed: %v\nOutput: %s", err, string(output)),
		}, err
	}

	// Get package info to determine version
	version, err := i.getNPMVersion(packageName, installDir)
	if err != nil {
		version = "unknown"
	}

	// Find the binary/script path
	binPath, err := i.findNPMBinary(packageName, installDir)
	if err != nil {
		return &InstallResult{
			Success: false,
			Error:   fmt.Sprintf("failed to find binary: %v", err),
		}, err
	}

	return &InstallResult{
		Name:        strings.ReplaceAll(packageName, "@", ""),
		Type:        storage.ServerTypeNPM,
		Package:     packageName,
		Version:     version,
		InstallPath: binPath,
		Success:     true,
	}, nil
}

// InstallFromPIP installs an MCP server from pip
func (i *Installer) InstallFromPIP(packageName string) (*InstallResult, error) {
	// Check if Python is available
	if err := i.checkPython(); err != nil {
		return &InstallResult{Success: false, Error: err.Error()}, err
	}

	// Create installation directory
	installDir := filepath.Join(i.cacheDir, "pip", strings.ReplaceAll(packageName, "/", "_"))
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return &InstallResult{Success: false, Error: fmt.Sprintf("failed to create install directory: %v", err)}, err
	}

	// Install the package using python3 -m pip
	cmd := exec.Command("python3", "-m", "pip", "install", "--target", installDir, packageName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Try with python if python3 fails
		cmd = exec.Command("python", "-m", "pip", "install", "--target", installDir, packageName)
		output, err = cmd.CombinedOutput()
		if err != nil {
			return &InstallResult{
				Success: false,
				Error:   fmt.Sprintf("pip install failed: %v\nOutput: %s", err, string(output)),
			}, err
		}
	}

	// Get package version
	version, err := i.getPIPVersion(packageName, installDir)
	if err != nil {
		version = "unknown"
	}

	// Find the entry point
	entryPoint, err := i.findPIPEntryPoint(packageName, installDir)
	if err != nil {
		return &InstallResult{
			Success: false,
			Error:   fmt.Sprintf("failed to find entry point: %v", err),
		}, err
	}

	return &InstallResult{
		Name:        strings.ReplaceAll(packageName, "-", "_"),
		Type:        storage.ServerTypePIP,
		Package:     packageName,
		Version:     version,
		InstallPath: entryPoint,
		Success:     true,
	}, nil
}

// InstallFromCustom installs from a custom source (git repo, local path, etc.)
func (i *Installer) InstallFromCustom(source string) (*InstallResult, error) {
	// Determine source type
	if strings.HasPrefix(source, "git@") || strings.HasPrefix(source, "https://") || strings.HasPrefix(source, "http://") {
		return i.installFromGit(source)
	} else if strings.HasPrefix(source, "/") || strings.HasPrefix(source, "./") || strings.HasPrefix(source, "../") {
		return i.installFromLocal(source)
	}

	return &InstallResult{
		Success: false,
		Error:   "unsupported custom source format",
	}, fmt.Errorf("unsupported custom source format: %s", source)
}

// checkNodeJS checks if Node.js is available
func (i *Installer) checkNodeJS() error {
	cmd := exec.Command("node", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Node.js is not installed or not in PATH")
	}
	return nil
}

// checkPython checks if Python is available
func (i *Installer) checkPython() error {
	cmd := exec.Command("python3", "--version")
	if err := cmd.Run(); err != nil {
		cmd = exec.Command("python", "--version")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("Python is not installed or not in PATH")
		}
	}

	// Check if pip is available
	cmd = exec.Command("python3", "-m", "pip", "--version")
	if err := cmd.Run(); err != nil {
		cmd = exec.Command("python", "-m", "pip", "--version")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("pip is not available")
		}
	}

	return nil
}

// getNPMVersion gets the version of an installed npm package
func (i *Installer) getNPMVersion(packageName, installDir string) (string, error) {
	packageJSON := filepath.Join(installDir, "lib", "node_modules", packageName, "package.json")
	data, err := os.ReadFile(packageJSON)
	if err != nil {
		return "", err
	}

	var pkg struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return "", err
	}

	return pkg.Version, nil
}

// findNPMBinary finds the binary path for an npm package
func (i *Installer) findNPMBinary(packageName, installDir string) (string, error) {
	binDir := filepath.Join(installDir, "bin")
	if runtime.GOOS == "windows" {
		binDir = filepath.Join(installDir, "bin")
	}

	// First check the package.json bin field
	packageJSON := filepath.Join(installDir, "lib", "node_modules", packageName, "package.json")
	data, err := os.ReadFile(packageJSON)
	if err == nil {
		var pkg struct {
			Bin map[string]string `json:"bin"`
		}
		if err := json.Unmarshal(data, &pkg); err == nil && len(pkg.Bin) > 0 {
			for binName, binPath := range pkg.Bin {
				// Check in .bin directory first (npm link location)
				fullPath := filepath.Join(installDir, "lib", "node_modules", ".bin", binName)
				if runtime.GOOS == "windows" {
					fullPath += ".cmd"
				}
				if _, err := os.Stat(fullPath); err == nil {
					return fullPath, nil
				}

				// Check direct path
				fullPath = filepath.Join(installDir, "lib", "node_modules", packageName, binPath)
				if _, err := os.Stat(fullPath); err == nil {
					return fmt.Sprintf("node %s", fullPath), nil
				}
			}
		}
	}

	// Look for common binary names as fallback
	possibleNames := []string{
		"mcp-server-" + strings.ToLower(strings.TrimPrefix(packageName, "@modelcontextprotocol/server-")),
		packageName,
		strings.TrimPrefix(packageName, "@"),
		strings.ReplaceAll(strings.TrimPrefix(packageName, "@"), "/", "-"),
		"mcp-server",
	}

	for _, name := range possibleNames {
		binPath := filepath.Join(binDir, name)
		if runtime.GOOS == "windows" {
			binPath += ".cmd"
		}

		if _, err := os.Stat(binPath); err == nil {
			return binPath, nil
		}
	}

	return "", fmt.Errorf("could not find binary for package %s", packageName)
}

// getPIPVersion gets the version of an installed pip package
func (i *Installer) getPIPVersion(packageName, installDir string) (string, error) {
	// Try to find version in metadata
	metadataDir := filepath.Join(installDir, packageName+"-"+strings.Split(packageName, "-")[0])
	if _, err := os.Stat(metadataDir); err == nil {
		metadataFile := filepath.Join(metadataDir, "METADATA")
		if data, err := os.ReadFile(metadataFile); err == nil {
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "Version: ") {
					return strings.TrimPrefix(line, "Version: "), nil
				}
			}
		}
	}

	// Try to find version in dist-info
	distInfoPattern := filepath.Join(installDir, packageName+"-*.dist-info")
	matches, err := filepath.Glob(distInfoPattern)
	if err == nil && len(matches) > 0 {
		metadataFile := filepath.Join(matches[0], "METADATA")
		if data, err := os.ReadFile(metadataFile); err == nil {
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "Version: ") {
					return strings.TrimPrefix(line, "Version: "), nil
				}
			}
		}
	}

	// Fallback: try to run the package to get version
	cmd := exec.Command("python3", "-c", fmt.Sprintf("import %s; print(%s.__version__)", packageName, packageName))
	if output, err := cmd.Output(); err == nil {
		return strings.TrimSpace(string(output)), nil
	}

	return "unknown", nil
}

// findPIPEntryPoint finds the entry point for a pip package
func (i *Installer) findPIPEntryPoint(packageName, installDir string) (string, error) {
	// Look for common entry points
	entryPoints := []string{
		filepath.Join(installDir, packageName, "__main__.py"),
		filepath.Join(installDir, packageName, "main.py"),
		filepath.Join(installDir, packageName, "cli.py"),
		filepath.Join(installDir, packageName, packageName+".py"),
	}

	for _, entryPoint := range entryPoints {
		if _, err := os.Stat(entryPoint); err == nil {
			return fmt.Sprintf("python3 %s", entryPoint), nil
		}
	}

	// Try to find console scripts
	consoleScriptsDir := filepath.Join(installDir, ".dist-info", packageName+"-"+strings.Split(packageName, "-")[0], "entry_points.txt")
	if data, err := os.ReadFile(consoleScriptsDir); err == nil {
		// Parse entry points - this is simplified
		content := string(data)
		if strings.Contains(content, "console_scripts") {
			return fmt.Sprintf("python3 -m %s", packageName), nil
		}
	}

	return "", fmt.Errorf("could not find entry point for package %s", packageName)
}

// installFromGit installs from a git repository
func (i *Installer) installFromGit(repoURL string) (*InstallResult, error) {
	// Extract repo name from URL
	repoName := strings.TrimSuffix(filepath.Base(repoURL), ".git")
	installDir := filepath.Join(i.cacheDir, "git", repoName)

	// Clone the repository
	if err := os.MkdirAll(filepath.Dir(installDir), 0755); err != nil {
		return &InstallResult{Success: false, Error: err.Error()}, err
	}

	cmd := exec.Command("git", "clone", repoURL, installDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		return &InstallResult{
			Success: false,
			Error:   fmt.Sprintf("git clone failed: %v\nOutput: %s", err, string(output)),
		}, err
	}

	// Try to install if there's a setup.py or package.json
	if _, err := os.Stat(filepath.Join(installDir, "package.json")); err == nil {
		// Node.js project
		if err := i.checkNodeJS(); err != nil {
			return &InstallResult{Success: false, Error: err.Error()}, err
		}

		cmd := exec.Command("npm", "install")
		cmd.Dir = installDir
		if err := cmd.Run(); err != nil {
			return &InstallResult{Success: false, Error: err.Error()}, err
		}

		return &InstallResult{
			Name:        repoName,
			Type:        storage.ServerTypeCustom,
			Package:     repoURL,
			Version:     "git",
			InstallPath: filepath.Join(installDir, "index.js"), // Assume main file
			Success:     true,
		}, nil
	} else if _, err := os.Stat(filepath.Join(installDir, "setup.py")); err == nil {
		// Python project
		if err := i.checkPython(); err != nil {
			return &InstallResult{Success: false, Error: err.Error()}, err
		}

		cmd := exec.Command("pip", "install", "-e", installDir)
		if err := cmd.Run(); err != nil {
			return &InstallResult{Success: false, Error: err.Error()}, err
		}

		return &InstallResult{
			Name:        repoName,
			Type:        storage.ServerTypeCustom,
			Package:     repoURL,
			Version:     "git",
			InstallPath: fmt.Sprintf("python3 -m %s", repoName),
			Success:     true,
		}, nil
	}

	return &InstallResult{
		Success: false,
		Error:   "no package.json or setup.py found in repository",
	}, fmt.Errorf("no package.json or setup.py found in repository")
}

// installFromLocal installs from a local path
func (i *Installer) installFromLocal(localPath string) (*InstallResult, error) {
	absPath, err := filepath.Abs(localPath)
	if err != nil {
		return &InstallResult{Success: false, Error: err.Error()}, err
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return &InstallResult{Success: false, Error: err.Error()}, fmt.Errorf("local path does not exist: %s", absPath)
	}

	repoName := filepath.Base(absPath)

	// Check for package.json (Node.js)
	if _, err := os.Stat(filepath.Join(absPath, "package.json")); err == nil {
		if err := i.checkNodeJS(); err != nil {
			return &InstallResult{Success: false, Error: err.Error()}, err
		}

		cmd := exec.Command("npm", "install")
		cmd.Dir = absPath
		if err := cmd.Run(); err != nil {
			return &InstallResult{Success: false, Error: err.Error()}, err
		}

		return &InstallResult{
			Name:        repoName,
			Type:        storage.ServerTypeCustom,
			Package:     localPath,
			Version:     "local",
			InstallPath: filepath.Join(absPath, "index.js"),
			Success:     true,
		}, nil
	}

	// Check for setup.py (Python)
	if _, err := os.Stat(filepath.Join(absPath, "setup.py")); err == nil {
		if err := i.checkPython(); err != nil {
			return &InstallResult{Success: false, Error: err.Error()}, err
		}

		cmd := exec.Command("pip", "install", "-e", absPath)
		if err := cmd.Run(); err != nil {
			return &InstallResult{Success: false, Error: err.Error()}, err
		}

		return &InstallResult{
			Name:        repoName,
			Type:        storage.ServerTypeCustom,
			Package:     localPath,
			Version:     "local",
			InstallPath: fmt.Sprintf("python3 -m %s", repoName),
			Success:     true,
		}, nil
	}

	return &InstallResult{
		Success: false,
		Error:   "no package.json or setup.py found in local path",
	}, fmt.Errorf("no package.json or setup.py found in local path")
}