import Script from "next/script";
import CheckoutClient from "@/components/CheckoutClient";

// Public hosted checkout. No auth / cookies. The vanilla qrcode.min.js asset
// (copied into /public in Task 8) exposes a global `QRCode`; loaded here once.
export default function PayPage({ params }: { params: { publicId: string } }) {
  return (
    <main className="min-h-screen bg-paycore-bg text-paycore-text flex items-start justify-center p-4">
      <Script src="/qrcode.min.js" strategy="afterInteractive" />
      <CheckoutClient publicId={params.publicId} />
    </main>
  );
}
