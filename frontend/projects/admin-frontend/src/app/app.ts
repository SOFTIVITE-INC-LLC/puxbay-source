import { Component, signal } from '@angular/core';
import { RouterOutlet } from '@angular/router';
import { AlertContainerComponent } from './layout/alert-container/alert-container';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [RouterOutlet, AlertContainerComponent],
  templateUrl: './app.html',
})
export class App {
  protected readonly title = signal('admin-frontend');
}
