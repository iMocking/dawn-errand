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
      boxShadow: {
        brand: '0 6px 16px rgba(255,107,44,.32)',
        'brand-hover': '0 8px 20px rgba(255,107,44,.4)',
        card: '0 2px 10px rgba(0,0,0,.04)',
      },
      backgroundImage: {
        'brand-gradient': 'linear-gradient(135deg,#FF8A3D,#FF6B2C)',
        'hero-gradient': 'linear-gradient(150deg,#FF8A3D 0%,#FF6B2C 55%,#F2531C 100%)',
      },
    },
  },
  plugins: [],
}
