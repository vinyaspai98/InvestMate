# ANTIGRAVITY AGENT EXECUTION BLUEPRINT

## AI Portfolio Scanner, Market Benchmarking & 20-Year Forecast Integration

**Target System:** InvestMate (Go Backend + GCP Firestore + Gemini AI + Angular Frontend)

## 1. REPOSITORY ARCHITECTURE & CAPABILITY AUDIT

### 1.1 Existing System Topology

* **Backend Stack:** Go (Golang) microservice utilizing Fiber/Gin REST HTTP handlers (`backend/cmd/main.go`).

* **Database & Storage:** Google Cloud Firestore SDK (`cloud.google.com/go/firestore`) with security rules in `backend/firestore.rules`.

* **AI Infrastructure:** Google Gemini API integration (`github.com/google/generative-ai-go/genai`) with system instructions stored in `backend/internal/handlers/prompts/gemini_cdsl_system_instruction.txt`.

* **OAuth & Gmail Sync:** Gmail OAuth token handlers (`backend/internal/handlers/gmail.go`) parsing attached e-CAS CDSL/NSDL PDF statements.

* **Frontend Stack:** Angular 18/19 SPA with Tailwind CSS, Chart.js, and Reactive Forms located in `frontend/src/app/`.

### 1.2 Identified Architectural Gaps

1. **Unstructured Data Ingestion:** Transactions are logged, but cost basis is not organized into an immutable FIFO (First-In, First-Out) tax lot ledger needed for capital gains harvesting.

2. **Lack of Active Share Audit:** Current fund tracking lacks benchmark comparison algorithms to detect closet-indexer funds charging active TER fees (1.5%-2.0%) with Active Share < 60%.

3. **Outdated Tax Assumptions:** Lacks compliance with India's Post-Budget 2024 tax regime (FY 2025-26 rules: ₹1.25L LTCG exemption, 12.5% LTCG rate, 20% STCG rate, and debt fund taxation shifts).

4. **Absence of Long-Term Forecasting:** No 20-year wealth projection model or stochastic Monte Carlo risk simulator to highlight net returns delta between unoptimized holdings and AI-recommended assets.

## 2. ANTIGRAVITY AGENT ROLES & WORKFLOW DIVISION

To execute this feature set efficiently, four specialized Antigravity agents will work in parallel with strict integration contracts.

```
[Agent Alpha: Go & AI Core] -------> [Go REST Endpoints & Firestore Schema]
                                                  |
[Agent Beta: Quant Engine] --------> [FIFO Ledger, Active Share & Tax Engine]
                                                  |
[Agent Gamma: Angular Frontend] ---> [Dashboard Components & Plotly/Chart.js]
                                                  |
[Agent Delta: Integration & QA] ---> [E2E Testing, Security & Benchmarks]

```

### Agent 1: Agent Alpha (Backend & AI Infrastructure Specialist)

* **Domain:** `backend/cmd/`, `backend/internal/handlers/`, `backend/pkg/config/`

* **Key Responsibilities:**

  1. Extend Gemini prompt templates for multi-page CAS parsing, ISIN extraction, and fund name normalization.

  2. Implement Go HTTP REST routes for portfolio scans, recommendations, and forecast generation.

  3. Manage Firestore collection schemas and atomic transaction writes for tax lots and health audits.

### Agent 2: Agent Beta (Quantitative & Tax Math Engine Specialist)

* **Domain:** `backend/internal/services/quant/`, `backend/internal/models/`

* **Key Responsibilities:**

  1. Build the immutable FIFO tax lot calculator with Post-Budget 2024 India tax harvesting rules.

  2. Develop the Active Share calculator and benchmark weight cross-matching index.

  3. Implement the Black-Litterman asset allocation optimizer and Geometric Brownian Motion (GBM) Monte Carlo simulator.

### Agent 3: Agent Gamma (Angular UI/UX & Data Visualization Specialist)

* **Domain:** `frontend/src/app/pages/`, `frontend/src/app/components/`, `frontend/src/app/services/`

