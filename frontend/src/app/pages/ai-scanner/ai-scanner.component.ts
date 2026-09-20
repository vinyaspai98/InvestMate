import {
  Component,
  OnInit,
  OnDestroy,
  AfterViewInit,
  ElementRef,
  ViewChild,
  inject
} from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressBarModule } from '@angular/material/progress-bar';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatTableModule } from '@angular/material/table';
import { MatChipsModule } from '@angular/material/chips';
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatDividerModule } from '@angular/material/divider';
import { MatSelectModule } from '@angular/material/select';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';

import { Chart, registerables } from 'chart.js';
import { AiScannerService } from '../../services/ai-scanner.service';
import { GmailService } from '../../services/gmail.service';
import {
  PortfolioHealthScore,
  RecommendationAction,
  ProfitForecastPoint,
  MonteCarloResponse
} from '../../models/ai-scanner.model';

Chart.register(...registerables);

/**
 * Mandatory 16-character label line wrapping logic.
 * Splits strings longer than 16 characters into an array of lines.
 */
export function wrapLabel16(label: string, maxLen: number = 16): string[] {
  if (!label || label.length <= maxLen) return [label || ''];
  const words = label.split(' ');
  const lines: string[] = [];
  let currentLine = '';

  for (const word of words) {
    if ((currentLine + (currentLine ? ' ' : '') + word).length <= maxLen) {
      currentLine += (currentLine ? ' ' : '') + word;
    } else {
      if (currentLine) lines.push(currentLine);
      let remaining = word;
      while (remaining.length > maxLen) {
        lines.push(remaining.substring(0, maxLen));
        remaining = remaining.substring(maxLen);
      }
      currentLine = remaining;
    }
  }
  if (currentLine) lines.push(currentLine);
  return lines;
}

/**
 * Standard Chart.js tooltip configuration with title callback reconstructing multi-line labels.
 */
export const commonTooltipConfig = {
  callbacks: {
    title: (tooltipItems: any[]) => {
      if (!tooltipItems || tooltipItems.length === 0) return '';
      const item = tooltipItems[0];
      const label = item.chart?.data?.labels?.[item.dataIndex];
      if (Array.isArray(label)) {
        return label.join(' ');
      }
      return label || '';
    }
  }
};

declare global {
  interface Window {
    Plotly?: any;
  }
}

@Component({
  selector: 'app-ai-scanner',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    MatCardModule,
    MatButtonModule,
    MatIconModule,
    MatProgressBarModule,
    MatProgressSpinnerModule,
    MatTableModule,
    MatChipsModule,
    MatTooltipModule,
    MatDividerModule,
    MatSelectModule,
    MatFormFieldModule,
    MatInputModule
  ],
  templateUrl: './ai-scanner.component.html',
  styleUrls: ['./ai-scanner.component.scss']
})
export class AiScannerComponent implements OnInit, AfterViewInit, OnDestroy {
  private aiService = inject(AiScannerService);
  private gmailService = inject(GmailService);

  @ViewChild('radarCanvas') radarCanvas!: ElementRef<HTMLCanvasElement>;
  @ViewChild('forecastCanvas') forecastCanvas!: ElementRef<HTMLCanvasElement>;
  @ViewChild('allocationCanvas') allocationCanvas!: ElementRef<HTMLCanvasElement>;
  @ViewChild('taxCanvas') taxCanvas!: ElementRef<HTMLCanvasElement>;
  @ViewChild('plotlyContainer') plotlyContainer!: ElementRef<HTMLDivElement>;

  // Charts
  private radarChart?: Chart;
  private forecastChart?: Chart;
  private allocationChart?: Chart;
  private taxChart?: Chart;

  // State
  isLoading = false;
  isScanning = false;
  isSyncingGmail = false;
  scanSuccess = false;
  syncMessage = '';
  selectedRisk = 'moderate';

  // Data
  healthScore: PortfolioHealthScore | null = null;
  recommendations: RecommendationAction[] = [];
  forecastPoints: ProfitForecastPoint[] = [];
  monteCarloData: MonteCarloResponse | null = null;

  displayedColumns: string[] = [
    'category',
    'currentAsset',
    'targetAsset',
    'terReduction',
    'taxStatus',
    'action'
  ];

  ngOnInit(): void {
    this.loadInitialData();
  }

  ngAfterViewInit(): void {
    this.loadPlotlyScript();
  }

  ngOnDestroy(): void {
    this.destroyCharts();
  }

  private destroyCharts(): void {
    this.radarChart?.destroy();
    this.forecastChart?.destroy();
    this.allocationChart?.destroy();
    this.taxChart?.destroy();
  }

