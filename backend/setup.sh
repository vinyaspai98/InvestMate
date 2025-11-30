#!/bin/bash

# InvestMate Backend Development Setup Script

echo "🚀 InvestMate Backend Setup"
echo "=========================="

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.24.6 or higher."
    exit 1
fi

echo "✅ Go version: $(go version)"

# Check if we're in the backend directory
if [ ! -f "go.mod" ]; then
    echo "❌ Please run this script from the backend directory"
    exit 1
fi

# Install dependencies
echo "📦 Installing Go dependencies..."
go mod tidy

# Check if .env file exists
if [ ! -f ".env" ]; then
    echo "📝 Creating .env file from template..."
    cp .env.example .env
    echo "⚠️  Please update .env file with your Firebase configuration"
fi

# Build the application
echo "🔨 Building the application..."
if go build -o bin/investmate-api cmd/main.go; then
    echo "✅ Build successful!"
else
    echo "❌ Build failed!"
    exit 1
fi

# Create configs directory for Firebase credentials
mkdir -p configs

echo ""
echo "🎉 Setup Complete!"
echo ""
echo "Next steps:"
echo "1. Update .env file with your Firebase project ID"
echo "2. Download Firebase service account JSON to configs/firebase-service-account.json"
echo "3. Run: go run cmd/main.go"
echo ""
echo "API will be available at: http://localhost:8080"
echo "Health check: http://localhost:8080/health"