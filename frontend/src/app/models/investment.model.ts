export interface Investment {
  id?: string;
  category: InvestmentCategory;
  amount: number;
  currentValue: number;
  date: string | Date;
  notes?: string;
  profitLoss?: number;
  profitLossPercentage?: number;
  
  // Stock-specific fields
  ticker?: string;
  name?: string;
  quantity?: number;
  buyPrice?: number;
  currentPrice?: number;
  
  // Mutual Fund-specific fields
  folioNumber?: string;
  units?: number;
  nav?: number;
  
  // FD-specific fields
  bankName?: string;
  fdNumber?: string;
  interestRate?: number;
  maturityDate?: string | Date;
  maturityAmount?: number;
  
  // Insurance-specific fields
  policyNumber?: string;
  insuranceCompany?: string;
  policyType?: string;
  coverageAmount?: number;
  premiumAmount?: number;
  premiumFrequency?: string;
  
  // Loan-specific fields (if needed separately)
  tenor?: number;
  outstandingBalance?: number;
}

export enum InvestmentCategory {
  STOCKS = 'stocks',
  MUTUAL_FUNDS = 'mutualFunds',
  FDS = 'fds',
  INSURANCE = 'insurance',
  LOANS = 'loans'
}

export interface CategorySummary {
  category: InvestmentCategory;
  totalInvested: number;
  currentValue: number;
  profitLoss: number;
  profitLossPercentage: number;
  count?: number;
}

export interface NetWorth {
  totalAssets: number;
  totalLiabilities: number;
  netWorth: number;
  lastUpdated?: Date;
}

export interface Transaction {
  id?: string;
  investmentId: string;
  category: InvestmentCategory;
  type: TransactionType;
  amount: number;
  date: string | Date;
  notes?: string;
  price?: number;
  quantity?: number;
  units?: number;
}

export enum TransactionType {
  BUY = 'buy',
  SELL = 'sell',
  DEPOSIT = 'deposit',
  WITHDRAWAL = 'withdrawal',
  PAYMENT = 'payment',
  DIVIDEND = 'dividend',
  INTEREST = 'interest'
}

export interface ChartData {
  labels: string[];
  values: number[];
  period?: ChartPeriod;
}

export enum ChartPeriod {
  ONE_MONTH = '1M',
  THREE_MONTHS = '3M',
  SIX_MONTHS = '6M',
  ONE_YEAR = '1Y',
  ALL = 'All'
}

// Helper interfaces for category-specific data
export interface StockData {
  ticker: string;
  companyName: string;
  quantity: number;
  buyPrice: number;
  currentPrice: number;
}

export interface MutualFundData {
  fundName: string;
  folioNumber: string;
  units: number;
  nav: number;
}

export interface FDData {
  bankName: string;
  fdNumber: string;
  interestRate: number;
  maturityDate: string;
  maturityAmount: number;
}

export interface InsuranceData {
  policyNumber: string;
  insuranceCompany: string;
  policyType: string;
  coverageAmount: number;
  premiumAmount: number;
  premiumFrequency: string;
  maturityDate?: string;
}

export interface LoanData {
  loanType: string;
  lender: string;
  interestRate: number;
  tenor: number;
  emiAmount: number;
  outstandingBalance: number;
  nextPaymentDate?: string;
}