import { Component, inject, output } from '@angular/core';
import { Router, RouterLink, RouterLinkActive } from '@angular/router';
import { MatToolbarModule } from '@angular/material/toolbar';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatBadgeModule } from '@angular/material/badge';
import { AuthService } from '../../core/services/auth.service';

@Component({
  selector: 'app-header',
  standalone: true,
  imports: [
    RouterLink,
    RouterLinkActive,
    MatToolbarModule,
    MatButtonModule,
    MatIconModule,
    MatBadgeModule,
  ],
  template: `
    <mat-toolbar color="primary">
      <span class="logo" routerLink="/">🛒 ShopHub Store</span>

      <span class="spacer"></span>

      <a mat-button routerLink="/" routerLinkActive="active-link" [routerLinkActiveOptions]="{ exact: true }">
        Home
      </a>

      @if (auth.isLoggedIn() && !auth.isAdmin()) {
        <a mat-button routerLink="/orders">My Orders</a>
      }

      @if (auth.isAdmin()) {
        <a mat-button routerLink="/admin">Admin</a>
      }

      @if (!auth.isAdmin()) {
        <button mat-icon-button routerLink="/cart" (click)="cartClick.emit()">
          <mat-icon>shopping_cart</mat-icon>
        </button>
      }

      @if (auth.isLoggedIn()) {
        <button mat-button (click)="logout()">
          <mat-icon>logout</mat-icon>
          Log out
        </button>
      } @else {
        <button mat-button routerLink="/login">
          <mat-icon>login</mat-icon>
          Sign in
        </button>
      }
    </mat-toolbar>
  `,
  styles: [`
    mat-toolbar { position: sticky; top: 0; z-index: 100; }
    .logo { font-size: 1.2rem; font-weight: 500; cursor: pointer; }
    .spacer { flex: 1 1 auto; }
    .active-link { font-weight: bold; }
  `],
})
export class HeaderComponent {
  auth = inject(AuthService);
  private router = inject(Router);
  cartClick = output<void>();

  logout(): void {
    this.auth.clearToken();
    this.router.navigate(['/']);
  }
}
