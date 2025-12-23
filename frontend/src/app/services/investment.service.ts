import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable, of, throwError } from 'rxjs';
import { catchError, map, shareReplay, tap } from 'rxjs/operators';
import { 
  Investment, 
  InvestmentCategory, 
  CategorySummary, 
  NetWorth, 
  Transaction,
  ChartData,
  ChartPeriod 
} from '../models/investment.model';
import { environment } from '../../environments/environment';

@Injectable({
  providedIn: 'root'
})
export class InvestmentService {
  private http = inject(HttpClient);
  private apiUrl = environment.apiBaseUrl;
  
  // Cache for frequently accessed data
  private netWorthCache$?: Observable<NetWorth>;
  private categorySummariesCache$?: Observable<CategorySummary[]>;

  constructor() {}

  // Transform frontend Investment to backend format
  private transformToBackendFormat(investment: Partial<Investment>): any {
    const backendPayload: any = {
      category: investment.category,
      name: investment.name || '',
      investedAmount: investment.amount || 0,
      currentValue: investment.currentValue || investment.amount || 0,
      notes: investment.notes || ''
    };

    // Add category-specific data
    const investmentData: any = {};
    
    switch (investment.category) {
      case InvestmentCategory.STOCKS:
        investmentData.ticker = investment.ticker || '';
        investmentData.companyName = investment.name || '';
        investmentData.quantity = investment.quantity || 0;
        investmentData.averagePrice = investment.buyPrice || 0;
        investmentData.currentPrice = investment.currentPrice || 0;
        break;
        
      case InvestmentCategory.MUTUAL_FUNDS:
        investmentData.fundName = investment.name || '';
        investmentData.folioNumber = investment.folioNumber || '';
        investmentData.units = investment.units || 0;
        investmentData.nav = investment.nav || 0;
        break;
        
      case InvestmentCategory.FDS:
        investmentData.bankName = investment.bankName || '';
        investmentData.fdNumber = investment.fdNumber || '';
        investmentData.interestRate = investment.interestRate || 0;
        investmentData.maturityDate = investment.maturityDate ? new Date(investment.maturityDate).toISOString() : null;
        investmentData.maturityAmount = investment.maturityAmount || 0;
        investmentData.tenure = investment.tenor || 0;
        break;
        
      case InvestmentCategory.INSURANCE:
        investmentData.policyNumber = investment.policyNumber || '';
        investmentData.insuranceCompany = investment.insuranceCompany || '';
        investmentData.insuranceType = investment.policyType || '';
        investmentData.sumAssured = investment.coverageAmount || 0;
        investmentData.premiumAmount = investment.premiumAmount || 0;
        investmentData.maturityDate = investment.maturityDate ? new Date(investment.maturityDate).toISOString() : null;
        investmentData.tenure = investment.tenor || 0;
        break;
    }

    backendPayload.investmentData = investmentData;
    return backendPayload;
  }

  // Transform backend response to frontend format
  private transformFromBackendFormat(backendData: any): Investment {
    const investment: Investment = {
      id: backendData.id || backendData._id,
      category: backendData.category,
      amount: backendData.investedAmount,
      currentValue: backendData.currentValue,
      date: backendData.createdAt || backendData.investmentDate || new Date().toISOString(),
      notes: backendData.notes,
      profitLoss: (backendData.currentValue || 0) - (backendData.investedAmount || 0),
      profitLossPercentage: backendData.investedAmount > 0 
        ? (((backendData.currentValue || 0) - (backendData.investedAmount || 0)) / backendData.investedAmount) * 100 
        : 0
    };

    // Add category-specific fields
    const data = backendData.investmentData || {};
    
    switch (backendData.category) {
      case InvestmentCategory.STOCKS:
        investment.ticker = data.ticker;
        investment.name = data.companyName || backendData.name;
        investment.quantity = data.quantity;
        investment.buyPrice = data.averagePrice;
        investment.currentPrice = data.currentPrice;
        break;
        
      case InvestmentCategory.MUTUAL_FUNDS:
        investment.name = data.fundName || backendData.name;
        investment.folioNumber = data.folioNumber;
        investment.units = data.units;
        investment.nav = data.nav;
        break;
        
      case InvestmentCategory.FDS:
        investment.name = backendData.name;
        investment.bankName = data.bankName;
        investment.fdNumber = data.fdNumber;
        investment.interestRate = data.interestRate;
        investment.maturityDate = data.maturityDate;
        investment.maturityAmount = data.maturityAmount;
        investment.tenor = data.tenure;
        break;
        
      case InvestmentCategory.INSURANCE:
        investment.name = backendData.name;
        investment.policyNumber = data.policyNumber;
        investment.insuranceCompany = data.insuranceCompany;
        investment.policyType = data.insuranceType;
        investment.coverageAmount = data.sumAssured;
        investment.premiumAmount = data.premiumAmount;
        investment.maturityDate = data.maturityDate;
        investment.tenor = data.tenure;
        break;
    }

    return investment;
  }

