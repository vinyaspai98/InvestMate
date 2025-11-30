import { Routes } from '@angular/router';
import { AuthGuard } from './guards/auth.guard';

export const routes: Routes = [
  { path: '', redirectTo: '/dashboard', pathMatch: 'full' },
  { 
    path: 'login', 
    loadComponent: () => import('./pages/auth/login/login.component').then(m => m.LoginComponent) 
  },
  { 
    path: 'signup', 
    loadComponent: () => import('./pages/auth/signup/signup.component').then(m => m.SignupComponent) 
  },
  {
    path: '',
    loadComponent: () => import('./components/layout/layout.component').then(m => m.LayoutComponent),
    canActivate: [AuthGuard],
    children: [
      { 
        path: 'dashboard', 
        loadComponent: () => import('./pages/dashboard/dashboard.component').then(m => m.DashboardComponent) 
      },
      { 
        path: 'investments/:category', 
        loadComponent: () => import('./pages/investment-detail/investment-detail.component').then(m => m.InvestmentDetailComponent) 
      },
      { 
        path: 'profile', 
        loadComponent: () => import('./pages/profile/profile.component').then(m => m.ProfileComponent) 
      },
      { 
        path: 'about', 
        loadComponent: () => import('./pages/about/about.component').then(m => m.AboutComponent) 
      }
    ]
  },
  { path: '**', redirectTo: '/dashboard' }
];
