import { Component, OnInit, ViewChild } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute, Router } from '@angular/router';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatTableModule, MatTableDataSource } from '@angular/material/table';
import { MatTabsModule } from '@angular/material/tabs';
import { MatChipsModule } from '@angular/material/chips';
import { MatDialogModule, MatDialog } from '@angular/material/dialog';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { BaseChartDirective } from 'ng2-charts';
import { ChartConfiguration, ChartOptions } from 'chart.js';
import { Observable, of, BehaviorSubject } from 'rxjs';
import { catchError, tap } from 'rxjs/operators';
import { InvestmentService } from '../../services/investment.service';
import { Investment, InvestmentCategory, CategorySummary, Transaction, ChartPeriod, ChartData } from '../../models/investment.model';
import { AddInvestmentDialogComponent } from '../../components/add-investment-dialog/add-investment-dialog.component';

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
    MatChipsModule,
    MatDialogModule,
    MatSnackBarModule,
    BaseChartDirective
  ],
  templateUrl: './investment-detail.component.html',
  styleUrls: ['./investment-detail.component.scss']
})
export class InvestmentDetailComponent implements OnInit {
  @ViewChild(BaseChartDirective) chart?: BaseChartDirective;
  
  category!: InvestmentCategory;
  
  // Use MatTableDataSource for better table handling
  investmentsDataSource = new MatTableDataSource<Investment>([]);
  transactionsDataSource = new MatTableDataSource<Transaction>([]);
  
  // Keep observables for other uses
  investments$: Observable<Investment[]> = of([]);
  categorySummary$!: Observable<CategorySummary>;
  transactions$: Observable<Transaction[]> = of([]);
  
  displayedColumns: string[] = ['name', 'amount', 'currentValue', 'profitLoss', 'date'];
  transactionColumns: string[] = ['type', 'amount', 'date', 'notes'];
  
  // Chart configuration
  selectedPeriod: ChartPeriod = ChartPeriod.THREE_MONTHS;
  chartPeriods = [
    { label: '1M', value: ChartPeriod.ONE_MONTH },
    { label: '3M', value: ChartPeriod.THREE_MONTHS },
    { label: '1Y', value: ChartPeriod.ONE_YEAR },
    { label: 'All', value: ChartPeriod.ALL }
  ];
  
  public lineChartData: ChartConfiguration<'line'>['data'] = {
    labels: [],
    datasets: [
      {
        data: [],
        label: 'Portfolio Value',
        fill: true,
        tension: 0.4,
        borderColor: '#2196f3',
        backgroundColor: 'rgba(33, 150, 243, 0.1)',
        pointBackgroundColor: '#2196f3',
        pointBorderColor: '#fff',
        pointHoverBackgroundColor: '#fff',
        pointHoverBorderColor: '#2196f3',
      }
    ]
  };

  public lineChartOptions: ChartOptions<'line'> = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: {
        display: true,
        position: 'top',
      },
      tooltip: {
        mode: 'index',
        intersect: false,
        callbacks: {
          label: (context) => {
            let label = context.dataset.label || '';
            if (label) {
              label += ': ';
            }
            if (context.parsed.y !== null) {
              label += this.formatCurrency(context.parsed.y);
            }
            return label;
          }
        }
      }
    },
    scales: {
      y: {
        beginAtZero: false,
        ticks: {
          callback: (value) => {
            return '₹' + (value as number).toLocaleString('en-IN');
          }
        }
      }
    }
  };
  
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
    private router: Router,
    private investmentService: InvestmentService,
    private dialog: MatDialog,
    private snackBar: MatSnackBar
  ) {}

  ngOnInit(): void {
    this.route.params.subscribe(params => {
      this.category = params['category'] as InvestmentCategory;
      this.loadData();
      this.loadChartData(this.selectedPeriod);
    });
    
    // Subscribe to observables to trigger data loading
    this.investments$.subscribe();
    this.transactions$.subscribe();
  }

  private loadData(): void {
    this.investments$ = this.investmentService.getInvestmentsByCategory(this.category).pipe(
      tap(investments => {
        this.investmentsDataSource.data = investments;
      }),
      catchError((error: any) => {
        console.error('Error loading investments:', error);
        this.showError('Failed to load investments. Please try again.');
        
        if (error.message?.includes('Unauthorized') || error.message?.includes('log in')) {
          this.router.navigate(['/login']);
        }
        
        this.investmentsDataSource.data = [];
        return of([]);
      })
    );
    
    this.categorySummary$ = this.investmentService.getCategorySummary(this.category).pipe(
      catchError((error: any) => {
        console.error('Error loading category summary:', error);
        return of({
          category: this.category,
          totalInvested: 0,
          currentValue: 0,
          profitLoss: 0,
          profitLossPercentage: 0,
          count: 0
        });
      })
    );
    
    this.transactions$ = this.investmentService.getTransactionsByCategory(this.category).pipe(
      tap(transactions => {
        this.transactionsDataSource.data = transactions;
      }),
      catchError((error: any) => {
        console.error('Error loading transactions:', error);
        this.transactionsDataSource.data = [];
        return of([]);
      })
    );
  }

  private loadChartData(period: ChartPeriod): void {
    this.investmentService.getChartData(this.category, period).subscribe({
      next: (data: ChartData) => {
        this.lineChartData.labels = data.labels;
        this.lineChartData.datasets[0].data = data.values;
        this.chart?.update();
      },
      error: (error: any) => {
        console.error('Error loading chart data:', error);
        this.showError('Failed to load chart data.');
        
        // Set empty chart data
        this.lineChartData.labels = [];
        this.lineChartData.datasets[0].data = [];
        this.chart?.update();
      }
    });
  }

  onPeriodChange(period: ChartPeriod): void {
    this.selectedPeriod = period;
    this.loadChartData(period);
  }

  get categoryDetail() {
    return this.categoryDetails[this.category];
  }

  formatCurrency(amount: number): string {
    return this.investmentService.formatCurrency(amount);
  }

  formatPercentage(percentage: number): string {
    return this.investmentService.formatPercentage(percentage);
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

  canAddInvestment(): boolean {
    return this.category !== InvestmentCategory.STOCKS && this.category !== InvestmentCategory.MUTUAL_FUNDS;
  }

  onAddInvestment(): void {
    const dialogRef = this.dialog.open(AddInvestmentDialogComponent, {
      width: '600px',
      data: {
        category: this.category,
        mode: 'add'
      }
    });

    dialogRef.afterClosed().subscribe(result => {
      if (result) {
        this.investmentService.createInvestment(result).subscribe({
          next: () => {
            // Reload data after successful creation
            this.loadData();
            this.loadChartData(this.selectedPeriod);
            this.showSuccess('Investment added successfully!');
          },
          error: (error: any) => {
            console.error('Error creating investment:', error);
            this.showError(error.message || 'Failed to create investment. Please try again.');
            
            if (error.message?.includes('Unauthorized') || error.message?.includes('log in')) {
              this.router.navigate(['/login']);
            }
          }
        });
      }
    });
  }

  private showSuccess(message: string): void {
    this.snackBar.open(message, 'Close', {
      duration: 3000,
      horizontalPosition: 'end',
      verticalPosition: 'top',
      panelClass: ['success-snackbar']
    });
  }

  private showError(message: string): void {
    this.snackBar.open(message, 'Close', {
      duration: 5000,
      horizontalPosition: 'end',
      verticalPosition: 'top',
      panelClass: ['error-snackbar']
    });
  }

  // Make Math available in template
  protected Math = Math;
}