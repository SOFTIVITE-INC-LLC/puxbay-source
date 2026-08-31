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
          navy: '#011f4b',
          deep: '#03396c',
          primary: '#005b96',
          steel: '#6497b1',
          ice: '#b3cde0',
        },
        primary: {
          50: '#f0f6fa',
          100: '#dbe9f4',
          200: '#b3cde0',
          300: '#8bb3ce',
          400: '#6497b1',
          500: '#005b96',
          600: '#034a7e',
          700: '#03396c',
          800: '#022953',
          900: '#011f4b',
          950: '#001330',
        },
        indigo: {
          50: '#f0f6fa',
          100: '#dbe9f4',
          200: '#b3cde0',
          300: '#8bb3ce',
          400: '#6497b1',
          500: '#005b96',
          600: '#034a7e',
          700: '#03396c',
          800: '#022953',
          900: '#011f4b',
          950: '#001330',
        },
        violet: {
          50: '#f0f6fa',
          100: '#dbe9f4',
          200: '#b3cde0',
          300: '#8bb3ce',
          400: '#6497b1',
          500: '#005b96',
          600: '#034a7e',
          700: '#03396c',
          800: '#022953',
          900: '#011f4b',
          950: '#001330',
        },
        purple: {
          50: '#f0f6fa',
          100: '#dbe9f4',
          200: '#b3cde0',
          300: '#8bb3ce',
          400: '#6497b1',
          500: '#03396c',
          600: '#03396c',
          700: '#011f4b',
          800: '#011f4b',
          900: '#011f4b',
          950: '#001330',
        },
        secondary: {
          500: '#6497b1',
        },
        accent: {
          500: '#b3cde0',
        }
      }
    },
  },
  plugins: [
    require('daisyui'),
  ],
}