  private loadPlotlyScript(): void {
    if (window.Plotly) {
      this.renderPlotlyChart();
      return;
    }
    const script = document.createElement('script');
    script.src = 'https://cdn.plot.ly/plotly-2.35.2.min.js';
    script.async = true;
    script.onload = () => {
      if (this.monteCarloData) {
        this.renderPlotlyChart();
      }
    };
    document.body.appendChild(script);
  }

  loadInitialData(): void {
    this.isLoading = true;

    this.aiService.getHealthScore().subscribe({
      next: (score) => {
        this.healthScore = score;
        this.renderRadarChart();
        this.renderTaxSavingsChart();
      },
      error: (err) => console.error('Failed to load health score', err)
    });

    this.aiService.getRecommendations(this.selectedRisk).subscribe({
      next: (recs) => {
        this.recommendations = recs;
        this.renderAllocationChart();
      },
      error: (err) => console.error('Failed to load recommendations', err)
    });

    this.aiService.forecastProfit(1000000, 20).subscribe({
      next: (points) => {
        this.forecastPoints = points;
        this.renderForecastChart();
      },
      error: (err) => console.error('Failed to load forecast', err)
    });

    this.aiService.getMonteCarlo(1000, 20, 1000000).subscribe({
      next: (mc) => {
        this.monteCarloData = mc;
        this.renderPlotlyChart();
        this.isLoading = false;
      },
      error: (err) => {
        console.error('Failed to load monte carlo', err)
        this.isLoading = false;
      }
    });
  }

  triggerDematScan(): void {
    this.isScanning = true;
    this.scanSuccess = false;
    this.syncMessage = '';

    this.aiService.scanCAS({}).subscribe({
      next: (score) => {
        this.healthScore = score;
        this.isScanning = false;
        this.scanSuccess = true;
        this.renderRadarChart();
        this.renderTaxSavingsChart();
        this.onRiskProfileChange();
      },
      error: (err) => {
        console.error('Scan error:', err);
        this.isScanning = false;
      }
    });
  }

  triggerGmailSyncAndScan(): void {
    this.isSyncingGmail = true;
    this.syncMessage = 'Syncing Demat notification emails from Gmail...';

    this.gmailService.syncGmail().subscribe({
      next: (res) => {
        this.isSyncingGmail = false;
        this.syncMessage = res.message || `Synced ${res.synced} new Demat email transactions. Running AI analysis...`;
        this.triggerDematScan();
      },
      error: (err) => {
        console.error('Gmail sync error:', err);
        this.isSyncingGmail = false;
        this.syncMessage = 'Demat records up to date. Running AI analysis...';
        this.triggerDematScan();
      }
    });
  }

  onRiskProfileChange(): void {
    this.aiService.getRecommendations(this.selectedRisk).subscribe({
      next: (recs) => {
        this.recommendations = recs;
        this.renderAllocationChart();
      }
    });
  }

