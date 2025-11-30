import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router } from '@angular/router';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatGridListModule } from '@angular/material/grid-list';
import { Observable } from 'rxjs';
import { MockDataService } from '../../services/mock-data.service';
import { CategorySummary, NetWorth, InvestmentCategory } from '../../models/investment.model';

@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [
    CommonModule,
    MatCardModule,
    MatButtonModule,
    MatIconModule,
    MatGridListModule
  ],
  templateUrl: './dashboard.component.html',
  styleUrls: ['./dashboard.component.scss']
})
export class DashboardComponent implements OnInit {
  netWorth$: Observable<NetWorth>;
  categorySummaries$: Observable<CategorySummary[]>;
  
  categoryDetails = [
    {
      category: InvestmentCategory.STOCKS,
      title: 'Stocks',
      icon: 'trending_up',
      color: '#4caf50',
      route: '/investments/stocks'
    },
    {
      category: InvestmentCategory.MUTUAL_FUNDS,
      title: 'Mutual Funds',
      icon: 'account_balance',
      color: '#2196f3',
      route: '/investments/mutual_funds'
    },
    {
      category: InvestmentCategory.FDS,
      title: 'Fixed Deposits',
      icon: 'savings',
      color: '#ff9800',
      route: '/investments/fds'
    },
    {
      category: InvestmentCategory.INSURANCE,
      title: 'Insurance',
      icon: 'security',
      color: '#9c27b0',
      route: '/investments/insurance'
    },
    {
      category: InvestmentCategory.LOANS,
      title: 'Loans',
      icon: 'credit_card',
      color: '#f44336',
      route: '/investments/loans'
    }
  ];

  constructor(
    private mockDataService: MockDataService,
    private router: Router
  ) {
    this.netWorth$ = this.mockDataService.getNetWorth();
    this.categorySummaries$ = this.mockDataService.getAllCategorySummaries();
  }

  ngOnInit(): void {
    // Component initialization
  }

  navigateToCategory(route: string): void {
    this.router.navigate([route]);
  }

  getCategoryDetails(category: InvestmentCategory) {
    return this.categoryDetails.find(detail => detail.category === category);
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

  isPositive(value: number): boolean {
    return value >= 0;
  }

  // Make Math available in template
  protected Math = Math;
}