# InvestMate - Financial Tracking Application

A comprehensive financial tracking Single-Page Application (SPA) designed specifically for the Indian financial market.

## Overview

InvestMate helps Indian investors track their complete financial portfolio including stocks, mutual funds, fixed deposits, insurance policies, and loans. The application provides a clean, card-based dashboard with real-time net worth calculations and detailed performance analytics.

## Features

### Core Functionality
- **Dashboard**: Overview of total net worth and category-wise investments
- **Portfolio Tracking**: Track stocks, mutual funds, FDs, insurance, and loans
- **Performance Analytics**: Profit/loss calculations and percentage tracking
- **Transaction History**: Detailed transaction logs for each category
- **Responsive Design**: Works seamlessly on desktop and mobile devices

### User Experience
- **Dark/Light Theme**: Toggle between dark and light modes
- **Material Design**: Clean, modern UI using Angular Material
- **Sidebar Navigation**: Easy access to all sections
- **Card-based Layout**: Information organized in intuitive cards
- **Interactive Charts**: Visual representation of performance (placeholder implemented)

### Security
- **Authentication**: Login/Signup with Firebase integration
- **Route Protection**: Protected routes with authentication guards
- **User Preferences**: Personalized settings and preferences

## Technology Stack

### Frontend
- **Angular 20+** with TypeScript
- **Angular Material** for UI components
- **RxJS** for reactive programming
- **SCSS** for styling
- **Chart.js** (ready for integration)

### Backend (Structure Created)
- **GoLang** with Gin framework
- **Firebase** for authentication and database
- **RESTful APIs** for data management
- **JWT** authentication
- **CORS** enabled for frontend integration

### Development Tools
- **Angular CLI** for project management
- **TypeScript** for type safety
- **ESLint** for code quality
- **Responsive design** with CSS Grid and Flexbox

## Project Structure

```
InvestMate1/
├── frontend/                 # Angular frontend application
│   ├── src/
│   │   ├── app/
│   │   │   ├── components/   # Reusable components
│   │   │   │   └── layout/   # Main layout with sidebar
│   │   │   ├── pages/        # Page components
│   │   │   │   ├── auth/     # Login/Signup pages
│   │   │   │   ├── dashboard/# Main dashboard
│   │   │   │   ├── investment-detail/ # Category details
│   │   │   │   ├── profile/  # User profile
│   │   │   │   └── about/    # About page
│   │   │   ├── services/     # Angular services
│   │   │   ├── models/       # TypeScript interfaces
│   │   │   └── guards/       # Route guards
│   │   └── styles.scss       # Global styles
├── backend/                  # GoLang backend (structure)
│   ├── cmd/                  # Application entry points
│   ├── internal/             # Private application code
│   │   ├── handlers/         # HTTP handlers
│   │   ├── models/           # Data models
│   │   └── middleware/       # HTTP middleware
│   └── pkg/                  # Public packages
└── .github/                  # Project documentation
```

## Getting Started

### Prerequisites
- Node.js (v18 or higher)
- Angular CLI (`npm install -g @angular/cli`)
- Go (v1.21 or higher) - for backend
- Firebase account (for production deployment)

### Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd InvestMate1
   ```

2. **Frontend Setup**
   ```bash
   cd frontend
   npm install
   ng serve
   ```
   The application will be available at `http://localhost:4200`

3. **Backend Setup** (Optional - for API integration)
   ```bash
   cd backend
   go mod tidy
   go run cmd/main.go
   ```
   The API will be available at `http://localhost:8080`

### Demo Credentials
For testing the application, use these demo credentials:
- **Email**: user@investmate.com
- **Password**: password

## Current Implementation Status

### ✅ Completed Features
- [x] Project setup with Angular and Go
- [x] Authentication module (Login/Signup)
- [x] Responsive sidebar navigation
- [x] Dashboard with net worth calculation
- [x] Investment detail pages for all categories
- [x] Profile management
- [x] Dark/Light theme system
- [x] Mocked data services
- [x] Responsive design
- [x] About page

### 🚧 In Progress
- [ ] Backend API integration
- [ ] Firebase authentication setup
- [ ] Chart visualizations
- [ ] Add/Edit investment modals

### 📋 Future Enhancements
- [ ] Real-time market data integration
- [ ] Push notifications
- [ ] Export functionality
- [ ] Advanced analytics
- [ ] Mobile app development

## Key Components

### Dashboard
- Net worth header card with total assets and liabilities
- Category summary cards with profit/loss calculations
- Quick action buttons for common tasks
- Navigation to detailed category views

### Investment Categories
- **Stocks**: Individual stock holdings with P&L tracking
- **Mutual Funds**: SIP and lump sum investments
- **Fixed Deposits**: Interest rate and maturity tracking
- **Insurance**: Policy value and premium tracking
- **Loans**: Outstanding balance and EMI tracking

### Data Models
- Investment entities with category-specific fields
- Transaction history for all operations
- User preferences and profile information
- Category summaries with aggregated metrics

## Development Guidelines

### Code Style
- Follow Angular style guide
- Use TypeScript strict mode
- Implement responsive design patterns
- Follow Material Design principles

### Data Flow
- Services handle all data operations
- Observables for reactive programming
- Mock data services for development
- Prepared for API integration

### Security
- Route guards for protected pages
- Authentication state management
- Input validation and sanitization
- HTTPS enforcement (production)

## Configuration

### Environment Variables
- `FIREBASE_CONFIG`: Firebase configuration
- `API_BASE_URL`: Backend API URL
- `ENVIRONMENT`: Development/Production mode

### Theme Customization
The application supports extensive theming through SCSS variables and Angular Material theme system.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## Support

For questions, issues, or suggestions:
- Create an issue in the repository
- Contact the development team
- Check the documentation in `.github/copilot-instructions.md`

## License

This project is developed for educational and demonstration purposes.

---

**InvestMate** - Your trusted financial companion for tracking investments in the Indian market.