import { Component, Inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, Validators, ReactiveFormsModule } from '@angular/forms';
import { MatDialogModule, MatDialogRef, MAT_DIALOG_DATA } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatSelectModule } from '@angular/material/select';
import { MatDatepickerModule } from '@angular/material/datepicker';
import { MatNativeDateModule } from '@angular/material/core';
import { MatIconModule } from '@angular/material/icon';
import { MatAutocompleteModule } from '@angular/material/autocomplete';
import { InvestmentCategory, TransactionType } from '../../models/investment.model';
import { AutocompleteService } from '../../services/autocomplete.service';
import { debounceTime, distinctUntilChanged, switchMap, catchError, of, Observable } from 'rxjs';

export interface InvestmentDialogData {
  category: InvestmentCategory;
  mode: 'add' | 'edit';
  investment?: any;
}

@Component({
  selector: 'app-add-investment-dialog',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    MatDialogModule,
    MatFormFieldModule,
    MatInputModule,
    MatButtonModule,
    MatSelectModule,
    MatDatepickerModule,
    MatNativeDateModule,
    MatIconModule,
    MatAutocompleteModule
  ],
  template: `
    <h2 mat-dialog-title>
      <mat-icon>{{ getCategoryIcon() }}</mat-icon>
      {{ data.mode === 'add' ? 'Add' : 'Edit' }} {{ getCategoryTitle() }}
    </h2>
    
    <mat-dialog-content>
      <form [formGroup]="investmentForm">
        <!-- Common Fields -->
        <mat-form-field appearance="outline" class="full-width">
          <mat-label>Name</mat-label>
          <input matInput formControlName="name" placeholder="Enter name">
          <mat-error *ngIf="investmentForm.get('name')?.hasError('required')">
            Name is required
          </mat-error>
        </mat-form-field>

        <!-- Stock-specific fields -->
        <mat-form-field appearance="outline" class="full-width" *ngIf="data.category === 'stocks'">
          <mat-label>Ticker Symbol</mat-label>
          <input matInput formControlName="ticker" placeholder="e.g., RELIANCE, TCS" [matAutocomplete]="auto">
          <mat-autocomplete #auto="matAutocomplete" (optionSelected)="onSymbolSelected($event.option.value)">
            <mat-option *ngFor="let match of filteredSymbols$ | async" [value]="match['1. symbol']">
              {{ match['1. symbol'] }} - {{ match['2. name'] }}
            </mat-option>
          </mat-autocomplete>
        </mat-form-field>

        <mat-form-field appearance="outline" class="full-width" *ngIf="data.category === 'stocks'">
          <mat-label>Quantity</mat-label>
          <input matInput type="number" formControlName="quantity" placeholder="Enter quantity">
          <mat-error *ngIf="investmentForm.get('quantity')?.hasError('required')">
            Quantity is required
          </mat-error>
        </mat-form-field>

        <mat-form-field appearance="outline" class="full-width">
          <mat-label>{{ getAmountLabel() }}</mat-label>
          <input matInput type="number" formControlName="amount" placeholder="Enter amount">
          <span matPrefix>₹&nbsp;</span>
          <mat-error *ngIf="investmentForm.get('amount')?.hasError('required')">
            Amount is required
          </mat-error>
          <mat-error *ngIf="investmentForm.get('amount')?.hasError('min')">
            Amount must be greater than 0
          </mat-error>
        </mat-form-field>

        <mat-form-field appearance="outline" class="full-width" *ngIf="data.category !== 'loans'">
          <mat-label>Current Value</mat-label>
          <input matInput type="number" formControlName="currentValue" placeholder="Enter current value">
          <span matPrefix>₹&nbsp;</span>
        </mat-form-field>

        <!-- Loan-specific fields -->
        <mat-form-field appearance="outline" class="full-width" *ngIf="data.category === 'loans'">
          <mat-label>Outstanding Balance</mat-label>
          <input matInput type="number" formControlName="outstandingBalance" placeholder="Enter outstanding balance">
          <span matPrefix>₹&nbsp;</span>
        </mat-form-field>

        <!-- Interest rate (for FDs and Loans) -->
        <mat-form-field appearance="outline" class="full-width" 
                       *ngIf="data.category === 'fds' || data.category === 'loans'">
          <mat-label>Interest Rate</mat-label>
          <input matInput type="number" formControlName="interestRate" placeholder="Enter rate" step="0.1">
          <span matSuffix>%</span>
        </mat-form-field>

        <!-- Tenor (for FDs and Insurance) -->
        <mat-form-field appearance="outline" class="full-width" 
                       *ngIf="data.category === 'fds' || data.category === 'insurance' || data.category === 'loans'">
          <mat-label>{{ getTenorLabel() }}</mat-label>
          <input matInput type="number" formControlName="tenor" placeholder="Enter duration">
          <span matSuffix>months</span>
        </mat-form-field>

        <!-- Maturity Date (for FDs) -->
        <mat-form-field appearance="outline" class="full-width" *ngIf="data.category === 'fds'">
          <mat-label>Maturity Date</mat-label>
          <input matInput [matDatepicker]="maturityPicker" formControlName="maturityDate">
          <mat-datepicker-toggle matSuffix [for]="maturityPicker"></mat-datepicker-toggle>
          <mat-datepicker #maturityPicker></mat-datepicker>
        </mat-form-field>

        <mat-form-field appearance="outline" class="full-width">
          <mat-label>Date</mat-label>
          <input matInput [matDatepicker]="datePicker" formControlName="date">
          <mat-datepicker-toggle matSuffix [for]="datePicker"></mat-datepicker-toggle>
          <mat-datepicker #datePicker></mat-datepicker>
          <mat-error *ngIf="investmentForm.get('date')?.hasError('required')">
            Date is required
          </mat-error>
        </mat-form-field>

        <mat-form-field appearance="outline" class="full-width">
          <mat-label>Notes</mat-label>
          <textarea matInput formControlName="notes" rows="3" placeholder="Add any additional notes"></textarea>
        </mat-form-field>
      </form>
    </mat-dialog-content>
    
    <mat-dialog-actions align="end">
      <button mat-button (click)="onCancel()">Cancel</button>
      <button mat-raised-button color="primary" (click)="onSave()" [disabled]="!investmentForm.valid">
        {{ data.mode === 'add' ? 'Add' : 'Update' }}
      </button>
    </mat-dialog-actions>
  `,
  styles: [`
    h2 {
      display: flex;
      align-items: center;
      gap: 0.5rem;
      margin: 0;
      padding: 1rem 1.5rem;
      border-bottom: 1px solid rgba(0, 0, 0, 0.1);
    }

    mat-dialog-content {
      padding: 1.5rem;
      max-height: 70vh;
      overflow-y: auto;
    }

    .full-width {
      width: 100%;
      margin-bottom: 1rem;
    }

    mat-dialog-actions {
      padding: 1rem 1.5rem;
      border-top: 1px solid rgba(0, 0, 0, 0.1);
      gap: 0.5rem;
    }

    form {
      min-width: 500px;
    }

    @media (max-width: 600px) {
      form {
        min-width: 300px;
      }
    }

    // Dark theme styles
    :host-context(.dark-theme) {
      h2 {
        background: #1e1e1e;
        color: #ffffff;
        border-bottom-color: rgba(255, 255, 255, 0.1);
      }

      mat-dialog-content {
        background: #1e1e1e;
        color: #ffffff;
      }

      mat-dialog-actions {
        background: #1e1e1e;
        border-top-color: rgba(255, 255, 255, 0.1);
      }

      // Material form field labels
      mat-form-field {
        .mat-mdc-form-field-label,
        .mdc-floating-label {
          color: rgba(255, 255, 255, 0.7) !important;
        }

        .mat-mdc-form-field-label.mdc-floating-label--float-above {
          color: rgba(255, 255, 255, 0.9) !important;
        }
      }
    }
  `]
})
export class AddInvestmentDialogComponent {
  investmentForm: FormGroup;
  filteredSymbols$: Observable<any[]> | undefined;

