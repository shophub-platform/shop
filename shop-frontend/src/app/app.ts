import { Component } from '@angular/core';
import { RouterOutlet } from '@angular/router';
import { HeaderComponent } from './layout/header/header.component';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [RouterOutlet, HeaderComponent],
  template: `
    <app-header />
    <main class="container page-content">
      <router-outlet />
    </main>
  `,
  styles: [`
    main { min-height: calc(100vh - 64px); }
  `],
})
export class App {}
