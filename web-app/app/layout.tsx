import "./globals.css";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "PayCore",
  description: "PayCore merchant dashboard & checkout",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="th">
      <body>{children}</body>
    </html>
  );
}
