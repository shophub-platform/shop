import { Component, inject, signal, OnInit } from '@angular/core';
import { RouterLink } from '@angular/router';
import { MatTableModule } from '@angular/material/table';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatSelectModule } from '@angular/material/select';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { DatePipe, DecimalPipe, SlicePipe } from '@angular/common';
import { OrderService } from '../../../core/services/order.service';
import { Order, OrderStatus, UpdateOrderStatusRequest } from '../../../core/models/order.model';

const STATUS_LABELS: Record<OrderStatus, string> = {
  PENDING_PAYMENT: 'Pending payment',
  PAID: 'Paid',
  PROCESSING: 'Processing',
  SHIPPED: 'Shipped',
  DELIVERED: 'Delivered',
  CANCELLED: 'Cancelled',
};

@Component({
  selector: 'app-admin-orders',
  standalone: true,
  imports: [
    RouterLink,
    MatTableModule,
    MatButtonModule,
    MatIconModule,
    MatSelectModule,
    MatFormFieldModule,
    MatProgressSpinnerModule,
    MatSnackBarModule,
    DatePipe,
    DecimalPipe,
    SlicePipe,
  ],
  template: `
    <div class="admin-header">
      <h1>Orders</h1>
      <a mat-button routerLink="/admin/items">
        <mat-icon>inventory_2</mat-icon> Items
      </a>
    </div>

    @if (loading()) {
      <div class="spinner-wrapper"><mat-spinner /></div>
    } @else if (orders().length === 0) {
      <div class="empty-state">
        <mat-icon>receipt_long</mat-icon>
        <p>No orders found.</p>
      </div>
    } @else {
      <table mat-table [dataSource]="orders()" class="orders-table">

        <ng-container matColumnDef="id">
          <th mat-header-cell *matHeaderCellDef>#</th>
          <td mat-cell *matCellDef="let row" class="mono">{{ row.id | slice:0:8 }}…</td>
        </ng-container>

        <ng-container matColumnDef="date">
          <th mat-header-cell *matHeaderCellDef>Date</th>
          <td mat-cell *matCellDef="let row">{{ row.createdAt | date:'dd.MM.yyyy HH:mm' }}</td>
        </ng-container>

        <ng-container matColumnDef="userId">
          <th mat-header-cell *matHeaderCellDef>User</th>
          <td mat-cell *matCellDef="let row" class="mono">{{ row.userId | slice:0:8 }}…</td>
        </ng-container>

        <ng-container matColumnDef="total">
          <th mat-header-cell *matHeaderCellDef>Total</th>
          <td mat-cell *matCellDef="let row">{{ row.total | number:'1.2-2' }} mUSDT</td>
        </ng-container>

        <ng-container matColumnDef="status">
          <th mat-header-cell *matHeaderCellDef>Status</th>
          <td mat-cell *matCellDef="let row">
            <span class="status-chip">{{ statusLabel(row.status) }}</span>
          </td>
        </ng-container>

        <ng-container matColumnDef="actions">
          <th mat-header-cell *matHeaderCellDef>Action</th>
          <td mat-cell *matCellDef="let row">
            @if (row.status === 'PAID') {
              <button mat-button color="primary" (click)="setStatus(row, 'PROCESSING')">
                Process
              </button>
              <button mat-button color="warn" (click)="setStatus(row, 'CANCELLED')">
                Cancel
              </button>
            }
            @if (row.status === 'PROCESSING') {
              <button mat-button color="primary" (click)="setStatus(row, 'SHIPPED')">
                Mark shipped
              </button>
            }
            @if (row.status === 'SHIPPED') {
              <button mat-button color="primary" (click)="setStatus(row, 'DELIVERED')">
                Mark delivered
              </button>
            }
          </td>
        </ng-container>

        <tr mat-header-row *matHeaderRowDef="columns"></tr>
        <tr mat-row *matRowDef="let row; columns: columns;"></tr>
      </table>
    }
  `,
  styles: [`
    .admin-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
    h1 { margin: 0; }
    .spinner-wrapper { display: flex; justify-content: center; padding: 64px; }
    .orders-table { width: 100%; }
    .empty-state { text-align: center; padding: 64px; color: #999; }
    .empty-state mat-icon { font-size: 64px; width: 64px; height: 64px; display: block; margin: 0 auto 16px; }
    .status-chip { padding: 4px 10px; background: #e3f2fd; border-radius: 12px; font-size: 0.8rem; }
    .mono { font-family: monospace; font-size: 0.85rem; }
  `],
})
export class AdminOrdersComponent implements OnInit {
  private orderSvc = inject(OrderService);
  private snackBar = inject(MatSnackBar);

  orders = signal<Order[]>([]);
  loading = signal(true);
  columns = ['id', 'date', 'userId', 'total', 'status', 'actions'];

  ngOnInit(): void {
    this.load();
  }

  load(): void {
    this.loading.set(true);
    this.orderSvc.list({ pageSize: 100 }).subscribe({
      next: (res) => { this.orders.set(res.orders); this.loading.set(false); },
      error: () => { this.loading.set(false); },
    });
  }

  statusLabel(status: OrderStatus): string {
    return STATUS_LABELS[status] ?? status;
  }

  setStatus(order: Order, status: UpdateOrderStatusRequest['status']): void {
    this.orderSvc.updateStatus(order.id, { status }).subscribe({
      next: (updated) => {
        this.orders.update((list) =>
          list.map((o) => (o.id === updated.id ? updated : o)),
        );
        this.snackBar.open('Status updated', 'OK', { duration: 2000 });
      },
      error: () => this.snackBar.open('Failed to update status', 'Close', { duration: 3000 }),
    });
  }
}
