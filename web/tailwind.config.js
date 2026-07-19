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
        brand: {
          50: '#f0f4ff',
          100: '#cce0ff',
          200: '#99b3ff',
          300: '#6680ff',
          400: '#334dff',
          500: '#001aff',
          600: '#0014cc',
          700: '#000f99',
          800: '#000a66',
          900: '#000533',
        }
      }
    },
  },
  plugins: [],
}