  // Dashboard APIs
  getNetWorth(): Observable<NetWorth> {
    if (!this.netWorthCache$) {
      this.netWorthCache$ = this.http.get<any>(`${this.apiUrl}/dashboard/summary`).pipe(
        map(response => ({
          totalAssets: response.netWorth.totalAssets,
          totalLiabilities: response.netWorth.totalLiabilities,
          netWorth: response.netWorth.netWorth
        })),
        shareReplay(1),
        catchError(this.handleError)
      );
    }
    return this.netWorthCache$;
  }

  getCategorySummaries(): Observable<CategorySummary[]> {
    if (!this.categorySummariesCache$) {
      this.categorySummariesCache$ = this.http.get<any>(`${this.apiUrl}/dashboard/summary`).pipe(
        map(response => {
          const summaries: CategorySummary[] = [];
          // Convert map to array
          Object.entries(response.categories).forEach(([category, data]: [string, any]) => {
            summaries.push({
              category: category as InvestmentCategory,
              totalInvested: data.investedAmount,
              currentValue: data.currentValue,
              profitLoss: data.currentValue - data.investedAmount,
              profitLossPercentage: data.profitLossPercentage,
              count: data.count
            });
          });
          
          // Add loans summary if available
          if (response.loans) {
            summaries.push({
              category: InvestmentCategory.LOANS,
              totalInvested: response.loans.principalAmount,
              currentValue: response.loans.outstandingBalance,
              profitLoss: response.loans.principalAmount - response.loans.outstandingBalance,
              profitLossPercentage: response.loans.principalAmount > 0 
                ? ((response.loans.principalAmount - response.loans.outstandingBalance) / response.loans.principalAmount) * 100 
                : 0,
              count: response.loans.count
            });
          }
          
          return summaries;
        }),
        shareReplay(1),
        catchError(this.handleError)
      );
    }
    return this.categorySummariesCache$;
  }

  getCategorySummary(category: InvestmentCategory): Observable<CategorySummary> {
    return this.http.get<any>(`${this.apiUrl}/investments/category/${category}/summary`).pipe(
      map(response => ({
        category: category,
        totalInvested: response.investedAmount,
        currentValue: response.currentValue,
        profitLoss: response.currentValue - response.investedAmount,
        profitLossPercentage: response.profitLossPercentage,
        count: response.count
      })),
      catchError(this.handleError)
    );
  }

  // Investment APIs
  getInvestments(): Observable<Investment[]> {
    return this.http.get<any[]>(`${this.apiUrl}/investments`).pipe(
      map(investments => investments.map(inv => this.transformFromBackendFormat(inv))),
      catchError(this.handleError)
    );
  }

  getInvestmentsByCategory(category: InvestmentCategory): Observable<Investment[]> {
    return this.http.get<any>(`${this.apiUrl}/investments/category/${category}`).pipe(
      map(response => {
        // Handle both wrapped and direct array responses
        const investments = response.investments || response || [];
        return Array.isArray(investments) 
          ? investments.map((inv: any) => this.transformFromBackendFormat(inv))
          : [];
      }),
      catchError(this.handleError)
    );
  }

