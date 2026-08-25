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
        primary: {
          500: '#0ea5e9',
        },
        secondary: {
          500: '#64748b',
        },
        accent: {
          500: '#f59e0b',
        }
      }
    },
  },
  plugins: [
    require('daisyui'),
  ],
}
