# InvestMate - Pending Tasks

**Project:** InvestMate - Financial Tracking Application  
**Last Updated:** December 13, 2025  
**Status:** In Development

---

## 🔴 High Priority - Core Functionality

### Backend Integration
- [x] Replace mock data services with actual API calls in `frontend/src/app/services/mock-data.service.ts`
- [x] Implement HTTP interceptor for adding JWT tokens to API requests
- [x] Create environment configuration files for API endpoints (`environment.ts`, `environment.prod.ts`)
- [ ] Add error handling and retry logic for failed API calls
- [ ] Implement data caching strategy using RxJS `shareReplay` or similar

### Firebase Authentication
- [x] Replace mock authentication in `frontend/src/app/services/auth.service.ts` with Firebase Auth SDK
- [x] Implement proper token refresh mechanism
- [x] Add Firebase configuration in `frontend/src/app/app.config.ts`
- [x] Handle Firebase authentication errors with user-friendly messages
- [ ] Implement email verification flow
- [x] Complete password reset functionality (currently stub in `AuthService`)

### Investment Detail Pages
- [x] Create `InvestmentDetailComponent` (route defined in `app.routes.ts` but component missing)
- [x] Implement line charts with time-series filters (1M, 3M, 1Y, All) using Chart.js or ngx-charts
- [x] Build transaction history table with sorting and filtering
- [x] Create FAB (Floating Action Button) for "Add Investment"
- [x] Implement modal/dialog forms for adding/editing investments with category-specific fields:
  - [x] Stock investment form (with ticker, quantity fields)
  - [x] Mutual fund form (with SIP/Lumpsum options)
  - [x] FD form (with interest rate, maturity date)
  - [x] Insurance form (with premium, coverage fields)
  - [x] Loan form (with EMI, tenure fields)

### Data Models & Validation
- [ ] Add form validation for all investment types based on `Investment` model
- [ ] Implement proper TypeScript interfaces for API responses
- [ ] Add client-side validation matching backend validation rules from `backend/internal/models/investment.go`

---

## 🟡 Medium Priority - User Experience

### Profile Management
- [ ] Implement actual profile update logic in `ProfileComponent.onUpdateProfile()`
- [ ] Connect theme toggle in preferences to `ThemeService`
- [ ] Implement password change functionality in `ProfileComponent.onChangePassword()`
- [ ] Add profile photo upload and display
- [ ] Save user preferences to Firestore

### Dashboard Enhancements
- [ ] Add real-time data refresh capability to `DashboardComponent`
- [ ] Implement category card animations and transitions
- [ ] Add loading skeletons for better perceived performance
- [ ] Create quick action buttons as per wireframe specifications
- [ ] Add summary statistics trends (up/down indicators)
- [ ] Implement Net Worth calculation formula: `(Investments + FDs + Insurance Value) - (Total Loans)`

### Navigation & Routing
- [ ] Enhance `AuthGuard` with role-based access control
- [ ] Add breadcrumb navigation component
- [ ] Implement route preloading strategy
- [ ] Add navigation state management for deep linking

### Responsive Design
- [ ] Test and optimize mobile layouts for all pages
- [ ] Improve `LayoutComponent` drawer behavior on tablets
- [ ] Add touch gestures for mobile navigation
- [ ] Optimize chart rendering for small screens

---

## 🟢 Low Priority - Polish & Enhancement

### Visualizations
- [ ] Integrate Chart.js or similar library for line charts
- [ ] Implement chart data point tooltips
- [ ] Add chart export functionality (PNG/PDF)
- [ ] Create donut/pie charts for asset allocation
- [ ] Add comparison charts between categories
- [ ] Implement historical tracking visualization

### About Page Functionality
- [ ] Connect "Contact Support" button in `AboutComponent` to actual email service
- [ ] Implement "Report an Issue" functionality
- [ ] Add FAQ section
- [ ] Create changelog/version history display

### Theme & Styling
- [ ] Ensure all Material components support dark mode
- [ ] Optimize SCSS variables for consistent theming
- [ ] Add custom color schemes beyond default Material palette
- [ ] Implement smooth theme transition animations
- [ ] Create print-friendly styles for reports
- [ ] Align all components with wireframe design specifications

---

## 🔵 Future Enhancements

### Advanced Features
- [ ] Implement bulk import/export functionality (CSV, Excel)
- [ ] Add transaction search and advanced filtering
- [ ] Create investment performance analytics dashboard
- [ ] Implement goal tracking and planning features
- [ ] Add tax calculation and reporting module
- [ ] Support for multiple portfolios per user

### Real-time Data Integration
- [ ] Integrate with stock market APIs (NSE/BSE) for live prices
- [ ] Add mutual fund NAV auto-fetch
- [ ] Implement push notifications for price alerts
- [ ] Create watchlist functionality
- [ ] Add automatic portfolio value updates

### Social & Collaboration
- [ ] Add portfolio sharing capabilities
- [ ] Implement comparison with benchmark indices (Nifty, Sensex)
- [ ] Create investment insights and recommendations engine
- [ ] Add community features or discussion forums

### Backend Enhancements
- [ ] Implement real-time stock price integration
- [ ] Add advanced portfolio analytics calculations
- [ ] Create email notification service
- [ ] Build data export functionality
- [ ] Add investment performance calculations
- [ ] Implement multi-currency support (with INR as default)
- [ ] Create bulk data import/export APIs
- [ ] Add transaction categorization and tagging

