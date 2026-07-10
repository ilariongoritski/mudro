import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";
import { Toaster } from "@/components/ui/toaster";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "Sweet Bonanza — Tumble Slot",
  description: "Sweet Bonanza style slot: 5×5 pay-anywhere grid, tumbling cascades, multiplier bombs and a gorgeous free-spins bonus mode.",
  keywords: ["slot", "casino", "Sweet Bonanza", "tumble", "cascade", "free spins", "Next.js"],
  authors: [{ name: "SweetBonanza" }],
  icons: {
    icon: "https://z-cdn.chatglm.cn/z-ai/static/logo.svg",
  },
  openGraph: {
    title: "Sweet Bonanza — Tumble Slot",
    description: "5×5 pay-anywhere slot with tumbling cascades, multiplier bombs and a free-spins bonus mode.",
    url: "https://chat.z.ai",
    siteName: "SweetBonanza",
    type: "website",
  },
  twitter: {
    card: "summary_large_image",
    title: "Sweet Bonanza — Tumble Slot",
    description: "5×5 pay-anywhere slot with tumbling cascades and a free-spins bonus mode.",
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body
        className={`${geistSans.variable} ${geistMono.variable} antialiased bg-background text-foreground`}
      >
        {children}
        <Toaster />
      </body>
    </html>
  );
}
