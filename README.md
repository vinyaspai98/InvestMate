# InvestMate - Intelligent Financial Tracking Platform

A comprehensive, full-stack personal portfolio and financial tracking application designed specifically for the Indian financial market. InvestMate combines an Angular 20 Material frontend with a Go (Gin) backend and Google Gemini AI to automate portfolio tracking across stocks, mutual funds, fixed deposits, insurance, and loans.

---

## 🌟 Highlights & Key Features

### 🤖 AI-Powered Gmail Sync
- **Automated CDSL Ingestion**: Connects with the user's Gmail via Google OAuth to retrieve Demat transaction emails from `services@cdslindia.co.in` (*"Transactions In Your Demat Account"*).
- **Gemini AI Extraction**: Leverages Google Gemini AI with structured schema instructions to accurately extract stock and mutual fund transactions (ISIN, ticker, company/fund name, quantity, transaction date, buy/sell action).
- **Automated Market Pricing**: Resolves real-time stock prices (BSE/NSE) using Yahoo Finance, Indian mutual fund NAVs via `mfapi.in`, and market data via Alpha Vantage.
- **Intelligent Holdings Update**: Matches ISINs against existing holdings, computes weighted-average purchase prices on buys, and deducts quantities on sells.

### 📊 Comprehensive Multi-Asset Dashboard
- **Net Worth Tracking**: Live calculation of total assets, liabilities, and overall net worth.
- **Category Summary Cards**: Visual cards for Stocks, Mutual Funds, Fixed Deposits, Insurance Policies, and Loans showing invested capital, current valuation, and absolute/percentage P&L.
- **Category Detail Pages**: Dedicated asset views with detailed holdings tables, transaction histories, and asset-specific attributes.

### 📈 Interactive Performance Analytics
- **Time-Series Charts**: Interactive line charts powered by `Chart.js` and `ng2-charts` with period filtering (`1M`, `3M`, `1Y`, `All`).
- **Dynamic Asset Valuation**: Portfolio timelines generated based on transaction histories and current market prices.

### ➕ Dynamic Investment Management
- **Category-Adaptive Dialogs**: Add/Edit modal dialogs (`AddInvestmentDialogComponent`) with dynamic form controls tailored to each asset type (e.g. Ticker & Quantity for Stocks, Folio & NAV for Mutual Funds, Interest Rate & Maturity for FDs).
- **Manual & Automated Entry**: Freely add manual holdings or let the Gmail sync automatically populate your demat investments.

### 🔒 Enterprise-Grade Security & Authentication
- **Google OAuth & Firebase Auth**: One-click Google Sign-In with requested `gmail.readonly` permission.
- **JWT Verification**: Functional Angular HTTP interceptor attaches Firebase ID tokens to all backend API calls, verified on the Go server.
- **User-Isolated Database**: Subcollection architecture in Firebase Firestore ensuring total privacy and isolation between users.

### 🎨 Modern Material Experience
- **Dark & Light Themes**: Seamless switching between dark and light modes with persistent user preferences.
- **Responsive Layout**: Designed for desktops, tablets, and mobile devices with collapsible navigation and Angular Material components.

---

## 🛠️ Technology Stack

### Frontend
- **Framework**: Angular 20 (Standalone Components, TypeScript)
- **UI & Layout**: Angular Material (Cards, Tables, Dialogs, Chips, Snackbars) & SCSS
- **State & Reactivity**: RxJS (Observables, `shareReplay` caching)
- **Charts**: Chart.js 4.x & ng2-charts 8.x
- **Authentication**: AngularFire / Firebase Auth SDK

### Backend
- **Language & Runtime**: Go 1.24+
- **HTTP Framework**: Gin Web Framework
- **Database**: Google Cloud Firestore
- **AI & LLM**: Google Gemini API (Structured JSON extraction)
- **Email Ingestion**: Gmail API via OAuth2
- **Market Data**: Yahoo Finance API, mfapi.in NAV API, Alpha Vantage API

---

## 📁 Repository Structure

