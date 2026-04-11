import { Injectable, inject } from '@angular/core';
import { Observable, BehaviorSubject, of, from, throwError } from 'rxjs';
import { map, catchError, tap, switchMap } from 'rxjs/operators';
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

@Injectable({
  providedIn: 'root'
})
export class AuthService {
  private auth = inject(Auth);
  private currentUserSubject = new BehaviorSubject<User | null>(null);
  public currentUser$ = this.currentUserSubject.asObservable();

  constructor() {
    // Subscribe to Firebase auth state changes
    user(this.auth).subscribe(firebaseUser => {
      if (firebaseUser) {
        this.mapFirebaseUserToUser(firebaseUser).then(user => {
          this.currentUserSubject.next(user);
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


  loginWithGoogle(): Observable<User> {
    const provider = new GoogleAuthProvider();
    // Add Gmail readonly scope for accessing emails
    provider.addScope('https://www.googleapis.com/auth/gmail.readonly');
    provider.setCustomParameters({
      prompt: 'select_account'
    });

    return from(signInWithPopup(this.auth, provider)).pipe(
      switchMap(credential => this.mapFirebaseUserToUser(credential.user)),
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
      contact: firebaseUser.phoneNumber || '',
      preferences: {
        currency: 'INR',
        theme: 'light',
        notifications: true
      }
    };
  }
}