---

## 🔒 Security & Performance

### Security
- [ ] Implement rate limiting on API endpoints
- [ ] Add request/response encryption for sensitive data
- [ ] Implement audit logging for sensitive operations
- [ ] Add input sanitization and XSS protection
- [ ] Implement CSRF protection
- [ ] Add security headers (CSP, HSTS, etc.)

### Performance
- [ ] Implement lazy loading for heavy modules
- [ ] Add service workers for offline functionality
- [ ] Optimize bundle size and implement code splitting
- [ ] Add proper caching strategies (browser cache, HTTP cache)
- [ ] Implement virtual scrolling for long lists
- [ ] Optimize database queries and indexing

---

## 🧪 Testing & Quality Assurance

### Unit Tests
- [ ] Add unit tests for all services (`auth.service.ts`, `mock-data.service.ts`, `theme.service.ts`)
- [ ] Add unit tests for all components
- [ ] Test backend handlers (`auth.go`, `dashboard.go`, `investment.go`, etc.)
- [ ] Achieve minimum 80% code coverage

### Integration Tests
- [ ] Test API integration between frontend and backend
- [ ] Test Firebase authentication flow
- [ ] Test data persistence with Firestore

### E2E Tests
- [ ] Create E2E tests using Cypress or Playwright
- [ ] Test complete user flows (signup, login, add investment, view dashboard)
- [ ] Test cross-browser compatibility

---

## 🚀 DevOps & Deployment

### CI/CD
- [ ] Set up CI/CD pipeline (GitHub Actions, GitLab CI, etc.)
- [ ] Implement automated testing in pipeline
- [ ] Add automated deployment to staging environment
- [ ] Implement blue-green deployment strategy

### Containerization & Orchestration
- [ ] Create Docker Compose for local development
- [ ] Optimize Dockerfile for production (backend already has Dockerfile)
- [ ] Add health check endpoints
- [ ] Implement logging and monitoring

### Deployment
- [ ] Configure Firebase hosting deployment for frontend
- [ ] Deploy backend to cloud platform (Google Cloud Run, AWS, etc.)
- [ ] Set up monitoring and error tracking (Sentry, DataDog, etc.)
- [ ] Implement automated backup strategies
- [ ] Set up CDN for static assets

---

## 📋 Documentation Tasks

### Code Documentation
- [ ] Complete API documentation for all backend endpoints
- [ ] Add JSDoc comments to all service methods
- [ ] Document component APIs and usage examples
- [ ] Add inline code comments for complex logic

### User Documentation
- [ ] Create user guide/help documentation
- [ ] Add onboarding tutorial for new users
- [ ] Create FAQ section
- [ ] Document all features with screenshots

### Developer Documentation
- [ ] Write deployment and maintenance guides
- [ ] Create architecture decision records (ADRs)
- [ ] Document development setup process
- [ ] Add contribution guidelines
- [ ] Create API integration guide for third-party developers

---

## 🔧 Technical Debt & Refactoring

### Code Quality
- [ ] Remove any unused dependencies from `package.json` and `go.mod`
- [ ] Add proper TypeScript strict mode compliance
- [ ] Implement proper error boundary components
- [ ] Refactor any duplicate code patterns
- [ ] Follow DRY (Don't Repeat Yourself) principles

### Accessibility
- [ ] Add proper ARIA labels to all interactive elements
- [ ] Implement keyboard navigation for all features
- [ ] Ensure proper color contrast ratios
- [ ] Add screen reader support
- [ ] Test with accessibility tools (axe, Lighthouse)

### Updates & Maintenance
- [ ] Update to latest Angular version
- [ ] Update to latest Go version
- [ ] Update all npm packages to latest stable versions
- [ ] Review and update Firebase SDK versions
- [ ] Update Material Angular to latest version

---

## 📊 Analytics & Monitoring

- [ ] Implement user analytics (Google Analytics, Mixpanel)
- [ ] Add error tracking and reporting
- [ ] Monitor API performance and response times
- [ ] Track user engagement metrics
- [ ] Implement application performance monitoring (APM)
- [ ] Set up alerts for critical errors

---

## 💼 Business & Compliance

- [ ] Add privacy policy page
- [ ] Add terms of service page
- [ ] Implement GDPR compliance features (data export, deletion)
- [ ] Add cookie consent banner
- [ ] Implement data retention policies

---

## Notes

- **Design Reference:** All UI implementations must strictly follow the wireframe design at:  
  https://uxpilot.ai/p/rf2g9IfblZbgRC9hijl0?wireframeId=T7Ep1G7T084jZfKIEyTe&fullscreen=true

- **Priority Order:** Complete High Priority tasks first to establish core functionality, then Medium Priority for enhanced UX, followed by Low Priority and Future Enhancements.

- **Current State:** 
  - Frontend structure is in place with routing and basic components
  - Backend has basic handlers and models defined
  - Currently using mock data for development
  - Authentication is stubbed but not integrated with Firebase
  - Investment detail pages are not yet implemented

- **Next Immediate Steps:**
  1. Implement Firebase Authentication integration
  2. Create Investment Detail Component with charts
  3. Replace mock data with API integration
  4. Implement form modals for adding investments

---

**Legend:**
- 🔴 High Priority - Critical for MVP
- 🟡 Medium Priority - Important for good UX
- 🟢 Low Priority - Nice to have
- 🔵 Future Enhancements - Post-MVP features
