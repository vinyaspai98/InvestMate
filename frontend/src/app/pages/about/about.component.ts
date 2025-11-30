import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';

@Component({
  selector: 'app-about',
  standalone: true,
  imports: [
    CommonModule,
    MatCardModule,
    MatButtonModule,
    MatIconModule
  ],
  template: `
    <div class="about-container">
      <div class="page-header">
        <h1>About InvestMate</h1>
        <p>Your trusted financial companion</p>
      </div>

      <div class="about-content">
        <!-- Mission Section -->
        <mat-card class="about-card">
          <mat-card-header>
            <mat-card-title>
              <mat-icon>rocket_launch</mat-icon>
              Our Mission
            </mat-card-title>
          </mat-card-header>
          <mat-card-content>
            <p>
              InvestMate is designed to simplify financial tracking for Indian investors. 
              We believe that everyone deserves access to powerful, easy-to-use tools 
              that help them make informed investment decisions and track their financial journey.
            </p>
            <p>
              Our platform provides a comprehensive view of your portfolio across various 
              asset classes including stocks, mutual funds, fixed deposits, insurance, and loans.
            </p>
          </mat-card-content>
        </mat-card>

        <!-- Features Section -->
        <mat-card class="about-card">
          <mat-card-header>
            <mat-card-title>
              <mat-icon>star</mat-icon>
              Key Features
            </mat-card-title>
          </mat-card-header>
          <mat-card-content>
            <div class="features-grid">
              <div class="feature-item">
                <mat-icon>dashboard</mat-icon>
                <h3>Comprehensive Dashboard</h3>
                <p>View your complete financial picture at a glance</p>
              </div>
              
              <div class="feature-item">
                <mat-icon>trending_up</mat-icon>
                <h3>Portfolio Tracking</h3>
                <p>Track stocks, mutual funds, FDs, and more</p>
              </div>
              
              <div class="feature-item">
                <mat-icon>analytics</mat-icon>
                <h3>Performance Analytics</h3>
                <p>Detailed charts and profit/loss analysis</p>
              </div>
              
              <div class="feature-item">
                <mat-icon>security</mat-icon>
                <h3>Secure & Private</h3>
                <p>Your financial data is encrypted and secure</p>
              </div>
            </div>
          </mat-card-content>
        </mat-card>

        <!-- Technology Section -->
        <mat-card class="about-card">
          <mat-card-header>
            <mat-card-title>
              <mat-icon>code</mat-icon>
              Technology Stack
            </mat-card-title>
          </mat-card-header>
          <mat-card-content>
            <div class="tech-stack">
              <div class="tech-category">
                <h4>Frontend</h4>
                <ul>
                  <li>Angular with TypeScript</li>
                  <li>Angular Material UI</li>
                  <li>Responsive Design</li>
                  <li>Progressive Web App</li>
                </ul>
              </div>
              
              <div class="tech-category">
                <h4>Backend</h4>
                <ul>
                  <li>GoLang (Gin Framework)</li>
                  <li>RESTful APIs</li>
                  <li>JWT Authentication</li>
                  <li>Microservices Architecture</li>
                </ul>
              </div>
              
              <div class="tech-category">
                <h4>Database & Cloud</h4>
                <ul>
                  <li>Firebase Firestore</li>
                  <li>Firebase Authentication</li>
                  <li>Cloud Functions</li>
                  <li>Real-time Sync</li>
                </ul>
              </div>
            </div>
          </mat-card-content>
        </mat-card>

        <!-- Contact Section -->
        <mat-card class="about-card">
          <mat-card-header>
            <mat-card-title>
              <mat-icon>contact_mail</mat-icon>
              Get in Touch
            </mat-card-title>
          </mat-card-header>
          <mat-card-content>
            <p>
              Have questions, suggestions, or need support? We'd love to hear from you!
            </p>
            <div class="contact-actions">
              <button mat-raised-button color="primary">
                <mat-icon>email</mat-icon>
                Contact Support
              </button>
              <button mat-button color="accent">
                <mat-icon>bug_report</mat-icon>
                Report an Issue
              </button>
            </div>
          </mat-card-content>
        </mat-card>

        <!-- Version Info -->
        <div class="version-info">
          <p>Version 1.0.0 | Built with ❤️ for Indian Investors</p>
        </div>
      </div>
    </div>
  `,
  styles: [`
    .about-container {
      max-width: 900px;
      margin: 0 auto;
      padding: 0;
    }

    .page-header {
      text-align: center;
      margin-bottom: 3rem;

      h1 {
        font-size: 2.5rem;
        font-weight: 600;
        color: #333;
        margin: 0 0 0.5rem 0;
      }

      p {
        font-size: 1.2rem;
        color: #666;
        margin: 0;
      }
    }

    .about-content {
      display: flex;
      flex-direction: column;
      gap: 2rem;
    }

    .about-card {
      border-radius: 12px;
      border: 1px solid rgba(0, 0, 0, 0.1);

      mat-card-header {
        padding: 1.5rem 1.5rem 0 1.5rem;

        mat-card-title {
          display: flex;
          align-items: center;
          gap: 0.75rem;
          font-size: 1.3rem;
          font-weight: 600;
          color: #333;

          mat-icon {
            color: #3f51b5;
          }
        }
      }

      mat-card-content {
        padding: 1.5rem;

        p {
          line-height: 1.6;
          color: #555;
          margin-bottom: 1rem;

          &:last-child {
            margin-bottom: 0;
          }
        }
      }
    }

    .features-grid {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
      gap: 2rem;
      margin-top: 1rem;

      .feature-item {
        text-align: center;

        mat-icon {
          font-size: 2.5rem;
          color: #3f51b5;
          margin-bottom: 0.5rem;
        }

        h3 {
          font-size: 1.1rem;
          font-weight: 600;
          color: #333;
          margin: 0.5rem 0;
        }

        p {
          font-size: 0.9rem;
          color: #666;
          margin: 0;
        }
      }
    }

    .tech-stack {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
      gap: 2rem;
      margin-top: 1rem;

      .tech-category {
        h4 {
          font-size: 1.1rem;
          font-weight: 600;
          color: #333;
          margin: 0 0 0.75rem 0;
          border-bottom: 2px solid #3f51b5;
          padding-bottom: 0.25rem;
          display: inline-block;
        }

        ul {
          list-style: none;
          padding: 0;
          margin: 0;

          li {
            padding: 0.5rem 0;
            color: #555;
            position: relative;
            padding-left: 1.5rem;

            &:before {
              content: '•';
              color: #3f51b5;
              position: absolute;
              left: 0;
              font-weight: bold;
            }
          }
        }
      }
    }

    .contact-actions {
      display: flex;
      gap: 1rem;
      margin-top: 1.5rem;
      flex-wrap: wrap;

      button {
        display: flex;
        align-items: center;
        gap: 0.5rem;
      }
    }

    .version-info {
      text-align: center;
      padding: 2rem 0;
      color: #666;
      font-size: 0.9rem;
      border-top: 1px solid rgba(0, 0, 0, 0.1);
      margin-top: 1rem;
    }

    // Dark theme
    :host-context(.dark-theme) {
      .page-header {
        h1 {
          color: #ffffff;
        }

        p {
          color: #b0b0b0;
        }
      }

      .about-card {
        background: #1e1e1e;
        border-color: rgba(255, 255, 255, 0.1);

        mat-card-header {
          mat-card-title {
            color: #ffffff;
          }
        }

        mat-card-content {
          p {
            color: #b0b0b0;
          }
        }

        .features-grid {
          .feature-item {
            h3 {
              color: #ffffff;
            }

            p {
              color: #b0b0b0;
            }
          }
        }

        .tech-stack {
          .tech-category {
            h4 {
              color: #ffffff;
            }

            ul li {
              color: #b0b0b0;
            }
          }
        }
      }

      .version-info {
        color: #b0b0b0;
        border-color: rgba(255, 255, 255, 0.1);
      }
    }

    // Responsive design
    @media (max-width: 768px) {
      .page-header {
        h1 {
          font-size: 2rem;
        }

        p {
          font-size: 1rem;
        }
      }

      .features-grid {
        grid-template-columns: 1fr;
      }

      .tech-stack {
        grid-template-columns: 1fr;
      }

      .contact-actions {
        flex-direction: column;

        button {
          width: 100%;
          justify-content: center;
        }
      }
    }
  `]
})
export class AboutComponent {}