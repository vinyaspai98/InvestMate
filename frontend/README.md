# InvestMate Frontend

A modern, responsive Single-Page Application (SPA) built with Angular 20 and Angular Material, tailored for tracking personal finances and investment portfolios in the Indian market.

## Overview

InvestMate Frontend provides an intuitive, card-based interface for tracking multiple asset classes (Stocks, Mutual Funds, Fixed Deposits, Insurance, Loans) with real-time portfolio valuation, interactive Chart.js analytics, dynamic dialogs for manual entry, and automated synchronization with CDSL Demat transaction emails via Google OAuth and Gemini AI.

## Key Features

### 📊 Portfolio Dashboard
- **Net Worth Overview**: Real-time calculation of total assets, liabilities, and overall net worth.
- **Category Summary Cards**: Visual cards for Stocks, Mutual Funds, Fixed Deposits, Insurance Policies, and Loans displaying invested amounts, current values, and profit/loss metrics.
- **Quick Navigation**: Seamless drill-down from the dashboard into individual category detail views.

### 📈 Detailed Investment Views & Analytics
- **Interactive Performance Charts**: Line charts powered by `Chart.js` and `ng2-charts` with dynamic time-frame filters (`1M`, `3M`, `1Y`, `All`).
- **Holdings & Transaction Tables**: Angular Material data tables showing individual investments, buy prices, current valuations, P&L percentages, and historical transactions.
- **Asset-Specific Metrics**: Custom metrics for each asset type (e.g. tickers & quantities for stocks, folio numbers & NAV for mutual funds, interest rates & maturity dates for FDs, coverage & premiums for insurance, outstanding balances for loans).

### ➕ Dynamic Add & Edit Investment Dialogs
- **Category-Adaptive Forms**: Reactive forms (`AddInvestmentDialogComponent`) that adjust required fields based on the selected asset class.
- **Input Validation**: Strict validation for amounts, rates, dates, and required metadata.
- **CRUD Operations**: Support for adding new investments and editing existing holdings.

### 🔄 Gmail Sync for CDSL Transactions
- **One-Click Sync**: Trigger Gmail Demat transaction synchronization directly from the Profile page.
- **Automated Processing**: Imports buy/sell transactions sent by CDSL (`services@cdslindia.co.in`) parsed by backend Gemini AI.
- **Real-Time Status**: Visual indicators, spinner animations, and snackbar alerts indicating synced transaction counts.

### 🔐 Authentication & Security
- **Google Sign-In**: Authenticate using Google OAuth with requested `https://www.googleapis.com/auth/gmail.readonly` scope.
- **Automatic Token Injection**: Functional HTTP interceptor (`authInterceptor`) automatically injects Firebase JWT Bearer tokens into all backend API calls.
- **Route Guards**: `AuthGuard` protects dashboard, investments, and profile routes against unauthenticated access.

### 🎨 User Experience & Theming
- **Dark & Light Mode**: Built-in theme switcher service with persistent user preferences.
- **Responsive Layout**: Sidebar navigation and responsive grid layout adaptable across mobile, tablet, and desktop screens.

---

## Technology Stack

- **Framework**: [Angular 20](https://angular.dev/) (Standalone Components, TypeScript)
- **UI Components**: [Angular Material](https://material.angular.io/) (Cards, Buttons, Tables, Dialogs, Chips, Snackbars)
- **State & Reactivity**: [RxJS](https://rxjs.dev/)
- **Charts**: [Chart.js](https://www.chartjs.org/) & [ng2-charts](https://github.com/valor-software/ng2-charts)
- **Authentication**: [AngularFire](https://github.com/angular/angularfire) / Firebase Auth SDK
- **Styling**: SCSS with custom Angular Material palettes

---

## Project Structure

```
frontend/
├── src/
│   ├── app/
│   │   ├── components/
│   │   │   ├── add-investment-dialog/   # Modal for adding & editing investments
│   │   │   └── layout/                  # Main shell layout with sidebar and header
│   │   ├── guards/
│   │   │   └── auth.guard.ts            # Angular route protection guard
│   │   ├── interceptors/
│   │   │   └── auth.interceptor.ts      # HTTP interceptor injecting Firebase JWT
│   │   ├── models/
│   │   │   ├── investment.model.ts      # Investment, category, loan & chart interfaces
│   │   │   └── user.model.ts            # User profile and auth interfaces
│   │   ├── pages/
│   │   │   ├── about/                   # About InvestMate page
│   │   │   ├── auth/                    # Login and Signup pages (Google OAuth)
│   │   │   ├── dashboard/               # Main portfolio summary dashboard
│   │   │   ├── investment-detail/       # Category detail view with charts and tables
│   │   │   └── profile/                 # User preferences and Gmail sync trigger
│   │   ├── services/
│   │   │   ├── auth.service.ts          # Firebase Auth & Google Sign-In service
│   │   │   ├── gmail.service.ts         # Backend Gmail sync trigger service
│   │   │   ├── investment.service.ts    # REST API client for investments & charts
│   │   │   ├── mock-data.service.ts     # Offline mock data fallback
│   │   │   └── theme.service.ts         # Dark/Light mode theme state management
│   │   ├── app.config.ts                # Application providers & Firebase initialization
│   │   ├── app.routes.ts                # Route definitions with guards
│   │   └── app.ts                       # Root application component
│   ├── environments/
│   │   ├── environment.ts               # Development configuration
│   │   └── environment.prod.ts          # Production configuration
│   ├── index.html                       # HTML root template
│   ├── main.ts                          # Application bootstrap entry point
│   └── styles.scss                      # Global styles and Material theme overrides
├── angular.json                         # Angular CLI configuration
├── package.json                         # Dependencies and build scripts
└── tsconfig.json                        # TypeScript compiler options
```

---

## Getting Started

### Prerequisites
- **Node.js**: v18.0.0 or higher (v20+ recommended)
- **npm**: v9.0.0 or higher
- **Angular CLI**: Installed globally (`npm install -g @angular/cli`) or run via `npx`

### 1. Install Dependencies
```bash
cd frontend
npm install
```

### 2. Configure Environment
Check `src/environments/environment.ts` (and `environment.prod.ts` for production builds) and verify your Firebase project credentials and backend API URL:

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

### 3. Start Development Server
```bash
npm start
# or: ng serve
```

Navigate to `http://localhost:4200/`. The app will automatically reload on any source code changes.

### 4. Build for Production
```bash
npm run build
```
Production artifacts will be compiled into the `dist/` directory with ahead-of-time (AOT) compilation and optimization.

### 5. Running Tests
```bash
npm test
```
Executes unit tests via the Karma test runner.

---

## Core Services Reference

| Service | Responsibility |
|---|---|
| `AuthService` | Handles Google Sign-In with popup, acquires Gmail OAuth access token, syncs with backend `/auth/verify-token`, and maintains current user state. |
| `InvestmentService` | Communicates with Go backend REST APIs for net worth, categories, loans, transactions, and historical chart data with RxJS `shareReplay` caching. |
| `GmailService` | Calls backend `POST /api/v1/gmail/sync` to trigger CDSL email retrieval and Gemini AI transaction parsing. |
| `ThemeService` | Toggles between light and dark themes by applying CSS classes to the document body. |
| `AuthInterceptor` | Asynchronously resolves the current Firebase user ID token and appends `Authorization: Bearer <token>` to all HTTP requests. |
