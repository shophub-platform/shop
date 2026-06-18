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
import { Web3Service } from '../../core/services/web3.service';
import { ShopConfigService } from '../../core/services/shop-config.service';
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
            <mat-card-title>Order payment</mat-card-title>
            <mat-card-subtitle>ID: {{ order()!.id }}</mat-card-subtitle>
          </mat-card-header>

          <mat-card-content>
            <div class="order-items">
              @for (item of order()!.items; track item.id) {
                <div class="order-item">
                  <span>{{ item.itemName }} × {{ item.quantity }}</span>
                  <span>{{ item.subtotal | number:'1.2-2' }} mUSDT</span>
                </div>
              }
            </div>

            <div class="total-row">
              <strong>Total to pay:</strong>
              <span class="total-amount">{{ order()!.total | number:'1.2-2' }} mUSDT</span>
            </div>

            @if (order()!.status === 'PAID') {
              <div class="paid-banner">
                <mat-icon>check_circle</mat-icon>
                <span>Payment confirmed!</span>
              </div>
              @if (order()!.txHash) {
                <p class="tx-hash">
                  TxHash:
                  <a [href]="'https://sepolia.etherscan.io/tx/' + order()!.txHash"
                     target="_blank" rel="noopener">
                    {{ order()!.txHash!.slice(0, 20) }}...
                  </a>
                </p>
              }
            } @else {
              @if (!web3Svc.isAvailable()) {
                <div class="warning-banner">
                  <mat-icon>warning</mat-icon>
                  <span>MetaMask not detected. Please install it to pay.</span>
                </div>
              } @else {
                <p class="info-text">
                  Click "Pay with MetaMask" to transfer
                  <strong>{{ order()!.total | number:'1.2-2' }} mUSDT</strong>
                  on Sepolia testnet. MetaMask will open for confirmation.
                </p>
              }
            }

            @if (statusMessage()) {
              <p class="status-msg">{{ statusMessage() }}</p>
            }
          </mat-card-content>

          <mat-card-actions>
            @if (order()!.status === 'PENDING_PAYMENT') {
              <button mat-raised-button color="primary"
                (click)="payWithMetaMask()"
                [disabled]="paying() || !web3Svc.isAvailable()">
                @if (paying()) {
                  <mat-spinner diameter="20" />
                } @else {
                  <mat-icon>account_balance_wallet</mat-icon>
                }
                Pay with MetaMask
              </button>
            } @else {
              <a mat-raised-button color="primary" routerLink="/orders">
                My orders
              </a>
            }
            <a mat-button routerLink="/">Back to store</a>
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
    .warning-banner {
      display: flex; align-items: center; gap: 8px;
      padding: 12px; background: #fff3e0; border-radius: 8px;
      color: #e65100; margin: 16px 0;
    }
    .tx-hash { font-size: 0.75rem; word-break: break-all; color: #555; }
    .info-text { color: #666; line-height: 1.6; margin: 16px 0; }
    .status-msg { color: #555; font-size: 0.85rem; font-style: italic; }
    mat-card-actions { display: flex; gap: 12px; flex-wrap: wrap; }
  `],
})
export class CheckoutComponent implements OnInit {
  private route = inject(ActivatedRoute);
  private router = inject(Router);
  private orderSvc = inject(OrderService);
  private snackBar = inject(MatSnackBar);
  web3Svc = inject(Web3Service);
  private shopConfigSvc = inject(ShopConfigService);

  order = signal<Order | null>(null);
  loading = signal(true);
  paying = signal(false);
  statusMessage = signal<string | null>(null);

  ngOnInit(): void {
    const id = this.route.snapshot.paramMap.get('orderId')!;
    this.orderSvc.getById(id).subscribe({
      next: (o) => { this.order.set(o); this.loading.set(false); },
      error: () => { this.loading.set(false); this.router.navigate(['/']); },
    });
  }

  async payWithMetaMask(): Promise<void> {
    this.paying.set(true);
    this.statusMessage.set(null);

    try {
      // 1. Load shop wallet + contract address from backend
      this.statusMessage.set('Loading shop configuration...');
      const config = await this.shopConfigSvc.get();

      // 2. Connect MetaMask
      this.statusMessage.set('Connecting to MetaMask...');
      await this.web3Svc.connect();

      // 3. Make sure we are on Sepolia
      this.statusMessage.set('Switching to Sepolia testnet...');
      await this.web3Svc.switchToSepolia();

      // 4. Send the USDT transfer — MetaMask popup opens here
      this.statusMessage.set('Waiting for MetaMask confirmation...');
      const txHash = await this.web3Svc.transferUSDT(
        config.contractAddress,
        config.walletAddress,
        this.order()!.total,
      );

      // 5. Notify backend — order status becomes PAID
      this.statusMessage.set('Confirming payment on server...');
      const updated = await this.orderSvc.confirmPayment(this.order()!.id, { txHash }).toPromise();
      this.order.set(updated!);

      this.paying.set(false);
      this.statusMessage.set(null);
      this.snackBar.open('Payment confirmed!', 'OK', { duration: 5000 });

    } catch (err: any) {
      this.paying.set(false);
      this.statusMessage.set(null);
      const msg = err?.info?.error?.message ?? err?.message ?? 'Payment failed';
      this.snackBar.open(msg, 'Close', { duration: 7000 });
    }
  }
}
