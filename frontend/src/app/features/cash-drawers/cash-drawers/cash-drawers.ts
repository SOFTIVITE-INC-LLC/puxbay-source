import { Component, inject, OnInit, signal } from '@angular/core';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { CashDrawerService } from '../../../core/services/cash-drawer.service';
import { ToastrService } from 'ngx-toastr';

@Component({
  selector: 'app-cash-drawers',
  standalone: true,
  imports: [CommonModule, FormsModule, AppCurrencyPipe],
  templateUrl: './cash-drawers.html',
})
export class CashDrawers implements OnInit {
  drawerService = inject(CashDrawerService);
  private toastr = inject(ToastrService);

  isOpenShiftModalOpen = signal(false);
  isCloseShiftModalOpen = signal(false);
  isFloatModalOpen = signal(false);
  floatType = signal<'add' | 'remove'>('add');

  openingBalance = signal(0);
  actualBalance = signal(0);
  floatAmount = signal(0);
  shiftNotes = signal('');

  ngOnInit() {
    this.drawerService.getDrawers().subscribe();
  }

  openShift() {
    if (this.openingBalance() < 0) return;
    this.drawerService.openShift(this.openingBalance(), this.shiftNotes()).subscribe({
      next: () => {
        this.toastr.success('Shift opened successfully!');
        this.isOpenShiftModalOpen.set(false);
        this.openingBalance.set(0);
        this.shiftNotes.set('');
      },
      error: () => this.toastr.error('Failed to open shift.')
    });
  }

  closeShift() {
    const active = this.drawerService.activeDrawer();
    if (!active) return;
    this.drawerService.closeShift(active.id, this.actualBalance(), this.shiftNotes()).subscribe({
      next: (res) => {
        this.toastr.success('Shift closed successfully!');
        this.isCloseShiftModalOpen.set(false);
      },
      error: () => this.toastr.error('Failed to close shift.')
    });
  }

  addRemoveFloat(type: 'add' | 'remove') {
    const active = this.drawerService.activeDrawer();
    if (!active || this.floatAmount() <= 0) return;
    const action = type === 'add' ? 
      this.drawerService.addFloat(active.id, this.floatAmount()) :
      this.drawerService.removeFloat(active.id, this.floatAmount());
    action.subscribe({
      next: () => {
        this.toastr.success(`Float ${type === 'add' ? 'added' : 'removed'} successfully!`);
        this.isFloatModalOpen.set(false);
        this.floatAmount.set(0);
      },
      error: () => this.toastr.error('Float operation failed.')
    });
  }
}
