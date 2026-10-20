import { Component, OnInit, OnDestroy, ChangeDetectorRef } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Subscription } from 'rxjs';
import { FormsModule, ReactiveFormsModule, FormBuilder, FormGroup, Validators } from '@angular/forms';
import { MatCardModule } from '@angular/material/card';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatSelectModule } from '@angular/material/select';
import { MatSlideToggleModule } from '@angular/material/slide-toggle';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { AuthService } from '../../services/auth.service';
import { ThemeService } from '../../services/theme.service';
import { GmailService } from '../../services/gmail.service';
import { InvestmentService } from '../../services/investment.service';
import { User } from '../../models/user.model';

@Component({
  selector: 'app-profile',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    ReactiveFormsModule,
    MatCardModule,
    MatFormFieldModule,
    MatInputModule,
    MatButtonModule,
    MatIconModule,
    MatSelectModule,
    MatSlideToggleModule,
    MatSnackBarModule,
    MatProgressSpinnerModule
  ],
  template: `
    <div class="profile-container">
      <div class="page-header">
        <h1>Profile Settings</h1>
        <p>Manage your account and preferences</p>
      </div>

      <div class="profile-content">
        <!-- Personal Details -->
        <mat-card class="profile-card">
          <mat-card-header>
            <mat-card-title>Personal Details</mat-card-title>
          </mat-card-header>
          <mat-card-content>
            <form [formGroup]="profileForm" (ngSubmit)="onUpdateProfile()">
              <mat-form-field appearance="outline" class="full-width">
                <mat-label>Full Name</mat-label>
                <input matInput formControlName="name" placeholder="Enter your name">
                <mat-icon matSuffix>person</mat-icon>
              </mat-form-field>

              <mat-form-field appearance="outline" class="full-width">
                <mat-label>Email</mat-label>
                <input matInput formControlName="email" placeholder="Enter your email" readonly>
                <mat-icon matSuffix>email</mat-icon>
              </mat-form-field>

              <mat-form-field appearance="outline" class="full-width">
                <mat-label>Contact Number</mat-label>
                <input matInput formControlName="contact" placeholder="Enter your contact number">
                <mat-icon matSuffix>phone</mat-icon>
              </mat-form-field>

              <mat-form-field appearance="outline" class="full-width">
                <mat-label>PAN Card Number</mat-label>
                <input matInput formControlName="panNumber" placeholder="Enter your PAN (e.g. ABCDE1234F)"
                       style="text-transform: uppercase" maxlength="10">
                <mat-icon matSuffix>badge</mat-icon>
                <mat-hint>Required to decrypt your CDSL CAS Statement</mat-hint>
                <mat-error *ngIf="profileForm.get('panNumber')?.hasError('pattern')">
                  Please enter a valid 10-character PAN (e.g. ABCDE1234F)
                </mat-error>
              </mat-form-field>

              <div class="form-actions">
                <button mat-raised-button color="primary" type="submit" [disabled]="!profileForm.valid">
                  <mat-icon>save</mat-icon>
                  Update Profile
                </button>
              </div>
            </form>
          </mat-card-content>
        </mat-card>

        <!-- Preferences -->
        <mat-card class="profile-card">
          <mat-card-header>
            <mat-card-title>Preferences</mat-card-title>
          </mat-card-header>
          <mat-card-content>
            <form [formGroup]="preferencesForm" (ngSubmit)="onUpdatePreferences()">
              <mat-form-field appearance="outline" class="full-width">
                <mat-label>Currency</mat-label>
                <mat-select formControlName="currency">
                  <mat-option value="INR">Indian Rupee (₹)</mat-option>
                  <mat-option value="USD">US Dollar ($)</mat-option>
                  <mat-option value="EUR">Euro (€)</mat-option>
                </mat-select>
              </mat-form-field>

              <div class="preference-row">
                <span class="preference-label">Dark Mode</span>
                <mat-slide-toggle formControlName="darkMode" (change)="onThemeToggle($event)">
                  {{ (themeService.isDarkTheme$ | async) ? 'Enabled' : 'Disabled' }}
                </mat-slide-toggle>
              </div>

              <div class="preference-row">
                <span class="preference-label">Notifications</span>
                <mat-slide-toggle formControlName="notifications">
                  {{ preferencesForm.get('notifications')?.value ? 'Enabled' : 'Disabled' }}
                </mat-slide-toggle>
              </div>

              <div class="form-actions">
                <button mat-raised-button color="primary" type="submit">
                  <mat-icon>save</mat-icon>
                  Save Preferences
                </button>
              </div>
            </form>
          </mat-card-content>
        </mat-card>

        <!-- Gmail Sync -->
        <mat-card class="profile-card">
          <mat-card-header>
            <mat-card-title>Gmail & CDSL Sync</mat-card-title>
          </mat-card-header>
          <mat-card-content>
            <p class="sync-description">Hybrid portfolio sync from CDSL Consolidated Account Statement (CAS) PDF and transaction emails</p>
            <div class="sync-info">
              <mat-icon>info</mat-icon>
              <span>Scans "CDSL Consolidated Account Statement (CAS)" with PAN password, extracts full portfolio (stocks, mutual funds, ETFs), and syncs subsequent Demat transactions.</span>
            </div>
            <div class="form-actions">
              <button mat-raised-button color="primary" (click)="onSyncGmail()" [disabled]="isSyncing">
                <mat-spinner *ngIf="isSyncing" diameter="20"></mat-spinner>
                <mat-icon *ngIf="!isSyncing">sync</mat-icon>
                {{ isSyncing ? 'Syncing...' : 'Sync Now' }}
              </button>
            </div>
            <div *ngIf="casEffectiveDate" class="effective-date-info">
              <mat-icon>event_available</mat-icon>
              <span>CAS Effective Date: {{ casEffectiveDate | date:'MMM d, y' }}</span>
            </div>
            <div *ngIf="lastRefresh" class="last-refresh-info">
              <mat-icon>schedule</mat-icon>
              <span>Last Refresh: {{ lastRefresh | date:'h:mm:ss a, MMM d, y' }}</span>
              <span *ngIf="syncMode" class="sync-mode-tag">{{ syncMode }}</span>
            </div>
          </mat-card-content>
        </mat-card>

        <!-- Security -->
        <mat-card class="profile-card">
          <mat-card-header>
            <mat-card-title>Security</mat-card-title>
          </mat-card-header>
          <mat-card-content>
            <p class="security-note">
              <mat-icon>lock</mat-icon>
              Password management is handled securely by Google. Visit your Google Account settings to update your password.
            </p>
          </mat-card-content>
        </mat-card>
      </div>
    </div>
  `,
  styles: [`
    .profile-container {
      max-width: 800px;
      margin: 0 auto;
      padding: 0;
    }

    .page-header {
      margin-bottom: 2rem;

      h1 {
        font-size: 2rem;
        font-weight: 600;
        color: #333;
        margin: 0 0 0.5rem 0;
      }

      p {
        font-size: 1rem;
        color: #666;
        margin: 0;
      }
    }

    .profile-content {
      display: flex;
      flex-direction: column;
      gap: 2rem;
    }

    .profile-card {
      border-radius: 12px;
      border: 1px solid rgba(0, 0, 0, 0.1);

      mat-card-header {
        padding: 1.5rem 1.5rem 0 1.5rem;

        mat-card-title {
          font-size: 1.3rem;
          font-weight: 600;
          color: #333;
        }
      }

      mat-card-content {
        padding: 1.5rem;
      }
    }

    .full-width {
      width: 100%;
      margin-bottom: 1rem;
    }

    .preference-row {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 1rem 0;
      border-bottom: 1px solid rgba(0, 0, 0, 0.1);

      &:last-child {
        border-bottom: none;
      }

      .preference-label {
        font-weight: 500;
        color: #333;
      }
    }

    .sync-description {
      margin: 0 0 1rem 0;
      color: #666;
    }

    .sync-info {
      display: flex;
      align-items: flex-start;
      gap: 8px;
      margin-bottom: 1.5rem;
      padding: 12px;
      background: rgba(103, 126, 234, 0.1);
      border-radius: 8px;
      color: #5a67d8;
      font-size: 0.875rem;

      mat-icon {
        font-size: 20px;
        width: 20px;
        height: 20px;
      }
    }

    .security-note {
      display: flex;
      align-items: center;
      gap: 12px;
      margin: 0;
      padding: 16px;
      background: rgba(255, 193, 7, 0.1);
      border-radius: 8px;
      color: #f57c00;

      mat-icon {
        font-size: 24px;
        width: 24px;
        height: 24px;
      }
    }

    .form-actions {
      margin-top: 1.5rem;
      display: flex;
      justify-content: flex-start;

      button {
        display: flex;
        align-items: center;
        gap: 0.5rem;
      }
    }

    .effective-date-info {
      display: flex;
      align-items: center;
      gap: 8px;
      margin-top: 1rem;
      font-size: 0.875rem;
      color: #2e7d32;
      font-weight: 500;

      mat-icon {
        font-size: 18px;
        width: 18px;
        height: 18px;
        color: #2e7d32;
      }
    }

    .last-refresh-info {
      display: flex;
      align-items: center;
      gap: 8px;
      margin-top: 0.5rem;
      font-size: 0.875rem;
      color: #555;
      font-weight: 500;

      mat-icon {
        font-size: 18px;
        width: 18px;
        height: 18px;
        color: #666;
      }
    }

    .sync-mode-tag {
      margin-left: 8px;
      padding: 2px 8px;
      background: #e0e7ff;
      color: #3730a3;
      border-radius: 12px;
      font-size: 0.75rem;
      font-weight: 600;
      text-transform: uppercase;
    }

    // Dark theme
    :host-context(.dark-theme) {
      .page-header {
        h1 {
          color: #ffffff;
        }

        p {
          color: #b0b0b0;
        }
      }

      .profile-card {
        background: #1e1e1e;
        border-color: rgba(255, 255, 255, 0.1);

        mat-card-header {
          mat-card-title {
            color: #ffffff;
          }
        }

        .preference-row {
          border-color: rgba(255, 255, 255, 0.1);

          .preference-label {
            color: #ffffff;
          }
        }

        .sync-description {
          color: #b0b0b0;
        }

        .sync-info {
          background: rgba(103, 126, 234, 0.2);
          color: #a5b4fc;
        }

        .security-note {
          background: rgba(255, 193, 7, 0.2);
          color: #ffb74d;
        }

        .last-refresh-info {
          color: #b0b0b0;

          mat-icon {
            color: #9e9e9e;
          }
        }
      }

      // Material form fields
      mat-form-field {
        .mat-mdc-input-element,
        input.mat-mdc-input-element {
          color: #ffffff !important;
        }

        .mat-mdc-select-value {
          color: #ffffff !important;
        }

        .mat-mdc-select-arrow {
          color: rgba(255, 255, 255, 0.7) !important;
        }

        .mat-mdc-form-field-label,
        .mdc-floating-label {
          color: rgba(255, 255, 255, 0.7) !important;
        }

        .mat-mdc-form-field-label.mdc-floating-label--float-above {
          color: rgba(255, 255, 255, 0.9) !important;
        }

        [matPrefix],
        [matSuffix],
        .mat-icon {
          color: rgba(255, 255, 255, 0.7);
        }

        .mdc-notched-outline__leading,
        .mdc-notched-outline__notch,
        .mdc-notched-outline__trailing {
          border-color: rgba(255, 255, 255, 0.38) !important;
        }

        &:hover:not(.mat-form-field-disabled):not(.mat-focused) {
          .mdc-notched-outline__leading,
          .mdc-notched-outline__notch,
          .mdc-notched-outline__trailing {
            border-color: rgba(255, 255, 255, 0.87) !important;
          }

          .mat-mdc-floating-label,
          .mdc-floating-label,
          .mat-mdc-form-field-label {
            color: rgba(255, 255, 255, 0.9) !important;
          }
        }

        &.mat-focused {
          .mdc-notched-outline__leading,
          .mdc-notched-outline__notch,
          .mdc-notched-outline__trailing {
            border-color: #90caf9 !important;
          }

          .mat-mdc-floating-label,
          .mdc-floating-label,
          .mat-mdc-form-field-label {
            color: #90caf9 !important;
          }
        }
      }
    }
  `]
})
export class ProfileComponent implements OnInit, OnDestroy {
  profileForm: FormGroup;
  preferencesForm: FormGroup;
  currentUser: User | null = null;
  isSyncing = false;
  lastRefresh: Date | null = null;
  casEffectiveDate: Date | null = null;
  syncMode: string | null = null;
  private userSubscription?: Subscription;