* **Key Responsibilities:**

  1. Design responsive Tailwind Angular components for AI Portfolio Scanning, Forecast Graphs, and Comparison Matrices.

  2. Implement Chart.js Canvas visualizations with mandatory 16-character label line wrapping and tooltip title callbacks.

  3. Build 3D/WebGL Plotly Monte Carlo sequence-of-returns stress test visualizations.

### Agent 4: Agent Delta (Integration, Testing & Security Specialist)

* **Domain:** `backend/internal/handlers/*_test.go`, E2E test suites, deployment scripts

* **Key Responsibilities:**

  1. Create mock CAS datasets and validate FIFO tax lot accuracy against edge cases (e.g., partial sales, SIP bonus issues).

  2. Verify CORS, OAuth scopes, and Firestore security rules to prevent cross-tenant data leakage.

  3. Perform performance profiling to ensure API responses resolve in under 1.2 seconds.

## 3. GO DATA STRUCTURES & FIRESTORE SCHEMAS

### 3.1 Go Struct Definitions (`backend/internal/models/ai_scanner.go`)

```
package models

import "time"

// TaxLot represents an immutable FIFO tax purchase lot for a asset.
type TaxLot struct {
	ID            string    `json:"id" firestore:"id"`
	UserID        string    `json:"user_id" firestore:"user_id"`
	ISIN          string    `json:"isin" firestore:"isin"`
	SchemeName    string    `json:"scheme_name" firestore:"scheme_name"`
	AssetType     string    `json:"asset_type" firestore:"asset_type"` // Equity, Debt, Gold, Arbitrage
	Units         float64   `json:"units" firestore:"units"`
	PurchaseNAV   float64   `json:"purchase_nav" firestore:"purchase_nav"`
	PurchaseDate  time.Time `json:"purchase_date" firestore:"purchase_date"`
	TotalCost     float64   `json:"total_cost" firestore:"total_cost"`
	IsLTCG        bool      `json:"is_ltcg" firestore:"is_ltcg"`       // >12 months for equity
	UnrealizedGain float64  `json:"unrealized_gain" firestore:"unrealized_gain"`
}

// ActiveShareAudit stores benchmark overlap and fee drag analysis.
type ActiveShareAudit struct {
	ISIN             string  `json:"isin"`
	SchemeName       string  `json:"scheme_name"`
	BenchmarkIndex   string  `json:"benchmark_index"`   // Nifty 50, Nifty Next 50, Midcap 150
	ActiveShareScore float64 `json:"active_share_score"` // Percentage 0.0 - 100.0
	CurrentTER       float64 `json:"current_ter"`         // e.g. 1.85%
	TargetIndexTER   float64 `json:"target_index_ter"`   // e.g. 0.08%
	IsClosetIndexer  bool    `json:"is_closet_indexer"`  // True if ActiveShare < 60.0 && TER > 1.0%
	AnnualFeeBleed   float64 `json:"annual_fee_bleed"`
}

// PortfolioHealthScore aggregates portfolio diagnostics (0 - 100 scale).
type PortfolioHealthScore struct {
	OverallScore      int                `json:"overall_score"`
	ActiveShareScore  int                `json:"active_share_score"`
	TaxEfficiencyScore int               `json:"tax_efficiency_score"`
	UlcerIndexScore   int                `json:"ulcer_index_score"`
	FactorBalanceScore int               `json:"factor_balance_score"`
	FeeControlScore   int                `json:"fee_control_score"`
	AuditDetails      []ActiveShareAudit `json:"audit_details"`
}

// ProfitForecastPoint represents projected wealth at a timeline milestone.
type ProfitForecastPoint struct {
	Year                   int     `json:"year"`
	UnoptimizedValue       float64 `json:"unoptimized_value"`
	UnoptimizedNetProfit   float64 `json:"unoptimized_net_profit"`
	OptimizedValue         float64 `json:"optimized_value"`
	OptimizedNetProfit     float64 `json:"optimized_net_profit"`
	NetProfitDelta         float64 `json:"net_profit_delta"`
}

// RecommendationAction defines side-by-side asset migration steps.
type RecommendationAction struct {
	AssetCategory      string  `json:"asset_category"`
	CurrentAsset       string  `json:"current_asset"`
	CurrentAllocation  float64 `json:"current_allocation"`
	TargetAsset        string  `json:"target_asset"`
	TargetAllocation   float64 `json:"target_allocation"`
	TERReduction       float64 `json:"ter_reduction"`
	TaxEfficiencyStatus string `json:"tax_efficiency_status"`
	ActionTrigger      string  `json:"action_trigger"` // e.g., "STP over 6 months", "Immediate Harvest"
}

```

