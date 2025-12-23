import { Injectable, inject } from '@angular/core';
import { Observable, BehaviorSubject, of, from, throwError } from 'rxjs';
import { map, catchError, tap, switchMap } from 'rxjs/operators';
import { 
  Auth, 
  signInWithEmailAndPassword, 
  createUserWithEmailAndPassword,
  signOut,
  sendPasswordResetEmail,
  updatePassword,
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

  login(authRequest: AuthRequest): Observable<User> {
    return from(
      signInWithEmailAndPassword(this.auth, authRequest.email, authRequest.password)
    ).pipe(
      switchMap(credential => this.mapFirebaseUserToUser(credential.user)),
      tap(user => this.currentUserSubject.next(user)),
      catchError(error => {
        console.error('Login error:', error);
        let errorMessage = 'Login failed. Please try again.';
        
        switch (error.code) {
          case 'auth/user-not-found':
            errorMessage = 'No account found with this email address.';
            break;
          case 'auth/wrong-password':
            errorMessage = 'Incorrect password. Please try again.';
            break;
          case 'auth/invalid-email':
            errorMessage = 'Invalid email address format.';
            break;
          case 'auth/user-disabled':
            errorMessage = 'This account has been disabled.';
            break;
          case 'auth/too-many-requests':
            errorMessage = 'Too many failed login attempts. Please try again later.';
            break;
        }
        
        return throwError(() => new Error(errorMessage));
      })
    );
  }

  signup(signupRequest: SignupRequest): Observable<User> {
    return from(
      createUserWithEmailAndPassword(this.auth, signupRequest.email, signupRequest.password)
    ).pipe(
      switchMap(credential => {
        // Update display name in Firebase
        return from(updateProfile(credential.user, { 
          displayName: signupRequest.name 
        })).pipe(
          map(() => credential.user)
        );
      }),
      switchMap(firebaseUser => this.mapFirebaseUserToUser(firebaseUser)),
      tap(user => this.currentUserSubject.next(user)),
      catchError(error => {
        console.error('Signup error:', error);
        let errorMessage = 'Signup failed. Please try again.';
        
        switch (error.code) {
          case 'auth/email-already-in-use':
            errorMessage = 'An account with this email already exists.';
            break;
          case 'auth/invalid-email':
            errorMessage = 'Invalid email address format.';
            break;
          case 'auth/weak-password':
            errorMessage = 'Password is too weak. Please use a stronger password.';
            break;
        }
        
        return throwError(() => new Error(errorMessage));
      })
    );
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

  resetPassword(email: string): Observable<boolean> {
    return from(sendPasswordResetEmail(this.auth, email)).pipe(
      map(() => true),
      catchError(error => {
        console.error('Password reset error:', error);
        let errorMessage = 'Failed to send password reset email.';
        
        switch (error.code) {
          case 'auth/user-not-found':
            errorMessage = 'No account found with this email address.';
            break;
          case 'auth/invalid-email':
            errorMessage = 'Invalid email address format.';
            break;
        }
        
        return throwError(() => new Error(errorMessage));
      })
    );
  }

  changePassword(currentPassword: string, newPassword: string): Observable<boolean> {
    const firebaseUser = this.auth.currentUser;
    
    if (!firebaseUser || !firebaseUser.email) {
      return throwError(() => new Error('No authenticated user'));
    }

    // Re-authenticate user before changing password
    return from(
      signInWithEmailAndPassword(this.auth, firebaseUser.email, currentPassword)
    ).pipe(
      switchMap(() => from(updatePassword(firebaseUser, newPassword))),
      map(() => true),
      catchError(error => {
        console.error('Change password error:', error);
        let errorMessage = 'Failed to change password.';
        
        switch (error.code) {
          case 'auth/wrong-password':
            errorMessage = 'Current password is incorrect.';
            break;
          case 'auth/weak-password':
            errorMessage = 'New password is too weak.';
            break;
          case 'auth/requires-recent-login':
            errorMessage = 'Please log in again before changing your password.';
            break;
        }
        
        return throwError(() => new Error(errorMessage));
      })
    );
  }

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