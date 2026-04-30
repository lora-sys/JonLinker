import type { Metadata } from 'next';
import { Inter, Plus_Jakarta_Sans, Geist } from 'next/font/google';
import './globals.css';
import { cn } from "@/lib/utils";

const geist = Geist({subsets:['latin'],variable:'--font-sans'});

const inter = Inter({
  subsets: ['latin'],
  variable: '--font-inter',
});

const plusJakartaSans = Plus_Jakarta_Sans({
  subsets: ['latin'],
  variable: '--font-jakarta',
});

export const metadata: Metadata = {
  title: 'JobLinker - A2A Agent Recruitment Platform',
  description: 'Intelligent agent-to-agent recruitment platform connecting seekers and recruiters',
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en" className={cn(inter.variable, plusJakartaSans.variable, "font-sans", geist.variable)} suppressHydrationWarning={true}>
      <body className="antialiased font-sans">
        {children}
      </body>
    </html>
  );
}
