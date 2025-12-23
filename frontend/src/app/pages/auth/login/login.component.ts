import { Component, ChangeDetectorRef } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule, ReactiveFormsModule, FormBuilder, FormGroup, Validators } from '@angular/forms';
import { Router, RouterModule } from '@angular/router';
import { MatCardModule } from '@angular/material/card';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { AuthService } from '../../../services/auth.service';
import { ThemeService } from '../../../services/theme.service';

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    ReactiveFormsModule,
    RouterModule,
    MatCardModule,
    MatFormFieldModule,
    MatInputModule,
    MatButtonModule,
    MatIconModule,
    MatSnackBarModule
  ],
  templateUrl: './login.component.html',
  styleUrls: ['./login.component.scss']
})
export class LoginComponent {
  loginForm: FormGroup;
  isLoading = false;
  hidePassword = true;

  constructor(
    private fb: FormBuilder,
    private authService: AuthService,
    private router: Router,
    private snackBar: MatSnackBar,
    public themeService: ThemeService,
    private cdr: ChangeDetectorRef
  ) {
    this.loginForm = this.fb.group({
      email: ['user@investmate.com', [Validators.required, Validators.email]],
      password: ['password', [Validators.required, Validators.minLength(6)]]
    });
  }

  onLogin(): void {
    if (this.loginForm.valid) {
      this.isLoading = true;
      this.cdr.detectChanges(); // Trigger change detection immediately
      
      const { email, password } = this.loginForm.value;

      this.authService.login({ email, password }).subscribe({
        next: (user) => {
          this.snackBar.open('Login successful!', 'Close', { duration: 3000 });
          this.router.navigate(['/dashboard']);
          this.isLoading = false;
          this.cdr.detectChanges();
        },
        error: (error) => {
          this.snackBar.open('Login failed. Please check your credentials.', 'Close', { duration: 5000 });
          this.isLoading = false;
          this.cdr.detectChanges();
        }
      });
    } else {
      this.markFormGroupTouched();
    }
  }

  onResetPassword(): void {
    const email = this.loginForm.get('email')?.value;
    if (!email) {
      this.snackBar.open('Please enter your email address first.', 'Close', { duration: 3000 });
      return;
    }

    this.authService.resetPassword(email).subscribe({
      next: () => {
        this.snackBar.open('Password reset email sent!', 'Close', { duration: 3000 });
      },
      error: () => {
        this.snackBar.open('Failed to send reset email.', 'Close', { duration: 3000 });
      }
    });
  }

  toggleTheme(): void {
    this.themeService.toggleTheme();
  }

  private markFormGroupTouched(): void {
    Object.keys(this.loginForm.controls).forEach(key => {
      const control = this.loginForm.get(key);
      control?.markAsTouched();
    });
  }

  getErrorMessage(field: string): string {
    const control = this.loginForm.get(field);
    if (control?.hasError('required')) {
      return `${field.charAt(0).toUpperCase() + field.slice(1)} is required`;
    }
    if (control?.hasError('email')) {
      return 'Please enter a valid email address';
    }
    if (control?.hasError('minlength')) {
      return `${field.charAt(0).toUpperCase() + field.slice(1)} must be at least 6 characters`;
    }
    return '';
  }
}