import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { map } from 'rxjs/operators';
import { environment } from '../../../environments/environment';
import {
  Order,
  OrdersResponse,
  ConfirmPaymentRequest,
  UpdateOrderStatusRequest,
  OrderFilters,
} from '../models/order.model';

interface ApiResponse<T> {
  success: boolean;
  data: T;
}

@Injectable({ providedIn: 'root' })
export class OrderService {
  private http = inject(HttpClient);
  private base = `${environment.apiUrl}/orders`;

  list(filters: OrderFilters = {}): Observable<OrdersResponse> {
    let params = new HttpParams();
    if (filters.page) params = params.set('page', filters.page);
    if (filters.pageSize) params = params.set('pageSize', filters.pageSize);
    if (filters.status) params = params.set('status', filters.status);
    return this.http
      .get<ApiResponse<OrdersResponse>>(this.base, { params })
      .pipe(map((r) => r.data));
  }

  getById(id: string): Observable<Order> {
    return this.http
      .get<ApiResponse<Order>>(`${this.base}/${id}`)
      .pipe(map((r) => r.data));
  }

  create(): Observable<Order> {
    return this.http
      .post<ApiResponse<Order>>(this.base, {})
      .pipe(map((r) => r.data));
  }

  confirmPayment(id: string, body: ConfirmPaymentRequest): Observable<Order> {
    return this.http
      .post<ApiResponse<Order>>(`${this.base}/${id}/confirm`, body)
      .pipe(map((r) => r.data));
  }

  updateStatus(id: string, body: UpdateOrderStatusRequest): Observable<Order> {
    return this.http
      .patch<ApiResponse<Order>>(`${this.base}/${id}/status`, body)
      .pipe(map((r) => r.data));
  }
}
