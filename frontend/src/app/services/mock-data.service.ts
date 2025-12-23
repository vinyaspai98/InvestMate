import { Injectable } from '@angular/core';
import { Observable, of } from 'rxjs';
import { Investment, InvestmentCategory, CategorySummary, NetWorth, Transaction, TransactionType } from '../models/investment.model';

@Injectable({
  providedIn: 'root'
})
export class MockDataService {

  private mockInvestments: Investment[] = [
    // Stocks
    {
      id: '1',
      category: InvestmentCategory.STOCKS,
      name: 'Reliance Industries',
      ticker: 'RELIANCE',
      amount: 50000,
      currentValue: 55000,
      date: new Date('2024-01-15'),
      notes: 'Long term investment'
    },
    {
      id: '2',
      category: InvestmentCategory.STOCKS,
      name: 'TCS',
      ticker: 'TCS',
      amount: 75000,
      currentValue: 82000,
      date: new Date('2024-02-20'),
      notes: 'IT sector investment'
    },
    // Mutual Funds
    {
      id: '3',
      category: InvestmentCategory.MUTUAL_FUNDS,
      name: 'SBI Bluechip Fund',
      amount: 100000,
      currentValue: 108000,
      date: new Date('2024-01-10'),
      notes: 'SIP investment'
    },
    {
      id: '4',
      category: InvestmentCategory.MUTUAL_FUNDS,
      name: 'HDFC Mid Cap Opportunities',
      amount: 60000,
      currentValue: 65000,
      date: new Date('2024-03-05'),
      notes: 'Growth focused'
    },
    // FDs
    {
      id: '5',
      category: InvestmentCategory.FDS,
      name: 'HDFC Bank FD',
      amount: 200000,
      currentValue: 210000,
      date: new Date('2024-01-01'),
      interestRate: 6.5,
      tenor: 12,
      maturityDate: new Date('2025-01-01'),
      notes: 'Safe investment'
    },
    // Insurance
    {
      id: '6',
      category: InvestmentCategory.INSURANCE,
      name: 'LIC Jeevan Anand',
      amount: 50000,
      currentValue: 52000,
      date: new Date('2023-06-15'),
      tenor: 20,
      notes: 'Life insurance policy'
    },
    // Loans
    {
      id: '7',
      category: InvestmentCategory.LOANS,
      name: 'Home Loan',
      amount: 2500000,
      currentValue: 2500000,
      outstandingBalance: 2200000,
      date: new Date('2023-01-01'),
      interestRate: 8.5,
      tenor: 240,
      notes: 'Home purchase'
    }
  ];

  private mockTransactions: Transaction[] = [
    {
      id: 't1',
      investmentId: '1',
      type: TransactionType.BUY,
      amount: 50000,
      date: new Date('2024-01-15'),
      notes: 'Initial purchase',
      category: InvestmentCategory.STOCKS
    },
    {
      id: 't2',
      investmentId: '3',
      type: TransactionType.DEPOSIT,
      amount: 10000,
      date: new Date('2024-01-10'),
      notes: 'Monthly SIP',
      category: InvestmentCategory.MUTUAL_FUNDS
    },
    {
      id: 't3',
      investmentId: '7',
      type: TransactionType.PAYMENT,
      amount: 25000,
      date: new Date('2024-09-01'),
      notes: 'EMI payment',
      category: InvestmentCategory.LOANS
    }
  ];

  getInvestmentsByCategory(category: InvestmentCategory): Observable<Investment[]> {
    const investments = this.mockInvestments.filter(inv => inv.category === category);
    return of(investments);
  }

  getAllInvestments(): Observable<Investment[]> {
    return of(this.mockInvestments);
  }

  getCategorySummary(category: InvestmentCategory): Observable<CategorySummary> {
    const investments = this.mockInvestments.filter(inv => inv.category === category);
    const totalInvested = investments.reduce((sum, inv) => sum + inv.amount, 0);
    const currentValue = category === InvestmentCategory.LOANS 
      ? investments.reduce((sum, inv) => sum + (inv.outstandingBalance || inv.currentValue), 0)
      : investments.reduce((sum, inv) => sum + inv.currentValue, 0);
    
    const profitLoss = category === InvestmentCategory.LOANS 
      ? totalInvested - currentValue  // For loans, less outstanding is better
      : currentValue - totalInvested;
    
    const profitLossPercentage = totalInvested > 0 ? (profitLoss / totalInvested) * 100 : 0;

    return of({
      category,
      totalInvested,
      currentValue,
      profitLoss,
      profitLossPercentage
    });
  }

  getAllCategorySummaries(): Observable<CategorySummary[]> {
    const categories = Object.values(InvestmentCategory);
    const summaries: CategorySummary[] = [];

    categories.forEach(category => {
      this.getCategorySummary(category).subscribe(summary => {
        summaries.push(summary);
      });
    });

    return of(summaries);
  }

  getNetWorth(): Observable<NetWorth> {
    const assets = this.mockInvestments
      .filter(inv => inv.category !== InvestmentCategory.LOANS)
      .reduce((sum, inv) => sum + inv.currentValue, 0);
    
    const liabilities = this.mockInvestments
      .filter(inv => inv.category === InvestmentCategory.LOANS)
      .reduce((sum, inv) => sum + (inv.outstandingBalance || inv.currentValue), 0);

    return of({
      totalAssets: assets,
      totalLiabilities: liabilities,
      netWorth: assets - liabilities
    });
  }

  getTransactionsByInvestmentId(investmentId: string): Observable<Transaction[]> {
    const transactions = this.mockTransactions.filter(t => t.investmentId === investmentId);
    return of(transactions);
  }

  getTransactionsByCategory(category: InvestmentCategory): Observable<Transaction[]> {
    const investmentIds = this.mockInvestments
      .filter(inv => inv.category === category)
      .map(inv => inv.id);
    
    const transactions = this.mockTransactions.filter(t => investmentIds.includes(t.investmentId));
    return of(transactions);
  }

  addInvestment(investment: Omit<Investment, 'id'>): Observable<Investment> {
    const newInvestment: Investment = {
      ...investment,
      id: Date.now().toString()
    };
    this.mockInvestments.push(newInvestment);
    return of(newInvestment);
  }

  addTransaction(transaction: Omit<Transaction, 'id'>): Observable<Transaction> {
    const newTransaction: Transaction = {
      ...transaction,
      id: Date.now().toString()
    };
    this.mockTransactions.push(newTransaction);
    return of(newTransaction);
  }

  // Mock chart data for historical tracking
  getChartData(category: InvestmentCategory, period: string): Observable<any> {
    const months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct'];
    const baseValue = category === InvestmentCategory.LOANS ? 2200000 : 100000;
    
    const data = months.map((month, index) => {
      const variation = Math.random() * 10000 - 5000; // Random variation
      return baseValue + (index * 5000) + variation;
    });

    return of({
      labels: months,
      datasets: [{
        label: `${category} Value`,
        data: data,
        borderColor: '#3f51b5',
        backgroundColor: 'rgba(63, 81, 181, 0.1)',
        fill: true
      }]
    });
  }
}