// Package promptpay builds Thai QR Payment (PromptPay) payloads.
//
// The payload conforms to the EMVCo QR Code Specification for Payment Systems —
// Merchant Presented Mode (MPM), which the Bank of Thailand / ITMX Thai QR
// standard is based on. Data is encoded as TLV: 2-digit ID, 2-digit length,
// value; the whole string is terminated by a CRC-16/CCITT-FALSE checksum in
// tag 63.
package promptpay

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

// Target identifies the PromptPay proxy the funds go to.
type Target struct {
	MobileNo   string // e.g. "0812345678" (mutually exclusive with the others)
	NationalID string // 13-digit citizen / tax id
	EWalletID  string // 15-digit e-wallet id
}

// TLV encodes one EMVCo tag.
func tlv(id, value string) string {
	return fmt.Sprintf("%s%02d%s", id, len(value), value)
}

// normalizeMobile converts a Thai mobile number to the PromptPay AID format:
// "0066" + number without the leading zero, zero-padded to 13 digits.
func normalizeMobile(m string) string {
	m = strings.TrimSpace(m)
	m = strings.ReplaceAll(m, "-", "")
	if strings.HasPrefix(m, "0") {
		m = m[1:]
	}
	return fmt.Sprintf("%013s", "66"+m) // AID prefix handled by caller
}

// Build returns the full PromptPay MPM payload string (encode this into a QR).
//
// dynamic=true embeds the amount and marks the QR one-time (point of initiation
// "12"); dynamic=false produces a reusable static QR ("11") with no amount.
// ref1 is an optional merchant reference stored in tag 62 (Additional Data).
func Build(t Target, amount decimal.Decimal, dynamic bool, ref1 string) (string, error) {
	var acct string
	switch {
	case t.MobileNo != "":
		acct = tlv("00", "A000000677010111") + tlv("01", normalizeMobile(t.MobileNo))
	case t.NationalID != "":
		acct = tlv("00", "A000000677010111") + tlv("02", t.NationalID)
	case t.EWalletID != "":
		acct = tlv("00", "A000000677010111") + tlv("03", t.EWalletID)
	default:
		return "", fmt.Errorf("promptpay: no target proxy provided")
	}

	poi := "11" // static
	if dynamic {
		poi = "12" // dynamic (one-time)
	}

	var b strings.Builder
	b.WriteString(tlv("00", "01"))            // payload format indicator
	b.WriteString(tlv("01", poi))             // point of initiation method
	b.WriteString(tlv("29", acct))            // merchant account info (PromptPay)
	b.WriteString(tlv("53", "764"))           // currency THB (ISO 4217 numeric)
	if dynamic {
		b.WriteString(tlv("54", amount.StringFixed(2))) // transaction amount
	}
	b.WriteString(tlv("58", "TH")) // country code
	if ref1 != "" {
		b.WriteString(tlv("62", tlv("01", ref1))) // additional data -> bill/ref
	}

	payload := b.String() + "6304"          // tag 63 id+len, value filled next
	return payload + crc16(payload), nil
}

// crc16 computes CRC-16/CCITT-FALSE (poly 0x1021, init 0xFFFF) and returns it as
// a 4-char uppercase hex string, per the EMVCo spec for tag 63.
func crc16(s string) string {
	crc := 0xFFFF
	for _, c := range []byte(s) {
		crc ^= int(c) << 8
		for i := 0; i < 8; i++ {
			if crc&0x8000 != 0 {
				crc = (crc << 1) ^ 0x1021
			} else {
				crc <<= 1
			}
			crc &= 0xFFFF
		}
	}
	return fmt.Sprintf("%04X", crc)
}
