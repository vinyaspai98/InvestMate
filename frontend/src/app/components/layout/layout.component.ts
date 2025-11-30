import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule, Router } from '@angular/router';
import { MatSidenavModule } from '@angular/material/sidenav';
import { MatToolbarModule } from '@angular/material/toolbar';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatListModule } from '@angular/material/list';
import { MatMenuModule } from '@angular/material/menu';
import { MatDividerModule } from '@angular/material/divider';
import { BreakpointObserver, Breakpoints } from '@angular/cdk/layout';
import { Observable } from 'rxjs';
import { map, shareReplay } from 'rxjs/operators';
import { AuthService } from '../../services/auth.service';
import { ThemeService } from '../../services/theme.service';
import { User } from '../../models/user.model';

@Component({
  selector: 'app-layout',
  standalone: true,
  imports: [
    CommonModule,
    RouterModule,
    MatSidenavModule,
    MatToolbarModule,
    MatButtonModule,
    MatIconModule,
    MatListModule,
    MatMenuModule,
    MatDividerModule
  ],
  templateUrl: './layout.component.html',
  styleUrls: ['./layout.component.scss']
})
export class LayoutComponent implements OnInit {
  user$: Observable<User | null>;
  currentUser: User | null = null;
  isHandset$: Observable<boolean>;

  navigationItems = [
    {
      label: 'Dashboard',
      icon: 'dashboard',
      route: '/dashboard',
      active: true
    },
    {
      label: 'Stocks',
      icon: 'trending_up',
      route: '/investments/stocks',
      active: false
    },
    {
      label: 'Mutual Funds',
      icon: 'account_balance',
      route: '/investments/mutual_funds',
      active: false
    },
    {
      label: 'Fixed Deposits',
      icon: 'savings',
      route: '/investments/fds',
      active: false
    },
    {
      label: 'Insurance',
      icon: 'security',
      route: '/investments/insurance',
      active: false
    },
    {
      label: 'Loans',
      icon: 'credit_card',
      route: '/investments/loans',
      active: false
    },
    {
      label: 'Profile',
      icon: 'person',
      route: '/profile',
      active: false
    },
    {
      label: 'About Us',
      icon: 'info',
      route: '/about',
      active: false
    }
  ];

  constructor(
    private breakpointObserver: BreakpointObserver,
    private authService: AuthService,
    private themeService: ThemeService,
    private router: Router
  ) {
    this.user$ = this.authService.currentUser$;
    this.isHandset$ = this.breakpointObserver.observe(Breakpoints.Handset)
      .pipe(
        map(result => result.matches),
        shareReplay()
      );
  }

  ngOnInit(): void {
    this.user$.subscribe(user => {
      this.currentUser = user;
    });
  }

  onNavigate(route: string): void {
    // Update active state
    this.navigationItems.forEach(item => {
      item.active = item.route === route;
    });
    
    this.router.navigate([route]);
  }

  toggleTheme(): void {
    this.themeService.toggleTheme();
  }

  onLogout(): void {
    this.authService.logout();
    this.router.navigate(['/login']);
  }

  get isDarkTheme$(): Observable<boolean> {
    return this.themeService.isDarkTheme$;
  }

  get userName(): string {
    return this.currentUser?.name || 'User';
  }

  get userEmail(): string {
    return this.currentUser?.email || '';
  }
}