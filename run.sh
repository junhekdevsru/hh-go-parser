#!/bin/bash

# HH.ru Resume Parser - Build and Run Script

set -e

echo "🚀 HH.ru Resume Parser Setup"
echo "================================"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21 or higher."
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | grep -o 'go[0-9]\+\.[0-9]\+' | cut -d'o' -f2)
REQUIRED_VERSION="1.21"

if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]; then 
    echo "❌ Go version $GO_VERSION is too old. Please install Go $REQUIRED_VERSION or higher."
    exit 1
fi

echo "✅ Go version $GO_VERSION detected"

# Build the application
echo "🔨 Building application..."
go mod tidy
go build -o hh-parser main.go

if [ $? -eq 0 ]; then
    echo "✅ Build successful!"
else
    echo "❌ Build failed!"
    exit 1
fi

# Create necessary directories
mkdir -p output logs keywords

echo "📁 Created directories: output, logs, keywords"

# Check if API token is provided
if [ -z "$HH_API_TOKEN" ]; then
    echo ""
    echo "⚠️  API Token Required"
    echo "================================"
    echo "To use this parser, you need an hh.ru API token."
    echo ""
    echo "1. Register at https://dev.hh.ru/"
    echo "2. Create a new application"
    echo "3. Get your API token"
    echo "4. Set the environment variable:"
    echo "   export HH_API_TOKEN='your_token_here'"
    echo ""
    echo "Or pass it directly with -token flag:"
    echo "   ./hh-parser -token='your_token' -keywords='Go' -format='json'"
    echo ""
else
    echo "✅ API token found in environment"
    echo ""
    echo "🎯 Quick Start Examples:"
    echo "================================"
    echo ""
    echo "Parse Go developers in Moscow (JSON):"
    echo "./hh-parser -token=\"\$HH_API_TOKEN\" -keywords=\"Go,Golang\" -city=\"Moscow\" -format=\"json\""
    echo ""
    echo "Export to CSV:"
    echo "./hh-parser -token=\"\$HH_API_TOKEN\" -keywords=\"Backend,API\" -format=\"csv\" -output=\"results.csv\""
    echo ""
    echo "Generate SQL script:"
    echo "./hh-parser -token=\"\$HH_API_TOKEN\" -keywords=\"Go developer\" -format=\"sql\""
    echo ""
    echo "Use keywords file:"
    echo "./hh-parser -token=\"\$HH_API_TOKEN\" -keywords-file=\"keywords.json\" -format=\"json\""
    echo ""
fi

echo "📚 Documentation: README.md"
echo "🐳 Docker: docker-compose up"
echo "📋 Keywords file: keywords.json (sample included)"
echo ""
echo "🎉 Setup complete! Use ./hh-parser -h for help."
