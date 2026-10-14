/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}"
  ],
  theme: {
    extend: {
      colors: {
        // Go-chat theme colors - modern, vibrant palette
        'primary': '#6366F1', // Indigo - more vibrant than before
        'primary-light': '#818CF8', // Lighter indigo for hover states
        'primary-dark': '#4F46E5', // Darker indigo for active states
        'secondary': '#FBBF24', // Amber - warmer yellow
        'secondary-light': '#FCD34D', // Lighter amber
        'secondary-dark': '#F59E0B', // Darker amber
        'success': '#10B981', // Emerald green
        'success-light': '#34D399', // Lighter green
        'success-dark': '#059669', // Darker green
        'danger': '#EF4444', // Red
        'danger-light': '#F87171', // Lighter red
        'danger-dark': '#DC2626', // Darker red
        'info': '#06B6D4', // Sky blue
        'info-light': '#67E8F9', // Lighter cyan
        'info-dark': '#0891B2', // Darker cyan
        'warning': '#F59E0B', // Amber
        'warning-light': '#FBBF24', // Lighter amber
        'warning-dark': '#D97706', // Darker amber
        'dark': '#0F172A', // Very dark blue-gray (navy)
        'darker': '#020617', // Even darker
        'darkest': '#000000', // Pure black for text contrast
        'light': '#F8FAFC', // Slate 50 - off-white
        'lighter': '#F1F5F9', // Slate 100
        'border': '#334155', // Slate 600 - softer border
        'border-light': '#475569', // Slate 700
        'input-bg': '#1E293B', // Slate 800 - input background
        'input-bg-light': '#2D3748', // Slightly lighter input bg
        'card-bg': '#111827', // Gray 900 - card backgrounds
        'card-bg-light': '#1F2937', // Gray 800
      },
      boxShadow: {
        'sm': '0 1px 3px 0 rgba(0, 0, 0, 0.1), 0 1px 2px 0 rgba(0, 0, 0, 0.06)',
        'md': '0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06)',
        'lg': '0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05)',
        'xl': '0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04)',
        'inner': 'inset 0 2px 4px 0 rgba(0, 0, 0, 0.06)',
        'none': 'none',
      },
      borderRadius: {
        'xs': '0.125rem',
        'sm': '0.25rem',
        'md': '0.375rem',
        'lg': '0.5rem',
        'xl': '0.75rem',
        '2xl': '1rem',
        '3xl': '1.25rem',
        'full': '9999px',
      },
      transitionDuration: {
        '150': '150ms',
        '200': '200ms',
        '300': '300ms',
        '500': '500ms',
      },
      transitionTimingFunction: {
        'spring': 'cubic-bezier(0.4, 0, 0.2, 1)',
        'bounce': 'cubic-bezier(0.68, -0.55, 0.265, 1.55)',
      },
      fontFamily: {
        sans: ['-apple-system', 'BlinkMacSystemFont', 'Segoe UI', 'Roboto', 'Helvetica Neue', 'Arial', 'sans-serif'],
        mono: ['SFMono-Regular', 'Menlo', 'Monaco', 'Consolas', 'Liberation Mono', 'Courier New', 'monospace'],
      },
      fontSize: {
        'xs': ['0.75rem', { lineHeight: '1rem' }],
        'sm': ['0.875rem', { lineHeight: '1.25rem' }],
        'base': ['1rem', { lineHeight: '1.5rem' }],
        'lg': ['1.125rem', { lineHeight: '1.75rem' }],
        'xl': ['1.25rem', { lineHeight: '1.75rem' }],
        '2xl': ['1.5rem', { lineHeight: '2rem' }],
        '3xl': ['1.875rem', { lineHeight: '2.25rem' }],
        '4xl': ['2.25rem', { lineHeight: '2.5rem' }],
        '5xl': ['3rem', { lineHeight: '1' }],
        '6xl': ['3.75rem', { lineHeight: '1' }],
        '7xl': ['4.5rem', { lineHeight: '1' }],
        '8xl': ['6rem', { lineHeight: '1' }],
        '9xl': ['8rem', { lineHeight: '1' }],
      },
      spacing: {
        'px': '1px',
        '0': '0px',
        '0.5': '0.125rem',
        '1': '0.25rem',
        '1.5': '0.375rem',
        '2': '0.5rem',
        '2.5': '0.625rem',
        '3': '0.75rem',
        '3.5': '0.875rem',
        '4': '1rem',
        '5': '1.25rem',
        '6': '1.5rem',
        '7': '1.75rem',
        '8': '2rem',
        '9': '2.25rem',
        '10': '2.5rem',
        '11': '2.75rem',
        '12': '3rem',
        '14': '3.5rem',
        '16': '4rem',
        '20': '5rem',
        '24': '6rem',
        '28': '7rem',
        '32': '8rem',
        '36': '9rem',
        '40': '10rem',
        '44': '11rem',
        '48': '12rem',
        '52': '13rem',
        '56': '14rem',
        '60': '15rem',
        '64': '16rem',
        '72': '18rem',
        '80': '20rem',
        '96': '24rem',
      },
      zIndex: {
        'auto': 'auto',
        '0': '0',
        '10': '10',
        '20': '20',
        '30': '30',
        '40': '40',
        '50': '50',
        '60': '60',
        '70': '70',
        '80': '80',
        '90': '90',
        '100': '100',
        '1000': '1000',
        '10000': '10000',
      }
    },
  },
  plugins: [],
}