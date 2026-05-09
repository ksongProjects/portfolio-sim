import type { Metadata } from "next";
import { Inter, Work_Sans } from "next/font/google";
import "./globals.css";
import { Navbar } from "@/components/navbar";
import { Toaster } from "sonner";

const inter = Inter({
  variable: "--font-inter",
  subsets: ["latin"],
  display: "swap",
});

const workSans = Work_Sans({
  variable: "--font-work-sans",
  subsets: ["latin"],
  display: "swap",
  weight: ["400", "500", "600"],
});

export const metadata: Metadata = {
  title: "Portfolio Sim",
  description: "Portfolio simulation platform",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className={`${inter.variable} ${workSans.variable}`}>
      <body className="min-h-full flex flex-col bg-surface text-on-background antialiased">
        <Navbar />
        <main className="flex-1 flex flex-col">
          {children}
        </main>
        <Toaster position="top-right" />
      </body>
    </html>
  );
}