```
InvestMate/
├── frontend/                          # Angular 20 frontend SPA
│   ├── src/
│   │   ├── app/
│   │   │   ├── components/
│   │   │   │   ├── add-investment-dialog/ # Modal for adding/editing investments
│   │   │   │   └── layout/                # App layout, header, and sidebar
│   │   │   ├── guards/
│   │   │   │   └── auth.guard.ts          # Route protection guard
│   │   │   ├── interceptors/
│   │   │   │   └── auth.interceptor.ts    # Firebase JWT bearer token injector
│   │   │   ├── models/
│   │   │   │   ├── investment.model.ts    # Investment, category, loan & chart models
│   │   │   │   └── user.model.ts          # User and preferences models
│   │   │   ├── pages/
│   │   │   │   ├── about/                 # About page
│   │   │   │   ├── auth/                  # Login and Signup (Google OAuth)
│   │   │   │   ├── dashboard/             # Portfolio net worth dashboard
│   │   │   │   ├── investment-detail/     # Detailed category views & charts
│   │   │   │   └── profile/               # Profile settings & Gmail sync trigger
│   │   │   ├── services/
│   │   │   │   ├── auth.service.ts        # Firebase Auth & Google Sign-In
│   │   │   │   ├── gmail.service.ts       # Gmail sync trigger service
│   │   │   │   ├── investment.service.ts  # REST API client for investments & charts
│   │   │   │   └── theme.service.ts       # Dark/Light theme state
│   │   │   ├── app.config.ts              # Firebase & HTTP providers
│   │   │   └── app.routes.ts              # Application route declarations
│   │   ├── environments/                  # Environment configurations
│   │   └── styles.scss                    # Global styling and Material theme
│   └── package.json
│
├── backend/                           # Go REST API
│   ├── cmd/
│   │   └── main.go                    # Entry point & router configuration
│   ├── internal/
│   │   ├── constants/                 # Collection names and constants
│   │   ├── handlers/                  # HTTP route handlers
│   │   │   ├── auth.go                # Token verification & profile
│   │   │   ├── chart.go               # Time-series chart handlers
│   │   │   ├── dashboard.go           # Net worth & dashboard summaries
│   │   │   ├── gmail.go               # Gmail sync, CDSL parsing & Gemini AI
│   │   │   ├── investment.go          # Investment CRUD & category summaries
│   │   │   ├── loan.go                # Loan CRUD & EMI calculations
│   │   │   ├── transaction.go         # Transaction ledger handlers
│   │   │   └── prompts/               # Gemini prompt system instructions
│   │   ├── middleware/
│   │   │   └── auth.go                # Firebase JWT verification middleware
│   │   └── models/                    # Go structs for Firestore documents
│   ├── pkg/
│   │   └── config/                    # Config loader and Firebase initialization
│   ├── configs/                       # Service account keys (git-ignored)
│   ├── Dockerfile                     # Docker container definition
│   ├── go.mod                         # Go module dependencies
│   └── .env.example                   # Environment variable template
│
└── README.md                          # Main project documentation (this file)
```

---

## 🚀 Getting Started

### Prerequisites
- **Node.js**: v18.0.0 or higher
- **Angular CLI**: `npm install -g @angular/cli`
- **Go**: v1.24 or higher
- **Firebase Project**: Firestore and Authentication (Google provider) enabled
- **Google Cloud / Gemini API Key**: For AI email extraction
- **Alpha Vantage API Key**: (Optional) For fallback market quotes

---

### 1. Backend Setup

1. **Navigate to the backend directory**:
   ```bash
   cd backend
   cp .env.example .env
   ```

2. **Configure `.env`**:
   Fill in your Firebase project ID, credentials path, and API keys:
   ```env
   PORT=8080
   ENVIRONMENT=development
   FIREBASE_PROJECT_ID=your-firebase-project-id
   FIREBASE_CREDENTIALS_PATH=./configs/firebase-service-account.json
   CORS_ALLOWED_ORIGINS=http://localhost:4200,http://localhost:3000
   GEMINI_API_KEY=your-gemini-api-key
   ALPHA_VANTAGE_API_KEY=your-alpha-vantage-api-key
   ```

3. **Install Go packages and run**:
   ```bash
   go mod tidy
   go run cmd/main.go
   ```
   The backend API will run on `http://localhost:8080`.

---

### 2. Frontend Setup

1. **Navigate to the frontend directory**:
   ```bash
   cd frontend
   npm install
   ```

2. **Configure Environment**:
   Update `frontend/src/environments/environment.ts` with your Firebase web configuration:
   ```typescript
   export const environment = {
     production: false,
     apiBaseUrl: 'http://localhost:8080/api/v1',
     firebase: {
       apiKey: "YOUR_FIREBASE_API_KEY",
       authDomain: "YOUR_PROJECT_ID.firebaseapp.com",
       projectId: "YOUR_PROJECT_ID",
       storageBucket: "YOUR_PROJECT_ID.firebasestorage.app",
       messagingSenderId: "YOUR_SENDER_ID",
       appId: "YOUR_APP_ID",
       measurementId: "YOUR_MEASUREMENT_ID"
     }
   };
   ```

3. **Start Angular Dev Server**:
   ```bash
   npm start
   ```
   Access the application at `http://localhost:4200`.

---

## 📋 Implementation Status

### ✅ Completed & Active Features
- [x] Full-stack architecture with Angular 20 and Go 1.24+ (Gin)
- [x] Firebase Authentication with Google Sign-In and Gmail scopes
- [x] JWT token validation middleware and HTTP interceptor
- [x] User-isolated Firestore subcollections for data security
- [x] Portfolio Dashboard with live net worth, assets, and liabilities calculations
- [x] Category detail pages for Stocks, Mutual Funds, Fixed Deposits, Insurance, and Loans
- [x] Interactive performance line charts (1M, 3M, 1Y, All) with Chart.js & ng2-charts
- [x] Dynamic Add & Edit investment modals with category-specific validation
- [x] Automated Gmail sync for CDSL Demat transaction emails
- [x] Google Gemini AI integration for structured email transaction parsing
- [x] Real-time market pricing via Yahoo Finance (BSE/NSE) and Indian MF NAVs via mfapi.in
- [x] ISIN matching with automated purchase averaging and sell transaction processing
- [x] Dark & Light theme switching with Angular Material
- [x] Profile management and manual Gmail sync trigger UI

### 🔮 Future Roadmap
- [ ] Automated scheduled background sync for Gmail transactions
- [ ] CAMS & KFintech Mutual Fund CAS statement PDF parsing
- [ ] Dividend tracking and automated payout logging
- [ ] CSV/Excel portfolio import and export
- [ ] Mobile application using Ionic or Flutter

---

## 📄 License

This project is licensed for educational and personal financial management use.