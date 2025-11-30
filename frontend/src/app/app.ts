import { Component, signal, OnInit } from '@angular/core';
import { RouterOutlet } from '@angular/router';
import { ThemeService } from './services/theme.service';

@Component({
  selector: 'app-root',
  imports: [RouterOutlet],
  templateUrl: './app.html',
  styleUrl: './app.scss'
})
export class App implements OnInit {
  protected readonly title = signal('InvestMate');

  constructor(private themeService: ThemeService) {}

  ngOnInit(): void {
    // Initialize theme service to apply saved theme
  }
}
