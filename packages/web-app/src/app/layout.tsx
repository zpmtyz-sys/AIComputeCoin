import type { Metadata } from "next";
import { Inter, JetBrains_Mono } from "next/font/google";
import "./globals.css";
import Header from "@/components/layout/Header";
import Sidebar from "@/components/layout/Sidebar";
import MobileNav from "@/components/layout/MobileNav";

const inter = Inter({
  subsets: ["latin"],
  variable: "--font-inter",
});

const jetbrainsMono = JetBrains_Mono({
  subsets: ["latin"],
  variable: "--font-jetbrains",
});

export const metadata: Metadata = {
  title: "ComputeCoin - Compute Power Exchange",
  description:
    "The world's first compute power exchange - trade GPU compute like financial assets",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en" className={`dark ${inter.variable} ${jetbrainsMono.variable}`}>
      <body className="font-sans bg-background text-foreground">
        <Header />
        <Sidebar />
        <main className="pt-14 md:pl-56 pb-14 md:pb-0 min-h-screen">
          {children}
        </main>
        <MobileNav />
      </body>
    </html>
  );
}
