import Script from "next/script";
import type { Metadata } from "next";
import { ErrorBoundary } from "@/components/ErrorBoundary";
import "./globals.css";

export const metadata: Metadata = {
  title: "Sweet Bonanza | MUDRO Casino",
  description: "Play Sweet Bonanza with real balance via Telegram",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body className="bg-black text-white">
        <Script src="https://telegram.org/js/telegram-web-app.js" strategy="beforeInteractive" />
        <ErrorBoundary>
          {children}
        </ErrorBoundary>
      </body>
    </html>
  );
}
