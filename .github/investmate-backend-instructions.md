# InvestMate: Backend API Development Instructions

This document outlines the technical specifications for developing the **InvestMate GoLang Backend API** that integrates with Firebase Firestore and supports the wireframe-based Angular frontend.

## 1. Backend Architecture Overview

* **Framework:** GoLang with Gin HTTP framework
* **Database:** Firebase Firestore (subcollection-based schema)
* **Authentication:** Firebase Auth JWT token validation
* **Architecture:** Stateless REST API with real-time calculations
* **Deployment:** Cloud-ready containerized application
* **Design Compliance:** Must support all wireframe features including dashboard cards, investment detail pages, transaction tables, and historical charts

---

## 2. Firebase Firestore Schema

### 2.1. Database Structure (No Portfolio Snapshots)
```
users/{userId}
├── (user profile fields)
├── investments/{investmentId}
├── loans/{loanId}
└── transactions/{transactionId}
```

### 2.2. User Profile Collection
```json
{
  "users": {
    "{userId}": {
      "email": "user@example.com",
      "name": "John Doe",
      "phoneNumber": "+91-9876543210",
      "preferences": {
        "currency": "INR",
        "theme": "dark",
        "notifications": {
          "email": true,
          "push": false
        }
      },
      "createdAt": "2024-01-15T10:30:00Z",
      "updatedAt": "2024-01-20T14:22:00Z",
      "isActive": true
    }
  }
}
```

### 2.3. Investments Subcollection (Unified Schema)
```json
{
  "investments": {
    "{investmentId}": {
      "category": "stocks",
      "name": "Reliance Industries Ltd",
      "investedAmount": 50000.00,
      "currentValue": 55000.00,
      "createdAt": "2024-01-15T10:30:00Z",
      "updatedAt": "2024-01-20T14:22:00Z",
      "isActive": true,
      "notes": "Blue chip stock investment",
      
      "investmentData": {
        "ticker": "RELIANCE",
        "quantity": 100,
        "averagePrice": 500.00,
        "currentPrice": 550.00,
        "exchange": "NSE"
      }
    }
  }
}
```

### 2.4. Loans Subcollection
```json
{
  "loans": {
    "{loanId}": {
      "loanType": "home_loan",
      "lenderName": "HDFC Bank",
      "principalAmount": 2500000.00,
      "outstandingBalance": 2200000.00,
      "interestRate": 8.5,
      "emi": 25000.00,
      "tenure": 240,
      "remainingTenure": 210,
      "startDate": "2023-01-01T00:00:00Z",
      "maturityDate": "2043-01-01T00:00:00Z",
      "createdAt": "2024-01-15T10:30:00Z",
      "updatedAt": "2024-01-20T14:22:00Z",
      "isActive": true,
      "notes": "Home loan for primary residence"
    }
  }
}
```

### 2.5. Transactions Subcollection
```json
{
  "transactions": {
    "{transactionId}": {
      "relatedId": "{investmentId|loanId}",
      "category": "stocks",
      "type": "buy",
      "amount": 10000.00,
      "quantity": 20,
      "price": 500.00,
      "date": "2024-01-15T10:30:00Z",
      "description": "Bought 20 shares of RELIANCE at ₹500",
      "fees": 50.00,
      "createdAt": "2024-01-15T10:30:00Z",
      "notes": "Regular monthly investment"
    }
  }
}
```

---

## 3. API Endpoints (Supporting Wireframe Features)

### 3.1. Authentication Endpoints
```
POST   /api/v1/auth/verify-token     # Verify Firebase JWT token
GET    /api/v1/auth/user            # Get authenticated user profile
```

### 3.2. Dashboard Endpoints (Net Worth Calculation)
```
GET    /api/v1/dashboard/summary    # Dashboard summary with category cards
```

