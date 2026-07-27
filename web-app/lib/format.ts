// formatMoney renders integer minor units (สตางค์) as a THB amount string.
export function formatMoney(amountMinor: number, currency = "THB"): string {
  const major = amountMinor / 100;
  return new Intl.NumberFormat("th-TH", { style: "currency", currency }).format(major);
}

// formatDecimalMoney renders a MAJOR-unit decimal (baht) as a THB amount string.
// domain.Payment.{Amount,CapturedAmount,RefundedAmount} are shopspring decimals
// serialized as quoted JSON strings (e.g. "100.00") — NOT minor units. Use this
// for payment amounts; use formatMoney for minor-unit values (stats volume,
// payment_links.amount_minor).
export function formatDecimalMoney(major: string | number, currency = "THB"): string {
  const n = typeof major === "number" ? major : parseFloat(major);
  const safe = Number.isFinite(n) ? n : 0;
  return new Intl.NumberFormat("th-TH", { style: "currency", currency }).format(safe);
}
