import { Component, Input } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-empty-state',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './empty-state.component.html'
})
export class EmptyStateComponent {
  @Input() icon = 'inbox';
  @Input() title = 'No data available';
  @Input() description = 'There is currently no data to display in this section.';
}
