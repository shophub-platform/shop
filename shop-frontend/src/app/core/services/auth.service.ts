import { Injectable, signal, computed, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, tap } from 'rxjs';
import { environment } from '../../../environments/environment';
import { JwtPayload } from '../models/auth.model';

export interface User {
  id: string;
  email: string;
  role: 'USER' | 'ADMIN';
  createdAt: string;
  updatedAt: string;
}

interface LoginResponse {
  token: string;
  user: User;
}

@Injectable({ providedIn: 'root' })
export class AuthService {
  private http = inject(HttpClient);
  private readonly TOKEN_KEY = 'shop_token';
  private readonly api = `${environment.apiUrl}/auth`;

  private _token = signal<string | null>(localStorage.getItem(this.TOKEN_KEY));

  readonly token = this._token.asReadonly();

  readonly isLoggedIn = computed(() => !!this._token());

  readonly currentUser = computed<JwtPayload | null>(() => {
    const token = this._token();
    if (!token) return null;
    try {
      const payload = token.split('.')[1];
      return JSON.parse(atob(payload)) as JwtPayload;
    } catch {
      return null;
    }
  });

  readonly isAdmin = computed(() => this.currentUser()?.role === 'ADMIN');

  login(email: string, password: string): Observable<LoginResponse> {
    return this.http.post<LoginResponse>(`${this.api}/login`, { email, password }).pipe(
      tap((res) => this.setToken(res.token)),
    );
  }

  register(email: string, password: string): Observable<User> {
    return this.http.post<User>(`${this.api}/register`, { email, password });
  }

  setToken(token: string): void {
    localStorage.setItem(this.TOKEN_KEY, token);
    this._token.set(token);
  }

  clearToken(): void {
    localStorage.removeItem(this.TOKEN_KEY);
    this._token.set(null);
  }
}