  getInvestmentById(id: string): Observable<Investment> {
    return this.http.get<any>(`${this.apiUrl}/investments/detail/${id}`).pipe(
      map(investment => this.transformFromBackendFormat(investment)),
      catchError(this.handleError)
    );
  }

  createInvestment(investment: Partial<Investment>): Observable<Investment> {
    this.clearCache();
    const backendPayload = this.transformToBackendFormat(investment);
    
    console.log('Creating investment with payload:', backendPayload);
    
    return this.http.post<any>(`${this.apiUrl}/investments`, backendPayload).pipe(
      map(response => this.transformFromBackendFormat(response)),
      catchError(this.handleError)
    );
  }

  updateInvestment(id: string, investment: Partial<Investment>): Observable<Investment> {
    this.clearCache();
    const backendPayload = this.transformToBackendFormat(investment);
    
    return this.http.put<any>(`${this.apiUrl}/investments/${id}`, backendPayload).pipe(
      map(response => this.transformFromBackendFormat(response)),
      catchError(this.handleError)
    );
  }

  deleteInvestment(id: string): Observable<void> {
    this.clearCache();
    return this.http.delete<void>(`${this.apiUrl}/investments/${id}`).pipe(
      catchError(this.handleError)
    );
  }

  // ...existing transaction, loan, and chart methods...
  // Transaction APIs
  getTransactions(): Observable<Transaction[]> {
    return this.http.get<Transaction[]>(`${this.apiUrl}/transactions`).pipe(
      catchError(this.handleError)
    );
  }

  getTransactionsByCategory(category: InvestmentCategory): Observable<Transaction[]> {
    return this.http.get<Transaction[]>(`${this.apiUrl}/transactions/category/${category}`).pipe(
      catchError(this.handleError)
    );
  }

  getTransactionsByInvestment(investmentId: string): Observable<Transaction[]> {
    return this.http.get<Transaction[]>(`${this.apiUrl}/transactions/related/${investmentId}`).pipe(
      catchError(this.handleError)
    );
  }

  createTransaction(transaction: Partial<Transaction>): Observable<Transaction> {
    this.clearCache();
    
    const backendPayload = {
      relatedId: transaction.investmentId,
      category: transaction.category,
      type: transaction.type,
      amount: transaction.amount,
      quantity: transaction.quantity,
      price: transaction.price,
      date: transaction.date ? new Date(transaction.date).toISOString() : new Date().toISOString(),
      description: transaction.notes || '',
      fees: 0,
      notes: transaction.notes || ''
    };
    
    return this.http.post<Transaction>(`${this.apiUrl}/transactions`, backendPayload).pipe(
      catchError(this.handleError)
    );
  }

  updateTransaction(id: string, transaction: Partial<Transaction>): Observable<Transaction> {
    this.clearCache();
    return this.http.put<Transaction>(`${this.apiUrl}/transactions/${id}`, transaction).pipe(
      catchError(this.handleError)
    );
  }

  deleteTransaction(id: string): Observable<void> {
    this.clearCache();
    return this.http.delete<void>(`${this.apiUrl}/transactions/${id}`).pipe(
      catchError(this.handleError)
    );
  }

  // Loan APIs
  getLoans(): Observable<Investment[]> {
    return this.http.get<any[]>(`${this.apiUrl}/loans`).pipe(
      map(loans => loans.map(loan => this.transformLoanFromBackendFormat(loan))),
      catchError(this.handleError)
    );
  }

  getLoanById(id: string): Observable<Investment> {
    return this.http.get<any>(`${this.apiUrl}/loans/${id}`).pipe(
      map(loan => this.transformLoanFromBackendFormat(loan)),
      catchError(this.handleError)
    );
  }

  createLoan(loan: Partial<Investment>): Observable<Investment> {
    this.clearCache();
    const backendPayload = this.transformLoanToBackendFormat(loan);
    
    return this.http.post<any>(`${this.apiUrl}/loans`, backendPayload).pipe(
      map(response => this.transformLoanFromBackendFormat(response)),
      catchError(this.handleError)
    );
  }

