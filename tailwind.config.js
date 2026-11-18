/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./view/**/*.{html,js,templ}",
    "./public/*.js"
  ],
  theme: {
    extend: {
      colors: {
        background: "#ffffff",
        foreground: "#171717",
        card: "#ffffff",
        muted: "#f5f5f5",
        "muted-foreground": "#737373",
        primary: "#171717",
        "primary-foreground": "#ffffff",
        input: "#d4d4d4",
        ring: "#737373",
        accent: "#f5f5f5",
        "accent-foreground": "#171717",
      },
    },
  },
  plugins: [],
}