  constructor(
    private fb: FormBuilder,
    private authService: AuthService,
    public themeService: ThemeService,
    private gmailService: GmailService,
    private investmentService: InvestmentService,
    private snackBar: MatSnackBar,
    private cdr: ChangeDetectorRef
  ) {
    this.profileForm = this.fb.group({
      name: ['', [Validators.required]],
      email: ['', [Validators.required, Validators.email]],
      contact: [''],
      panNumber: ['', [Validators.pattern(/^[A-Za-z]{5}[0-9]{4}[A-Za-z]{1}$/)]]
    });

    this.preferencesForm = this.fb.group({
      currency: ['INR'],
      darkMode: [false],
      notifications: [true]
    });
  }

  ngOnInit(): void {
    this.userSubscription = this.authService.currentUser$.subscribe(user => {
      if (user) {
        this.currentUser = user;
        if (user.lastGmailSync && !this.lastRefresh) {
          this.lastRefresh = new Date(user.lastGmailSync);
        }
        if (user.casEffectiveDate && !this.casEffectiveDate) {
          this.casEffectiveDate = new Date(user.casEffectiveDate);
        }
        this.profileForm.patchValue({
          name: user.name,
          email: user.email,
          contact: user.phoneNumber || user.contact || '',
          panNumber: user.panNumber || ''
        });

        const notificationsVal = typeof user.preferences?.notifications === 'boolean'
          ? user.preferences.notifications
          : !!user.preferences?.notifications?.email;

        this.preferencesForm.patchValue({
          currency: user.preferences?.currency || 'INR',
          darkMode: user.preferences?.theme === 'dark',
          notifications: notificationsVal
        });
        this.cdr.markForCheck();
      }
    });
  }

