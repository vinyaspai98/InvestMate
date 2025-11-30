# InvestMate Backend Development Setup Script for Windows

Write-Host "🚀 InvestMate Backend Setup" -ForegroundColor Green
Write-Host "==========================" -ForegroundColor Green

# Check if Go is installed
try {
    $goVersion = go version
    Write-Host "✅ Go version: $goVersion" -ForegroundColor Green
} catch {
    Write-Host "❌ Go is not installed. Please install Go 1.24.6 or higher." -ForegroundColor Red
    exit 1
}

# Check if we're in the backend directory
if (-not (Test-Path "go.mod")) {
    Write-Host "❌ Please run this script from the backend directory" -ForegroundColor Red
    exit 1
}

# Install dependencies
Write-Host "📦 Installing Go dependencies..." -ForegroundColor Yellow
go mod tidy

# Check if .env file exists
if (-not (Test-Path ".env")) {
    Write-Host "📝 Creating .env file from template..." -ForegroundColor Yellow
    Copy-Item ".env.example" ".env"
    Write-Host "⚠️  Please update .env file with your Firebase configuration" -ForegroundColor Yellow
}

# Build the application
Write-Host "🔨 Building the application..." -ForegroundColor Yellow
if (-not (Test-Path "bin")) {
    New-Item -ItemType Directory -Path "bin"
}

$buildResult = go build -o "bin/investmate-api.exe" cmd/main.go
if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Build successful!" -ForegroundColor Green
} else {
    Write-Host "❌ Build failed!" -ForegroundColor Red
    exit 1
}

# Create configs directory for Firebase credentials
if (-not (Test-Path "configs")) {
    New-Item -ItemType Directory -Path "configs"
}

Write-Host ""
Write-Host "🎉 Setup Complete!" -ForegroundColor Green
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Cyan
Write-Host "1. Update .env file with your Firebase project ID" -ForegroundColor White
Write-Host "2. Download Firebase service account JSON to configs/firebase-service-account.json" -ForegroundColor White
Write-Host "3. Run: go run cmd/main.go" -ForegroundColor White
Write-Host ""
Write-Host "API will be available at: http://localhost:8080" -ForegroundColor Cyan
Write-Host "Health check: http://localhost:8080/health" -ForegroundColor Cyan