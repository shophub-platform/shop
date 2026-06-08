import { Component, inject, signal, OnInit } from '@angular/core';
import { RouterLink } from '@angular/router';
import { FormControl, ReactiveFormsModule } from '@angular/forms';
import { debounceTime, distinctUntilChanged } from 'rxjs/operators';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatCheckboxModule } from '@angular/material/checkbox';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatPaginatorModule, PageEvent } from '@angular/material/paginator';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { DecimalPipe } from '@angular/common';
import { ItemService } from '../../../core/services/item.service';
import { CartService } from '../../../core/services/cart.service';
import { AuthService } from '../../../core/services/auth.service';
import { Item, PaginatedMeta } from '../../../core/models/item.model';

@Component({
  selector: 'app-items-list',
  standalone: true,
  imports: [
    RouterLink,
    ReactiveFormsModule,
    MatCardModule,
    MatButtonModule,
    MatIconModule,
    MatFormFieldModule,
    MatInputModule,
    MatCheckboxModule,
    MatProgressSpinnerModule,
    MatPaginatorModule,
    MatSnackBarModule,
    DecimalPipe,
  ],
  template: `
    <div class="filters">
      <mat-form-field appearance="outline">
        <mat-label>Search</mat-label>
        <input matInput [formControl]="searchCtrl" placeholder="Item name..." />
        <mat-icon matSuffix>search</mat-icon>
      </mat-form-field>

      <mat-form-field appearance="outline">
        <mat-label>Min. price (USDT)</mat-label>
        <input matInput type="number" [formControl]="minPriceCtrl" />
      </mat-form-field>

      <mat-form-field appearance="outline">
        <mat-label>Max. price (USDT)</mat-label>
        <input matInput type="number" [formControl]="maxPriceCtrl" />
      </mat-form-field>

      <mat-checkbox [formControl]="inStockCtrl">In stock only</mat-checkbox>
    </div>

    @if (loading()) {
      <div class="spinner-wrapper">
        <mat-spinner />
      </div>
    } @else if (items().length === 0) {
      <div class="empty-state">
        <mat-icon>inventory_2</mat-icon>
        <p>No items match your search.</p>
      </div>
    } @else {
      <div class="items-grid">
        @for (item of items(); track item.id) {
          <mat-card class="item-card">
            @if (item.imageUrl) {
              <img mat-card-image [src]="item.imageUrl" [alt]="item.name" />
            } @else {
              <div class="no-image">
                <mat-icon>image_not_supported</mat-icon>
              </div>
            }
            <mat-card-content>
              <h3 class="item-name">{{ item.name }}</h3>
              @if (item.description) {
                <p class="item-desc">{{ item.description }}</p>
              }
              <p class="item-price">{{ item.price | number:'1.2-2' }} USDT</p>
              <p class="item-stock" [class.out-of-stock]="item.stock === 0">
                {{ item.stock > 0 ? 'In stock: ' + item.stock : 'Out of stock' }}
              </p>
            </mat-card-content>
            <mat-card-actions>
              <a mat-button [routerLink]="['/items', item.id]">Details</a>
              @if (auth.isLoggedIn() && !auth.isAdmin() && item.stock > 0) {
                <button mat-raised-button color="primary" (click)="addToCart(item)">
                  <mat-icon>add_shopping_cart</mat-icon>
                  Add to cart
                </button>
              }
            </mat-card-actions>
          </mat-card>
        }
      </div>

      @if (meta()) {
        <mat-paginator
          [length]="meta()!.total"
          [pageSize]="meta()!.pageSize"
          [pageIndex]="meta()!.page - 1"
          [pageSizeOptions]="[12, 24, 48]"
          (page)="onPage($event)"
          showFirstLastButtons>
        </mat-paginator>
      }
    }
  `,
  styles: [`
    .filters {
      display: flex;
      flex-wrap: wrap;
      gap: 16px;
      margin-bottom: 24px;
      align-items: center;
    }
    .items-grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
      gap: 20px;
      margin-bottom: 24px;
    }
    .item-card { display: flex; flex-direction: column; height: 100%; }
    .item-card img { height: 200px; object-fit: cover; }
    .no-image {
      height: 200px;
      display: flex;
      align-items: center;
      justify-content: center;
      background: #f0f0f0;
      color: #aaa;
    }
    .no-image mat-icon { font-size: 48px; width: 48px; height: 48px; }
    mat-card-content { flex: 1; }
    .item-name { margin: 8px 0 4px; font-size: 1rem; font-weight: 500; }
    .item-desc { color: #666; font-size: 0.85rem; margin: 0 0 8px; }
    .item-price { font-size: 1.1rem; font-weight: bold; color: #1565c0; margin: 4px 0; }
    .item-stock { font-size: 0.85rem; color: #4caf50; margin: 0; }
    .out-of-stock { color: #f44336; }
    .spinner-wrapper { display: flex; justify-content: center; padding: 64px; }
    .empty-state {
      text-align: center;
      padding: 64px;
      color: #999;
    }
    .empty-state mat-icon { font-size: 64px; width: 64px; height: 64px; }
  `],
})
export class ItemsListComponent implements OnInit {
  private itemSvc = inject(ItemService);
  private cartSvc = inject(CartService);
  private snackBar = inject(MatSnackBar);
  auth = inject(AuthService);

  items = signal<Item[]>([]);
  meta = signal<PaginatedMeta | null>(null);
  loading = signal(false);

  searchCtrl = new FormControl('');
  minPriceCtrl = new FormControl<number | null>(null);
  maxPriceCtrl = new FormControl<number | null>(null);
  inStockCtrl = new FormControl(false);

  private currentPage = 1;

  ngOnInit(): void {
    this.load();

    this.searchCtrl.valueChanges
      .pipe(debounceTime(400), distinctUntilChanged())
      .subscribe(() => { this.currentPage = 1; this.load(); });

    this.minPriceCtrl.valueChanges
      .pipe(debounceTime(400))
      .subscribe(() => { this.currentPage = 1; this.load(); });

    this.maxPriceCtrl.valueChanges
      .pipe(debounceTime(400))
      .subscribe(() => { this.currentPage = 1; this.load(); });

    this.inStockCtrl.valueChanges
      .subscribe(() => { this.currentPage = 1; this.load(); });
  }

  load(): void {
    this.loading.set(true);
    this.itemSvc.list({
      page: this.currentPage,
      pageSize: 12,
      search: this.searchCtrl.value || undefined,
      minPrice: this.minPriceCtrl.value ?? undefined,
      maxPrice: this.maxPriceCtrl.value ?? undefined,
      inStock: this.inStockCtrl.value ?? undefined,
    }).subscribe({
      next: (res) => {
        this.items.set(res.items);
        this.meta.set(res.meta);
        this.loading.set(false);
      },
      error: () => {
        this.loading.set(false);
        this.snackBar.open('Failed to load items', 'Close', { duration: 3000 });
      },
    });
  }

  onPage(event: PageEvent): void {
    this.currentPage = event.pageIndex + 1;
    this.load();
  }

  addToCart(item: Item): void {
    this.cartSvc.addItem({ itemId: item.id, quantity: 1 }).subscribe({
      next: () => this.snackBar.open(`"${item.name}" added to cart`, 'OK', { duration: 2000 }),
      error: () => this.snackBar.open('Failed to add to cart', 'Close', { duration: 3000 }),
    });
  }
}