  ngOnDestroy(): void {
    this.userSubscription?.unsubscribe();
  }

  onUpdateProfile(): void {
    if (this.profileForm.valid) {
      const { name, contact, panNumber } = this.profileForm.value;

      this.authService.updateUserProfile({
        name,
        phoneNumber: contact,
        panNumber: panNumber ? panNumber.trim().toUpperCase() : undefined
      }).subscribe({
        next: (updated) => {
          if (this.currentUser) {
            this.currentUser.panNumber = updated?.panNumber || panNumber?.toUpperCase();
          }
          this.snackBar.open('Profile updated successfully!', 'Close', { duration: 3000 });
          this.cdr.markForCheck();
        },
        error: () => {
          this.snackBar.open('Failed to update profile.', 'Close', { duration: 3000 });
          this.cdr.markForCheck();
        }
      });
    }
  }

  onUpdatePreferences(): void {
    const currency = this.preferencesForm.get('currency')?.value;
    const notificationsEnabled = !!this.preferencesForm.get('notifications')?.value;
    const isDark = !!this.preferencesForm.get('darkMode')?.value;

    this.authService.updateUserPreferences({
      currency,
      theme: isDark ? 'dark' : 'light',
      notifications: { email: notificationsEnabled, push: false }
    }).subscribe({
      next: () => {
        this.snackBar.open('Preferences saved!', 'Close', { duration: 3000 });
        this.cdr.markForCheck();
      },
      error: () => {
        this.snackBar.open('Failed to save preferences.', 'Close', { duration: 3000 });
        this.cdr.markForCheck();
      }
    });
  }

