package handlers

import (
	"context"
	"net/http"

	"investmate-backend/internal/models"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/iterator"
)

type DashboardHandler struct {
	firestoreClient *firestore.Client
}

func NewDashboardHandler(firestoreClient *firestore.Client) *DashboardHandler {
	return &DashboardHandler{
		firestoreClient: firestoreClient,
	}
}

// GetDashboardSummary returns the dashboard summary with net worth and category summaries
func (h *DashboardHandler) GetDashboardSummary(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get investments summary
	investments, err := h.getAllUserInvestments(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get investments"})
		return
	}

	// Get loans summary
	loans, err := h.getAllUserLoans(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get loans"})
		return
	}

	// Calculate category summaries
	categorySummaries := h.calculateCategorySummaries(investments)

	// Calculate loan summary
	loanSummary := h.calculateLoanSummary(loans)

	// Calculate net worth
	netWorth := h.calculateNetWorth(categorySummaries, loanSummary)

	// Build dashboard response
	dashboard := models.DashboardSummary{
		NetWorth:   netWorth,
		Categories: categorySummaries,
		Loans:      loanSummary,
	}

	c.JSON(http.StatusOK, dashboard)
}

// Helper methods

func (h *DashboardHandler) getAllUserInvestments(ctx context.Context, userID string) ([]models.Investment, error) {
	var investments []models.Investment

	iter := h.firestoreClient.Collection("users").Doc(userID).Collection("investments").
		Where("isActive", "==", true).Documents(ctx)

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var investment models.Investment
		if err := doc.DataTo(&investment); err != nil {
			return nil, err
		}
		investment.ID = doc.Ref.ID
		investments = append(investments, investment)
	}

	return investments, nil
}

func (h *DashboardHandler) getAllUserLoans(ctx context.Context, userID string) ([]models.Loan, error) {
	var loans []models.Loan

	iter := h.firestoreClient.Collection("users").Doc(userID).Collection("loans").
		Where("isActive", "==", true).Documents(ctx)

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var loan models.Loan
		if err := doc.DataTo(&loan); err != nil {
			return nil, err
		}
		loan.ID = doc.Ref.ID
		loans = append(loans, loan)
	}

	return loans, nil
}

func (h *DashboardHandler) calculateCategorySummaries(investments []models.Investment) map[models.InvestmentCategory]models.CategorySummary {
	summaries := make(map[models.InvestmentCategory]models.CategorySummary)

	// Initialize all categories
	categories := []models.InvestmentCategory{
		models.CategoryStocks,
		models.CategoryMutualFunds,
		models.CategoryFDs,
		models.CategoryInsurance,
	}

	for _, category := range categories {
		summaries[category] = models.CategorySummary{
			InvestedAmount:       0,
			CurrentValue:         0,
			ProfitLossPercentage: 0,
			Count:                0,
		}
	}

	// Calculate summaries
	for _, investment := range investments {
		if investment.Category == models.CategoryLoans {
			continue // Skip loans as they're handled separately
		}

		summary := summaries[investment.Category]
		summary.InvestedAmount += investment.InvestedAmount
		summary.CurrentValue += investment.CurrentValue
		summary.Count++

		// Calculate profit/loss percentage
		if summary.InvestedAmount > 0 {
			profitLoss := summary.CurrentValue - summary.InvestedAmount
			summary.ProfitLossPercentage = (profitLoss / summary.InvestedAmount) * 100
		}

		summaries[investment.Category] = summary
	}

	return summaries
}

func (h *DashboardHandler) calculateLoanSummary(loans []models.Loan) models.LoanSummary {
	var totalPrincipal, totalOutstanding float64
	count := len(loans)

	for _, loan := range loans {
		totalPrincipal += loan.PrincipalAmount
		totalOutstanding += loan.OutstandingBalance
	}

	return models.LoanSummary{
		PrincipalAmount:    totalPrincipal,
		OutstandingBalance: totalOutstanding,
		Count:              count,
	}
}

func (h *DashboardHandler) calculateNetWorth(categorySummaries map[models.InvestmentCategory]models.CategorySummary, loanSummary models.LoanSummary) models.NetWorthSummary {
	var totalAssets float64

	// Sum all investment categories
	for _, summary := range categorySummaries {
		totalAssets += summary.CurrentValue
	}

	totalLiabilities := loanSummary.OutstandingBalance
	netWorth := totalAssets - totalLiabilities

	return models.NetWorthSummary{
		TotalAssets:      totalAssets,
		TotalLiabilities: totalLiabilities,
		NetWorth:         netWorth,
		Currency:         "INR",
	}
}