### 3.2 Firestore Collection Schema Hierarchy

```
users/ {userId}
  │
  ├── portfolio_scans/ {scanId}
  │     ├── scanned_at: Timestamp
  │     ├── health_score: Map (OverallScore, ActiveShareScore, etc.)
  │     └── summary: Map (TotalValue, TERBleed, PotentialTaxSavings)
  │
  ├── fifo_tax_lots/ {lotId}
  │     ├── isin: String
  │     ├── scheme_name: String
  │     ├── units: Number
  │     ├── purchase_nav: Number
  │     ├── purchase_date: Timestamp
  │     └── is_ltcg: Boolean
  │
  ├── ai_recommendations/ {recId}
  │     ├── generated_at: Timestamp
  │     ├── actions: Array of RecommendationAction
  │     └── tax_alpha_projected: Number
  │
  └── forecast_cache/ {forecastId}
        ├── base_capital: Number
        ├── timeline_years: Number (20)
        └── forecast_points: Array of ProfitForecastPoint

```

## 4. REST API SPECIFICATIONS

| **Method** | **Endpoint Route** | **Request Payload** | **Response Data Payload** | **Description** | 
| `POST` | `/api/v1/ai/scan-cas` | `{ "file_bytes": "...", "password": "..." }` | `PortfolioHealthScore` struct | Ingests e-CAS PDF, builds FIFO lots, and runs closet-index audit. | 
| `GET` | `/api/v1/ai/health-score` | None (Auth Token) | `PortfolioHealthScore` struct | Retrieves cached diagnostic health breakdown for user. | 
| `POST` | `/api/v1/ai/recommendations` | `{ "target_risk": "moderate" }` | `[]RecommendationAction` | Generates side-by-side rebalancing plan using Black-Litterman rules. | 
| `POST` | `/api/v1/ai/forecast-profit` | `{ "initial_base": 1000000, "years": 20 }` | `[]ProfitForecastPoint` | Computes 20-year wealth projections comparing current vs target portfolios. | 
| `GET` | `/api/v1/ai/monte-carlo` | `?simulations=1000&years=20` | `{ "trajectories": [[...]], "survival_rate": 94.8 }` | Generates WebGL stochastic simulation trajectories for drawdowns. | 

## 5. MATHEMATICAL & ALGORITHMIC FORMULAS

### 5.1 Active Share Audit Calculation

Active Share ($AS$) measures the proportion of portfolio holdings that differ from the benchmark index:

$$
AS = \frac{1}{2} \sum_{i=1}^{N} \left\vert{} w_{i, \text{fund}} - w_{i, \text{index}} \right\vert{}
$$

* **Threshold Flag:** If $AS < 60.0\%$ and Fund TER $> 1.0\%$, flag as a **Closet Indexer**. Recommend migrating to Direct Index ETFs ($TER \approx 0.08\%$).

### 5.2 Post-Budget 2024 India Tax Loss Harvesting Rules (FY 2025-26)

1. **Equity LTCG Harvesting:**

   * Exemption Threshold: Up to ₹1,25,000 of LTCG per financial year is tax-free.

   * Tax Rate above threshold: $12.5\%$.

   * **Algorithm:** Identify long-term equity lots with unrealized gains $\le ₹125,000$. Harvest by selling and immediately re-investing to raise cost basis.