  onThemeToggle(event: any): void {
    this.themeService.setDarkTheme(event.checked);
  }

  onSyncGmail(): void {
    const currentPAN = this.profileForm.get('panNumber')?.value || this.currentUser?.panNumber;
    if (!currentPAN || currentPAN.trim() === '') {
      this.snackBar.open('Please enter and save your PAN Card Number before syncing your CDSL CAS.', 'Close', { duration: 4000 });
      return;
    }

    this.isSyncing = true;
    this.cdr.detectChanges();

    this.gmailService.syncGmail().subscribe({
      next: (response) => {
        this.isSyncing = false;
        if (response?.lastGmailSync) {
          this.lastRefresh = new Date(response.lastGmailSync);
        } else {
          this.lastRefresh = new Date();
        }
        if (response?.portfolioEffectiveDate) {
          this.casEffectiveDate = new Date(response.portfolioEffectiveDate);
        }
        if (response?.syncMode) {
          this.syncMode = response.syncMode;
        }
        if (this.currentUser) {
          this.currentUser.lastGmailSync = this.lastRefresh.toISOString();
          if (this.casEffectiveDate) {
            this.currentUser.casEffectiveDate = this.casEffectiveDate.toISOString();
          }
        }
        this.investmentService.clearCache();
        const displayMsg = response?.message || 'Portfolio sync completed successfully!';
        this.snackBar.open(displayMsg, 'Close', { duration: 4000 });
        this.cdr.detectChanges();
      },
      error: (err) => {
        this.isSyncing = false;
        const errMsg = err?.error?.error || err?.message || 'Failed to sync Gmail. Please check your PAN and try again.';
        this.snackBar.open(errMsg, 'Close', { duration: 5000 });
        this.cdr.detectChanges();
      }
    });
  }
}