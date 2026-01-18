import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule, ReactiveFormsModule, FormBuilder, FormGroup, Validators } from '@angular/forms';
import { MatCardModule } from '@angular/material/card';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatSelectModule } from '@angular/material/select';
import { MatSlideToggleModule } from '@angular/material/slide-toggle';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { AuthService } from '../../services/auth.service';
import { ThemeService } from '../../services/theme.service';
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
    MatSnackBarModule
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

        <!-- Security -->
        <mat-card class="profile-card">
          <mat-card-header>
            <mat-card-title>Security</mat-card-title>
          </mat-card-header>
          <mat-card-content>
            <form [formGroup]="passwordForm" (ngSubmit)="onChangePassword()">
              <mat-form-field appearance="outline" class="full-width">
                <mat-label>Current Password</mat-label>
                <input matInput type="password" formControlName="currentPassword">
                <mat-icon matSuffix>lock</mat-icon>
              </mat-form-field>

              <mat-form-field appearance="outline" class="full-width">
                <mat-label>New Password</mat-label>
                <input matInput type="password" formControlName="newPassword">
                <mat-icon matSuffix>lock_open</mat-icon>
              </mat-form-field>

              <mat-form-field appearance="outline" class="full-width">
                <mat-label>Confirm New Password</mat-label>
                <input matInput type="password" formControlName="confirmPassword">
                <mat-icon matSuffix>lock_open</mat-icon>
              </mat-form-field>

              <div class="form-actions">
                <button mat-raised-button color="accent" type="submit" [disabled]="!passwordForm.valid">
                  <mat-icon>security</mat-icon>
                  Change Password
                </button>
              </div>
            </form>
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
export class ProfileComponent implements OnInit {
  profileForm: FormGroup;
  preferencesForm: FormGroup;
  passwordForm: FormGroup;
  currentUser: User | null = null;

  constructor(
    private fb: FormBuilder,
    private authService: AuthService,
    public themeService: ThemeService,
    private snackBar: MatSnackBar
  ) {
    this.profileForm = this.fb.group({
      name: ['', [Validators.required]],
      email: ['', [Validators.required, Validators.email]],
      contact: ['']
    });

    this.preferencesForm = this.fb.group({
      currency: ['INR'],
      darkMode: [false],
      notifications: [true]
    });

    this.passwordForm = this.fb.group({
      currentPassword: ['', [Validators.required]],
      newPassword: ['', [Validators.required, Validators.minLength(6)]],
      confirmPassword: ['', [Validators.required]]
    });
  }

  ngOnInit(): void {
    this.authService.currentUser$.subscribe(user => {
      if (user) {
        this.currentUser = user;
        this.profileForm.patchValue({
          name: user.name,
          email: user.email,
          contact: user.contact || ''
        });
        this.preferencesForm.patchValue({
          currency: user.preferences.currency,
          darkMode: user.preferences.theme === 'dark',
          notifications: user.preferences.notifications
        });
      }
    });
  }

  onUpdateProfile(): void {
    if (this.profileForm.valid && this.currentUser) {
      const updatedUser = {
        ...this.currentUser,
        ...this.profileForm.value
      };

      this.authService.updateUser(updatedUser).subscribe({
        next: () => {
          this.snackBar.open('Profile updated successfully!', 'Close', { duration: 3000 });
        },
        error: () => {
          this.snackBar.open('Failed to update profile.', 'Close', { duration: 3000 });
        }
      });
    }
  }

  onUpdatePreferences(): void {
    if (this.currentUser) {
      const updatedUser = {
        ...this.currentUser,
        preferences: {
          ...this.currentUser.preferences,
          currency: this.preferencesForm.get('currency')?.value,
          notifications: this.preferencesForm.get('notifications')?.value
        }
      };

      this.authService.updateUser(updatedUser).subscribe({
        next: () => {
          this.snackBar.open('Preferences saved!', 'Close', { duration: 3000 });
        },
        error: () => {
          this.snackBar.open('Failed to save preferences.', 'Close', { duration: 3000 });
        }
      });
    }
  }

  onThemeToggle(event: any): void {
    this.themeService.setDarkTheme(event.checked);
  }

  onChangePassword(): void {
    if (this.passwordForm.valid) {
      const { currentPassword, newPassword } = this.passwordForm.value;

      this.authService.changePassword(currentPassword, newPassword).subscribe({
        next: () => {
          this.snackBar.open('Password changed successfully!', 'Close', { duration: 3000 });
          this.passwordForm.reset();
        },
        error: () => {
          this.snackBar.open('Failed to change password.', 'Close', { duration: 3000 });
        }
      });
    }
  }
}