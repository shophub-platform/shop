import { Component, inject, signal, OnInit } from '@angular/core';
import { RouterLink } from '@angular/router';
import { MatTableModule } from '@angular/material/table';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatChipsModule } from '@angular/material/chips';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { DecimalPipe, DatePipe, SlicePipe } from '@angular/common';
import { OrderService } from '../../core/services/order.service';
import { Order, OrderStatus } from '../../core/models/order.model';

const STATUS_LABELS: Record<OrderStatus, string> = {
  PENDING_PAYMENT: 'Pending payment',
  PAID: 'Paid',
  PROCESSING: 'Processing',
  SHIPPED: 'Shipped',
  DELIVERED: 'Delivered',
  CANCELLED: 'Cancelled',
};

const STATUS_COLORS: Record<OrderStatus, string> = {
  PENDING_PAYMENT: '#ff9800',
  PAID: '#4caf50',
  PROCESSING: '#2196f3',
  SHIPPED: '#9c27b0',
  DELIVERED: '#4caf50',
  CANCELLED: '#f44336',
};

@Component({
  selector: 'app-orders',
  standalone: true,
  imports: [
    RouterLink,
    MatTableModule,
    MatButtonModule,
    MatIconModule,
    MatChipsModule,
    MatProgressSpinnerModule,
    MatSnackBarModule,
    DecimalPipe,
    DatePipe,
    SlicePipe,
  ],
  template: `
    <h1>My orders</h1>

    @if (loading()) {
      <div class="spinner-wrapper"><mat-spinner /></div>
    } @else if (orders().length === 0) {
      <div class="empty-state">
        <mat-icon>receipt_long</mat-icon>
        <p>You have no orders.</p>
        <a mat-raised-button color="primary" routerLink="/">Start shopping</a>
      </div>
    } @else {
      <table mat-table [dataSource]="orders()" class="orders-table">

        <ng-container matColumnDef="id">
          <th mat-header-cell *matHeaderCellDef>#</th>
          <td mat-cell *matCellDef="let row">{{ row.id | slice:0:8 }}…</td>
        </ng-container>

        <ng-container matColumnDef="date">
          <th mat-header-cell *matHeaderCellDef>Date</th>
          <td mat-cell *matCellDef="let row">{{ row.createdAt | date:'dd.MM.yyyy HH:mm' }}</td>
        </ng-container>

        <ng-container matColumnDef="items">
          <th mat-header-cell *matHeaderCellDef>Items</th>
          <td mat-cell *matCellDef="let row">
            {{ row.items.map(i => i.itemName).join(', ') }}
          </td>
        </ng-container>

        <ng-container matColumnDef="total">
          <th mat-header-cell *matHeaderCellDef>Total</th>
          <td mat-cell *matCellDef="let row">{{ row.total | number:'1.2-2' }} USDT</td>
        </ng-container>

        <ng-container matColumnDef="status">
          <th mat-header-cell *matHeaderCellDef>Status</th>
          <td mat-cell *matCellDef="let row">
            <span class="status-chip" [style.background]="statusColor(row.status)">
              {{ statusLabel(row.status) }}
            </span>
          </td>
        </ng-container>

        <ng-container matColumnDef="actions">
          <th mat-header-cell *matHeaderCellDef></th>
          <td mat-cell *matCellDef="let row">
            @if (row.status === 'PENDING_PAYMENT') {
              <a mat-button color="primary" [routerLink]="['/checkout', row.id]">
                Pay
              </a>
            }
          </td>
        </ng-container>

        <tr mat-header-row *matHeaderRowDef="columns"></tr>
        <tr mat-row *matRowDef="let row; columns: columns;"></tr>
      </table>
    }
  `,
  styles: [`
    h1 { margin-bottom: 24px; }
    .spinner-wrapper { display: flex; justify-content: center; padding: 64px; }
    .empty-state { text-align: center; padding: 64px; color: #999; }
    .empty-state mat-icon { font-size: 64px; width: 64px; height: 64px; display: block; margin: 0 auto 16px; }
    .orders-table { width: 100%; }
    .status-chip {
      padding: 4px 10px; border-radius: 12px; color: white;
      font-size: 0.8rem; font-weight: 500;
    }
  `],
})
export class OrdersComponent implements OnInit {
  private orderSvc = inject(OrderService);
  private snackBar = inject(MatSnackBar);

  orders = signal<Order[]>([]);
  loading = signal(true);

  columns = ['id', 'date', 'items', 'total', 'status', 'actions'];

  ngOnInit(): void {
    this.orderSvc.list().subscribe({
      next: (res) => { this.orders.set(res.orders); this.loading.set(false); },
      error: () => {
        this.loading.set(false);
        this.snackBar.open('Failed to load orders', 'Close', { duration: 3000 });
      },
    });
  }

  statusLabel(status: OrderStatus): string {
    return STATUS_LABELS[status] ?? status;
  }

  statusColor(status: OrderStatus): string {
    return STATUS_COLORS[status] ?? '#999';
  }
}
