import type { Config } from "tailwindcss";

const config: Config = {
  content: ["./app/**/*.{ts,tsx}", "./components/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        paycore: {
          // surfaces (light theme — mirrors public/assets/paycore.css :root)
          bg: "#f5f7fa",
          surface: "#ffffff",
          surface2: "#fbfcfe",
          // text
          text: "#0f1729",
          text2: "#374151",
          muted: "#6b7280",
          // lines
          line: "#e5e7eb",
          line2: "#eef1f5",
          // accent
          primary: "#185fa5",
          primaryHover: "#0c447c",
          accentBg: "#e6f1fb",
          accentInk: "#0c447c",
          // semantic
          success: "#0f6e56",
          successBg: "#e1f5ee",
          danger: "#a32d2d",
          dangerBg: "#fcebeb",
          warn: "#8a5a0b",
          warnBg: "#fdf3e1",
        },
      },
      borderRadius: { xl2: "1rem" },
      boxShadow: {
        card: "0 1px 3px rgba(16,23,41,.08),0 1px 2px rgba(16,23,41,.04)",
        cardlg: "0 4px 12px rgba(16,23,41,.10),0 2px 4px rgba(16,23,41,.05)",
      },
    },
  },
  plugins: [],
};
export default config;