  updateLoan(id: string, loan: Partial<Investment>): Observable<Investment> {
    this.clearCache();
    const backendPayload = this.transformLoanToBackendFormat(loan);
    
    return this.http.put<any>(`${this.apiUrl}/loans/${id}`, backendPayload).pipe(
      map(response => this.transformLoanFromBackendFormat(response)),
      catchError(this.handleError)
    );
  }

  deleteLoan(id: string): Observable<void> {
    this.clearCache();
    return this.http.delete<void>(`${this.apiUrl}/loans/${id}`).pipe(
      catchError(this.handleError)
    );
  }

  // Helper methods for loan transformations
  private transformLoanToBackendFormat(loan: Partial<Investment>): any {
    return {
      loanType: loan.name || 'personal_loan',
      lenderName: loan.bankName || '',
      principalAmount: loan.amount || 0,
      outstandingBalance: loan.outstandingBalance || loan.amount || 0,
      interestRate: loan.interestRate || 0,
      emi: (loan.amount || 0) / (loan.tenor || 1), // Simple EMI calculation
      tenure: loan.tenor || 0,
      remainingTenure: loan.tenor || 0,
      startDate: loan.date ? new Date(loan.date).toISOString() : new Date().toISOString(),
      maturityDate: loan.maturityDate ? new Date(loan.maturityDate).toISOString() : new Date().toISOString(),
      notes: loan.notes || ''
    };
  }

  private transformLoanFromBackendFormat(backendData: any): Investment {
    return {
      id: backendData.id,
      category: InvestmentCategory.LOANS,
      name: backendData.loanType,
      bankName: backendData.lenderName,
      amount: backendData.principalAmount,
      currentValue: backendData.outstandingBalance,
      outstandingBalance: backendData.outstandingBalance,
      interestRate: backendData.interestRate,
      tenor: backendData.tenure,
      date: backendData.startDate,
      maturityDate: backendData.maturityDate,
      notes: backendData.notes,
      profitLoss: backendData.principalAmount - backendData.outstandingBalance,
      profitLossPercentage: backendData.principalAmount > 0 
        ? ((backendData.principalAmount - backendData.outstandingBalance) / backendData.principalAmount) * 100 
        : 0
    };
  }

  // Chart Data APIs
  getChartData(category: InvestmentCategory, period: ChartPeriod): Observable<ChartData> {
    return this.http.get<any>(`${this.apiUrl}/charts/${category}/${period}`).pipe(
      map(response => ({
        labels: response.data.map((point: any) => point.date),
        values: response.data.map((point: any) => point.value),
        period: response.period
      })),
      catchError(this.handleError)
    );
  }

  // Helper Methods
  private clearCache(): void {
    this.netWorthCache$ = undefined;
    this.categorySummariesCache$ = undefined;
  }

  private handleError(error: any): Observable<never> {
    console.error('API Error:', error);
    
    let errorMessage = 'An error occurred. Please try again later.';
    
    if (error.error instanceof ErrorEvent) {
      // Client-side error
      errorMessage = `Error: ${error.error.message}`;
    } else if (error.status) {
      // Server-side error
      switch (error.status) {
        case 400:
          errorMessage = error.error?.error || 'Invalid request. Please check your input.';
          break;
        case 401:
          errorMessage = 'Unauthorized. Please log in again.';
          break;
        case 403:
          errorMessage = 'You do not have permission to perform this action.';
          break;
        case 404:
          errorMessage = 'Resource not found.';
          break;
        case 500:
          errorMessage = 'Server error. Please try again later.';
          break;
        default:
          errorMessage = error.error?.message || errorMessage;
      }
    }
    
    return throwError(() => new Error(errorMessage));
  }

  // Formatting helpers
  formatCurrency(amount: number): string {
    return new Intl.NumberFormat('en-IN', {
      style: 'currency',
      currency: 'INR',
      minimumFractionDigits: 0,
      maximumFractionDigits: 0
    }).format(amount);
  }

  formatPercentage(percentage: number): string {
    const sign = percentage >= 0 ? '+' : '';
    return `${sign}${percentage.toFixed(2)}%`;
  }

  calculateProfitLoss(invested: number, current: number): number {
    if (invested === 0) return 0;
    return ((current - invested) / invested) * 100;
  }
}