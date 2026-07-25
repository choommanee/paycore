// formatMoney renders integer minor units (สตางค์) as a THB amount string.
export function formatMoney(amountMinor: number, currency = "THB"): string {
  const major = amountMinor / 100;
  return new Intl.NumberFormat("th-TH", { style: "currency", currency }).format(major);
}
