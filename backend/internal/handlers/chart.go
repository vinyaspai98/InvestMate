package handlers

import (
	"context"
	"net/http"
	"time"

	"investmate-backend/internal/models"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/iterator"
)

type ChartHandler struct {
	firestoreClient *firestore.Client
}

func NewChartHandler(firestoreClient *firestore.Client) *ChartHandler {
	return &ChartHandler{
		firestoreClient: firestoreClient,
	}
}

// GetCategoryChart gets chart data for a specific category with time period filtering
func (h *ChartHandler) GetCategoryChart(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	category := c.Param("category")
	period := c.Param("period")

	if category == "" || period == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Category and period parameters are required"})
		return
	}

	// Validate period
	if !isValidPeriod(period) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid period. Use: 1M, 3M, 1Y, All"})
		return
	}

	chartData, err := h.generateCategoryChart(c.Request.Context(), userID, models.InvestmentCategory(category), period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate chart data"})
		return
	}

	c.JSON(http.StatusOK, chartData)
}

// GetPortfolioChart gets overall portfolio chart data with time period filtering
func (h *ChartHandler) GetPortfolioChart(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	period := c.Param("period")
	if period == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Period parameter is required"})
		return
	}

	// Validate period
	if !isValidPeriod(period) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid period. Use: 1M, 3M, 1Y, All"})
		return
	}

	chartData, err := h.generatePortfolioChart(c.Request.Context(), userID, period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate portfolio chart data"})
		return
	}

	c.JSON(http.StatusOK, chartData)
}

// Helper methods

