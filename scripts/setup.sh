#!/bin/bash
# Setup script for MAGE-X development

set -e

echo "🚀 Setting up MAGE-X development environment..."

# Check Go version
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
REQUIRED_VERSION="1.24"

if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]; then 
    echo "❌ Go version $REQUIRED_VERSION or higher is required (found $GO_VERSION)"
    exit 1
fi

echo "✅ Go version $GO_VERSION"

# Install Mage
echo "📦 Installing Mage..."
go install github.com/magefile/mage@latest

# Download dependencies
echo "📦 Downloading dependencies..."
go mod download

# Run initial build to verify
echo "🔨 Running initial build..."
mage build

# Install development tools
echo "🛠️  Installing development tools..."
mage tools:install

# Run tests
echo "🧪 Running tests..."
mage test

echo "✅ Setup complete! You can now use MAGE-X."
echo ""
echo "Try these commands:"
echo "  mage help        # List all available tasks (beautiful format)"
echo "  mage build       # Build the project"
echo "  mage test        # Run tests"
echo "  mage lint        # Run linter"
echo ""
echo "See README.md for more information."