  constructor(
    private fb: FormBuilder,
    private autocompleteService: AutocompleteService,
    public dialogRef: MatDialogRef<AddInvestmentDialogComponent>,
    @Inject(MAT_DIALOG_DATA) public data: InvestmentDialogData
  ) {
    this.investmentForm = this.createForm();

    if (data.mode === 'edit' && data.investment) {
      this.investmentForm.patchValue(data.investment);
    }

    this.setupSymbolAutocomplete();
  }

  private setupSymbolAutocomplete(): void {
    if (this.data.category === InvestmentCategory.STOCKS) {
      this.filteredSymbols$ = this.investmentForm.get('ticker')?.valueChanges.pipe(
        debounceTime(300),
        distinctUntilChanged(),
        switchMap(value => {
          if (value && value.length >= 2) {
            return this.autocompleteService.searchSymbols(value).pipe(
              catchError(() => of([]))
            );
          }
          return of([]);
        })
      );
    }
  }

  onSymbolSelected(symbol: string): void {
    const date = this.investmentForm.get('date')?.value;
    const dateStr = date ? new Date(date).toISOString().split('T')[0] : '';
    const investedAmount = this.investmentForm.get('amount')?.value || 0;

    this.autocompleteService.getPriceData(symbol, dateStr).subscribe(data => {
      let quantity = 0;
      let currentPrice = 0;

      // Handle current price
      if (data.quote && data.quote['Global Quote']) {
        currentPrice = parseFloat(data.quote['Global Quote']['05. price']);
      }

      // Handle historical price for quantity calculation
      if (dateStr && data.history && data.history['Time Series (Daily)']) {
        const history = data.history['Time Series (Daily)'];
        // The exact date might not be in the keys (e.g. if it's a weekend)
        // We'd ideally find the closest previous date, but for now we'll check exact match
        if (history[dateStr]) {
          const historicalPrice = parseFloat(history[dateStr]['4. close']);
          if (historicalPrice > 0 && investedAmount > 0) {
            quantity = investedAmount / historicalPrice;
          }
        }
      }

      // Fallback: if quantity still 0 but we have currentPrice, use that as a proxy for quantity
      if (quantity === 0 && currentPrice > 0 && investedAmount > 0) {
        quantity = investedAmount / currentPrice;
      }

      if (quantity > 0) {
        this.investmentForm.patchValue({
          quantity: quantity,
          currentValue: currentPrice * quantity
        });
      } else if (currentPrice > 0) {
        this.investmentForm.patchValue({
          currentValue: currentPrice * (this.investmentForm.get('quantity')?.value || 1)
        });
      }
    });
  }

