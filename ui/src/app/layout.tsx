import type { Metadata } from 'next'
import './globals.css'

export const metadata: Metadata = {
  title: 'SearchJet Admin',
  description: 'Bilingual search engine admin for the UAE',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  )
}
