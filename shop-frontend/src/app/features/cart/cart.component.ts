import { Component, inject, signal, computed, OnInit } from '@angular/core';
import { Router, RouterLink } from '@angular/router';
import { MatTableModule } from '@angular/material/table';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { MatDividerModule } from '@angular/material/divider';
import { DecimalPipe } from '@angular/common';
import { CartService } from '../../core/services/cart.service';
import { OrderService } from '../../core/services/order.service';
import { Web3Service } from '../../core/services/web3.service';
import { Cart, CartItem } from '../../core/models/cart.model';

@Component({
  selector: 'app-cart',
  standalone: true,
  imports: [
    RouterLink,
    MatTableModule,
    MatButtonModule,
    MatIconModule,
    MatProgressSpinnerModule,
    MatSnackBarModule,
    MatDividerModule,
    DecimalPipe,
  ],
  template: `
    <h1>Cart</h1>

    @if (loading()) {
      <div class="spinner-wrapper"><mat-spinner /></div>
    } @else if (!cart() || !cart()!.items?.length) {
      <div class="empty-state">
        <mat-icon>shopping_cart</mat-icon>
        <p>Your cart is empty.</p>
        <a mat-raised-button color="primary" routerLink="/">Browse items</a>
      </div>
    } @else {
      <table mat-table [dataSource]="cart()!.items" class="cart-table">

        <ng-container matColumnDef="item">
          <th mat-header-cell *matHeaderCellDef>Item</th>
          <td mat-cell *matCellDef="let row">
            <a [routerLink]="['/items', row.item.id]">{{ row.item.name }}</a>
          </td>
        </ng-container>

        <ng-container matColumnDef="price">
          <th mat-header-cell *matHeaderCellDef>Price</th>
          <td mat-cell *matCellDef="let row">{{ row.item.price | number:'1.2-2' }} mUSDT</td>
        </ng-container>

        <ng-container matColumnDef="quantity">
          <th mat-header-cell *matHeaderCellDef>Quantity</th>
          <td mat-cell *matCellDef="let row">
            <div class="qty-controls">
              <button mat-icon-button (click)="changeQty(row, row.quantity - 1)"
                [disabled]="row.quantity <= 1">
                <mat-icon>remove</mat-icon>
              </button>
              <span>{{ row.quantity }}</span>
              <button mat-icon-button (click)="changeQty(row, row.quantity + 1)"
                [disabled]="row.quantity >= row.item.stock">
                <mat-icon>add</mat-icon>
              </button>
            </div>
          </td>
        </ng-container>

        <ng-container matColumnDef="subtotal">
          <th mat-header-cell *matHeaderCellDef>Subtotal</th>
          <td mat-cell *matCellDef="let row">{{ row.item.price * row.quantity | number:'1.2-2' }} mUSDT</td>
        </ng-container>

        <ng-container matColumnDef="actions">
          <th mat-header-cell *matHeaderCellDef></th>
          <td mat-cell *matCellDef="let row">
            <button mat-icon-button color="warn" (click)="removeItem(row)">
              <mat-icon>delete</mat-icon>
            </button>
          </td>
        </ng-container>

        <tr mat-header-row *matHeaderRowDef="columns"></tr>
        <tr mat-row *matRowDef="let row; columns: columns;"></tr>
      </table>

      <mat-divider />

      <div class="cart-footer">
        <div class="total">
          <strong>Total:</strong>
          <span class="total-amount">{{ total() | number:'1.2-2' }} mUSDT</span>
        </div>

        <div class="footer-actions">
          <button mat-button color="warn" (click)="clearCart()">
            <mat-icon>delete_sweep</mat-icon>
            Clear cart
          </button>
          <button mat-raised-button color="primary" (click)="checkout()" [disabled]="ordering()">
            <mat-icon>payment</mat-icon>
            Place order
          </button>
        </div>
      </div>
    }
  `,
  styles: [`
    h1 { margin-bottom: 24px; }
    .spinner-wrapper { display: flex; justify-content: center; padding: 64px; }
    .empty-state { text-align: center; padding: 64px; color: #999; }
    .empty-state mat-icon { font-size: 64px; width: 64px; height: 64px; display: block; margin: 0 auto 16px; }
    .cart-table { width: 100%; margin-bottom: 16px; }
    .qty-controls { display: flex; align-items: center; gap: 8px; }
    .cart-footer {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 16px 0;
      flex-wrap: wrap;
      gap: 16px;
    }
    .total { display: flex; align-items: center; gap: 12px; font-size: 1.1rem; }
    .total-amount { font-size: 1.4rem; font-weight: bold; color: #1565c0; }
    .footer-actions { display: flex; gap: 12px; }
  `],
})
export class CartComponent implements OnInit {
  private cartSvc = inject(CartService);
  private orderSvc = inject(OrderService);
  private web3Svc = inject(Web3Service);
  private router = inject(Router);
  private snackBar = inject(MatSnackBar);

  cart = signal<Cart | null>(null);
  loading = signal(true);
  ordering = signal(false);

  total = computed(() =>
    this.cart()?.items.reduce((sum, row) => sum + row.item.price * row.quantity, 0) ?? 0
  );

  columns = ['item', 'price', 'quantity', 'subtotal', 'actions'];

  ngOnInit(): void {
    this.loadCart();
  }

  loadCart(): void {
    this.loading.set(true);
    this.cartSvc.get().subscribe({
      next: (cart) => { this.cart.set(cart); this.loading.set(false); },
      error: () => { this.loading.set(false); },
    });
  }

  changeQty(row: CartItem, newQty: number): void {
    if (newQty < 1) return;
    this.cartSvc.updateItem(row.item.id, { quantity: newQty }).subscribe({
      next: (cart) => this.cart.set(cart),
      error: () => this.snackBar.open('Failed to update quantity', 'Close', { duration: 3000 }),
    });
  }

  removeItem(row: CartItem): void {
    this.cartSvc.removeItem(row.item.id).subscribe({
      next: (cart) => this.cart.set(cart),
      error: () => this.snackBar.open('Failed to remove item', 'Close', { duration: 3000 }),
    });
  }

  clearCart(): void {
    this.cartSvc.clear().subscribe({
      next: () => this.loadCart(),
      error: () => this.snackBar.open('Error', 'Close', { duration: 3000 }),
    });
  }

  checkout(): void {
    this.ordering.set(true);
    this.getWalletAddress().then((walletFrom) => {
      this.orderSvc.create({ walletFrom }).subscribe({
        next: (order) => {
          this.ordering.set(false);
          this.router.navigate(['/checkout', order.id]);
        },
        error: () => {
          this.ordering.set(false);
          this.snackBar.open('Failed to create order', 'Close', { duration: 3000 });
        },
      });
    });
  }

  private async getWalletAddress(): Promise<string | null> {
    if (!this.web3Svc.isAvailable()) return null;
    try {
      return await this.web3Svc.connect();
    } catch {
      return null;
    }
  }
}