**Response Structure:**
```json
{
  "netWorth": {
    "totalAssets": 500000.00,
    "totalLiabilities": 2200000.00,
    "netWorth": -1700000.00,
    "currency": "INR"
  },
  "categories": {
    "stocks": {
      "investedAmount": 200000.00,
      "currentValue": 220000.00,
      "profitLossPercentage": 10.0,
      "count": 5
    },
    "mutualFunds": {
      "investedAmount": 150000.00,
      "currentValue": 165000.00,
      "profitLossPercentage": 10.0,
      "count": 3
    },
    "fds": {
      "investedAmount": 100000.00,
      "currentValue": 105000.00,
      "profitLossPercentage": 5.0,
      "count": 2
    },
    "insurance": {
      "investedAmount": 50000.00,
      "currentValue": 60000.00,
      "profitLossPercentage": 20.0,
      "count": 1
    },
    "loans": {
      "principalAmount": 2500000.00,
      "outstandingBalance": 2200000.00,
      "count": 1
    }
  }
}
```

### 3.3. Investment Detail Page Endpoints
```
GET    /api/v1/investments/:category          # Get investments by category
GET    /api/v1/investments/:category/summary  # Category aggregate metrics
GET    /api/v1/investments/:id               # Get specific investment details
POST   /api/v1/investments                   # Create investment via FAB modal
PUT    /api/v1/investments/:id               # Update investment
DELETE /api/v1/investments/:id               # Delete investment
```

### 3.4. Transaction Table Support
```
GET    /api/v1/transactions/:category        # Get transactions for category detail page
GET    /api/v1/transactions/:relatedId       # Get transactions for specific investment/loan
POST   /api/v1/transactions                 # Add transaction via modal forms
```

### 3.5. Historical Charts (1M, 3M, 1Y, All filters)
```
GET    /api/v1/charts/:category/:period     # Get chart data for category with time filter
GET    /api/v1/charts/portfolio/:period     # Get overall portfolio chart data
```

**Chart Response Structure:**
```json
{
  "period": "3M",
  "data": [
    {
      "date": "2024-01-15",
      "value": 45000.00
    },
    {
      "date": "2024-02-15", 
      "value": 48000.00
    }
  ]
}
```

### 3.6. Loan Management
```
GET    /api/v1/loans                # Get all user loans
GET    /api/v1/loans/:id            # Get specific loan details
POST   /api/v1/loans                # Create loan via FAB modal
PUT    /api/v1/loans/:id            # Update loan
DELETE /api/v1/loans/:id            # Delete loan
```

### 3.7. Profile Management
```
GET    /api/v1/users/profile        # Get user profile
PUT    /api/v1/users/profile        # Update user profile
PUT    /api/v1/users/preferences    # Update theme/currency preferences
```

---

## 4. Firebase Security Rules

```javascript
rules_version = '2';
service cloud.firestore {
  match /databases/{database}/documents {
    match /users/{userId} {
      allow read, write: if request.auth != null && request.auth.uid == userId;
      
      match /investments/{investmentId} {
        allow read, write: if request.auth != null && request.auth.uid == userId;
      }
      
      match /loans/{loanId} {
        allow read, write: if request.auth != null && request.auth.uid == userId;
      }
      
      match /transactions/{transactionId} {
        allow read, write: if request.auth != null && request.auth.uid == userId;
      }
    }
  }
}
```

---

## 5. Environment Configuration

```
FIREBASE_PROJECT_ID=your-project-id
FIREBASE_CREDENTIALS_PATH=./configs/firebase-service-account.json
PORT=8080
GIN_MODE=release
LOG_LEVEL=info
CORS_ALLOWED_ORIGINS=https://your-frontend-domain.com
```

---

## 6. Key Implementation Principles

### 6.1. Stateless Design
* No server-side sessions or state storage
* All user context derived from Firebase JWT tokens
* Horizontal scaling support

### 6.2. Real-Time Calculations
* Dashboard net worth calculated on each request
* Category summaries aggregated from subcollection data
* Historical charts generated from transaction history

### 6.3. Firestore Query Optimization
* Single-field queries to avoid composite indexes
* Application-level filtering for complex conditions
* Subcollection isolation for user data security

### 6.4. Indian Financial Market Support
* INR currency as default
* Support for NSE/BSE stock tickers
* Indian mutual fund and banking institutions
* FD and insurance product types specific to India

This backend implementation provides a robust foundation for the InvestMate Angular frontend while maintaining complete statelessness and Firebase integration.