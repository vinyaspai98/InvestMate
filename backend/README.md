# InvestMate Backend API

A high-performance Go REST API for the InvestMate financial tracking platform, powered by the Gin web framework, Firebase Firestore, Google Gemini AI, and Gmail API integration.

---

## 🏗️ Architecture

- **Language & Framework**: Go (v1.24+) with [Gin](https://gin-gonic.com/) web framework
- **Database**: [Firebase Firestore](https://firebase.google.com/docs/firestore) using user-isolated subcollections
- **Authentication**: Firebase Authentication JWT verification via Google OAuth
- **AI Transaction Extraction**: Google Gemini API structured parsing for CDSL Demat transaction emails
- **Email Ingestion**: Gmail API via OAuth2 access tokens for automated inbox sync
- **Market Data Providers**:
  - **Yahoo Finance API**: BSE (`.BO`) and NSE (`.NS`) real-time and historical stock quotes
  - **mfapi.in**: Real-time NAV quotes for Indian mutual funds
  - **Alpha Vantage API**: Daily stock quotes and market data fallback
- **Architecture**: Stateless RESTful API with automated ISIN matching, weighted buy/sell calculations, and portfolio aggregation

---

## 🚀 Features

- **Firebase JWT Authentication**: Validates Firebase ID tokens and handles Google OAuth tokens.
- **AI-Powered Gmail Sync (`/api/v1/gmail/sync`)**:
  - Connects to the user's Gmail using OAuth access tokens.
  - Queries emails from `services@cdslindia.co.in` with subject *"Transactions In Your Demat Account"*.
  - Extracts transaction tables from email HTML.
  - Uses Google Gemini AI with structured system instructions (`gemini_cdsl_system_instruction.txt`) to extract transaction metadata (category, stock/fund name, ISIN, ticker, quantity, buy/sell type, transaction date).
  - Automatically fetches live prices and NAVs from Yahoo Finance and mfapi.in.
  - Matches ISINs with existing portfolio holdings, computes weighted-average buy price, or deducts sold quantities.
- **Portfolio & Net Worth Calculations**: Real-time aggregation of total assets, liabilities, and category-level profit/loss.
- **Investment Management (CRUD)**: Complete lifecycle management for Stocks, Mutual Funds, Fixed Deposits, and Insurance policies.
- **Loan Management**: Tracking principal, outstanding balances, interest rates, tenures, and monthly EMI metrics.
- **Transaction Ledger**: Logs all financial operations (`buy`, `sell`, `deposit`, `withdrawal`, `payment`) tied to specific assets.
- **Historical Chart Data**: Dynamically generated portfolio valuation timelines filtered by period (`1M`, `3M`, `1Y`, `All`).
- **User Preferences**: Currency formatting, dark/light theme options, and notification settings.

---

## 📁 Project Structure

```
backend/
├── cmd/
│   └── main.go                 # Application entry point, router & server initialization
├── internal/
│   ├── constants/
│   │   └── constants.go        # Global constants and collection names
│   ├── handlers/               # HTTP request handlers
│   │   ├── auth.go             # Authentication and user profile handlers
│   │   ├── chart.go            # Historical chart data handlers
│   │   ├── dashboard.go        # Dashboard summary and net worth calculation
│   │   ├── gmail.go            # Gmail OAuth, CDSL parsing & Gemini AI sync
│   │   ├── gmail_test.go       # Unit tests for Gmail parsing and Gemini extraction
│   │   ├── investment.go       # Investment CRUD and category summary handlers
│   │   ├── loan.go             # Loan management and EMI calculation handlers
│   │   ├── transaction.go      # Transaction ledger handlers
│   │   └── prompts/
│   │       └── gemini_cdsl_system_instruction.txt # Gemini prompt schema for CDSL emails
│   ├── middleware/
│   │   └── auth.go             # Firebase JWT verification middleware
│   └── models/
│       ├── investment.go       # Models for investments, loans, transactions, charts
│       └── user.go             # User profile, preferences and auth models
├── pkg/
│   └── config/
│       └── config.go           # Environment variables and Firebase initialization
├── configs/
│   └── firebase-service-account.json # Firebase service account key (git-ignored)
├── firestore.rules             # Firestore security rules
├── Dockerfile                  # Production container definition
├── .env.example                # Template for environment variables
├── go.mod                      # Go module definitions
├── go.sum                      # Go module checksums
└── README.md                   # This documentation file
```

---

## 🔧 Setup & Installation

### Prerequisites

- **Go**: 1.24.6 or higher
- **Firebase Project**: Firestore Database and Firebase Authentication enabled
- **Firebase Service Account**: Service account JSON credentials file
- **Google Cloud / Gemini API Key**: For AI extraction
- **Alpha Vantage API Key**: (Optional/Recommended) For market data fallback

### 1. Clone & Setup Environment

```bash
cd backend
cp .env.example .env
```

### 2. Configure Environment Variables

Update your `.env` file with the appropriate credentials:

```env
# Server Configuration
PORT=8080
ENVIRONMENT=development
GIN_MODE=debug
LOG_LEVEL=info

# Firebase Configuration
FIREBASE_PROJECT_ID=your-firebase-project-id
FIREBASE_CREDENTIALS_PATH=./configs/firebase-service-account.json

# CORS Configuration
CORS_ALLOWED_ORIGINS=http://localhost:4200,http://localhost:3000

# Gmail API Configuration
GMAIL_CLIENT_ID=your-google-oauth-client-id
GMAIL_CLIENT_SECRET=your-google-oauth-client-secret
GMAIL_REDIRECT_URI=http://localhost:4200/auth/callback

# AI & Market Data API Keys
GEMINI_API_KEY=your-gemini-api-key
ALPHA_VANTAGE_API_KEY=your-alpha-vantage-api-key
```

### 3. Install Dependencies

```bash
go mod tidy
```

### 4. Place Firebase Service Account Credentials

Download your service account JSON from Firebase Console (`Project Settings > Service Accounts > Generate new private key`) and save it to:
```
backend/configs/firebase-service-account.json
```
(Or set `FIREBASE_CREDENTIALS_PATH` to point to its path).

### 5. Run the Server

```bash
go run cmd/main.go
```

The API server will listen on `http://localhost:8080`.

---

## 📡 API Endpoints Reference

All endpoints except `/health` and `/api/v1/auth/verify-token` require a valid Firebase ID token passed in the header:
```
Authorization: Bearer <FIREBASE_ID_TOKEN>
```

### Health Check
- `GET /health` - Server status and current timestamp

### Authentication & Users
- `POST /api/v1/auth/verify-token` - Verify Firebase JWT token (accepts `{ "token": "...", "gmailAccessToken": "..." }`)
- `GET /api/v1/auth/user` - Retrieve authenticated user profile
- `GET /api/v1/users/profile` - Get user profile and preferences
- `PUT /api/v1/users/profile` - Update user name or phone number
- `PUT /api/v1/users/preferences` - Update user currency, theme, or notification preferences

### Dashboard
- `GET /api/v1/dashboard/summary` - Aggregate net worth, assets, liabilities, and category summaries

### Investments
- `GET /api/v1/investments/category/:category` - Get investments by category (`stocks`, `mutualFunds`, `fds`, `insurance`)
- `GET /api/v1/investments/category/:category/summary` - Get summary metrics for a specific category
- `GET /api/v1/investments/detail/:id` - Get specific investment by ID
- `POST /api/v1/investments` - Create a new investment
- `PUT /api/v1/investments/:id` - Update existing investment
- `DELETE /api/v1/investments/:id` - Delete an investment

### Loans
- `GET /api/v1/loans/` - Get all loans
- `GET /api/v1/loans/summary` - Get aggregated loan summary (principal, outstanding, count)
- `GET /api/v1/loans/:id` - Get specific loan by ID
- `POST /api/v1/loans` - Create a new loan
- `PUT /api/v1/loans/:id` - Update existing loan
- `DELETE /api/v1/loans/:id` - Delete a loan

### Transactions
- `GET /api/v1/transactions/` - Get all transactions
- `GET /api/v1/transactions/category/:category` - Get transactions by category
- `GET /api/v1/transactions/related/:relatedId` - Get transactions tied to an investment or loan
- `POST /api/v1/transactions/` - Record a new transaction

### Historical Charts
- `GET /api/v1/charts/:category/:period` - Get valuation timeline for a category (`1M`, `3M`, `1Y`, `All`)
- `GET /api/v1/charts/portfolio/:period` - Get overall portfolio valuation timeline

### Gmail & AI Sync
- `POST /api/v1/gmail/sync` - Ingest CDSL Demat transaction emails from Gmail, parse via Gemini AI, resolve market quotes, and update holdings

---

## 🗄️ Database Schema

### Firestore User-Isolated Subcollections

```
users/{userId}/
├── (user profile fields: email, name, preferences, gmailAccessToken, lastGmailSync)
├── investments/{investmentId}
│   ├── category: "stocks" | "mutualFunds" | "fds" | "insurance"
│   ├── name: string
│   ├── investedAmount: number
│   ├── currentValue: number
│   ├── investmentData: { isin, ticker, quantity, averagePrice, ... }
│   └── updatedAt: timestamp
├── loans/{loanId}
│   ├── name: string
│   ├── principalAmount: number
│   ├── outstandingBalance: number
│   ├── interestRate: number
│   ├── emiAmount: number
│   └── tenureMonths: number
└── transactions/{transactionId}
    ├── category: string
    ├── type: "buy" | "sell" | "deposit" | "withdrawal" | "payment"
    ├── amount: number
    ├── relatedId: string
    └── transactionDate: timestamp
```

---

## 🧪 Testing

```bash
# Run unit tests
go test ./...

# Run unit tests with coverage
go test -cover ./...

# Run Gmail handler tests specifically
go test -v ./internal/handlers -run TestGmail
```

---

## 📦 Build & Containerization

### Compile Binary
```bash
go build -o bin/investmate-api cmd/main.go
```

### Docker
```bash
# Build image
docker build -t investmate-api .

# Run container
docker run -p 8080:8080 --env-file .env investmate-api
```