import type { Config } from "tailwindcss";

const config: Config = {
  content: ["./app/**/*.{ts,tsx}", "./components/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        paycore: {
          bg: "#0b0f1a",
          surface: "#141a2b",
          primary: "#3b82f6",
          primaryHover: "#2563eb",
          text: "#e5e9f0",
          muted: "#94a3b8",
        },
      },
      borderRadius: { xl2: "1rem" },
    },
  },
  plugins: [],
};
export default config;