2. **Short-Term Capital Loss (STCL) Matching:**

   * Match STCL against Short-Term Capital Gains (STCG taxed at $20\%$) to reduce net liability.

3. **Debt Allocation Shift:**

   * Taxable debt funds (taxed at slab rate) are systematically migrated to **Arbitrage Funds** (taxed as equity: $12.5\%$ LTCG) and **Sovereign Gold Bonds (SGBs)** (tax-free upon maturity).

### 5.3 20-Year Compound Financial Growth Equation

Net portfolio accumulation over $t$ years considering expense ratio savings ($\Delta TER$) and tax alpha recovery ($\alpha_{\text{tax}}$):

$$
V(t) = P_0 \times \prod_{k=1}^{t} \left(1 + R_{\text{gross}} - TER_{\text{target}} + \alpha_{\text{tax}}\right)
$$

* **Unoptimized CAGR (**$R_{\text{net, old}}$**):** $10.2\%$ p.a. (high TER bleed, unharvested gains, tax drag).

* **Optimized AI CAGR (**$R_{\text{net, target}}$**):** $14.1\%$ p.a. (net spread of $+3.9\%$ p.a.).

## 6. PHASED EXECUTION ROADMAP FOR ANTIGRAVITY AGENTS

```
Phase 1: Ingestion & FIFO Lot Engine  [Backend Go + Firestore]
  └── Phase 2: Active Share & Audit   [Quant Math Engine]
        └── Phase 3: Tax Alpha & BL   [Rebalancing Engine]
              └── Phase 4: Forecast   [Monte Carlo Engine]
                    └── Phase 5: UI   [Angular Frontend Components]

```

### Phase 1: Multimodal Ingestion & FIFO Lot Ledger (Week 1)

* **Agent Alpha:** Update `backend/internal/handlers/gmail.go` and `prompts/gemini_cdsl_system_instruction.txt` to parse multi-page transactions, extract ISIN codes, and write normalized records to Firestore.

* **Agent Beta:** Implement `backend/internal/services/quant/fifo_ledger.go` to compute cost basis, track purchase dates, and mark LTCG eligibility based on holding period.

* **Deliverable:** Fully functional `/api/v1/ai/scan-cas` endpoint populating `fifo_tax_lots`.

### Phase 2: Active Share Audit & Health Diagnostics (Week 2)

* **Agent Beta:** Implement `backend/internal/services/quant/active_share.go`. Load index constituent weights for Nifty 50, Nifty Next 50, and Nifty Midcap 150. Calculate $AS$ scores and annual fee bleed.

* **Agent Alpha:** Expose `/api/v1/ai/health-score` endpoint returning the 5-dimension diagnostic radar payload (Active Share, Tax Efficiency, Ulcer Index, Factor Balance, TER Control).

* **Deliverable:** Active Share audit engine correctly identifying closet indexers.

### Phase 3: Post-Budget Tax Alpha & Rebalancing Engine (Week 3)

* **Agent Beta:** Implement `backend/internal/services/quant/tax_rebalancer.go`. Implement Black-Litterman optimization combined with Gemini market sentiment inputs. Apply 5/25 threshold drift models for cash-flow SIP rebalancing.

* **Agent Alpha:** Expose `/api/v1/ai/recommendations` route returning actionable side-by-side asset migration steps.

* **Deliverable:** Tax harvesting and Black-Litterman asset rebalancing API endpoints.

### Phase 4: 20-Year Profit Forecast & Monte Carlo Simulation (Week 4)

* **Agent Beta:** Build `backend/internal/services/quant/forecast_engine.go`. Implement multi-year compound interest calculations and Geometric Brownian Motion (GBM) Monte Carlo path simulations (1000+ paths).

* **Agent Alpha:** Expose `/api/v1/ai/forecast-profit` and `/api/v1/ai/monte-carlo` endpoints.

