/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,ts,tsx}'],
  theme: {
    extend: {
      colors: {
        primary: {
          DEFAULT: '#FF6B2C',
          light: '#FF8A3D',
          dark: '#F2531C',
          bg: '#FFF3E8',
        },
        page: '#F5F6F8',
        ink: {
          1: '#1D2129',
          2: '#4E5969',
          3: '#86909C',
          4: '#C9CDD4',
        },
        line: '#E5E6EB',
        head: '#F7F8FA',
        sem: {
          red: '#FF4D4F',
          blue: '#2B7FFF',
          green: '#00B578',
          yellow: '#D48806',
          purple: '#722ED1',
        },
      },
      fontFamily: {
        sans: [
          '-apple-system',
          'BlinkMacSystemFont',
          'PingFang SC',
          'Microsoft YaHei',
          'sans-serif',
        ],
      },
      borderRadius: {
        card: '12px',
        tag: '4px',
      },
    },
  },
  plugins: [],
}