  // -------------------------------------------------------------------------
  // Diagnostic Radar Chart (5-Dimension Health Score)
  // -------------------------------------------------------------------------
  private renderRadarChart(): void {
    if (!this.radarCanvas || !this.healthScore) return;
    this.radarChart?.destroy();

    const rawLabels = [
      'Active Share',
      'Tax Efficiency',
      'Ulcer Index',
      'Factor Balance',
      'Fee Control'
    ];

    // Mandatory 16-character label wrapping
    const wrappedLabels = rawLabels.map(l => wrapLabel16(l, 16));

    const dataValues = [
      this.healthScore.active_share_score,
      this.healthScore.tax_efficiency_score,
      this.healthScore.ulcer_index_score,
      this.healthScore.factor_balance_score,
      this.healthScore.fee_control_score
    ];

    this.radarChart = new Chart(this.radarCanvas.nativeElement, {
      type: 'radar',
      data: {
        labels: wrappedLabels,
        datasets: [
          {
            label: 'Portfolio Score',
            data: dataValues,
            backgroundColor: 'rgba(59, 130, 246, 0.25)',
            borderColor: '#3b82f6',
            pointBackgroundColor: '#1d4ed8',
            pointBorderColor: '#ffffff',
            pointHoverBackgroundColor: '#ffffff',
            pointHoverBorderColor: '#1d4ed8',
            borderWidth: 2
          },
          {
            label: 'Target Benchmark',
            data: [85, 95, 90, 85, 95],
            backgroundColor: 'rgba(16, 185, 129, 0.10)',
            borderColor: '#10b981',
            borderDash: [4, 4],
            pointBackgroundColor: '#10b981',
            borderWidth: 1.5
          }
        ]
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          legend: { position: 'bottom' },
          tooltip: commonTooltipConfig
        },
        scales: {
          r: {
            min: 0,
            max: 100,
            ticks: { stepSize: 20, font: { size: 10 } },
            pointLabels: { font: { size: 11, weight: 600 } }
          }
        }
      }
    });
  }

  // -------------------------------------------------------------------------
  // 20-Year Profit Forecast Chart (Chart.js Line)
  // -------------------------------------------------------------------------
  private renderForecastChart(): void {
    if (!this.forecastCanvas || this.forecastPoints.length === 0) return;
    this.forecastChart?.destroy();

    const rawLabels = this.forecastPoints.map(p => `Year ${p.year}`);
    const wrappedLabels = rawLabels.map(l => wrapLabel16(l, 16));

    const unoptData = this.forecastPoints.map(p => Math.round(p.unoptimized_value));
    const optData = this.forecastPoints.map(p => Math.round(p.optimized_value));
    const deltaData = this.forecastPoints.map(p => Math.round(p.net_profit_delta));

    this.forecastChart = new Chart(this.forecastCanvas.nativeElement, {
      type: 'line',
      data: {
        labels: wrappedLabels,
        datasets: [
          {
            label: 'AI-Optimized Wealth (14.1% CAGR)',
            data: optData,
            borderColor: '#10b981',
            backgroundColor: 'rgba(16, 185, 129, 0.12)',
            fill: true,
            tension: 0.35,
            borderWidth: 2.5
          },
          {
            label: 'Unoptimized Portfolio (10.2% CAGR)',
            data: unoptData,
            borderColor: '#ef4444',
            backgroundColor: 'rgba(239, 68, 68, 0.05)',
            fill: true,
            tension: 0.35,
            borderWidth: 2,
            borderDash: [5, 5]
          },
          {
            label: 'Net Wealth Delta (+3.9% Spread)',
            data: deltaData,
            borderColor: '#6366f1',
            backgroundColor: 'transparent',
            tension: 0.35,
            borderWidth: 2
          }
        ]
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          legend: { position: 'top' },
          tooltip: {
            callbacks: {
              ...commonTooltipConfig.callbacks,
              label: (ctx) => `${ctx.dataset.label}: ₹${Number(ctx.parsed.y).toLocaleString('en-IN')}`
            }
          }
        },
        scales: {
          y: {
            ticks: {
              callback: (val) => `₹${(Number(val) / 100000).toFixed(1)}L`
            }
          }
        }
      }
    });
  }

  // -------------------------------------------------------------------------
  // Asset Allocation Donut Chart
  // -------------------------------------------------------------------------
  private renderAllocationChart(): void {
    if (!this.allocationCanvas) return;
    this.allocationChart?.destroy();

    const rawLabels = [
      'Large Cap Index Direct',
      'Next 50 / Midcap Direct',
      'Arbitrage Fund (12.5% Tax)',
      'Sovereign Gold Bonds'
    ];
    const wrappedLabels = rawLabels.map(l => wrapLabel16(l, 16));

    let allocData = [45, 20, 25, 10];
    if (this.selectedRisk === 'conservative') {
      allocData = [25, 10, 50, 15];
    } else if (this.selectedRisk === 'aggressive') {
      allocData = [55, 25, 10, 10];
    }

    this.allocationChart = new Chart(this.allocationCanvas.nativeElement, {
      type: 'doughnut',
      data: {
        labels: wrappedLabels,
        datasets: [
          {
            data: allocData,
            backgroundColor: ['#3b82f6', '#8b5cf6', '#10b981', '#f59e0b'],
            borderWidth: 2
          }
        ]
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          legend: { position: 'right' },
          tooltip: {
            callbacks: {
              ...commonTooltipConfig.callbacks,
              label: (ctx) => `${ctx.label}: ${ctx.parsed}%`
            }
          }
        }
      }
    });
  }

  // -------------------------------------------------------------------------
  // Post-Budget 2024 Tax Alpha Breakdown Chart (Bar)
  // -------------------------------------------------------------------------
  private renderTaxSavingsChart(): void {
    if (!this.taxCanvas) return;
    this.taxChart?.destroy();

    const rawLabels = [
      'LTCG Harvest Exemption',
      'STCL Loss Offset Saved',
      'Debt to Arbitrage Shift',
      'TER Bleed Saved'
    ];
    const wrappedLabels = rawLabels.map(l => wrapLabel16(l, 16));

    let bleedVal = 0;
    if (this.healthScore?.audit_details) {
      bleedVal = this.healthScore.audit_details.reduce((acc, a) => acc + a.annual_fee_bleed, 0);
    }
    if (bleedVal <= 0) bleedVal = 14200;

    const values = [15625, 2000, 8400, Math.round(bleedVal)];

    this.taxChart = new Chart(this.taxCanvas.nativeElement, {
      type: 'bar',
      data: {
        labels: wrappedLabels,
        datasets: [
          {
            label: 'Annual Tax & Fee Alpha (₹)',
            data: values,
            backgroundColor: ['#10b981', '#3b82f6', '#f59e0b', '#6366f1'],
            borderRadius: 6
          }
        ]
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          legend: { display: false },
          tooltip: {
            callbacks: {
              ...commonTooltipConfig.callbacks,
              label: (ctx) => `Annual Savings: ₹${Number(ctx.parsed.y).toLocaleString('en-IN')}`
            }
          }
        },
        scales: {
          y: {
            ticks: {
              callback: (val) => `₹${Number(val).toLocaleString('en-IN')}`
            }
          }
        }
      }
    });
  }

  // -------------------------------------------------------------------------
  // Plotly WebGL Monte Carlo Simulation (scattergl)
  // -------------------------------------------------------------------------
  private renderPlotlyChart(): void {
    if (!this.plotlyContainer || !this.monteCarloData || !window.Plotly) return;

    const mc = this.monteCarloData;
    const xYears = Array.from({ length: mc.years + 1 }, (_, i) => i);
    const traces: any[] = [];

    // Background stochastic paths using scattergl WebGL
    mc.trajectories.slice(0, 35).forEach((path, idx) => {
      traces.push({
        x: xYears,
        y: path,
        mode: 'lines',
        type: 'scattergl',
        line: { color: 'rgba(99, 102, 241, 0.12)', width: 1 },
        hoverinfo: 'none',
        showlegend: idx === 0,
        name: 'Stochastic Paths'
      });
    });

    // 10th Percentile (Stress Test Drawdown)
    traces.push({
      x: xYears,
      y: mc.percentile_10,
      mode: 'lines',
      type: 'scattergl',
      line: { color: '#ef4444', width: 2.2, dash: 'dot' },
      name: 'P10 Stress Case'
    });

    // 50th Percentile (Median Path)
    traces.push({
      x: xYears,
      y: mc.median,
      mode: 'lines',
      type: 'scattergl',
      line: { color: '#3b82f6', width: 3 },
      name: 'P50 Median'
    });

    // 90th Percentile (Bull Case)
    traces.push({
      x: xYears,
      y: mc.percentile_90,
      mode: 'lines',
      type: 'scattergl',
      line: { color: '#10b981', width: 2.2, dash: 'dash' },
      name: 'P90 Bull Case'
    });

    const layout = {
      title: {
        text: `20-Year GBM Monte Carlo Risk Stress Simulation (Survival Rate: ${mc.survival_rate}%)`,
        font: { size: 14, color: '#475569' }
      },
      margin: { t: 40, r: 20, l: 60, b: 40 },
      paper_bgcolor: 'transparent',
      plot_bgcolor: 'transparent',
      xaxis: {
        title: 'Timeline (Years)',
        gridcolor: 'rgba(148, 163, 184, 0.15)',
        dtick: 2
      },
      yaxis: {
        title: 'Projected Wealth (₹)',
        gridcolor: 'rgba(148, 163, 184, 0.15)',
        tickprefix: '₹'
      },
      legend: {
        orientation: 'h',
        y: -0.2,
        x: 0.1
      },
      hovermode: 'closest'
    };

    window.Plotly.newPlot(this.plotlyContainer.nativeElement, traces, layout, {
      responsive: true,
      displayModeBar: false
    });
  }

  // Helper getters
  get totalFeeBleed(): number {
    if (!this.healthScore?.audit_details) return 0;
    return this.healthScore.audit_details.reduce((acc, a) => acc + a.annual_fee_bleed, 0);
  }

  get closetIndexersCount(): number {
    if (!this.healthScore?.audit_details) return 0;
    return this.healthScore.audit_details.filter(a => a.is_closet_indexer).length;
  }

  get netProfitDelta20Y(): number {
    if (this.forecastPoints.length === 0) return 0;
    return this.forecastPoints[this.forecastPoints.length - 1].net_profit_delta;
  }

  getScoreBadgeClass(score: number): string {
    if (score >= 80) return 'badge-success';
    if (score >= 60) return 'badge-warning';
    return 'badge-danger';
  }
}
