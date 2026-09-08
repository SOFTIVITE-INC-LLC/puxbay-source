import { ToastService } from '../../../core/services/toast';
import { Component, inject, OnInit, computed } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { StorefrontService } from '../../../core/services/storefront.service';
import { SettingsService } from '../../../core/services/settings.service';
import { TenantStore } from '../../../core/services/tenant.store';

export interface ThemePreset {
  name: string;
  primary_color: string;
  secondary_color: string;
  accent_color: string;
  background_color: string;
  card_radius: string;
  button_radius: string;
  font_family: string;
}

@Component({
  selector: 'app-storefront',
  standalone: true,
  imports: [FormsModule],
  templateUrl: './storefront.html',
  styles: `
    .glass-panel {
      background: rgba(255, 255, 255, 0.05);
      backdrop-filter: blur(10px);
      border: 1px solid rgba(255, 255, 255, 0.1);
    }
    .dark .glass-panel {
      background: rgba(0, 0, 0, 0.2);
    }
  `,
})
export class Storefront implements OnInit {
  toastService = inject(ToastService);
  storefrontService = inject(StorefrontService);
  settingsService = inject(SettingsService);
  tenantStore = inject(TenantStore);

  subdomain = computed(() => this.tenantStore.subdomain() || 'yourstore');

  fontOptions = [
    { label: 'Roboto (Clean & Balanced)', value: 'Roboto' },
    { label: 'Inter (Modern UI)', value: 'Inter' },
    { label: 'Outfit (Trendy & Premium)', value: 'Outfit' },
    { label: 'Poppins (Geometric & Friendly)', value: 'Poppins' },
    { label: 'Plus Jakarta Sans (Crisp & High-End)', value: 'Plus Jakarta Sans' },
    { label: 'Montserrat (Bold & Elegant)', value: 'Montserrat' },
    { label: 'Playfair Display (Luxury Serif)', value: 'Playfair Display' },
  ];

  cardRadiusOptions = [
    { label: 'Sharp (0px)', value: '0px', hint: 'Classic MultiShop style' },
    { label: 'Subtle (4px)', value: '4px', hint: 'Minimal curvature' },
    { label: 'Rounded (8px)', value: '8px', hint: 'Modern standard' },
    { label: 'Smooth (12px)', value: '12px', hint: 'Contemporary soft' },
    { label: 'Curved (16px)', value: '16px', hint: 'Card pop style' },
    { label: 'Pill/Oval (24px)', value: '24px', hint: 'Playful rounded' },
  ];

  buttonRadiusOptions = [
    { label: 'Sharp (0px)', value: '0px' },
    { label: 'Subtle (4px)', value: '4px' },
    { label: 'Rounded (8px)', value: '8px' },
    { label: 'Smooth (12px)', value: '12px' },
    { label: 'Full Pill (9999px)', value: '9999px' },
  ];

  popularPrimaryColors = [
    '#FFD333', // MultiShop Gold
    '#2563EB', // Sapphire Blue
    '#059669', // Emerald Green
    '#7C3AED', // Royal Violet
    '#E11D48', // Ruby Rose
    '#EA580C', // Sunset Orange
    '#0891B2', // Cyan Teal
    '#18181B', // Carbon Black
  ];

  popularSecondaryColors = [
    '#3D464D', // MultiShop Dark Charcoal
    '#0F172A', // Slate Navy
    '#1E1B4B', // Deep Indigo
    '#14532D', // Deep Forest
    '#18181B', // Zinc Dark
    '#2E1065', // Midnight Purple
  ];

  themePresets: ThemePreset[] = [
    {
      name: 'MultiShop Gold',
      primary_color: '#FFD333',
      secondary_color: '#3D464D',
      accent_color: '#FFD333',
      background_color: '#F5F5F5',
      card_radius: '0px',
      button_radius: '0px',
      font_family: 'Roboto',
    },
    {
      name: 'Modern Sapphire',
      primary_color: '#2563EB',
      secondary_color: '#0F172A',
      accent_color: '#38BDF8',
      background_color: '#F8FAFC',
      card_radius: '12px',
      button_radius: '8px',
      font_family: 'Inter',
    },
    {
      name: 'Emerald Organic',
      primary_color: '#059669',
      secondary_color: '#064E3B',
      accent_color: '#10B981',
      background_color: '#F0FDF4',
      card_radius: '16px',
      button_radius: '9999px',
      font_family: 'Plus Jakarta Sans',
    },
    {
      name: 'Royal Violet',
      primary_color: '#7C3AED',
      secondary_color: '#1E1B4B',
      accent_color: '#C084FC',
      background_color: '#FAF5FF',
      card_radius: '12px',
      button_radius: '8px',
      font_family: 'Outfit',
    },
    {
      name: 'Ruby Vibrant',
      primary_color: '#E11D48',
      secondary_color: '#18181B',
      accent_color: '#FB7185',
      background_color: '#FFF1F2',
      card_radius: '8px',
      button_radius: '6px',
      font_family: 'Poppins',
    },
    {
      name: 'Sunset Amber',
      primary_color: '#D97706',
      secondary_color: '#292524',
      accent_color: '#F59E0B',
      background_color: '#FFFBEB',
      card_radius: '6px',
      button_radius: '6px',
      font_family: 'Montserrat',
    },
    {
      name: 'Clean Minimal',
      primary_color: '#18181B',
      secondary_color: '#27272A',
      accent_color: '#71717A',
      background_color: '#FAFAFA',
      card_radius: '4px',
      button_radius: '4px',
      font_family: 'Inter',
    },
  ];

  ngOnInit() {
    this.storefrontService.getSettings().subscribe(settings => {
      // Ensure defaults if not set
      if (settings) {
        if (!settings.primary_color) settings.primary_color = '#FFD333';
        if (!settings.secondary_color) settings.secondary_color = '#3D464D';
        if (!settings.accent_color) settings.accent_color = '#FFD333';
        if (!settings.background_color) settings.background_color = '#F5F5F5';
        if (!settings.card_radius) settings.card_radius = '0px';
        if (!settings.button_radius) settings.button_radius = '0px';
        if (!settings.font_family) settings.font_family = 'Roboto';
      }
    });
  }

  applyThemePreset(preset: ThemePreset) {
    const s = this.storefrontService.settings();
    if (!s) return;
    s.primary_color = preset.primary_color;
    s.secondary_color = preset.secondary_color;
    s.accent_color = preset.accent_color;
    s.background_color = preset.background_color;
    s.card_radius = preset.card_radius;
    s.button_radius = preset.button_radius;
    s.font_family = preset.font_family;
    this.toastService.showSuccess(`Applied "${preset.name}" theme preset! Click Save Changes to apply.`);
  }

  saveSettings() {
    const s = this.storefrontService.settings();
    if (s) {
      this.storefrontService.updateSettings(s).subscribe(() => {
        this.toastService.showSuccess('Storefront settings updated successfully.');
      });
    }
  }
}
