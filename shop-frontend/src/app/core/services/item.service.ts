import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { map } from 'rxjs/operators';
import { environment } from '../../../environments/environment';
import {
  Item,
  ItemsResponse,
  CreateItemRequest,
  UpdateItemRequest,
  ItemFilters,
} from '../models/item.model';

interface ApiResponse<T> {
  success: boolean;
  data: T;
}

@Injectable({ providedIn: 'root' })
export class ItemService {
  private http = inject(HttpClient);
  private base = `${environment.apiUrl}/items`;

  list(filters: ItemFilters = {}): Observable<ItemsResponse> {
    let params = new HttpParams();
    if (filters.page) params = params.set('page', filters.page);
    if (filters.pageSize) params = params.set('pageSize', filters.pageSize);
    if (filters.search) params = params.set('search', filters.search);
    if (filters.minPrice != null) params = params.set('minPrice', filters.minPrice);
    if (filters.maxPrice != null) params = params.set('maxPrice', filters.maxPrice);
    if (filters.inStock != null) params = params.set('inStock', filters.inStock);
    return this.http
      .get<ApiResponse<ItemsResponse>>(this.base, { params })
      .pipe(map((r) => r.data));
  }

  getById(id: string): Observable<Item> {
    return this.http
      .get<ApiResponse<Item>>(`${this.base}/${id}`)
      .pipe(map((r) => r.data));
  }

  create(body: CreateItemRequest): Observable<Item> {
    return this.http
      .post<ApiResponse<Item>>(this.base, body)
      .pipe(map((r) => r.data));
  }

  update(id: string, body: UpdateItemRequest): Observable<Item> {
    return this.http
      .patch<ApiResponse<Item>>(`${this.base}/${id}`, body)
      .pipe(map((r) => r.data));
  }

  delete(id: string): Observable<void> {
    return this.http.delete<void>(`${this.base}/${id}`);
  }
}
