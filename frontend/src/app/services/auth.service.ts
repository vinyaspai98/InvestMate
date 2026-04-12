import { Injectable, inject } from '@angular/core';
import { Observable, BehaviorSubject, of, from, throwError } from 'rxjs';
import { map, catchError, tap, switchMap } from 'rxjs/operators';
import { HttpClient } from '@angular/common/http';
import {
  Auth,
  signInWithPopup,
  GoogleAuthProvider,
  signOut,
  updateProfile,
  User as FirebaseUser,
  user
} from '@angular/fire/auth';
import { User, AuthRequest, SignupRequest } from '../models/user.model';
import { environment } from '../../environments/environment';

@Injectable({
  providedIn: 'root'
})
export class AuthService {
  private auth = inject(Auth);
  private http = inject(HttpClient);
  private currentUserSubject = new BehaviorSubject<User | null>(null);
  public currentUser$ = this.currentUserSubject.asObservable();
  private apiUrl = environment.apiBaseUrl || 'http://localhost:8080/api/v1';

  constructor() {
    // Subscribe to Firebase auth state changes
    user(this.auth).subscribe(firebaseUser => {
      if (firebaseUser) {
        // Sync with backend on initial load/state change
        this.syncWithBackend(firebaseUser).subscribe({
          next: (user) => this.currentUserSubject.next(user),
          error: (err) => console.error('Failed to sync user with backend:', err)
        });
      } else {
        this.currentUserSubject.next(null);
      }
    });
  }

  async getIdToken(): Promise<string | null> {
    const user = this.auth.currentUser;
    if (user) {
      try {
        return await user.getIdToken();
      } catch (error) {
        console.error('Error getting ID token:', error);
        return null;
      }
    }
    return null;
  }

  private syncWithBackend(firebaseUser: FirebaseUser, gmailAccessToken?: string): Observable<User> {
    return from(firebaseUser.getIdToken()).pipe(
      switchMap(token => {
        return this.http.post<{ user: User }>(`${this.apiUrl}/auth/verify-token`, {
          token,
          gmailAccessToken
        }).pipe(
          map(response => response.user),
          catchError(error => {
            console.error('Backend sync error:', error);
            // Fallback to local mapping if backend fails, but this might lead to 500s later
            return from(this.mapFirebaseUserToUser(firebaseUser));
          })
        );
      })
    );
  }

  loginWithGoogle(): Observable<User> {
    const provider = new GoogleAuthProvider();
    // Add Gmail readonly scope for accessing emails
    provider.addScope('https://www.googleapis.com/auth/gmail.readonly');
    provider.setCustomParameters({
      prompt: 'select_account'
    });

    return from(signInWithPopup(this.auth, provider)).pipe(
      switchMap(result => {
        const credential = GoogleAuthProvider.credentialFromResult(result);
        const gmailAccessToken = credential?.accessToken || undefined;
        return this.syncWithBackend(result.user, gmailAccessToken);
      }),
      tap(user => this.currentUserSubject.next(user)),
      catchError(error => {
        console.error('Google Sign-In error:', error);
        let errorMessage = 'Sign in failed. Please try again.';

        switch (error.code) {
          case 'auth/popup-closed-by-user':
            errorMessage = 'Sign in cancelled.';
            break;
          case 'auth/popup-blocked':
            errorMessage = 'Pop-up blocked. Please allow pop-ups for this site.';
            break;
          case 'auth/cancelled-popup-request':
            errorMessage = 'Sign in cancelled.';
            break;
          case 'auth/account-exists-with-different-credential':
            errorMessage = 'An account already exists with the same email address.';
            break;
        }

        return throwError(() => new Error(errorMessage));
      })
    );
  }

  // Legacy methods kept for backward compatibility but not used
  login(authRequest: AuthRequest): Observable<User> {
    // Redirect to Google Sign-In
    return this.loginWithGoogle();
  }

  signup(signupRequest: SignupRequest): Observable<User> {
    // Redirect to Google Sign-In
    return this.loginWithGoogle();
  }

  logout(): Observable<void> {
    return from(signOut(this.auth)).pipe(
      tap(() => this.currentUserSubject.next(null)),
      catchError(error => {
        console.error('Logout error:', error);
        return throwError(() => new Error('Logout failed. Please try again.'));
      })
    );
  }

  isAuthenticated(): boolean {
    return this.currentUserSubject.value !== null;
  }

  getCurrentUser(): User | null {
    return this.currentUserSubject.value;
  }

  updateUser(user: User): Observable<User> {
    const firebaseUser = this.auth.currentUser;

    if (!firebaseUser) {
      return throwError(() => new Error('No authenticated user'));
    }

    return from(updateProfile(firebaseUser, {
      displayName: user.name
    })).pipe(
      map(() => {
        const updatedUser: User = {
          ...user,
          id: firebaseUser.uid,
          email: firebaseUser.email || user.email
        };
        this.currentUserSubject.next(updatedUser);
        return updatedUser;
      }),
      catchError(error => {
        console.error('Update user error:', error);
        return throwError(() => new Error('Failed to update profile. Please try again.'));
      })
    );
  }

  // Password methods removed - using Google OAuth

  // Helper method to map Firebase User to our User model
  private async mapFirebaseUserToUser(firebaseUser: FirebaseUser): Promise<User> {
    return {
      id: firebaseUser.uid,
      email: firebaseUser.email || '',
      name: firebaseUser.displayName || 'User',
      preferences: {
        currency: 'INR',
        theme: 'light',
        notifications: true
      }
    };
  }
}