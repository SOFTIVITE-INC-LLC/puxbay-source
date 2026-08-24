import { Component, ViewEncapsulation } from '@angular/core';
import { RouterModule } from '@angular/router';
import { CommonModule } from '@angular/common';

export interface FeatureItem {
  icon: string;
  title: string;
  description: string;
}

export interface FeatureCategory {
  id: string;
  title: string;
  subtitle: string;
  features: FeatureItem[];
}

@Component({
  selector: 'app-features',
  standalone: true,
  imports: [RouterModule, CommonModule],
  templateUrl: './features.html',
  styleUrls: ['./features.css'],
  encapsulation: ViewEncapsulation.None,
})
export class Features {
  categories: FeatureCategory[] = [
    {
      id: 'pos',
      title: 'Point of Sale (POS)',
      subtitle: 'Fast, reliable, and built for modern retail.',
      features: [
        { icon: 'fa-solid fa-wifi', title: 'Offline Mode', description: 'Keep selling even when the internet goes down. Auto-syncs when online.' },
        { icon: 'fa-solid fa-code-branch', title: 'Multi-Branch Management', description: 'Manage multiple stores or branches from a single unified dashboard.' },
        { icon: 'fa-solid fa-barcode', title: 'Barcode Scanning', description: 'Lightning-fast checkout with native barcode scanner support.' },
        { icon: 'fa-solid fa-users-gear', title: 'Staff Shifts & Permissions', description: 'Control cash registers, track employee shifts, and set granular access levels.' },
        { icon: 'fa-solid fa-receipt', title: 'Custom Digital Receipts', description: 'Print custom receipts or send them instantly via SMS/Email to customers.' }
      ]
    },
    {
      id: 'inventory',
      title: 'Inventory Management',
      subtitle: 'Complete visibility and control over your stock.',
      features: [
        { icon: 'fa-solid fa-triangle-exclamation', title: 'Low Stock Alerts', description: 'Get notified instantly when stock runs low before you miss a sale.' },
        { icon: 'fa-solid fa-warehouse', title: 'Multi-Warehouse Tracking', description: 'Track and transfer stock across different warehouses and branch locations.' },
        { icon: 'fa-solid fa-truck-field', title: 'Supplier POs', description: 'Generate professional Purchase Orders and manage supplier relationships.' },
        { icon: 'fa-solid fa-layer-group', title: 'Variants & Serial Numbers', description: 'Manage complex products with size/color variants and unique serial tracking.' },
        { icon: 'fa-solid fa-boxes-stacked', title: 'Batch & Expiry Tracking', description: 'Track product batches and expiry dates to minimize waste and loss.' }
      ]
    },
    {
      id: 'ecommerce',
      title: 'E-Commerce Storefront',
      subtitle: 'Take your physical store online in one click.',
      features: [
        { icon: 'fa-solid fa-globe', title: 'Free Custom Subdomain', description: 'Instantly launch your store at yourname.puxbay.com for free.' },
        { icon: 'fa-solid fa-box-open', title: 'Order Fulfillment Workflow', description: 'Streamlined picking, packing, and dispatching for online orders.' },
        { icon: 'fa-solid fa-mobile-screen-button', title: 'Mobile-Optimized Cart', description: 'A seamless, friction-free checkout experience on any device.' },
        { icon: 'fa-solid fa-motorcycle', title: 'Delivery Integrations', description: 'Manage in-house drivers or integrate with third-party logistics partners.' },
        { icon: 'fa-brands fa-stripe-s', title: 'Global Payment Gateways', description: 'Accept cards, Apple Pay, and local payments via Paystack and Stripe.' }
      ]
    },
    {
      id: 'crm',
      title: 'Customer Management (CRM)',
      subtitle: 'Build relationships that drive repeat business.',
      features: [
        { icon: 'fa-solid fa-medal', title: 'Loyalty Programs', description: 'Reward your best customers with points and automated discounts.' },
        { icon: 'fa-solid fa-clock-rotate-left', title: 'Purchase History', description: 'View complete cross-channel purchase history for every customer.' },
        { icon: 'fa-solid fa-comment-sms', title: 'SMS & Email Marketing', description: 'Send targeted broadcast campaigns directly to your customer segments.' },
        { icon: 'fa-solid fa-chart-pie', title: 'Customer Segmentation', description: 'Automatically categorize customers based on spending habits and frequency.' },
        { icon: 'fa-solid fa-money-bill-transfer', title: 'Store Credit & Tabs', description: 'Allow trusted customers to run tabs or issue store credit for returns.' }
      ]
    },
    {
      id: 'analytics',
      title: 'Analytics & Reporting',
      subtitle: 'Data-driven insights to grow your bottom line.',
      features: [
        { icon: 'fa-solid fa-chart-line', title: 'Real-time Dashboards', description: 'Monitor sales, traffic, and inventory value in real-time from anywhere.' },
        { icon: 'fa-solid fa-file-invoice-dollar', title: 'Tax Reporting', description: 'Automated tax calculations and exportable reports for accounting.' },
        { icon: 'fa-solid fa-brain', title: 'AI-Powered Insights', description: 'Predictive analytics highlighting your best-selling items and slow movers.' },
        { icon: 'fa-solid fa-cash-register', title: 'Shift & Drawer Reports', description: 'Detailed cash management reports to prevent shrinkage and theft.' },
        { icon: 'fa-solid fa-scale-balanced', title: 'Profit & Loss Statements', description: 'Understand your true margins with comprehensive financial reporting.' }
      ]
    }
  ];

  scrollToCategory(id: string) {
    const el = document.getElementById(id);
    if (el) {
      el.scrollIntoView({ behavior: 'smooth', block: 'start' });
    }
  }
}
