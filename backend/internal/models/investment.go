package models

import (
	"time"
)

// Investment represents an investment document in Firestore
type Investment struct {
	ID             string             `json:"id" firestore:"-"`
	Category       InvestmentCategory `json:"category" firestore:"category"`
	Name           string             `json:"name" firestore:"name"`
	InvestedAmount float64            `json:"investedAmount" firestore:"investedAmount"`
	CurrentValue   float64            `json:"currentValue" firestore:"currentValue"`
	CreatedAt      time.Time          `json:"createdAt" firestore:"createdAt"`
	UpdatedAt      time.Time          `json:"updatedAt" firestore:"updatedAt"`
	IsActive       bool               `json:"isActive" firestore:"isActive"`
	Notes          string             `json:"notes" firestore:"notes"`
	InvestmentData InvestmentData     `json:"investmentData" firestore:"investmentData"`
}

// InvestmentData contains category-specific investment data
type InvestmentData struct {
	// Stock specific fields
	Ticker       string  `json:"ticker,omitempty" firestore:"ticker,omitempty"`
	Quantity     float64 `json:"quantity,omitempty" firestore:"quantity,omitempty"`
	AveragePrice float64 `json:"averagePrice,omitempty" firestore:"averagePrice,omitempty"`
	CurrentPrice float64 `json:"currentPrice,omitempty" firestore:"currentPrice,omitempty"`
	Exchange     string  `json:"exchange,omitempty" firestore:"exchange,omitempty"`

	// Mutual Fund specific fields
	FundName string  `json:"fundName,omitempty" firestore:"fundName,omitempty"`
	NAV      float64 `json:"nav,omitempty" firestore:"nav,omitempty"`
	Units    float64 `json:"units,omitempty" firestore:"units,omitempty"`

	// FD/Insurance specific fields
	InterestRate float64    `json:"interestRate,omitempty" firestore:"interestRate,omitempty"`
	MaturityDate *time.Time `json:"maturityDate,omitempty" firestore:"maturityDate,omitempty"`
	Tenure       int        `json:"tenure,omitempty" firestore:"tenure,omitempty"` // in months

	// Insurance specific fields
	PremiumAmount float64 `json:"premiumAmount,omitempty" firestore:"premiumAmount,omitempty"`
	SumAssured    float64 `json:"sumAssured,omitempty" firestore:"sumAssured,omitempty"`
	PolicyNumber  string  `json:"policyNumber,omitempty" firestore:"policyNumber,omitempty"`
	InsuranceType string  `json:"insuranceType,omitempty" firestore:"insuranceType,omitempty"`
}

// Loan represents a loan document in Firestore
type Loan struct {
	ID                 string    `json:"id" firestore:"-"`
	LoanType           string    `json:"loanType" firestore:"loanType"`
	LenderName         string    `json:"lenderName" firestore:"lenderName"`
	PrincipalAmount    float64   `json:"principalAmount" firestore:"principalAmount"`
	OutstandingBalance float64   `json:"outstandingBalance" firestore:"outstandingBalance"`
	InterestRate       float64   `json:"interestRate" firestore:"interestRate"`
	EMI                float64   `json:"emi" firestore:"emi"`
	Tenure             int       `json:"tenure" firestore:"tenure"`                   // total tenure in months
	RemainingTenure    int       `json:"remainingTenure" firestore:"remainingTenure"` // remaining months
	StartDate          time.Time `json:"startDate" firestore:"startDate"`
	MaturityDate       time.Time `json:"maturityDate" firestore:"maturityDate"`
	CreatedAt          time.Time `json:"createdAt" firestore:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt" firestore:"updatedAt"`
	IsActive           bool      `json:"isActive" firestore:"isActive"`
	Notes              string    `json:"notes" firestore:"notes"`
}

// Transaction represents a transaction document in Firestore
type Transaction struct {
	ID          string             `json:"id" firestore:"-"`
	RelatedID   string             `json:"relatedId" firestore:"relatedId"` // investment or loan ID
	Category    InvestmentCategory `json:"category" firestore:"category"`
	Type        TransactionType    `json:"type" firestore:"type"`
	Amount      float64            `json:"amount" firestore:"amount"`
	Quantity    float64            `json:"quantity,omitempty" firestore:"quantity,omitempty"`
	Price       float64            `json:"price,omitempty" firestore:"price,omitempty"`
	Date        time.Time          `json:"date" firestore:"date"`
	Description string             `json:"description" firestore:"description"`
	Fees        float64            `json:"fees,omitempty" firestore:"fees,omitempty"`
	CreatedAt   time.Time          `json:"createdAt" firestore:"createdAt"`
	Notes       string             `json:"notes" firestore:"notes"`
}

// Enums
type InvestmentCategory string

const (
	CategoryStocks      InvestmentCategory = "stocks"
	CategoryMutualFunds InvestmentCategory = "mutualFunds"
	CategoryFDs         InvestmentCategory = "fds"
	CategoryInsurance   InvestmentCategory = "insurance"
	CategoryLoans       InvestmentCategory = "loans"
)

