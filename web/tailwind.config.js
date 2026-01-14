/** @type {import('tailwindcss').Config} */
export default {
  content: ['./src/**/*.{html,js,svelte,ts}'],
  theme: {
    extend: {
      colors: {
        // Terminal green theme
        terminal: {
          bg: '#0a0a0a',
          'bg-light': '#121212',
          green: '#00ff00',
          'green-dim': '#00cc00',
          'green-bright': '#33ff33',
          amber: '#ffb000',
          'amber-dim': '#cc8c00',
          red: '#ff3333',
          cyan: '#00ffff',
          purple: '#cc00ff',
          gray: '#333333',
          'gray-light': '#666666',
        },
      },
      fontFamily: {
        mono: [
          'JetBrains Mono',
          'Fira Code',
          'SF Mono',
          'Monaco',
          'Inconsolata',
          'Roboto Mono',
          'Source Code Pro',
          'monospace',
        ],
      },
      animation: {
        'cursor-blink': 'blink 1s step-end infinite',
        'scanline': 'scanline 8s linear infinite',
        'flicker': 'flicker 0.15s infinite',
        'glow': 'glow 2s ease-in-out infinite alternate',
      },
      keyframes: {
        blink: {
          '0%, 100%': { opacity: '1' },
          '50%': { opacity: '0' },
        },
        scanline: {
          '0%': { transform: 'translateY(-100%)' },
          '100%': { transform: 'translateY(100vh)' },
        },
        flicker: {
          '0%': { opacity: '0.97' },
          '50%': { opacity: '1' },
          '100%': { opacity: '0.98' },
        },
        glow: {
          '0%': { textShadow: '0 0 5px #00ff00, 0 0 10px #00ff00' },
          '100%': { textShadow: '0 0 10px #00ff00, 0 0 20px #00ff00, 0 0 30px #00ff00' },
        },
      },
      boxShadow: {
        'terminal': '0 0 10px rgba(0, 255, 0, 0.3)',
        'terminal-hover': '0 0 20px rgba(0, 255, 0, 0.5)',
      },
    },
  },
  plugins: [],
};
