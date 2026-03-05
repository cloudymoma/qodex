/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        dark: {
          bg: '#0a0a0a',
          'bg-secondary': '#1a1a1a',
          'bg-tertiary': '#2a2a2a',
          border: '#333333',
          text: '#e5e5e5',
          'text-secondary': '#a3a3a3',
        },
        accent: {
          primary: '#3b82f6',
          secondary: '#8b5cf6',
          success: '#10b981',
          warning: '#f59e0b',
          error: '#ef4444',
        },
      },
    },
  },
  plugins: [],
};
