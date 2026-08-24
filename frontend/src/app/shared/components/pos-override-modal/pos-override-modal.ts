import { Component, EventEmitter, Output, input } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';

@Component({
  selector: 'app-pos-override-modal',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './pos-override-modal.html',
})
export class PosOverrideModalComponent {
  actionName = input<string>('perform this action');
  @Output() confirm = new EventEmitter<string>();
  @Output() cancel = new EventEmitter<void>();

  pin: string = '';

  onKeypadPress(num: number) {
    if (this.pin.length < 4) {
      this.pin += num.toString();
    }
  }

  onBackspace() {
    this.pin = this.pin.slice(0, -1);
  }

  onConfirm() {
    if (this.pin.length === 4) {
      this.confirm.emit(this.pin);
      this.pin = '';
    }
  }

  onCancel() {
    this.pin = '';
    this.cancel.emit();
  }
}
