import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute } from '@angular/router';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatTableModule } from '@angular/material/table';
import { MatTabsModule } from '@angular/material/tabs';
import { MatChipsModule } from '@angular/material/chips';
import { Observable } from 'rxjs';
import { MockDataService } from '../../services/mock-data.service';
import { Investment, InvestmentCategory, CategorySummary, Transaction } from '../../models/investment.model';

@Component({
  selector: 'app-investment-detail',
  standalone: true,
  imports: [
    CommonModule,
    MatCardModule,
    MatButtonModule,
    MatIconModule,
    MatTableModule,
    MatTabsModule,
    MatChipsModule
  ],
  templateUrl: './investment-detail.component.html',
  styleUrls: ['./investment-detail.component.scss']
})
export class InvestmentDetailComponent implements OnInit {
  category!: InvestmentCategory;
  investments$!: Observable<Investment[]>;
  categorySummary$!: Observable<CategorySummary>;
  transactions$!: Observable<Transaction[]>;
  
  displayedColumns: string[] = ['name', 'amount', 'currentValue', 'profitLoss', 'date'];
  transactionColumns: string[] = ['type', 'amount', 'date', 'notes'];
  
  categoryDetails: { [key: string]: any } = {
    [InvestmentCategory.STOCKS]: {
      title: 'Stocks',
      icon: 'trending_up',
      color: '#4caf50'
    },
    [InvestmentCategory.MUTUAL_FUNDS]: {
      title: 'Mutual Funds',
      icon: 'account_balance',
      color: '#2196f3'
    },
    [InvestmentCategory.FDS]: {
      title: 'Fixed Deposits',
      icon: 'savings',
      color: '#ff9800'
    },
    [InvestmentCategory.INSURANCE]: {
      title: 'Insurance',
      icon: 'security',
      color: '#9c27b0'
    },
    [InvestmentCategory.LOANS]: {
      title: 'Loans',
      icon: 'credit_card',
      color: '#f44336'
    }
  };

  constructor(
    private route: ActivatedRoute,
    private mockDataService: MockDataService
  ) {}

  ngOnInit(): void {
    this.route.params.subscribe(params => {
      this.category = params['category'] as InvestmentCategory;
      this.loadData();
    });
  }

  private loadData(): void {
    this.investments$ = this.mockDataService.getInvestmentsByCategory(this.category);
    this.categorySummary$ = this.mockDataService.getCategorySummary(this.category);
    this.transactions$ = this.mockDataService.getTransactionsByCategory(this.category);
  }

  get categoryDetail() {
    return this.categoryDetails[this.category];
  }

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

  calculateProfitLoss(investment: Investment): number {
    if (this.category === InvestmentCategory.LOANS) {
      return investment.amount - (investment.outstandingBalance || investment.currentValue);
    }
    return investment.currentValue - investment.amount;
  }

  calculateProfitLossPercentage(investment: Investment): number {
    const profitLoss = this.calculateProfitLoss(investment);
    return investment.amount > 0 ? (profitLoss / investment.amount) * 100 : 0;
  }

  isPositive(value: number): boolean {
    return value >= 0;
  }

  onAddInvestment(): void {
    // TODO: Open modal to add new investment
    console.log('Add investment for category:', this.category);
  }

  // Make Math available in template
  protected Math = Math;
}