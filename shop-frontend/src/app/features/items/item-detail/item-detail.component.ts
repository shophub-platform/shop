import { Component, inject, signal, OnInit } from '@angular/core';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { FormControl, ReactiveFormsModule, Validators } from '@angular/forms';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { DecimalPipe } from '@angular/common';
import { ItemService } from '../../../core/services/item.service';
import { CartService } from '../../../core/services/cart.service';
import { AuthService } from '../../../core/services/auth.service';
import { Item } from '../../../core/models/item.model';

@Component({
  selector: 'app-item-detail',
  standalone: true,
  imports: [
    RouterLink,
    ReactiveFormsModule,
    MatCardModule,
    MatButtonModule,
    MatIconModule,
    MatFormFieldModule,
    MatInputModule,
    MatProgressSpinnerModule,
    MatSnackBarModule,
    DecimalPipe,
  ],
  template: `
    @if (loading()) {
      <div class="spinner-wrapper"><mat-spinner /></div>
    } @else if (item()) {
      <div class="detail-layout">
        <div class="image-section">
          @if (item()!.imageUrl) {
            <img [src]="item()!.imageUrl" [alt]="item()!.name" class="detail-image" />
          } @else {
            <div class="no-image">
              <mat-icon>image_not_supported</mat-icon>
            </div>
          }
        </div>

        <div class="info-section">
          <a mat-button routerLink="/" class="back-btn">
            <mat-icon>arrow_back</mat-icon> Back
          </a>

          <h1>{{ item()!.name }}</h1>

          @if (item()!.description) {
            <p class="description">{{ item()!.description }}</p>
          }

          <p class="price">{{ item()!.price | number:'1.2-2' }} USDT</p>

          <p class="stock" [class.out-of-stock]="item()!.stock === 0">
            {{ item()!.stock > 0 ? 'In stock: ' + item()!.stock : 'Out of stock' }}
          </p>

          @if (auth.isLoggedIn() && !auth.isAdmin() && item()!.stock > 0) {
            <div class="quantity-row">
              <mat-form-field appearance="outline" class="qty-field">
                <mat-label>Quantity</mat-label>
                <input matInput type="number" [formControl]="qtyCtrl" min="1" [max]="item()!.stock" />
              </mat-form-field>

              <button mat-raised-button color="primary" (click)="addToCart()" [disabled]="qtyCtrl.invalid || adding()">
                <mat-icon>add_shopping_cart</mat-icon>
                Add to cart
              </button>
            </div>
          } @else if (!auth.isLoggedIn()) {
            <p class="login-hint">
              <a routerLink="/login">Sign in</a> to add items to your cart.
            </p>
          }
        </div>
      </div>
    } @else {
      <div class="empty-state">
        <mat-icon>error_outline</mat-icon>
        <p>Item not found.</p>
        <a mat-button routerLink="/">Back to list</a>
      </div>
    }
  `,
  styles: [`
    .spinner-wrapper { display: flex; justify-content: center; padding: 64px; }
    .detail-layout {
      display: grid;
      grid-template-columns: 1fr 1fr;
      gap: 32px;
      max-width: 900px;
      margin: 0 auto;
    }
    @media (max-width: 600px) {
      .detail-layout { grid-template-columns: 1fr; }
    }
    .detail-image { width: 100%; border-radius: 8px; object-fit: cover; max-height: 400px; }
    .no-image {
      width: 100%; height: 300px; background: #f0f0f0; border-radius: 8px;
      display: flex; align-items: center; justify-content: center; color: #aaa;
    }
    .no-image mat-icon { font-size: 64px; width: 64px; height: 64px; }
    .info-section { display: flex; flex-direction: column; gap: 12px; }
    .back-btn { margin-bottom: 8px; padding-left: 0; }
    h1 { margin: 0; font-size: 1.6rem; }
    .description { color: #555; line-height: 1.6; }
    .price { font-size: 1.8rem; font-weight: bold; color: #1565c0; margin: 0; }
    .stock { color: #4caf50; margin: 0; }
    .out-of-stock { color: #f44336; }
    .quantity-row { display: flex; gap: 16px; align-items: center; flex-wrap: wrap; }
    .qty-field { width: 120px; }
    .login-hint { color: #666; }
    .login-hint a { color: #1565c0; }
    .empty-state { text-align: center; padding: 64px; color: #999; }
    .empty-state mat-icon { font-size: 64px; width: 64px; height: 64px; }
  `],
})
export class ItemDetailComponent implements OnInit {
  private route = inject(ActivatedRoute);
  private router = inject(Router);
  private itemSvc = inject(ItemService);
  private cartSvc = inject(CartService);
  private snackBar = inject(MatSnackBar);
  auth = inject(AuthService);

  item = signal<Item | null>(null);
  loading = signal(true);
  adding = signal(false);

  qtyCtrl = new FormControl(1, [Validators.required, Validators.min(1)]);

  ngOnInit(): void {
    const id = this.route.snapshot.paramMap.get('id')!;
    this.itemSvc.getById(id).subscribe({
      next: (item) => { this.item.set(item); this.loading.set(false); },
      error: () => { this.loading.set(false); },
    });
  }

  addToCart(): void {
    const item = this.item();
    if (!item) return;
    this.adding.set(true);
    this.cartSvc.addItem({ itemId: item.id, quantity: this.qtyCtrl.value! }).subscribe({
      next: () => {
        this.adding.set(false);
        this.snackBar.open(`"${item.name}" added to cart`, 'Go to cart', { duration: 3000 })
          .onAction().subscribe(() => this.router.navigate(['/cart']));
      },
      error: () => {
        this.adding.set(false);
        this.snackBar.open('Failed to add to cart', 'Close', { duration: 3000 });
      },
    });
  }
}
