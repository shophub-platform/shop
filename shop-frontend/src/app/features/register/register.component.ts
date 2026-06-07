import { Component, inject, signal } from '@angular/core';
import { AbstractControl, FormBuilder, ReactiveFormsModule, ValidationErrors, Validators } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { MatCardModule } from '@angular/material/card';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { AuthService } from '../../core/services/auth.service';

function passwordsMatch(control: AbstractControl): ValidationErrors | null {
  const password = control.get('password')?.value;
  const confirm = control.get('confirmPassword')?.value;
  return password && confirm && password !== confirm ? { passwordsMismatch: true } : null;
}

@Component({
  selector: 'app-register',
  standalone: true,
  imports: [
    ReactiveFormsModule,
    RouterLink,
    MatCardModule,
    MatFormFieldModule,
    MatInputModule,
    MatButtonModule,
    MatIconModule,
    MatSnackBarModule,
    MatProgressSpinnerModule,
  ],
  template: `
    <div class="register-wrapper">
      <mat-card class="register-card">
        <mat-card-header>
          <mat-card-title>Registracija</mat-card-title>
          <mat-card-subtitle>Kreirajte novi nalog</mat-card-subtitle>
        </mat-card-header>

        <mat-card-content>
          <form [formGroup]="form" (ngSubmit)="submit()">
            <mat-form-field appearance="outline" class="full-width">
              <mat-label>Email</mat-label>
              <input matInput type="email" formControlName="email" autocomplete="email" />
              @if (form.get('email')?.hasError('required')) {
                <mat-error>Email je obavezan</mat-error>
              }
              @if (form.get('email')?.hasError('email')) {
                <mat-error>Unesite validan email</mat-error>
              }
            </mat-form-field>

            <mat-form-field appearance="outline" class="full-width">
              <mat-label>Lozinka</mat-label>
              <input matInput [type]="showPassword() ? 'text' : 'password'"
                formControlName="password" autocomplete="new-password" />
              <button mat-icon-button matSuffix type="button"
                (click)="showPassword.set(!showPassword())">
                <mat-icon>{{ showPassword() ? 'visibility_off' : 'visibility' }}</mat-icon>
              </button>
              @if (form.get('password')?.hasError('required')) {
                <mat-error>Lozinka je obavezna</mat-error>
              }
              @if (form.get('password')?.hasError('minlength')) {
                <mat-error>Lozinka mora imati najmanje 8 karaktera</mat-error>
              }
            </mat-form-field>

            <mat-form-field appearance="outline" class="full-width">
              <mat-label>Potvrda lozinke</mat-label>
              <input matInput [type]="showPassword() ? 'text' : 'password'"
                formControlName="confirmPassword" autocomplete="new-password" />
              @if (form.hasError('passwordsMismatch') && form.get('confirmPassword')?.dirty) {
                <mat-error>Lozinke se ne podudaraju</mat-error>
              }
            </mat-form-field>

            <button mat-raised-button color="primary" type="submit"
              [disabled]="form.invalid || loading()" class="full-width">
              @if (loading()) {
                <mat-spinner diameter="20" />
              } @else {
                <mat-icon>person_add</mat-icon>
                Registrujte se
              }
            </button>
          </form>
        </mat-card-content>

        <mat-card-actions>
          <span>Već imate nalog? <a routerLink="/login">Prijavite se</a></span>
        </mat-card-actions>
      </mat-card>
    </div>
  `,
  styles: [`
    .register-wrapper {
      display: flex;
      justify-content: center;
      align-items: center;
      min-height: 60vh;
    }
    .register-card { width: 100%; max-width: 440px; padding: 16px; }
    .full-width { width: 100%; margin-bottom: 12px; }
    mat-card-actions { padding: 8px 16px 16px; font-size: 14px; }
    mat-card-actions a { color: var(--mat-primary-500); text-decoration: none; font-weight: 500; }
    button[type="submit"] { margin-top: 4px; }
  `],
})
export class RegisterComponent {
  private fb = inject(FormBuilder);
  private auth = inject(AuthService);
  private router = inject(Router);
  private snackBar = inject(MatSnackBar);

  loading = signal(false);
  showPassword = signal(false);

  form = this.fb.group(
    {
      email: ['', [Validators.required, Validators.email]],
      password: ['', [Validators.required, Validators.minLength(8)]],
      confirmPassword: ['', Validators.required],
    },
    { validators: passwordsMatch },
  );

  submit(): void {
    if (this.form.invalid || this.loading()) return;

    this.loading.set(true);
    const { email, password } = this.form.value;

    this.auth.register(email!, password!).subscribe({
      next: () => {
        this.snackBar.open('Nalog kreiran! Prijavite se.', 'Zatvori', { duration: 3000 });
        this.router.navigate(['/login']);
      },
      error: (err) => {
        const msg = err.status === 409
          ? 'Email je već u upotrebi'
          : 'Greška pri registraciji, pokušajte ponovo';
        this.snackBar.open(msg, 'Zatvori', { duration: 4000 });
        this.loading.set(false);
      },
    });
  }
}
