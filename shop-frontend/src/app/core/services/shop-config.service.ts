import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { firstValueFrom } from 'rxjs';
import { environment } from '../../../environments/environment';

export interface ShopConfig {
  walletAddress: string;
  contractAddress: string;
}

interface ApiResponse<T> {
  success: boolean;
  data: T;
}

@Injectable({ providedIn: 'root' })
export class ShopConfigService {
  private http = inject(HttpClient);
  private cached: ShopConfig | null = null;

  async get(): Promise<ShopConfig> {
    if (this.cached) return this.cached;
    const resp = await firstValueFrom(
      this.http.get<ApiResponse<ShopConfig>>(`${environment.apiUrl}/config`),
    );
    this.cached = resp.data;
    return this.cached;
  }
}
