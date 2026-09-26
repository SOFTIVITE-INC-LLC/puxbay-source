/** @type {import('tailwindcss').Config} */
module.exports = {
  darkMode: 'class',
  content: [
    "./src/**/*.{html,ts}",
    "./projects/storefront/src/**/*.{html,ts}",
    "./projects/admin-frontend/src/**/*.{html,ts}"
  ],
  theme: {
    extend: {
      animation: {
        shine: 'shine 1s forwards',
      },
      keyframes: {
        shine: {
          '100%': { left: '200%' },
        },
      },
      colors: {
        puxbay: {
          navy: '#050811',
          deep: '#0B101D',
          surface: '#121B2D',
          primary: '#2563EB',
          steel: '#0284C7',
          ice: '#06B6D4',
          cyan: '#06B6D4',
        },
        primary: {
          50: '#F0F9FF',
          100: '#E0F2FE',
          200: '#BAE6FD',
          300: '#7DD3FC',
          400: '#38BDF8',
          500: '#2563EB',
          600: '#1D4ED8',
          700: '#1E40AF',
          800: '#121B2D',
          900: '#0B101D',
          950: '#050811',
        },
        indigo: {
          50: '#F0F9FF',
          100: '#E0F2FE',
          200: '#BAE6FD',
          300: '#7DD3FC',
          400: '#38BDF8',
          500: '#2563EB',
          600: '#1D4ED8',
          700: '#1E40AF',
          800: '#1E293B',
          900: '#0F172A',
          950: '#0B132B',
        },
        violet: {
          50: '#F0F9FF',
          100: '#E0F2FE',
          200: '#BAE6FD',
          300: '#7DD3FC',
          400: '#38BDF8',
          500: '#2563EB',
          600: '#1D4ED8',
          700: '#1E40AF',
          800: '#1E293B',
          900: '#0F172A',
          950: '#0B132B',
        },
        purple: {
          50: '#F0F9FF',
          100: '#E0F2FE',
          200: '#BAE6FD',
          300: '#7DD3FC',
          400: '#38BDF8',
          500: '#1E40AF',
          600: '#1D4ED8',
          700: '#1E293B',
          800: '#0F172A',
          900: '#0B132B',
          950: '#082F49',
        },
        cyan: {
          50: '#ECFEFF',
          100: '#CFFAFE',
          200: '#A5F3FC',
          300: '#67E8F9',
          400: '#22D3EE',
          500: '#06B6D4',
          600: '#0891B2',
          700: '#0E7490',
          800: '#155E75',
          900: '#164E63',
          950: '#083344',
        },
        secondary: {
          500: '#06B6D4',
        },
        accent: {
          500: '#06B6D4',
        }
      }
    },
  },
  plugins: [
    require('daisyui'),
  ],
}