  private createForm(): FormGroup {
    const formConfig: any = {
      name: ['', Validators.required],
      amount: [0, [Validators.required, Validators.min(0.01)]],
      date: [new Date(), Validators.required],
      notes: ['']
    };

    // Add category-specific fields
    switch (this.data.category) {
      case InvestmentCategory.STOCKS:
        formConfig.ticker = [''];
        formConfig.quantity = [0, [Validators.required, Validators.min(0.01)]];
        formConfig.currentValue = [0, [Validators.required, Validators.min(0)]];
        break;

      case InvestmentCategory.MUTUAL_FUNDS:
        formConfig.currentValue = [0, [Validators.required, Validators.min(0)]];
        break;

      case InvestmentCategory.FDS:
        formConfig.currentValue = [0, [Validators.required, Validators.min(0)]];
        formConfig.interestRate = [0, [Validators.required, Validators.min(0)]];
        formConfig.tenor = [0, [Validators.required, Validators.min(1)]];
        formConfig.maturityDate = [''];
        break;

      case InvestmentCategory.INSURANCE:
        formConfig.currentValue = [0, [Validators.required, Validators.min(0)]];
        formConfig.tenor = [0, [Validators.required, Validators.min(1)]];
        break;

      case InvestmentCategory.LOANS:
        formConfig.currentValue = [0, [Validators.required, Validators.min(0)]];
        formConfig.outstandingBalance = [0, [Validators.required, Validators.min(0)]];
        formConfig.interestRate = [0, [Validators.required, Validators.min(0)]];
        formConfig.tenor = [0, [Validators.required, Validators.min(1)]];
        break;
    }

    return this.fb.group(formConfig);
  }

  getCategoryTitle(): string {
    const titles: { [key: string]: string } = {
      [InvestmentCategory.STOCKS]: 'Stock',
      [InvestmentCategory.MUTUAL_FUNDS]: 'Mutual Fund',
      [InvestmentCategory.FDS]: 'Fixed Deposit',
      [InvestmentCategory.INSURANCE]: 'Insurance',
      [InvestmentCategory.LOANS]: 'Loan'
    };
    return titles[this.data.category] || 'Investment';
  }

  getCategoryIcon(): string {
    const icons: { [key: string]: string } = {
      [InvestmentCategory.STOCKS]: 'trending_up',
      [InvestmentCategory.MUTUAL_FUNDS]: 'account_balance',
      [InvestmentCategory.FDS]: 'savings',
      [InvestmentCategory.INSURANCE]: 'security',
      [InvestmentCategory.LOANS]: 'credit_card'
    };
    return icons[this.data.category] || 'add_circle';
  }

  getAmountLabel(): string {
    return this.data.category === InvestmentCategory.LOANS ? 'Principal Amount' : 'Investment Amount';
  }

  getTenorLabel(): string {
    if (this.data.category === InvestmentCategory.LOANS) {
      return 'Loan Tenure';
    } else if (this.data.category === InvestmentCategory.INSURANCE) {
      return 'Policy Tenure';
    }
    return 'Tenure';
  }

  onSave(): void {
    if (this.investmentForm.valid) {
      const { name, amount, date, notes, currentValue, ...rest } = this.investmentForm.value;

      // Map form fields to the structure expected by the backend
      const formValue = {
        category: this.data.category,
        name: name,
        investedAmount: amount,
        currentValue: currentValue || 0,
        notes: notes,
        investmentData: {
          ...rest
        }
      };

      // Specifically for stocks, ensure ticker and quantity are set
      if (this.data.category === InvestmentCategory.STOCKS) {
        formValue.investmentData.ticker = this.investmentForm.value.ticker;
        formValue.investmentData.quantity = this.investmentForm.value.quantity;
      }

      this.dialogRef.close(formValue);
    }
  }

  onCancel(): void {
    this.dialogRef.close();
  }
}
