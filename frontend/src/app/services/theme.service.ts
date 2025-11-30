import { Injectable } from '@angular/core';
import { BehaviorSubject } from 'rxjs';

@Injectable({
  providedIn: 'root'
})
export class ThemeService {
  private isDarkTheme = new BehaviorSubject<boolean>(false);
  public isDarkTheme$ = this.isDarkTheme.asObservable();

  constructor() {
    // Check for saved theme preference or default to light theme
    const savedTheme = localStorage.getItem('theme');
    if (savedTheme) {
      this.setDarkTheme(savedTheme === 'dark');
    } else {
      // Check system preference
      const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
      this.setDarkTheme(prefersDark);
    }
  }

  setDarkTheme(isDark: boolean): void {
    this.isDarkTheme.next(isDark);
    localStorage.setItem('theme', isDark ? 'dark' : 'light');
    
    // Apply theme to document body
    if (isDark) {
      document.body.classList.add('dark-theme');
      document.body.classList.remove('light-theme');
    } else {
      document.body.classList.add('light-theme');
      document.body.classList.remove('dark-theme');
    }
  }

  toggleTheme(): void {
    this.setDarkTheme(!this.isDarkTheme.value);
  }

  getCurrentTheme(): 'light' | 'dark' {
    return this.isDarkTheme.value ? 'dark' : 'light';
  }
}