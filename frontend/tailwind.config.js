/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        'bg-primary': '#0D0D0D',
        'bg-secondary': '#1A1A1A',
        'bg-card': '#242424',
        'bg-hover': '#2D2D2D',
        'border': '#333333',
        'text-primary': '#FFFFFF',
        'text-secondary': '#A0A0A0',
        'text-muted': '#666666',
        'accent-blue': '#3B82F6',
        'accent-green': '#22C55E',
        'accent-red': '#EF4444',
        'accent-yellow': '#F59E0B',
      },
      fontFamily: {
        'primary': ['Space Grotesk', 'sans-serif'],
        'secondary': ['Inter', 'sans-serif'],
      },
    },
  },
  plugins: [],
}
