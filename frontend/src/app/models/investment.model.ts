export interface Investment {
  id: string;
  category: InvestmentCategory;
  name: string;
  ticker?: string; // For stocks
  amount: number;
  currentValue: number;
  date: Date;
  notes?: string;
  interestRate?: number; // For FDs and loans
  tenor?: number; // For FDs and insurance
  maturityDate?: Date;
  outstandingBalance?: number; // For loans
}

export enum InvestmentCategory {
  STOCKS = 'stocks',
  MUTUAL_FUNDS = 'mutual_funds',
  FDS = 'fds',
  INSURANCE = 'insurance',
  LOANS = 'loans'
}

export interface CategorySummary {
  category: InvestmentCategory;
  totalInvested: number;
  currentValue: number;
  profitLossPercentage: number;
  profitLoss: number;
}

export interface NetWorth {
  totalAssets: number;
  totalLiabilities: number;
  netWorth: number;
}

export interface Transaction {
  id: string;
  investmentId: string;
  type: TransactionType;
  amount: number;
  date: Date;
  notes?: string;
}

export enum TransactionType {
  BUY = 'buy',
  SELL = 'sell',
  DEPOSIT = 'deposit',
  WITHDRAWAL = 'withdrawal',
  PAYMENT = 'payment' // For loans
}