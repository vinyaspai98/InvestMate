import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router } from '@angular/router';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatGridListModule } from '@angular/material/grid-list';
import { Observable, of } from 'rxjs';
import { tap, catchError } from 'rxjs/operators';
import { InvestmentService } from '../../services/investment.service';
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
  isLoading = true;
  error: string | null = null;
  
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
      route: '/investments/mutualfunds'
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
    private investmentService: InvestmentService,
    private router: Router
  ) {
    this.netWorth$ = this.investmentService.getNetWorth().pipe(
      tap(() => this.isLoading = false),
      catchError(error => {
        console.error('Error loading net worth:', error);
        this.error = error.message || 'Failed to load dashboard data. Please try again.';
        this.isLoading = false;
        
        // If 401 Unauthorized, redirect to login
        if (error.message?.includes('Unauthorized') || error.message?.includes('log in')) {
          this.router.navigate(['/login']);
        }
        
        // Return empty data to prevent breaking the UI
        return of({ totalAssets: 0, totalLiabilities: 0, netWorth: 0 });
      })
    );
    
    this.categorySummaries$ = this.investmentService.getCategorySummaries().pipe(
      catchError(error => {
        console.error('Error loading category summaries:', error);
        // Return empty array to prevent breaking the UI
        return of([]);
      })
    );
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
    return this.investmentService.formatCurrency(amount);
  }

  formatPercentage(percentage: number): string {
    return this.investmentService.formatPercentage(percentage);
  }

  isPositive(value: number): boolean {
    return value >= 0;
  }

  // Make Math available in template
  protected Math = Math;
}