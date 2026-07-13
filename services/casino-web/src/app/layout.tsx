import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Sweet Bonanza | MUDRO Casino",
  description: "Play Sweet Bonanza with real balance via Telegram",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body className="bg-black text-white">{children}</body>
    </html>
  );
}