* **Deliverable:** 20-year wealth projections and stress-test survival analytics.

### Phase 5: Angular UI Components & Interactive Charts (Week 5)

* **Agent Gamma:**

  1. Create `frontend/src/app/pages/ai-scanner/ai-scanner.component.ts` displaying KPI summary cards, radar health charts, and side-by-side asset tables.

  2. Implement Chart.js instances for competitor benchmarks, long-term profit forecasts, donut allocations, and stacked tax alpha savings. Apply 16-character label wrapping logic and tooltip title callbacks to all chart instances.

  3. Integrate Plotly WebGL scatter plot (`scattergl`) for the 20-year Monte Carlo trajectory chart.

* **Agent Delta:** Perform unit testing, cross-browser responsiveness checks, and E2E integration tests across all features.

## 7. READY-TO-DISPATCH ANTIGRAVITY AGENT PROMPTS

### Prompt for Agent Alpha (Backend & AI Infrastructure)

> "You are Agent Alpha. Read `backend/cmd/main.go`, `backend/internal/handlers/gmail.go`, and `backend/internal/handlers/prompts/gemini_cdsl_system_instruction.txt`. Extend the Go backend by adding `/api/v1/ai/scan-cas`, `/api/v1/ai/health-score`, `/api/v1/ai/recommendations`, and `/api/v1/ai/forecast-profit`. Implement the Go structs defined in `ANTIGRAVITY_AGENT_PLAN.md` and manage atomic Firestore batch writes under `users/{userId}/portfolio_scans` and `users/{userId}/fifo_tax_lots`. Ensure strict error handling and auth middleware integration."

### Prompt for Agent Beta (Quantitative & Tax Math Specialist)

> "You are Agent Beta. Your task is to build the quantitative financial engines in Go inside `backend/internal/services/quant/`. Implement: 1) FIFO tax lot manager with Post-Budget 2024 India tax rules (FY 2025-26 ₹1.25L LTCG exemption, 12.5% LTCG, 20% STCG, Arbitrage fund shifts); 2) Active Share calculator using benchmark index constituents; 3) Black-Litterman asset allocation optimizer; and 4) 20-year compound forecast & Geometric Brownian Motion Monte Carlo simulation engine. Ensure all math functions pass unit tests with 100% precision."

### Prompt for Agent Gamma (Angular Frontend & UI/UX Developer)

> "You are Agent Gamma. Your task is to build the Angular frontend interface for the AI Portfolio Scanner and Forecast System inside `frontend/src/app/pages/ai-scanner/`. Create responsive Tailwind components for KPI metrics, side-by-side comparison tables, and diagnostic radar charts. Build Chart.js visualizations for profit forecasts, health scores, and tax savings, ensuring all string labels utilize 16-character array wrapping and proper tooltip title callbacks. Embed a Plotly WebGL `scattergl` chart for the 20-year Monte Carlo simulation."

### Prompt for Agent Delta (Integration, Security & QA Specialist)

> "You are Agent Delta. Review all code generated by Agents Alpha, Beta, and Gamma. Create mock e-CAS PDF test payloads, verify FIFO tax lot calculations against edge cases, ensure Firestore security rules (`backend/firestore.rules`) block unauthorized tenant access, and conduct end-to-end testing to ensure all API endpoints respond in under 1.2 seconds."

```

---

### Summary of Blueprint & Next Steps

This plan provides a clear, structured roadmap for Antigravity AI agents:
1. **Agent Roles & Execution Sequence:** Clear division between Go backend API development, quantitative math engines, Angular Tailwind UI components, and QA testing.
2. **Data & API Contracts:** Go structs, Firestore collection models, and REST endpoints for feature delivery.
3. **Quantitative Mathematical Rules:** Active Share formulas, Post-Budget 2024 tax loss harvesting algorithms, Black-Litterman asset allocation logic, and 20-year Monte Carlo simulation specifications.
4. **Agent Prompts:** Direct, ready-to-run prompts to orchestrate Antigravity agents.

```