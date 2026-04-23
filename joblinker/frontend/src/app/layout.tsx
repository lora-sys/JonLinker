import type { Metadata } from 'next';
import './globals.css';

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
    <html lang="en">
      <body className="antialiased">
        {children}
      </body>
    </html>
  );
}
