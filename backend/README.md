# InvestMate Backend API

A robust GoLang REST API for the InvestMate financial tracking application, built with Gin web framework and Firebase Firestore.

## 🏗️ Architecture

- **Framework**: Gin HTTP framework
- **Database**: Firebase Firestore (subcollection-based schema)
- **Authentication**: Firebase Auth JWT token validation
- **Architecture**: Stateless REST API with real-time calculations
- **Design**: Supports wireframe-based Angular frontend

## 🚀 Features

- **Authentication**: Firebase JWT token verification
- **Dashboard**: Net worth calculation and category summaries
- **Investment Management**: CRUD operations for stocks, mutual funds, FDs, insurance
- **Loan Management**: Comprehensive loan tracking with EMI calculations
- **Transaction History**: Detailed transaction tracking and categorization
- **Chart Data**: Historical data with time period filtering (1M, 3M, 1Y, All)
- **User Profiles**: Profile and preference management

## 📁 Project Structure

```
backend/
├── cmd/
│   └── main.go                 # Application entry point
├── internal/
│   ├── handlers/              # HTTP request handlers
│   │   ├── auth.go           # Authentication handlers
│   │   ├── dashboard.go      # Dashboard summary handlers
│   │   ├── investment.go     # Investment CRUD handlers
│   │   ├── loan.go          # Loan management handlers
│   │   ├── transaction.go   # Transaction handlers
│   │   └── chart.go         # Chart data handlers
│   ├── middleware/           # HTTP middleware
│   │   └── auth.go          # Firebase JWT middleware
│   └── models/              # Data models and structs
│       ├── user.go          # User and preferences models
│       └── investment.go    # Investment, loan, transaction models
├── pkg/
│   └── config/              # Configuration management
│       └── config.go       # Firebase and app configuration
├── .env.example            # Environment variables template
├── go.mod                  # Go module dependencies
└── README.md              # This file
```

## 🔧 Setup & Installation

### Prerequisites

- Go 1.24.6 or higher
- Firebase project with Firestore enabled
- Firebase service account credentials

### 1. Clone and Setup

```bash
cd backend
cp .env.example .env
```

### 2. Configure Environment Variables

Edit `.env` file with your Firebase project details:

```env
FIREBASE_PROJECT_ID=your-firebase-project-id
FIREBASE_CREDENTIALS_PATH=./configs/firebase-service-account.json
PORT=8080
ENVIRONMENT=development
```

### 3. Install Dependencies

```bash
go mod tidy
```

### 4. Firebase Setup

1. Download your Firebase service account JSON file
2. Place it in `configs/firebase-service-account.json`
3. Or set `FIREBASE_CREDENTIALS_PATH` to the file location

### 5. Run the Application

```bash
go run cmd/main.go
```

The server will start on `http://localhost:8080`

## 📡 API Endpoints

### Authentication
- `POST /api/v1/auth/verify-token` - Verify Firebase JWT token
- `GET /api/v1/auth/user` - Get authenticated user profile

### Dashboard
- `GET /api/v1/dashboard/summary` - Get dashboard summary with net worth

### User Profile
- `GET /api/v1/users/profile` - Get user profile
- `PUT /api/v1/users/profile` - Update user profile
- `PUT /api/v1/users/preferences` - Update user preferences

### Investments
- `GET /api/v1/investments/:category` - Get investments by category
- `GET /api/v1/investments/:category/summary` - Get category summary
- `GET /api/v1/investments/:id` - Get specific investment
- `POST /api/v1/investments` - Create new investment
- `PUT /api/v1/investments/:id` - Update investment
- `DELETE /api/v1/investments/:id` - Delete investment

### Loans
- `GET /api/v1/loans` - Get all loans
- `GET /api/v1/loans/summary` - Get loan summary
- `GET /api/v1/loans/:id` - Get specific loan
- `POST /api/v1/loans` - Create new loan
- `PUT /api/v1/loans/:id` - Update loan
- `DELETE /api/v1/loans/:id` - Delete loan

### Transactions
- `GET /api/v1/transactions` - Get all transactions
- `GET /api/v1/transactions/:category` - Get transactions by category
- `GET /api/v1/transactions/related/:relatedId` - Get transactions for specific investment/loan
- `POST /api/v1/transactions` - Create new transaction

### Charts
- `GET /api/v1/charts/:category/:period` - Get category chart data
- `GET /api/v1/charts/portfolio/:period` - Get portfolio chart data

## 🗄️ Database Schema

### Firestore Collections Structure

```
users/{userId}/
├── (user profile fields)
├── investments/{investmentId}
├── loans/{loanId}
└── transactions/{transactionId}
```

### Investment Categories
- `stocks` - Stock investments
- `mutualFunds` - Mutual fund investments  
- `fds` - Fixed deposits
- `insurance` - Insurance policies

### Transaction Types
- `buy` - Purchase transactions
- `sell` - Sale transactions
- `deposit` - Deposit transactions
- `withdrawal` - Withdrawal transactions
- `payment` - Loan payment transactions

## 🔒 Security

- Firebase Authentication for JWT token verification
- User data isolation through subcollections
- Firestore security rules for data protection
- CORS configuration for frontend integration

## 🌍 Environment Configuration

### Development
```env
ENVIRONMENT=development
GIN_MODE=debug
LOG_LEVEL=info
CORS_ALLOWED_ORIGINS=http://localhost:4200,http://localhost:3000
```

### Production
```env
ENVIRONMENT=production
GIN_MODE=release
LOG_LEVEL=warn
CORS_ALLOWED_ORIGINS=https://your-production-domain.com
```

## 🧪 Testing

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Benchmark tests
go test -bench=. ./...
```

## 📦 Build & Deploy

### Build Binary
```bash
go build -o bin/investmate-api cmd/main.go
```

### Docker Deployment
```bash
# Build Docker image
docker build -t investmate-api .

# Run container
docker run -p 8080:8080 --env-file .env investmate-api
```

## 🔍 Health Check

The API provides a health check endpoint:
- `GET /health` - Returns server status and timestamp

## 🌟 Features Implemented

- ✅ Firebase Authentication integration
- ✅ Real-time net worth calculation
- ✅ Category-based investment management
- ✅ Comprehensive loan tracking
- ✅ Transaction history and management
- ✅ Historical chart data generation
- ✅ User profile and preferences
- ✅ CORS configuration for frontend
- ✅ Graceful server shutdown
- ✅ Environment-based configuration

## 🔮 Future Enhancements

- [ ] Real-time stock price integration
- [ ] Advanced portfolio analytics
- [ ] Email notifications
- [ ] Data export functionality
- [ ] Investment performance calculations
- [ ] Multi-currency support
- [ ] Bulk data import/export

## 📝 License

This project is part of the InvestMate financial tracking application.