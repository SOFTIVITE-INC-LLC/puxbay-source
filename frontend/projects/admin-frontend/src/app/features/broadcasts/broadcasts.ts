import { Component, OnInit, inject, signal } from '@angular/core';
import { CommonModule, TitleCasePipe } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { BroadcastService, Broadcast } from '../../services/broadcast.service';
import { Pipe, PipeTransform } from '@angular/core';

@Pipe({ name: 'broadcastCount', standalone: true })
export class BroadcastCountPipe implements PipeTransform {
  transform(broadcasts: Broadcast[], type: string): number {
    return broadcasts.filter(b => b.type === type).length;
  }
}

@Component({
  selector: 'app-broadcasts',
  standalone: true,
  imports: [CommonModule, FormsModule, TitleCasePipe, BroadcastCountPipe],
  templateUrl: './broadcasts.html',
})
export class BroadcastsComponent implements OnInit {
  private service = inject(BroadcastService);

  broadcasts = signal<Broadcast[]>([]);
  isLoading = signal(true);
  isModalOpen = signal(false);
  isSaving = signal(false);

  form = signal<Broadcast>({
    title: '',
    message: '',
    type: 'info',
    target_audience: 'all'
  });

  audiences = [
    { value: 'all', label: 'All Tenants', icon: 'groups' },
    { value: 'active', label: 'Active Only', icon: 'verified' },
    { value: 'trialing', label: 'Trialing', icon: 'hourglass_top' },
    { value: 'past_due', label: 'Past Due', icon: 'schedule' },
    { value: 'suspended', label: 'Suspended', icon: 'block' },
  ];

  ngOnInit() {
    this.loadBroadcasts();
  }

  loadBroadcasts() {
    this.isLoading.set(true);
    this.service.getBroadcasts().subscribe({
      next: (data) => {
        this.broadcasts.set(data || []);
        this.isLoading.set(false);
      },
      error: (err) => {
        console.error('Failed to load broadcasts', err);
        this.isLoading.set(false);
      }
    });
  }

  openModal() {
    this.form.set({ title: '', message: '', type: 'info', target_audience: 'all' });
    this.isModalOpen.set(true);
  }

  closeModal() {
    this.isModalOpen.set(false);
  }

  setFormField(field: keyof Broadcast, value: string) {
    this.form.update(f => ({ ...f, [field]: value }));
  }

  createBroadcast() {
    this.isSaving.set(true);
    this.service.createBroadcast(this.form()).subscribe({
      next: () => {
        this.isSaving.set(false);
        this.closeModal();
        this.loadBroadcasts();
      },
      error: (err) => {
        console.error('Failed to create broadcast', err);
        this.isSaving.set(false);
      }
    });
  }
}
