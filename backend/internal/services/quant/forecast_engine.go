package quant

import (
	"math"
	"math/rand"
	"sort"
	"time"

	"investmate-backend/internal/models"
)

const (
	UnoptimizedCAGR = 0.102 // 10.2% p.a. (high TER bleed, tax drag)
	OptimizedCAGR   = 0.141 // 14.1% p.a. (net spread +3.9% p.a.)
	DefaultYears    = 20
	DefaultBase     = 1000000.0 // ₹10,00,000 base capital
)

// Calculate20YearForecast computes wealth milestone projections comparing unoptimized vs optimized portfolios.
func Calculate20YearForecast(initialBase float64, years int) []models.ProfitForecastPoint {
	if initialBase <= 0 {
		initialBase = DefaultBase
	}
	if years <= 0 {
		years = DefaultYears
	}

	points := make([]models.ProfitForecastPoint, years)

	for y := 1; y <= years; y++ {
		unoptVal := initialBase * math.Pow(1+UnoptimizedCAGR, float64(y))
		unoptProfit := unoptVal - initialBase

		optVal := initialBase * math.Pow(1+OptimizedCAGR, float64(y))
		optProfit := optVal - initialBase

		delta := optProfit - unoptProfit

		points[y-1] = models.ProfitForecastPoint{
			Year:                 y,
			UnoptimizedValue:     math.Round(unoptVal*100) / 100,
			UnoptimizedNetProfit: math.Round(unoptProfit*100) / 100,
			OptimizedValue:       math.Round(optVal*100) / 100,
			OptimizedNetProfit:   math.Round(optProfit*100) / 100,
			NetProfitDelta:       math.Round(delta*100) / 100,
		}
	}

	return points
}

// MonteCarloConfig contains parameters for Geometric Brownian Motion simulation.
type MonteCarloConfig struct {
	InitialBase float64
	Years       int
	Simulations int
	Mu          float64 // Expected annual return (drift)
	Sigma       float64 // Annual volatility
}

// RunMonteCarloSimulation executes a stochastic GBM simulation of wealth paths.
func RunMonteCarloSimulation(cfg MonteCarloConfig) models.MonteCarloResponse {
	if cfg.InitialBase <= 0 {
		cfg.InitialBase = DefaultBase
	}
	if cfg.Years <= 0 {
		cfg.Years = DefaultYears
	}
	if cfg.Simulations <= 0 {
		cfg.Simulations = 1000
	}
	if cfg.Mu <= 0 {
		cfg.Mu = OptimizedCAGR // 14.1% drift for AI-optimized portfolio
	}
	if cfg.Sigma <= 0 {
		cfg.Sigma = 0.155 // 15.5% annual volatility (standard Indian equity index)
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	dt := 1.0 // 1-year discrete steps
	driftStep := (cfg.Mu - 0.5*cfg.Sigma*cfg.Sigma) * dt
	volStep := cfg.Sigma * math.Sqrt(dt)

	// Matrix of paths: [Simulations][Years+1]
	allPaths := make([][]float64, cfg.Simulations)
	survivedCount := 0

	for s := 0; s < cfg.Simulations; s++ {
		path := make([]float64, cfg.Years+1)
		path[0] = cfg.InitialBase
		curVal := cfg.InitialBase

		for t := 1; t <= cfg.Years; t++ {
			// Box-Muller standard normal random number
			z := r.NormFloat64()
			curVal = curVal * math.Exp(driftStep+volStep*z)
			if curVal < 0 {
				curVal = 0
			}
			path[t] = math.Round(curVal*100) / 100
		}

		allPaths[s] = path
		if path[cfg.Years] >= cfg.InitialBase {
			survivedCount++
		}
	}

	survivalRate := (float64(survivedCount) / float64(cfg.Simulations)) * 100.0
	survivalRate = math.Round(survivalRate*10) / 10

	// Calculate P10, Median (P50), P90 across all steps
	p10 := make([]float64, cfg.Years+1)
	median := make([]float64, cfg.Years+1)
	p90 := make([]float64, cfg.Years+1)

	for t := 0; t <= cfg.Years; t++ {
		valsAtT := make([]float64, cfg.Simulations)
		for s := 0; s < cfg.Simulations; s++ {
			valsAtT[s] = allPaths[s][t]
		}
		sort.Float64s(valsAtT)

		idx10 := int(float64(cfg.Simulations) * 0.10)
		idx50 := int(float64(cfg.Simulations) * 0.50)
		idx90 := int(float64(cfg.Simulations) * 0.90)

		p10[t] = math.Round(valsAtT[idx10]*100) / 100
		median[t] = math.Round(valsAtT[idx50]*100) / 100
		p90[t] = math.Round(valsAtT[idx90]*100) / 100
	}

	// Select representative sample paths (e.g. 50 paths) for frontend WebGL rendering
	sampleCount := 50
	if sampleCount > cfg.Simulations {
		sampleCount = cfg.Simulations
	}
	sampleStep := cfg.Simulations / sampleCount
	sampleTrajectories := make([][]float64, 0, sampleCount)

	for i := 0; i < sampleCount; i++ {
		sampleTrajectories = append(sampleTrajectories, allPaths[i*sampleStep])
	}

	return models.MonteCarloResponse{
		Trajectories: sampleTrajectories,
		SurvivalRate: survivalRate,
		Percentile10: p10,
		Median:       median,
		Percentile90: p90,
		Years:        cfg.Years,
	}
}
