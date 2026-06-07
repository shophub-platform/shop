import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { map } from 'rxjs/operators';
import { environment } from '../../../environments/environment';
import { Cart, AddToCartRequest, UpdateCartItemRequest } from '../models/cart.model';

interface ApiResponse<T> {
  success: boolean;
  data: T;
}

@Injectable({ providedIn: 'root' })
export class CartService {
  private http = inject(HttpClient);
  private base = `${environment.apiUrl}/cart`;

  get(): Observable<Cart> {
    return this.http.get<ApiResponse<Cart>>(this.base).pipe(map((r) => r.data));
  }

  addItem(body: AddToCartRequest): Observable<Cart> {
    return this.http
      .post<ApiResponse<Cart>>(`${this.base}/items`, body)
      .pipe(map((r) => r.data));
  }

  updateItem(itemId: string, body: UpdateCartItemRequest): Observable<Cart> {
    return this.http
      .put<ApiResponse<Cart>>(`${this.base}/items/${itemId}`, body)
      .pipe(map((r) => r.data));
  }

  removeItem(itemId: string): Observable<Cart> {
    return this.http
      .delete<ApiResponse<Cart>>(`${this.base}/items/${itemId}`)
      .pipe(map((r) => r.data));
  }

  clear(): Observable<void> {
    return this.http.delete<void>(this.base);
  }
}