func (h *ChartHandler) generateCategoryChart(ctx context.Context, userID string, category models.InvestmentCategory, period string) (*models.ChartResponse, error) {
	// Get time range for the period
	endDate := time.Now()
	startDate := getStartDateForPeriod(period, endDate)

	// Get transactions for the category within the time range
	transactions, err := h.getTransactionsInPeriod(ctx, userID, category, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Get current investments for the category
	investments, err := h.getInvestmentsByCategory(ctx, userID, category)
	if err != nil {
		return nil, err
	}

	// Generate chart data points
	dataPoints := h.calculateValueOverTime(transactions, investments, startDate, endDate, period)

	return &models.ChartResponse{
		Period: period,
		Data:   dataPoints,
	}, nil
}

func (h *ChartHandler) generatePortfolioChart(ctx context.Context, userID string, period string) (*models.ChartResponse, error) {
	// Get time range for the period
	endDate := time.Now()
	startDate := getStartDateForPeriod(period, endDate)

	// Get all transactions within the time range
	allTransactions, err := h.getAllTransactionsInPeriod(ctx, userID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Get all current investments
	allInvestments, err := h.getAllUserInvestments(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get all loans
	allLoans, err := h.getAllUserLoans(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Generate portfolio chart data points (net worth over time)
	dataPoints := h.calculateNetWorthOverTime(allTransactions, allInvestments, allLoans, startDate, endDate, period)

	return &models.ChartResponse{
		Period: period,
		Data:   dataPoints,
	}, nil
}

func (h *ChartHandler) getTransactionsInPeriod(ctx context.Context, userID string, category models.InvestmentCategory, startDate, endDate time.Time) ([]models.Transaction, error) {
	var transactions []models.Transaction

	iter := h.firestoreClient.Collection("users").Doc(userID).Collection("transactions").
		Where("category", "==", string(category)).
		Where("date", ">=", startDate).
		Where("date", "<=", endDate).
		OrderBy("date", firestore.Asc).
		Documents(ctx)

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var transaction models.Transaction
		if err := doc.DataTo(&transaction); err != nil {
			return nil, err
		}
		transaction.ID = doc.Ref.ID
		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

func (h *ChartHandler) getAllTransactionsInPeriod(ctx context.Context, userID string, startDate, endDate time.Time) ([]models.Transaction, error) {
	var transactions []models.Transaction

	iter := h.firestoreClient.Collection("users").Doc(userID).Collection("transactions").
		Where("date", ">=", startDate).
		Where("date", "<=", endDate).
		OrderBy("date", firestore.Asc).
		Documents(ctx)

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var transaction models.Transaction
		if err := doc.DataTo(&transaction); err != nil {
			return nil, err
		}
		transaction.ID = doc.Ref.ID
		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

func (h *ChartHandler) getInvestmentsByCategory(ctx context.Context, userID string, category models.InvestmentCategory) ([]models.Investment, error) {
	var investments []models.Investment

	iter := h.firestoreClient.Collection("users").Doc(userID).Collection("investments").
		Where("category", "==", string(category)).
		Where("isActive", "==", true).
		Documents(ctx)

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

func (h *ChartHandler) getAllUserInvestments(ctx context.Context, userID string) ([]models.Investment, error) {
	var investments []models.Investment

	iter := h.firestoreClient.Collection("users").Doc(userID).Collection("investments").
		Where("isActive", "==", true).
		Documents(ctx)

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

func (h *ChartHandler) getAllUserLoans(ctx context.Context, userID string) ([]models.Loan, error) {
	var loans []models.Loan

	iter := h.firestoreClient.Collection("users").Doc(userID).Collection("loans").
		Where("isActive", "==", true).
		Documents(ctx)

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

func (h *ChartHandler) calculateValueOverTime(transactions []models.Transaction, investments []models.Investment, startDate, endDate time.Time, period string) []models.ChartDataPoint {
	var dataPoints []models.ChartDataPoint

	// For simplicity, we'll create data points based on current investment values
	// In a real implementation, you'd calculate historical values based on transactions

	intervals := getTimeIntervals(startDate, endDate, period)

	// Calculate current total value for the category
	var totalCurrentValue float64
	for _, investment := range investments {
		totalCurrentValue += investment.CurrentValue
	}

	// Generate sample data points (in production, this would be calculated from historical data)
	for i, interval := range intervals {
		// Simulate value growth over time
		progressRatio := float64(i+1) / float64(len(intervals))
		value := totalCurrentValue * progressRatio

		dataPoints = append(dataPoints, models.ChartDataPoint{
			Date:  interval.Format("2006-01-02"),
			Value: value,
		})
	}

	return dataPoints
}

func (h *ChartHandler) calculateNetWorthOverTime(transactions []models.Transaction, investments []models.Investment, loans []models.Loan, startDate, endDate time.Time, period string) []models.ChartDataPoint {
	var dataPoints []models.ChartDataPoint

	intervals := getTimeIntervals(startDate, endDate, period)

	// Calculate current net worth
	var totalAssets, totalLiabilities float64
	for _, investment := range investments {
		totalAssets += investment.CurrentValue
	}
	for _, loan := range loans {
		totalLiabilities += loan.OutstandingBalance
	}
	currentNetWorth := totalAssets - totalLiabilities

	// Generate sample data points (in production, this would be calculated from historical data)
	for i, interval := range intervals {
		// Simulate net worth growth over time
		progressRatio := float64(i+1) / float64(len(intervals))
		netWorth := currentNetWorth * progressRatio

		dataPoints = append(dataPoints, models.ChartDataPoint{
			Date:  interval.Format("2006-01-02"),
			Value: netWorth,
		})
	}

	return dataPoints
}

// Utility functions

func isValidPeriod(period string) bool {
	validPeriods := []string{"1M", "3M", "1Y", "All"}
	for _, valid := range validPeriods {
		if period == valid {
			return true
		}
	}
	return false
}

func getStartDateForPeriod(period string, endDate time.Time) time.Time {
	switch period {
	case "1M":
		return endDate.AddDate(0, -1, 0)
	case "3M":
		return endDate.AddDate(0, -3, 0)
	case "1Y":
		return endDate.AddDate(-1, 0, 0)
	case "All":
		return endDate.AddDate(-10, 0, 0) // 10 years back as a reasonable "all" period
	default:
		return endDate.AddDate(0, -3, 0) // Default to 3 months
	}
}

func getTimeIntervals(startDate, endDate time.Time, period string) []time.Time {
	var intervals []time.Time

	var step time.Duration
	switch period {
	case "1M":
		step = 24 * time.Hour * 7 // Weekly intervals for 1 month
	case "3M":
		step = 24 * time.Hour * 7 // Weekly intervals for 3 months
	case "1Y":
		step = 24 * time.Hour * 30 // Monthly intervals for 1 year
	case "All":
		step = 24 * time.Hour * 30 // Monthly intervals
	default:
		step = 24 * time.Hour * 7 // Default weekly
	}

	current := startDate
	for current.Before(endDate) || current.Equal(endDate) {
		intervals = append(intervals, current)
		current = current.Add(step)
	}

	// Ensure we have at least a few data points
	if len(intervals) < 5 {
		duration := endDate.Sub(startDate)
		step = duration / 5
		intervals = nil
		current = startDate
		for i := 0; i < 5; i++ {
			intervals = append(intervals, current)
			current = current.Add(step)
		}
	}

	return intervals
}