type TransactionType string

const (
	TransactionTypeBuy        TransactionType = "buy"
	TransactionTypeSell       TransactionType = "sell"
	TransactionTypeDeposit    TransactionType = "deposit"
	TransactionTypeWithdrawal TransactionType = "withdrawal"
	TransactionTypePayment    TransactionType = "payment"
)

// Dashboard response models
type NetWorthSummary struct {
	TotalAssets      float64 `json:"totalAssets"`
	TotalLiabilities float64 `json:"totalLiabilities"`
	NetWorth         float64 `json:"netWorth"`
	Currency         string  `json:"currency"`
}

type CategorySummary struct {
	InvestedAmount       float64 `json:"investedAmount"`
	CurrentValue         float64 `json:"currentValue"`
	ProfitLossPercentage float64 `json:"profitLossPercentage"`
	Count                int     `json:"count"`
}

type LoanSummary struct {
	PrincipalAmount    float64 `json:"principalAmount"`
	OutstandingBalance float64 `json:"outstandingBalance"`
	Count              int     `json:"count"`
}

type DashboardSummary struct {
	NetWorth   NetWorthSummary                        `json:"netWorth"`
	Categories map[InvestmentCategory]CategorySummary `json:"categories"`
	Loans      LoanSummary                            `json:"loans"`
}

// Chart data models
type ChartDataPoint struct {
	Date  string  `json:"date"`
	Value float64 `json:"value"`
}

type ChartResponse struct {
	Period string           `json:"period"`
	Data   []ChartDataPoint `json:"data"`
}

// Request models for API endpoints
type CreateInvestmentRequest struct {
	Category       InvestmentCategory `json:"category" validate:"required,oneof=stocks mutualFunds fds insurance"`
	Name           string             `json:"name" validate:"required,min=1"`
	InvestedAmount float64            `json:"investedAmount" validate:"required,gt=0"`
	CurrentValue   float64            `json:"currentValue" validate:"required,gte=0"`
	Notes          string             `json:"notes"`
	InvestmentData InvestmentData     `json:"investmentData"`
}

type UpdateInvestmentRequest struct {
	Name           *string         `json:"name,omitempty" validate:"omitempty,min=1"`
	InvestedAmount *float64        `json:"investedAmount,omitempty" validate:"omitempty,gt=0"`
	CurrentValue   *float64        `json:"currentValue,omitempty" validate:"omitempty,gte=0"`
	Notes          *string         `json:"notes,omitempty"`
	InvestmentData *InvestmentData `json:"investmentData,omitempty"`
}

type CreateLoanRequest struct {
	LoanType           string    `json:"loanType" validate:"required,oneof=home_loan personal_loan car_loan education_loan business_loan"`
	LenderName         string    `json:"lenderName" validate:"required,min=1"`
	PrincipalAmount    float64   `json:"principalAmount" validate:"required,gt=0"`
	OutstandingBalance float64   `json:"outstandingBalance" validate:"required,gte=0"`
	InterestRate       float64   `json:"interestRate" validate:"required,gt=0"`
	EMI                float64   `json:"emi" validate:"required,gt=0"`
	Tenure             int       `json:"tenure" validate:"required,gt=0"`
	RemainingTenure    int       `json:"remainingTenure" validate:"required,gte=0"`
	StartDate          time.Time `json:"startDate" validate:"required"`
	MaturityDate       time.Time `json:"maturityDate" validate:"required"`
	Notes              string    `json:"notes"`
}

type UpdateLoanRequest struct {
	LenderName         *string  `json:"lenderName,omitempty" validate:"omitempty,min=1"`
	OutstandingBalance *float64 `json:"outstandingBalance,omitempty" validate:"omitempty,gte=0"`
	InterestRate       *float64 `json:"interestRate,omitempty" validate:"omitempty,gt=0"`
	EMI                *float64 `json:"emi,omitempty" validate:"omitempty,gt=0"`
	RemainingTenure    *int     `json:"remainingTenure,omitempty" validate:"omitempty,gte=0"`
	Notes              *string  `json:"notes,omitempty"`
}

type CreateTransactionRequest struct {
	RelatedID   string             `json:"relatedId" validate:"required"`
	Category    InvestmentCategory `json:"category" validate:"required,oneof=stocks mutualFunds fds insurance loans"`
	Type        TransactionType    `json:"type" validate:"required,oneof=buy sell deposit withdrawal payment"`
	Amount      float64            `json:"amount" validate:"required,gt=0"`
	Quantity    float64            `json:"quantity,omitempty" validate:"omitempty,gt=0"`
	Price       float64            `json:"price,omitempty" validate:"omitempty,gt=0"`
	Date        time.Time          `json:"date" validate:"required"`
	Description string             `json:"description" validate:"required,min=1"`
	Fees        float64            `json:"fees,omitempty" validate:"omitempty,gte=0"`
	Notes       string             `json:"notes"`
}
