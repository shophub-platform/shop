import { Component, inject, signal, OnInit } from '@angular/core';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { MatChipsModule } from '@angular/material/chips';
import { DecimalPipe } from '@angular/common';
import { OrderService } from '../../core/services/order.service';
import { Order } from '../../core/models/order.model';

@Component({
  selector: 'app-checkout',
  standalone: true,
  imports: [
    RouterLink,
    MatCardModule,
    MatButtonModule,
    MatIconModule,
    MatProgressSpinnerModule,
    MatSnackBarModule,
    MatChipsModule,
    DecimalPipe,
  ],
  template: `
    @if (loading()) {
      <div class="spinner-wrapper"><mat-spinner /></div>
    } @else if (order()) {
      <div class="checkout-wrapper">
        <mat-card class="checkout-card">
          <mat-card-header>
            <mat-icon mat-card-avatar>payment</mat-icon>
            <mat-card-title>Plaćanje porudžbine</mat-card-title>
            <mat-card-subtitle>ID: {{ order()!.id }}</mat-card-subtitle>
          </mat-card-header>

          <mat-card-content>
            <div class="order-items">
              @for (item of order()!.items; track item.id) {
                <div class="order-item">
                  <span>{{ item.itemName }} × {{ item.quantity }}</span>
                  <span>{{ item.subtotal | number:'1.2-2' }} USDT</span>
                </div>
              }
            </div>

            <div class="total-row">
              <strong>Ukupno za uplatu:</strong>
              <span class="total-amount">{{ order()!.total | number:'1.2-2' }} USDT</span>
            </div>

            @if (order()!.status === 'PAID') {
              <div class="paid-banner">
                <mat-icon>check_circle</mat-icon>
                <span>Plaćanje potvrđeno!</span>
              </div>
              @if (order()!.txHash) {
                <p class="tx-hash">TxHash: <code>{{ order()!.txHash }}</code></p>
              }
            } @else {
              <p class="info-text">
                Kliknite "Plati MetaMask-om" da biste inicirali blockchain transakciju
                na Sepolia mreži u USDT.
              </p>
            }
          </mat-card-content>

          <mat-card-actions>
            @if (order()!.status === 'PENDING_PAYMENT') {
              <button mat-raised-button color="primary"
                (click)="payWithMetaMask()" [disabled]="paying()">
                @if (paying()) {
                  <mat-spinner diameter="20" />
                } @else {
                  <mat-icon>account_balance_wallet</mat-icon>
                }
                Plati MetaMask-om
              </button>
            } @else {
              <a mat-raised-button color="primary" routerLink="/orders">
                Moje porudžbine
              </a>
            }
            <a mat-button routerLink="/">Nazad na prodavnicu</a>
          </mat-card-actions>
        </mat-card>
      </div>
    }
  `,
  styles: [`
    .spinner-wrapper { display: flex; justify-content: center; padding: 64px; }
    .checkout-wrapper { display: flex; justify-content: center; padding: 32px 0; }
    .checkout-card { width: 100%; max-width: 540px; padding: 16px; }
    .order-items { margin: 16px 0; }
    .order-item {
      display: flex; justify-content: space-between;
      padding: 8px 0; border-bottom: 1px solid #eee;
    }
    .total-row {
      display: flex; justify-content: space-between;
      align-items: center; padding: 16px 0; font-size: 1.05rem;
    }
    .total-amount { font-size: 1.5rem; font-weight: bold; color: #1565c0; }
    .paid-banner {
      display: flex; align-items: center; gap: 8px;
      padding: 12px; background: #e8f5e9; border-radius: 8px;
      color: #2e7d32; margin: 16px 0;
    }
    .tx-hash { font-size: 0.75rem; word-break: break-all; color: #555; }
    .info-text { color: #666; line-height: 1.6; margin: 16px 0; }
    mat-card-actions { display: flex; gap: 12px; flex-wrap: wrap; }
  `],
})
export class CheckoutComponent implements OnInit {
  private route = inject(ActivatedRoute);
  private router = inject(Router);
  private orderSvc = inject(OrderService);
  private snackBar = inject(MatSnackBar);

  order = signal<Order | null>(null);
  loading = signal(true);
  paying = signal(false);

  ngOnInit(): void {
    const id = this.route.snapshot.paramMap.get('orderId')!;
    this.orderSvc.getById(id).subscribe({
      next: (o) => { this.order.set(o); this.loading.set(false); },
      error: () => { this.loading.set(false); this.router.navigate(['/']); },
    });
  }

  async payWithMetaMask(): Promise<void> {
    this.snackBar.open(
      'Web3 integracija dolazi u F6 fazi. Za sada koristite /confirm endpoint direktno.',
      'OK',
      { duration: 5000 },
    );
  }
}
