package quant

import (
	"testing"
)

func TestCalculate20YearForecast_ValuesAndDelta(t *testing.T) {
	base := 1000000.0 // 10 Lakhs
	points := Calculate20YearForecast(base, 20)

	if len(points) != 20 {
		t.Fatalf("expected 20 forecast points, got %d", len(points))
	}

	for i, pt := range points {
		if pt.Year != i+1 {
			t.Errorf("expected year %d, got %d", i+1, pt.Year)
		}
		if pt.OptimizedValue <= pt.UnoptimizedValue {
			t.Errorf("year %d: expected optimized value > unoptimized value", pt.Year)
		}
		if pt.NetProfitDelta <= 0 {
			t.Errorf("year %d: expected positive net profit delta", pt.Year)
		}
	}

	// At Year 20:
	// Unoptimized: 1M * (1.102)^20 = ~6.96M (Net profit ~5.96M)
	// Optimized: 1M * (1.141)^20 = ~13.78M (Net profit ~12.78M)
	// Delta should be > 6M
	y20 := points[19]
	if y20.NetProfitDelta < 6000000.0 {
		t.Errorf("year 20 net profit delta too low: %.2f", y20.NetProfitDelta)
	}
}

func TestRunMonteCarloSimulation_ConvergenceAndSurvival(t *testing.T) {
	resp := RunMonteCarloSimulation(MonteCarloConfig{
		InitialBase: 1000000.0,
		Years:       20,
		Simulations: 1000,
	})

	if resp.Years != 20 {
		t.Errorf("expected 20 years, got %d", resp.Years)
	}
	if len(resp.Trajectories) == 0 {
		t.Errorf("expected non-empty trajectories")
	}
	if len(resp.Median) != 21 {
		t.Errorf("expected 21 points for median (0 to 20), got %d", len(resp.Median))
	}
	if len(resp.Percentile10) != 21 {
		t.Errorf("expected 21 points for p10, got %d", len(resp.Percentile10))
	}
	if len(resp.Percentile90) != 21 {
		t.Errorf("expected 21 points for p90, got %d", len(resp.Percentile90))
	}

	// Verify order: P10 <= Median <= P90 at year 20
	if resp.Percentile10[20] > resp.Median[20] {
		t.Errorf("p10 should be <= median, got p10=%.2f, median=%.2f", resp.Percentile10[20], resp.Median[20])
	}
	if resp.Median[20] > resp.Percentile90[20] {
		t.Errorf("median should be <= p90, got median=%.2f, p90=%.2f", resp.Median[20], resp.Percentile90[20])
	}

	// Survival rate with 14.1% drift over 20 years should be high (>85%)
	if resp.SurvivalRate < 80.0 {
		t.Errorf("expected survival rate >= 80%%, got %.2f%%", resp.SurvivalRate)
	}
}
