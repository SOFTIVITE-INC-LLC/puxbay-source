import { Component, OnInit, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { BroadcastService, Broadcast } from '../../services/broadcast.service';

@Component({
  selector: 'app-broadcasts',
  standalone: true,
  imports: [CommonModule, FormsModule],
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
