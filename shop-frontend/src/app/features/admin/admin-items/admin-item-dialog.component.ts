import { Component, inject, Inject } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import {
  MAT_DIALOG_DATA,
  MatDialogModule,
  MatDialogRef,
} from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { ItemService } from '../../../core/services/item.service';
import { Item } from '../../../core/models/item.model';

@Component({
  selector: 'app-admin-item-dialog',
  standalone: true,
  imports: [
    ReactiveFormsModule,
    MatDialogModule,
    MatFormFieldModule,
    MatInputModule,
    MatButtonModule,
    MatSnackBarModule,
  ],
  template: `
    <h2 mat-dialog-title>{{ isEdit ? 'Izmeni artikal' : 'Novi artikal' }}</h2>

    <mat-dialog-content>
      <form [formGroup]="form" class="dialog-form">
        <mat-form-field appearance="outline" class="full-width">
          <mat-label>Naziv</mat-label>
          <input matInput formControlName="name" />
          @if (form.get('name')?.hasError('required')) {
            <mat-error>Naziv je obavezan</mat-error>
          }
        </mat-form-field>

        <mat-form-field appearance="outline" class="full-width">
          <mat-label>Opis</mat-label>
          <textarea matInput formControlName="description" rows="3"></textarea>
        </mat-form-field>

        <mat-form-field appearance="outline" class="full-width">
          <mat-label>Cena (USDT)</mat-label>
          <input matInput type="number" formControlName="price" min="0.01" step="0.01" />
          @if (form.get('price')?.hasError('required')) {
            <mat-error>Cena je obavezna</mat-error>
          }
          @if (form.get('price')?.hasError('min')) {
            <mat-error>Cena mora biti veća od 0</mat-error>
          }
        </mat-form-field>

        <mat-form-field appearance="outline" class="full-width">
          <mat-label>Stanje (kom.)</mat-label>
          <input matInput type="number" formControlName="stock" min="0" />
          @if (form.get('stock')?.hasError('required')) {
            <mat-error>Stanje je obavezno</mat-error>
          }
        </mat-form-field>

        <mat-form-field appearance="outline" class="full-width">
          <mat-label>URL slike</mat-label>
          <input matInput formControlName="imageUrl" />
        </mat-form-field>
      </form>
    </mat-dialog-content>

    <mat-dialog-actions align="end">
      <button mat-button mat-dialog-close>Otkaži</button>
      <button mat-raised-button color="primary"
        (click)="submit()" [disabled]="form.invalid || saving">
        {{ isEdit ? 'Sačuvaj' : 'Kreiraj' }}
      </button>
    </mat-dialog-actions>
  `,
  styles: [`
    .dialog-form { display: flex; flex-direction: column; gap: 8px; padding-top: 8px; }
    .full-width { width: 100%; }
  `],
})
export class AdminItemDialogComponent {
  private fb = inject(FormBuilder);
  private itemSvc = inject(ItemService);
  private snackBar = inject(MatSnackBar);
  private dialogRef = inject(MatDialogRef<AdminItemDialogComponent>);

  data: Item | null = inject(MAT_DIALOG_DATA);

  isEdit = !!this.data;
  saving = false;

  form = this.fb.group({
    name: [this.data?.name ?? '', Validators.required],
    description: [this.data?.description ?? ''],
    price: [this.data?.price ?? null, [Validators.required, Validators.min(0.01)]],
    stock: [this.data?.stock ?? 0, [Validators.required, Validators.min(0)]],
    imageUrl: [this.data?.imageUrl ?? ''],
  });

  submit(): void {
    if (this.form.invalid) return;
    this.saving = true;
    const val = this.form.value;
    const body = {
      name: val.name!,
      description: val.description || null,
      price: val.price!,
      stock: val.stock!,
      imageUrl: val.imageUrl || null,
    };

    const req$ = this.isEdit
      ? this.itemSvc.update(this.data!.id, body)
      : this.itemSvc.create(body);

    req$.subscribe({
      next: () => {
        this.saving = false;
        this.snackBar.open(
          this.isEdit ? 'Artikal izmenjen' : 'Artikal kreiran',
          'OK',
          { duration: 2000 },
        );
        this.dialogRef.close(true);
      },
      error: () => {
        this.saving = false;
        this.snackBar.open('Greška pri čuvanju', 'Zatvori', { duration: 3000 });
      },
    });
  }
}
