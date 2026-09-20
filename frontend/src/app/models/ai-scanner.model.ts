export interface TaxLot {
  id: string;
  user_id: string;
  isin: string;
  scheme_name: string;
  asset_type: 'Equity' | 'Debt' | 'Gold' | 'Arbitrage';
  units: number;
  purchase_nav: number;
  purchase_date: string;
  total_cost: number;
  is_ltcg: boolean;
  unrealized_gain: number;
}

export interface ActiveShareAudit {
  isin: string;
  scheme_name: string;
  benchmark_index: string;
  active_share_score: number;
  current_ter: number;
  target_index_ter: number;
  is_closet_indexer: boolean;
  annual_fee_bleed: number;
}

export interface PortfolioHealthScore {
  overall_score: number;
  active_share_score: number;
  tax_efficiency_score: number;
  ulcer_index_score: number;
  factor_balance_score: number;
  fee_control_score: number;
  audit_details: ActiveShareAudit[];
  source?: string;
  transaction_count?: number;
  holdings_count?: number;
}

export interface ProfitForecastPoint {
  year: number;
  unoptimized_value: number;
  unoptimized_net_profit: number;
  optimized_value: number;
  optimized_net_profit: number;
  net_profit_delta: number;
}

export interface RecommendationAction {
  asset_category: string;
  current_asset: string;
  current_allocation: number;
  target_asset: string;
  target_allocation: number;
  ter_reduction: number;
  tax_efficiency_status: string;
  action_trigger: string;
}

export interface MonteCarloResponse {
  trajectories: number[][];
  survival_rate: number;
  percentile_10: number[];
  median: number[];
  percentile_90: number[];
  years: number;
}
