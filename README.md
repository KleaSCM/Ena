# 🌸 Ena VA 🌸

Ena is a virtual assistant that manages your system with care! Developed in Go, she handles all system operations including file management, terminal control, application management, system health monitoring, and more- Ena is NOT an AI

## ✨ Features

- 🖥️ **Comprehensive System Control**: Complete control over files, folders, terminals, and applications
- 🏥 **System Health Monitoring**: Real-time monitoring of CPU, memory, and disk usage
- 🔍 **Advanced Search Features**: File search and safe deletion capabilities
- ⚡ **System Operations**: Restart, shutdown, and sleep functionality
- 🎨 **Beautiful Interface**: Colorful and intuitive command-line interface
- 💕 **Gentle English Support**: Loving messages with care and attention

## 🚀 Installation & Running

### Prerequisites

- Go 1.21 or higher
- Linux, macOS, or Windows

### Build Instructions

```bash
# Clone the repository
git clone <repository-url>
cd Ena

# Install dependencies
go mod tidy

# Build
go build -o ena cmd/main.go
```

### 🎯 How to Run

#### Interactive Mode (Recommended)

```bash
# Start Ena and begin interactive mode
./ena

# Or
./ena --help  # Show help
```

#### Direct Command Execution

```bash
# Check system health status
./ena health

# Create a file
./ena file create /path/to/file.txt

# Start an application
./ena app start firefox

# Display system information
./ena system info
```

#### Example Session

```bash
# Start Ena
$ ./ena

🌸 EnaVA 🌸
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Hello! I'm Ena ✨ Let me help you with your system!

💡 Tip: Type 'help' to see what I can do!
💡 Tip: Type 'exit' to say goodbye...

Ena> health
🏥 System Health Report
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
💻 CPU Information:
   Model: AMD Ryzen 7 8845HS w/ Radeon 780M Graphics
   Cores: 16
   Usage: 8.8%
   Status: 🟢 Normal
...

Ena> file create test.txt
Created file "test.txt"! ✨

Ena> exit
See you next time! (╹◡╹)♡
```

## 📖 Usage

### Basic Operations

Ena can be used in two ways:

1. **Interactive Mode**: Start with `./ena` and enter commands interactively
2. **Direct Execution**: Run specific commands with `./ena <command>`

### Help and Support

```bash
# Show general help
./ena --help

# Show help for specific commands
./ena file --help
./ena app --help
./ena system --help
```

## 🎯 Command Reference

### 📁 File Operations

```bash
# Create a file
ena file create /path/to/file.txt

# Read a file
ena file read /path/to/file.txt

# Write to a file
ena file write /path/to/file.txt "Hello, World!"

# Copy a file
ena file copy /source.txt /dest.txt

# Move a file
ena file move /old.txt /new.txt

# Delete a file
ena file delete /path/to/file.txt

# Show file information
ena file info /path/to/file.txt
```

### 📂 Folder Operations

```bash
# Create a folder
ena folder create /path/to/folder

# List folder contents
ena folder list /path/to/folder

# Delete a folder
ena folder delete /path/to/folder

# Show folder information
ena folder info /path/to/folder
```

### 🖥️ Terminal Operations

```bash
# Open a new terminal
ena terminal open

# Close terminal
ena terminal close

# Execute a command
ena terminal execute "ls -la"

# Change directory
ena terminal cd /home/user
```

### 📱 Application Operations

```bash
# Start an application
ena app start firefox

# Stop an application
ena app stop firefox

# Restart an application
ena app restart firefox

# List running applications
ena app list

# Show application information
ena app info firefox
```

### ⚡ System Operations

```bash
# Restart system
ena system restart

# Shutdown system
ena system shutdown

# Put system to sleep
ena system sleep

# Show system information
ena system info
```

### 🏥 System Health Check

```bash
# Check system health status
ena health
```

### 🔍 Search & Delete

```bash
# Search for files
ena search "*.txt" /home/user

# Delete files
ena delete /path/to/file.txt
```

## 🏗️ Architecture

```
Ena/
├── cmd/                    # Main entry point
├── internal/               # Internal packages
│   ├── core/              # Core engine
│   ├── hooks/             # System hooks
│   ├── health/            # System health monitoring
│   └── utils/             # Utilities
├── pkg/                    # Public packages
│   ├── commands/          # Command definitions
│   └── system/            # System operations
├── Docs/                   # Documentation
└── Tests/                  # Test files
```

## 🛡️ Safety Features

- **Safe Mode**: Enabled by default, asks for confirmation before dangerous operations
- **Dangerous Command Detection**: Automatically detects commands that could harm the system
- **Error Handling**: Comprehensive error handling with user-friendly error messages

## 🎨 Customization

Ena's appearance and behavior can be customized through configuration files and environment variables.

## 🔧 Troubleshooting

### Common Issues

**Q: Build errors occur**
```bash
# Reinstall dependencies
go clean -modcache
go mod tidy
go build -o ena cmd/main.go
```

**Q: Terminal won't open**
- Check which terminal emulator is installed on your system
- Supported: gnome-terminal, xterm, konsole, xfce4-terminal, alacritty, kitty

**Q: Application won't start**
- Verify the application name is correct (e.g., firefox, chrome, vim)
- Check if the application is installed on your system

**Q: Permission errors with system operations**
```bash
# May require sudo privileges
sudo ./ena system restart
```


**Author**: KleaSCM  
**Email**: KleaSCM@gmail.com

---
