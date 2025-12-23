import { HttpInterceptorFn, HttpErrorResponse } from '@angular/common/http';
import { inject } from '@angular/core';
import { AuthService } from '../services/auth.service';
import { catchError, switchMap } from 'rxjs/operators';
import { throwError, from } from 'rxjs';

export const authInterceptor: HttpInterceptorFn = (req, next) => {
  const authService = inject(AuthService);

  // Skip auth for auth endpoints (backend doesn't have login/register endpoints)
  // Firebase handles authentication, backend only verifies tokens
  if (req.url.includes('/auth/verify-token')) {
    return next(req);
  }

  // Get token from Firebase
  return from(authService.getIdToken()).pipe(
    switchMap(token => {
      if (token) {
        // Clone request and add Authorization header
        const clonedReq = req.clone({
          setHeaders: {
            Authorization: `Bearer ${token}`,
            'Content-Type': 'application/json'
          }
        });
        return next(clonedReq);
      }
      // No token available - this will cause 401
      console.warn('No auth token available - request will fail with 401');
      return throwError(() => new Error('Authentication required. Please log in.'));
    }),
    catchError(error => {
      // Only catch errors related to GETTING the token, not HTTP response errors
      // If it's an HttpErrorResponse, it means the request was made and failed
      // Pass it through to be handled by the service layer
      if (error instanceof HttpErrorResponse) {
        console.error('HTTP Error in interceptor (passing through):', error.status, error.statusText);
        return throwError(() => error);
      }
      
      // This is an error getting the token itself (Firebase error)
      console.error('Error getting auth token from Firebase:', error);
      if (error.message && error.message.includes('Authentication required')) {
        return throwError(() => error);
      }
      return throwError(() => new Error('Failed to authenticate request. Please try logging in again.'));
    })
  );
};