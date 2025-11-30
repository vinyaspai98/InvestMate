import { Injectable } from '@angular/core';
import { Observable, BehaviorSubject, of } from 'rxjs';
import { User, AuthRequest, SignupRequest } from '../models/user.model';

@Injectable({
  providedIn: 'root'
})
export class AuthService {
  private currentUserSubject = new BehaviorSubject<User | null>(null);
  public currentUser$ = this.currentUserSubject.asObservable();

  // Mock user for demonstration
  private mockUser: User = {
    id: '1',
    email: 'user@investmate.com',
    name: 'John Doe',
    contact: '+91 9876543210',
    preferences: {
      currency: 'INR',
      theme: 'light',
      notifications: true
    }
  };

  constructor() {
    // Check for existing session
    const savedUser = localStorage.getItem('currentUser');
    if (savedUser) {
      this.currentUserSubject.next(JSON.parse(savedUser));
    }
  }

  login(authRequest: AuthRequest): Observable<User> {
    // Mock authentication - in real app, this would call Firebase Auth
    if (authRequest.email === 'user@investmate.com' && authRequest.password === 'password') {
      localStorage.setItem('currentUser', JSON.stringify(this.mockUser));
      this.currentUserSubject.next(this.mockUser);
      return of(this.mockUser);
    }
    throw new Error('Invalid credentials');
  }

  signup(signupRequest: SignupRequest): Observable<User> {
    // Mock signup - in real app, this would call Firebase Auth
    const newUser: User = {
      id: Date.now().toString(),
      email: signupRequest.email,
      name: signupRequest.name,
      preferences: {
        currency: 'INR',
        theme: 'light',
        notifications: true
      }
    };
    
    localStorage.setItem('currentUser', JSON.stringify(newUser));
    this.currentUserSubject.next(newUser);
    return of(newUser);
  }

  logout(): void {
    localStorage.removeItem('currentUser');
    this.currentUserSubject.next(null);
  }

  isAuthenticated(): boolean {
    return this.currentUserSubject.value !== null;
  }

  getCurrentUser(): User | null {
    return this.currentUserSubject.value;
  }

  updateUser(user: User): Observable<User> {
    localStorage.setItem('currentUser', JSON.stringify(user));
    this.currentUserSubject.next(user);
    return of(user);
  }

  resetPassword(email: string): Observable<boolean> {
    // Mock password reset
    console.log('Password reset requested for:', email);
    return of(true);
  }

  changePassword(currentPassword: string, newPassword: string): Observable<boolean> {
    // Mock password change
    console.log('Password change requested');
    return of(true);
  }
}