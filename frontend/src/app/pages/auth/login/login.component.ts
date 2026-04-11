import { Component, ChangeDetectorRef } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router, RouterModule } from '@angular/router';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { AuthService } from '../../../services/auth.service';
import { ThemeService } from '../../../services/theme.service';

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [
    CommonModule,
    RouterModule,
    MatCardModule,
    MatButtonModule,
    MatIconModule,
    MatSnackBarModule,
    MatProgressSpinnerModule
  ],
  templateUrl: './login.component.html',
  styleUrls: ['./login.component.scss']
})
export class LoginComponent {
  isLoading = false;

  constructor(
    private authService: AuthService,
    private router: Router,
    private snackBar: MatSnackBar,
    public themeService: ThemeService,
    private cdr: ChangeDetectorRef
  ) { }

  onGoogleSignIn(): void {
    this.isLoading = true;
    this.cdr.detectChanges();

    this.authService.loginWithGoogle().subscribe({
      next: (user) => {
        this.snackBar.open('Sign in successful!', 'Close', { duration: 3000 });
        this.router.navigate(['/dashboard']);
        this.isLoading = false;
        this.cdr.detectChanges();
      },
      error: (error) => {
        this.snackBar.open(error.message || 'Sign in failed. Please try again.', 'Close', { duration: 5000 });
        this.isLoading = false;
        this.cdr.detectChanges();
      }
    });
  }

  toggleTheme(): void {
    this.themeService.toggleTheme();
  }
}