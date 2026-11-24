import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "kilnscrub — Studio Kiln Scrubber Designer, Shop Plans & Live Controller",
  description:
    "Pure-Go single-binary kiln exhaust scrubber engineering designer, offline 3D CAD viewer, COTS shopping kit, shop plans, and live fail-safe PID controller.",
  icons: {
    icon: "/favicon.ico",
  },
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
