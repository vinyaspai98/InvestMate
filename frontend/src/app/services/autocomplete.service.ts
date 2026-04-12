import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, map } from 'rxjs';
import { environment } from '../../environments/environment';

@Injectable({
    providedIn: 'root'
})
export class AutocompleteService {
    private apiUrl = `${environment.apiBaseUrl}/investments`;

    constructor(private http: HttpClient) { }

    searchSymbols(keywords: string): Observable<any[]> {
        return this.http.get<any>(`${this.apiUrl}/search`, {
            params: { keywords }
        }).pipe(
            map(res => res.bestMatches || [])
        );
    }

    getPriceData(symbol: string, date?: string): Observable<any> {
        const params: any = {};
        if (date) {
            params.date = date;
        }
        return this.http.get<any>(`${this.apiUrl}/price/${symbol}`, { params });
    }
}
