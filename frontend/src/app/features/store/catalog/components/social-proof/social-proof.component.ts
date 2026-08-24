import { Component, OnDestroy, OnInit, signal } from '@angular/core';
import { CommonModule } from '@angular/common';

interface Notification {
  name: string;
  location: string;
  product: string;
  timeAgo: string;
  image: string;
}

@Component({
  selector: 'app-social-proof',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './social-proof.component.html'
})
export class SocialProofComponent implements OnInit, OnDestroy {
  isVisible = signal(false);
  currentNotification = signal<Notification | null>(null);
  
  private notifications: Notification[] = [
    { name: 'Sarah', location: 'New York', product: 'Premium Wireless Headphones', timeAgo: '2 mins ago', image: 'https://images.unsplash.com/photo-1505740420928-5e560c06d30e?w=100&h=100&fit=crop' },
    { name: 'Michael', location: 'London', product: 'Mechanical Keyboard', timeAgo: '5 mins ago', image: 'https://images.unsplash.com/photo-1595225476474-87563907a212?w=100&h=100&fit=crop' },
    { name: 'Emma', location: 'Sydney', product: 'Smart Watch Gen 4', timeAgo: '12 mins ago', image: 'https://images.unsplash.com/photo-1523275335684-37898b6baf30?w=100&h=100&fit=crop' },
    { name: 'David', location: 'Toronto', product: 'Noise Cancelling Earbuds', timeAgo: 'Just now', image: 'https://images.unsplash.com/photo-1590658268037-6bf12165a8df?w=100&h=100&fit=crop' }
  ];

  private timer: any;
  private interval: any;

  ngOnInit() {
    // Start showing notifications randomly
    this.scheduleNext();
  }

  ngOnDestroy() {
    if (this.timer) clearTimeout(this.timer);
    if (this.interval) clearInterval(this.interval);
  }

  private scheduleNext() {
    // Wait between 10 to 20 seconds before showing
    const waitTime = Math.floor(Math.random() * 10000) + 10000;
    
    this.timer = setTimeout(() => {
      this.showNotification();
    }, waitTime);
  }

  private showNotification() {
    const randomItem = this.notifications[Math.floor(Math.random() * this.notifications.length)];
    this.currentNotification.set(randomItem);
    this.isVisible.set(true);

    // Hide after 5 seconds
    setTimeout(() => {
      this.isVisible.set(false);
      this.scheduleNext();
    }, 5000);
  }

  close() {
    this.isVisible.set(false);
  }
}
