import { Component, inject, signal, OnInit } from '@angular/core';
import { MatTableModule } from '@angular/material/table';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatDialogModule, MatDialog } from '@angular/material/dialog';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { DecimalPipe } from '@angular/common';
import { RouterLink } from '@angular/router';
import { ItemService } from '../../../core/services/item.service';
import { Item } from '../../../core/models/item.model';
import { AdminItemDialogComponent } from './admin-item-dialog.component';

@Component({
  selector: 'app-admin-items',
  standalone: true,
  imports: [
    RouterLink,
    MatTableModule,
    MatButtonModule,
    MatIconModule,
    MatFormFieldModule,
    MatInputModule,
    MatDialogModule,
    MatProgressSpinnerModule,
    MatSnackBarModule,
    DecimalPipe,
  ],
  template: `
    <div class="admin-header">
      <h1>Item management</h1>
      <button mat-raised-button color="primary" (click)="openDialog()">
        <mat-icon>add</mat-icon>
        New item
      </button>
    </div>

    <a mat-button routerLink="/admin/orders">
      <mat-icon>receipt_long</mat-icon> Orders
    </a>

    @if (loading()) {
      <div class="spinner-wrapper"><mat-spinner /></div>
    } @else {
      <table mat-table [dataSource]="items()" class="items-table">

        <ng-container matColumnDef="name">
          <th mat-header-cell *matHeaderCellDef>Name</th>
          <td mat-cell *matCellDef="let row">{{ row.name }}</td>
        </ng-container>

        <ng-container matColumnDef="price">
          <th mat-header-cell *matHeaderCellDef>Price (USDT)</th>
          <td mat-cell *matCellDef="let row">{{ row.price | number:'1.2-2' }}</td>
        </ng-container>

        <ng-container matColumnDef="stock">
          <th mat-header-cell *matHeaderCellDef>Stock</th>
          <td mat-cell *matCellDef="let row" [class.low-stock]="row.stock < 5">
            {{ row.stock }}
          </td>
        </ng-container>

        <ng-container matColumnDef="actions">
          <th mat-header-cell *matHeaderCellDef></th>
          <td mat-cell *matCellDef="let row">
            <button mat-icon-button (click)="openDialog(row)">
              <mat-icon>edit</mat-icon>
            </button>
            <button mat-icon-button color="warn" (click)="delete(row)">
              <mat-icon>delete</mat-icon>
            </button>
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
    .items-table { width: 100%; margin-top: 16px; }
    .low-stock { color: #f44336; font-weight: bold; }
  `],
})
export class AdminItemsComponent implements OnInit {
  private itemSvc = inject(ItemService);
  private dialog = inject(MatDialog);
  private snackBar = inject(MatSnackBar);

  items = signal<Item[]>([]);
  loading = signal(true);
  columns = ['name', 'price', 'stock', 'actions'];

  ngOnInit(): void {
    this.load();
  }

  load(): void {
    this.loading.set(true);
    this.itemSvc.list({ pageSize: 100 }).subscribe({
      next: (res) => { this.items.set(res.items); this.loading.set(false); },
      error: () => { this.loading.set(false); },
    });
  }

  openDialog(item?: Item): void {
    const ref = this.dialog.open(AdminItemDialogComponent, {
      width: '500px',
      data: item ?? null,
    });
    ref.afterClosed().subscribe((saved) => {
      if (saved) this.load();
    });
  }

  delete(item: Item): void {
    if (!confirm(`Delete "${item.name}"?`)) return;
    this.itemSvc.delete(item.id).subscribe({
      next: () => {
        this.snackBar.open('Item deleted', 'OK', { duration: 2000 });
        this.load();
      },
      error: () => this.snackBar.open('Failed to delete', 'Close', { duration: 3000 }),
    });
  }
}
