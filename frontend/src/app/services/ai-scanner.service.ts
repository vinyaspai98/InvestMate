import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../environments/environment';
import {
  PortfolioHealthScore,
  RecommendationAction,
  ProfitForecastPoint,
  MonteCarloResponse
} from '../models/ai-scanner.model';

@Injectable({
  providedIn: 'root'
})
export class AiScannerService {
  private http = inject(HttpClient);
  private apiUrl = `${environment.apiBaseUrl}/ai`;

  scanCAS(payload: { file_bytes?: string; password?: string; text?: string } = {}): Observable<PortfolioHealthScore> {
    return this.http.post<PortfolioHealthScore>(`${this.apiUrl}/scan-cas`, payload);
  }

  getHealthScore(): Observable<PortfolioHealthScore> {
    return this.http.get<PortfolioHealthScore>(`${this.apiUrl}/health-score`);
  }

  getRecommendations(targetRisk: string = 'moderate'): Observable<RecommendationAction[]> {
    return this.http.post<RecommendationAction[]>(`${this.apiUrl}/recommendations`, { target_risk: targetRisk });
  }

  forecastProfit(initialBase: number = 1000000, years: number = 20): Observable<ProfitForecastPoint[]> {
    return this.http.post<ProfitForecastPoint[]>(`${this.apiUrl}/forecast-profit`, {
      initial_base: initialBase,
      years: years
    });
  }

  getMonteCarlo(simulations: number = 1000, years: number = 20, initialBase: number = 1000000): Observable<MonteCarloResponse> {
    const params = new HttpParams()
      .set('simulations', simulations.toString())
      .set('years', years.toString())
      .set('initial_base', initialBase.toString());

    return this.http.get<MonteCarloResponse>(`${this.apiUrl}/monte-carlo`, { params });
  }
}
