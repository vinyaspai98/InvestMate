import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../environments/environment';

export interface GmailSyncResponse {
    message: string;
    synced: number;
    total: number;
}

@Injectable({
    providedIn: 'root'
})
export class GmailService {
    private apiUrl = environment.apiBaseUrl || 'http://localhost:8080/api/v1';

    constructor(private http: HttpClient) { }

    syncGmail(): Observable<GmailSyncResponse> {
        return this.http.post<GmailSyncResponse>(`${this.apiUrl}/gmail/sync`, {});
    }
}
