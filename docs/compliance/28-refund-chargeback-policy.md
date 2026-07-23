# นโยบายการคืนเงินและการโต้แย้งรายการ (Chargeback) (ไทย)

> เอกสารประกอบการยื่นขอใบอนุญาต **การให้บริการรับชำระเงินด้วยวิธีการทางอิเล็กทรอนิกส์ (Full Acquiring)**
> ภายใต้ พ.ร.บ. ระบบการชำระเงิน พ.ศ. 2560 ต่อธนาคารแห่งประเทศไทย (ธปท.) — ทุนจดทะเบียนชำระแล้วเป้าหมาย 50 ล้านบาท
>
> เอกสารเลขที่: `COMP-28` · เวอร์ชัน 1.0 · เจ้าของเอกสาร: Head of Risk & Dispute Operations (ร่วมกับ Head of Settlement และ Compliance)
> เอกสารอ้างอิง: `COMPLIANCE-TH.md`, `ARCHITECTURE.md`, `ROADMAP.md`, `05-aml-kyc-cdd-policy.md`, `07-sar-str-procedure.md`, `09-pdpa-privacy-policy.md`, `11-data-retention-deletion.md`, `19-pci-dss-roadmap.md`
>
> **หมายเหตุ:** เอกสารนี้เป็นเอกสารเชิงนโยบายภายใน ไม่ใช่คำแนะนำทางกฎหมาย กรอบเวลาและสิทธิการโต้แย้งรายการ (chargeback rights) ที่มีผลผูกพันจริงกำหนดโดยกฎของ card scheme (Visa Core Rules, Mastercard Chargeback Guide) และสัญญากับธนาคารผู้รับเชื่อม (sponsor bank) ซึ่งอาจปรับปรุงเป็นระยะ ต้องยืนยันกับที่ปรึกษากฎหมายและ sponsor bank ก่อนยื่นและก่อนบังคับใช้

---

> ### ⚠️ ข้อสมมติและสิ่งที่ยังต้องยืนยัน (Assumptions / TODO)
> รายการต่อไปนี้ยังขึ้นกับคู่สัญญา/ผู้ให้บริการภายนอกที่ยังไม่สรุป — **ห้ามถือเป็นข้อเท็จจริงจนกว่าจะยืนยัน**:
> - **[TODO — Sponsor Bank / Acquirer]** ยังไม่ลงนามธนาคารผู้รับเชื่อม — ผู้กำหนด (ก) กรอบเวลา representment/pre-arbitration ที่บังคับใช้จริง (ข) อัตราค่าปรับ chargeback fee ต่อรายการ (ค) เกณฑ์ excessive chargeback program และ (ง) ข้อกำหนดเงินสำรอง/หลักประกัน (reserve) ที่ acquirer เรียกจากบริษัท
> - **[TODO — Card scheme rules version]** เวอร์ชันปัจจุบันของ Visa Core Rules & Visa Product and Service Rules, Mastercard Chargeback Guide/Security Rules and Procedures รวมทั้ง reason code set ต้องอ้างอิงฉบับล่าสุด ณ วันยื่น
> - **[TODO — Local scheme]** เงื่อนไข dispute สำหรับบัตรในประเทศ/local rails (เช่น NITMX/ITMX, บัตร PromptCard) และ PromptPay/QR (ThaiQR) ต้องยืนยันแยก เพราะกลไก chargeback ต่างจากบัตรระหว่างประเทศ
> - **[TODO — ทุนจดทะเบียน/เงินสำรอง]** โครงสร้าง merchant reserve (rolling reserve %, upfront reserve, capped reserve) และแหล่งเงินทุนรองรับความเสียหาย chargeback ต้องสอดคล้องกับ `02-financial-projections-capital.md` และเกณฑ์คงทุน ≥ 75%
> - **[TODO — ผู้ให้บริการ 3DS/ACS]** ผู้ให้บริการ EMV 3-D Secure (3DS Server/DS/ACS routing) กำหนด liability shift ที่ได้จริง
> - **[TODO — ชื่อบริษัท/ผู้บริหาร]** ชื่อ `[บริษัท / Company]`, ที่อยู่จดทะเบียน, ชื่อผู้รับผิดชอบ (Head of Risk & Dispute Ops) ต้องเติมค่าจริงก่อนยื่น

---

## 1. บทนำ วัตถุประสงค์ และขอบเขต

`[บริษัท / Company]` (ต่อไปนี้เรียก "บริษัท") ประกอบธุรกิจให้บริการรับชำระเงินด้วยบัตร (Full Acquiring Payment Gateway) จึงมีหน้าที่บริหารจัดการ **การคืนเงิน (refund)**, **การกลับรายการ (reversal/void)** และ **การโต้แย้งรายการ (dispute/chargeback)** ให้เป็นไปตามกฎของ card scheme, สัญญากับ sponsor bank และกฎหมายไทย

**วัตถุประสงค์ของนโยบายนี้:**
1. กำหนดกรอบเวลา (timeframes) ที่ชัดเจนสำหรับ refund, void, chargeback แต่ละขั้น
2. กำหนดกระบวนการกลับรายการ (reversal procedures) และการเดินสถานะใน state machine ของระบบ
3. กำหนดการจัดสรรความรับผิด (liability allocation) ระหว่างผู้ถือบัตร ธนาคารผู้ออกบัตร ร้านค้า บริษัท และ sponsor bank
4. กำหนดกลไกเงินสำรอง (reserves) และหลักประกันเพื่อรองรับความเสี่ยง chargeback/insolvency ของร้านค้า

**ขอบเขต:** ครอบคลุมธุรกรรมบัตรเครดิต/เดบิต (Visa, Mastercard และแบรนด์ที่รองรับ), การชำระผ่าน 3DS และธุรกรรม card-not-present (CNP)/card-present (CP) รวมถึงการเชื่อมโยงกับ dispute ของ local rails (QR/PromptPay) เท่าที่กลไกรองรับ (ดู `QR-PAYMENT.md`)

### 1.1 นิยามศัพท์
| คำ | ความหมาย |
|----|----------|
| **Refund (คืนเงิน)** | ร้านค้าคืนเงินให้ผู้ถือบัตรโดยสมัครใจ หลังจาก transaction ถูก capture แล้ว |
| **Void / Reversal (ยกเลิก/กลับรายการ)** | ยกเลิก authorization ก่อน capture/settlement — เงินไม่เคยถูกหักจริง |
| **Chargeback (การโต้แย้งรายการ)** | ผู้ถือบัตรโต้แย้งรายการผ่านธนาคารผู้ออกบัตร (issuer) → issuer ดึงเงินคืนจาก acquirer |
| **Representment (การนำรายการเสนอกลับ)** | acquirer/ร้านค้าโต้กลับ chargeback พร้อมหลักฐานว่ารายการถูกต้อง |
| **Pre-arbitration / Arbitration** | ขั้นตอนยกระดับข้อพิพาทเมื่อ representment ถูกโต้กลับ จนถึงให้ card scheme ตัดสิน |
| **Liability shift** | การโอนความรับผิดจาก acquirer/ร้านค้าไปยัง issuer เมื่อผ่าน EMV 3DS authentication สำเร็จ |
| **Reserve (เงินสำรอง)** | เงินที่บริษัทกันไว้จาก settlement ของร้านค้าเพื่อรองรับ chargeback/refund ในอนาคต |
| **Excessive chargeback merchant** | ร้านค้าที่มีอัตรา chargeback เกินเกณฑ์ของ card scheme (VDMP/VFMP, Mastercard ECM/ECP) |

---

## 2. การคืนเงิน (Refund Policy)

### 2.1 เงื่อนไขและหลักการ
- Refund ทำได้เฉพาะกับรายการที่มีสถานะ **`captured`** หรือ **`partial_refunded`** เท่านั้น (ดู state machine ใน `ARCHITECTURE.md` §5.4)
- ยอด refund สะสม **ต้องไม่เกิน** ยอดที่ capture (`Σ refunds ≤ captured_amount`) — บังคับด้วย invariant ในชั้น service (ARCHITECTURE §5.2)
- Refund เป็น **operation ที่ขยับเงิน** จึงต้องแนบ `Idempotency-Key` ทุกครั้ง เพื่อกันคืนเงินซ้ำจาก retry (ARCHITECTURE §2 หลักการ Idempotency)
- Refund คืนเข้าบัตรใบเดิมที่ใช้ชำระเสมอ **ห้าม** คืนเป็นเงินสด/โอนเข้าบัตรใบอื่น (กันการฟอกเงิน — ดู `05-aml-kyc-cdd-policy.md`)
- ทุก refund เขียน `ledger_entries(refund)` (append-only) และ `audit_log`

### 2.2 กรอบเวลา (Refund Timeframes)
| ขั้นตอน | กรอบเวลาเป้าหมาย (SLA) | หมายเหตุ |
|---------|----------------------|----------|
| ร้านค้ายื่นคำขอ refund → บริษัทส่งต่อ acquirer | ภายใน **1 ชั่วโมง** (อัตโนมัติ) / ภายในวันทำการเดียวกันหากต้องรีวิว | ผ่าน `POST /v1/payments/{id}/refund` |
| Refund ปรากฏใน settlement ของ acquirer | ภายใน **1–2 วันทำการ** | ขึ้นกับ cut-off ของ acquirer |
| เงินคืนเข้าบัญชี/วงเงินบัตรผู้ถือบัตร | โดยทั่วไป **5–15 วันทำการ** | issuer เป็นผู้กำหนดรอบ statement — บริษัทควบคุมไม่ได้ ต้องแจ้งผู้ถือบัตรตามจริง |
| การแจ้ง webhook `payment.refunded` ให้ร้านค้า | ภายใน **นาที** หลังยืนยัน (at-least-once + retry) | ตาม ARCHITECTURE §4 Webhook/Notifier |

> **หมายเหตุการคุ้มครองผู้บริโภค:** บริษัทต้องเปิดเผยกรอบเวลาคืนเงินโดยประมาณและช่องทางร้องเรียนต่อผู้ถือบัตร/ร้านค้าอย่างชัดเจน สอดคล้องกับแนวทางคุ้มครองผู้ใช้บริการทางการเงินของ ธปท.

### 2.3 Refund vs. Void — เลือกใช้เมื่อใด
| เงื่อนไข | ใช้ |
|----------|-----|
| ยังไม่ capture (ยังเป็น `authorized`) | **Void** — เร็วกว่า ไม่มีค่าธรรมเนียม ไม่กระทบ statement ผู้ถือบัตร |
| capture แล้ว / settle แล้ว | **Refund** |

**หลักการ:** หากยกเลิกก่อน settlement ให้ใช้ void เสมอ เพื่อลดต้นทุนและลดความสับสนของผู้ถือบัตร

---

## 3. การกลับรายการ (Reversal / Void Procedures)

### 3.1 Authorization Reversal (Void)
- ใช้ได้เฉพาะสถานะ **`authorized`** (ก่อน capture) เท่านั้น → ผลลัพธ์สถานะ **`voided`**
- Endpoint: `POST /v1/payments/{id}/void` → ส่ง reversal message ไป acquirer → เขียน `ledger_entries(void)`
- Full auth reversal ต้องส่งภายในกรอบเวลาที่ scheme กำหนด เพื่อคืนวงเงิน (open-to-buy) ของผู้ถือบัตรทันที **[TODO — ยืนยันกรอบเวลากับ sponsor bank/scheme]**

### 3.2 หลักการ "Fail closed" กับ reversal
ตาม ARCHITECTURE §2 หลักการ 7 — เมื่อไม่แน่ใจสถานะธุรกรรม (เช่น timeout จาก acquirer) ระบบต้อง **ถือว่ายังไม่สำเร็จ** และเข้ากระบวนการ reconciliation:
- หาก authorize timeout → บันทึกเป็น pending → reconcile กับ acquirer → หากพบว่า auth สำเร็จแต่ระบบไม่ได้รับ response → ส่ง **auto-reversal**
- ป้องกันการหักเงินซ้ำ/ค้างวงเงินผู้ถือบัตร (dangling authorization)

### 3.3 State machine ที่เกี่ยวข้อง
```
requires_action ──3DS ok──▶ authorized ──capture──▶ captured ──refund──▶ partial_refunded ──▶ refunded
       │                        │                        │
       └── fail ──▶ failed      └── void ──▶ voided       └── chargeback ──▶ disputed ──▶ (won/lost)
```
> สถานะ `disputed`, `dispute_won`, `dispute_lost` เป็นส่วนขยายของ state machine ใน ARCHITECTURE §5.4 สำหรับงาน chargeback (Phase 3: chargeback/dispute workflow)

---

## 4. การโต้แย้งรายการ (Chargeback Lifecycle)

### 4.1 วงจร chargeback (dual-message, Visa/Mastercard)
```
ผู้ถือบัตรโต้แย้ง → Issuer เปิด chargeback (reason code)
        ▼
Acquirer/บริษัท รับแจ้ง → แจ้งร้านค้า (webhook: dispute.opened) → เก็บหลักฐาน
        ▼
[ทางเลือก] Representment (โต้กลับพร้อมหลักฐาน)  ──── หรือ ──── ยอมรับ (accept liability)
        ▼
Issuer พิจารณา → หากไม่ยอมรับ → Pre-arbitration
        ▼
Arbitration (card scheme ตัดสิน) → ผลผูกพัน + ค่าธรรมเนียม arbitration
```

### 4.2 กรอบเวลา chargeback (Chargeback Timeframes)
> ค่าต่อไปนี้เป็น **ค่าอ้างอิงตามแนวปฏิบัติทั่วไปของ scheme** — **[TODO — ยืนยันตัวเลขจริงกับ sponsor bank ตาม Visa Core Rules / Mastercard Chargeback Guide ฉบับล่าสุด]**

| ขั้นตอน | ผู้กระทำ | กรอบเวลา (อ้างอิง) | SLA ภายในของบริษัท |
|---------|---------|-------------------|---------------------|
| Issuer เปิด chargeback | Issuer | โดยทั่วไปภายใน **120 วัน** นับจากวันทำรายการ/วันคาดว่าจะได้รับสินค้า (ต่างตาม reason code) | รับผ่าน acquirer แล้วบันทึกทันที |
| บริษัทแจ้งร้านค้า + เปิดเคส | บริษัท | — | ภายใน **1 วันทำการ** หลังรับแจ้ง (webhook `dispute.opened`) |
| ร้านค้าส่งหลักฐานให้บริษัท | Merchant | — | ภายใน **7 วันปฏิทิน** (กันชนก่อน scheme deadline) |
| บริษัทยื่น representment | บริษัท → acquirer | โดยทั่วไปภายใน **30 วัน** นับจากวันที่ chargeback | ต้องยื่นก่อน scheme deadline อย่างน้อย **3 วันทำการ** |
| Pre-arbitration | Issuer/Acquirer | โดยทั่วไปภายใน **30 วัน** | ทีม Dispute Ops ตัดสินใจ escalate/accept |
| Arbitration (scheme ตัดสิน) | Card scheme | ตามกฎ scheme | ต้องมีเหตุผลเชิงเศรษฐกิจ (มูลค่า > ค่าธรรมเนียม arbitration) |

> **หลักการเดดไลน์:** บริษัทตั้ง **internal deadline เร็วกว่า scheme deadline เสมอ** (buffer ≥ 3 วันทำการ) เพื่อลดความเสี่ยงแพ้เคสจากการยื่นล่าช้า ระบบตั้ง reminder อัตโนมัติที่ T-7, T-3, T-1 วัน

### 4.3 ตัวอย่าง Reason Code (หมวดหลัก)
| หมวด | ตัวอย่าง (Visa / Mastercard) | สาเหตุ | หลักฐานโต้กลับที่ต้องใช้ |
|------|------------------------------|--------|--------------------------|
| **Fraud** | Visa 10.4 / MC 4837 | ธุรกรรมไม่ได้รับอนุญาต (CNP fraud) | หลักฐาน 3DS/AVS/CVV, IP/device, ประวัติลูกค้า |
| **Authorization** | Visa 11.x / MC 4808 | ไม่มี/ไม่ถูกต้องของ authorization | auth_code, log การขออนุมัติ |
| **Processing error** | Visa 12.x / MC 4834 | หักซ้ำ/ยอดผิด/สกุลเงินผิด | ledger_entries, ใบเสร็จ, การ refund ที่ทำไปแล้ว |
| **Consumer dispute** | Visa 13.x / MC 4853 | ไม่ได้รับสินค้า/สินค้าไม่ตรง/ยกเลิกแล้ว | หลักฐานส่งมอบ, T&C, นโยบายคืนเงิน, การติดต่อลูกค้า |

### 4.4 กระบวนการเก็บและยื่นหลักฐาน (Compelling Evidence)
1. ระบบสร้าง dispute case ผูกกับ `payment_id` และดึงหลักฐานอัตโนมัติจาก ledger/audit_log/3DS result
2. ร้านค้าอัปโหลดหลักฐานเพิ่มเติมผ่าน merchant portal (invoice, proof of delivery, การสื่อสารกับลูกค้า)
3. ทีม Dispute Ops รวบรวมเป็น representment package → ยื่นผ่าน acquirer
4. **PDPA:** หลักฐานอาจมีข้อมูลส่วนบุคคล — จำกัดการเข้าถึงตาม least privilege และ retention ตาม `09-pdpa-privacy-policy.md`, `11-data-retention-deletion.md`
5. **PCI-DSS:** หลักฐานห้ามมี full PAN/SAD — ใช้ `card_last4` เท่านั้น (ARCHITECTURE §6, `19-pci-dss-roadmap.md`)

---

## 5. การจัดสรรความรับผิด (Liability Allocation)

### 5.1 หลักการ liability shift ด้วย EMV 3-D Secure
ธุรกรรม CNP ที่ผ่าน **EMV 3DS 2.x authentication สำเร็จ** โดยทั่วไปจะ **โอนความรับผิด fraud chargeback ไปยัง issuer** ("liability shift") ทำให้บริษัท/ร้านค้าไม่ต้องรับผิดในเคส fraud (reason code หมวด 10.x)

| สถานการณ์ | ผู้รับผิด (fraud chargeback) |
|-----------|------------------------------|
| ผ่าน 3DS สำเร็จ (fully authenticated) | **Issuer** (liability shift) |
| Issuer ไม่รองรับ 3DS / attempted | โดยทั่วไป **Issuer** (attempted liability shift) — ยืนยันตาม scheme |
| ร้านค้าไม่ส่งเข้ากระบวนการ 3DS (CNP) | **บริษัท/ร้านค้า** |
| Chargeback เชิง non-fraud (สินค้าไม่ตรง/ไม่ได้รับ) | **ร้านค้า** (ไม่ครอบคลุมโดย liability shift) |

> **[TODO]** ขอบเขต liability shift ที่มีผลจริงขึ้นกับผู้ให้บริการ 3DS และกฎ scheme ฉบับล่าสุด — ต้องยืนยัน

### 5.2 ตารางความรับผิดรวม (Liability Matrix)
| ประเภทข้อพิพาท | ความรับผิดเบื้องต้น | ผู้รับภาระทางการเงินสุดท้าย (หากแพ้เคส) |
|----------------|---------------------|------------------------------------------|
| Fraud + ผ่าน 3DS | Issuer | Issuer |
| Fraud + ไม่มี 3DS | Acquirer → ร้านค้า | ร้านค้า (บริษัทหักจาก settlement/reserve) |
| สินค้าไม่ได้รับ/ไม่ตรง | ร้านค้า | ร้านค้า |
| Processing error ของบริษัท | บริษัท | บริษัท |
| ร้านค้าล้มละลาย/ปิดกิจการ + chargeback ค้าง | บริษัท (ต่อ scheme/acquirer) | **บริษัท** (ชดเชยจาก reserve; ส่วนขาดเป็นความเสี่ยงของบริษัท) |

> **หลักการสำคัญ:** ต่อ card scheme และ sponsor bank **บริษัทคือผู้รับผิดชอบสุดท้าย (acquirer-of-record risk)** สำหรับ chargeback ของร้านค้าในเครือข่ายตน หากร้านค้าไม่สามารถชดใช้ได้ บริษัทต้องรับภาระ — นี่คือเหตุผลหลักของการกำหนด **reserve** และการคัดกรองร้านค้า (merchant underwriting)

### 5.3 ความสัมพันธ์กับสัญญาร้านค้า (Merchant Agreement)
- สัญญาร้านค้าต้องมีข้อ **chargeback liability & indemnification** ให้บริษัทมีสิทธิ **หักกลบ (offset)** ยอด chargeback/refund/ค่าปรับจาก settlement และ reserve ของร้านค้า
- สัญญาต้องระบุ **right of set-off**, การเรียกเงินเพิ่ม (debit) หากยอดค้างเกิน reserve, และเงื่อนไข termination กรณี excessive chargeback

---

## 6. เงินสำรองและการบริหารความเสี่ยงร้านค้า (Reserves & Merchant Risk)

### 6.1 ประเภทเงินสำรอง (Reserve Types)
| ประเภท | นิยาม | ใช้เมื่อ |
|--------|-------|----------|
| **Rolling reserve** | กันเงินร้อยละหนึ่งของยอดขายไว้เป็นช่วงเวลาหมุนเวียน (เช่น 5–10% นาน 90–180 วัน) แล้วทยอยปล่อยคืน | ร้านค้าความเสี่ยงปานกลาง–สูง, ธุรกิจใหม่ |
| **Upfront / capped reserve** | วางเงินก้อนหน้าจนถึงเพดานที่กำหนด | ร้านค้าความเสี่ยงสูง/มีประวัติ chargeback |
| **Ad-hoc / event reserve** | กันเพิ่มชั่วคราวเมื่อพบสัญญาณเสี่ยง (chargeback spike, ข่าวเชิงลบ) | มาตรการเฉพาะกรณี |

> **[TODO — sponsor bank/underwriting]** อัตรา % ระยะเวลา และเพดาน reserve ที่ใช้จริงกำหนดร่วมกับ acquirer/sponsor bank และนโยบาย underwriting ของบริษัท

### 6.2 เกณฑ์กำหนด reserve ตามระดับความเสี่ยงร้านค้า (ตัวอย่าง)
| ระดับความเสี่ยง | ตัวอย่างเกณฑ์ | Rolling reserve (อ้างอิง) | ระยะเวลา |
|-----------------|--------------|---------------------------|----------|
| ต่ำ | MCC ความเสี่ยงต่ำ, chargeback < 0.3% | 0% | — |
| ปานกลาง | ธุรกิจใหม่/ปริมาณผันผวน | 5% | 90 วัน |
| สูง | MCC สูง (travel, subscription, gaming), chargeback 0.6–0.9% | 10% | 180 วัน |
| สูงมาก | เกินเกณฑ์ scheme/มีประวัติ | ตามเจรจา + upfront | ตามเจรจา |

การจัดระดับความเสี่ยงร้านค้าให้สอดคล้องกับ `08-customer-risk-categorization.md` และการทำ CDD ตาม `05-aml-kyc-cdd-policy.md`

### 6.3 เกณฑ์ excessive chargeback และการดำเนินการ
| อัตรา chargeback ต่อเดือน | การดำเนินการ |
|---------------------------|--------------|
| < 0.5% | ปกติ — เฝ้าติดตาม |
| 0.5% – 0.9% | เตือน + วางแผนลด + อาจเพิ่ม reserve |
| ≥ 0.9% หรือเข้าเกณฑ์ VDMP/Mastercard ECM | เข้าโปรแกรมแก้ไข, เพิ่ม reserve, จำกัดปริมาณ |
| เกินเกณฑ์ต่อเนื่อง | ระงับ/ยกเลิกร้านค้า (offboarding) |

> เกณฑ์อ้างอิง Visa Dispute Monitoring Program (VDMP) และ Mastercard Excessive Chargeback Program (ECM/ECP) — **[TODO — ยืนยัน threshold ฉบับล่าสุด]**

### 6.4 การบันทึกทางบัญชีและ ledger
- Reserve บันทึกเป็น liability แยกใน ledger (ไม่ปะปนกับเงินทุนบริษัท) — เป็นเงินของร้านค้าที่บริษัทถือแทน
- ทุก hold/release ของ reserve เขียน `ledger_entries` และ `audit_log`
- ต้องกระทบยอด reserve กับ settlement ในงาน reconciliation (ARCHITECTURE §4, Phase 3)

---

## 7. บทบาทหน้าที่และการกำกับ (Roles & Governance)

### 7.1 RACI (สรุป)
| บทบาท | หน้าที่หลัก |
|-------|-------------|
| **Head of Risk & Dispute Ops** | เจ้าของนโยบายนี้, ตัดสินใจ escalate/accept, กำหนด reserve |
| **Dispute Operations Analyst** | จัดการเคส chargeback รายวัน, เตรียม representment, จับเดดไลน์ |
| **Head of Settlement** | หักกลบ chargeback/refund จาก settlement, บริหาร reserve |
| **Merchant Underwriting / Risk** | จัดระดับความเสี่ยง, กำหนด reserve แรกเข้า, monitoring |
| **Compliance / AML** | เชื่อม refund/chargeback ผิดปกติกับ `07-sar-str-procedure.md` (ปปง./AMLO) |
| **Legal** | สัญญาร้านค้า, ข้อกำหนด liability, การเชื่อมกับ sponsor bank |
| **Sponsor Bank / Acquirer (external)** | ช่องทางยื่น chargeback/representment ต่อ scheme, กำหนด reserve requirement |

### 7.2 การเชื่อมโยงกับหน่วยงานกำกับไทย
- **ธปท.** — refund/chargeback/reserve เป็นส่วนของการบริหารความเสี่ยงด้านปฏิบัติการและการคุ้มครองผู้ใช้บริการ ต้องพร้อมให้ตรวจ (on-site) และรายงานเป็นงวด
- **ปปง./AMLO** — รูปแบบ refund ผิดปกติ (เช่น refund ไปบัตรที่ไม่ใช่บัตรจ่าย, refund วนซ้ำ) อาจเป็นสัญญาณฟอกเงิน → เชื่อมกับ SAR/STR (`07-sar-str-procedure.md`)
- **PDPC (PDPA)** — หลักฐาน dispute ที่มีข้อมูลส่วนบุคคลต้องจัดการตามฐานทางกฎหมายและ retention (`09-`, `11-`)

### 7.3 การเก็บบันทึกและ retention
- เก็บบันทึก dispute/refund/reserve ทั้งหมด (append-only audit trail) ตามระยะเวลาที่กฎหมายและ scheme กำหนด — สอดคล้อง `11-data-retention-deletion.md` (โดยทั่วไป ≥ ระยะที่ ธปท./ปปง. กำหนด และ ≥ วงจร chargeback)
- ห้ามเก็บ full PAN/SAD ในบันทึกเหล่านี้ (PCI-DSS v4.0 Req 3)

---

## 8. ตัวชี้วัด (KPIs) และการทบทวน
| ตัวชี้วัด | เป้าหมาย |
|----------|----------|
| Chargeback rate (ต่อร้านค้า/ทั้งพอร์ต) | < 0.5% ต่อเดือน |
| Representment win rate | ≥ 40% (ปรับตามชนิดเคส) |
| Refund SLA (ส่งต่อ acquirer) | ≥ 99% ภายในวันทำการเดียวกัน |
| เคสยื่นทันเดดไลน์ | 100% (ไม่มีเคสแพ้จากยื่นสาย) |

> ทบทวนนโยบายนี้อย่างน้อย **ปีละครั้ง** หรือเมื่อ (ก) scheme rules เปลี่ยน (ข) sponsor bank เปลี่ยนเงื่อนไข (ค) chargeback rate เกินเกณฑ์

---

> **มาตรฐานที่อ้างอิง:** PCI-DSS v4.0 (PCI SSC), EMV 3-D Secure (EMV 3DS) 2.x, Visa Core Rules/VDMP, Mastercard Chargeback Guide/ECM, ISO 8583
> **กฎหมายไทยที่อ้างอิง:** พ.ร.บ. ระบบการชำระเงิน พ.ศ. 2560 (ธปท.), PDPA พ.ศ. 2562 (PDPC), พ.ร.บ. ปปง./AMLO
> เอกสารนี้ต้องได้รับการทบทวนโดยที่ปรึกษากฎหมายและ sponsor bank ก่อนบังคับใช้และก่อนยื่น

---
---

# Refund & chargeback policy: timeframes, reversal procedures, liability allocation, reserves (English)

> Supporting document for the application for an **Electronic Payment Acquiring Service (Full Acquiring)** license
> under the Payment Systems Act B.E. 2560 (2017), submitted to the Bank of Thailand (BOT / ธปท.) — target paid-up capital THB 50 million.
>
> Document no.: `COMP-28` · Version 1.0 · Owner: Head of Risk & Dispute Operations (with Head of Settlement and Compliance)
> Related: `COMPLIANCE-TH.md`, `ARCHITECTURE.md`, `ROADMAP.md`, `05-aml-kyc-cdd-policy.md`, `07-sar-str-procedure.md`, `09-pdpa-privacy-policy.md`, `11-data-retention-deletion.md`, `19-pci-dss-roadmap.md`
>
> **Note:** This is an internal policy document, not legal advice. The binding chargeback timeframes and dispute rights are set by card scheme rules (Visa Core Rules, Mastercard Chargeback Guide) and by the sponsor-bank agreement, which are periodically updated. Confirm with legal counsel and the sponsor bank before submission and before enforcement.

---

> ### ⚠️ Assumptions / TODO (unresolved external dependencies)
> The following depend on counterparties/vendors not yet finalized — **do not treat as fact until confirmed**:
> - **[TODO — Sponsor Bank / Acquirer]** No sponsoring bank signed yet — determines (a) the actual representment/pre-arbitration timeframes, (b) per-item chargeback fees, (c) the excessive-chargeback program thresholds, and (d) the reserve/collateral requirements the acquirer imposes on the company.
> - **[TODO — Card scheme rules version]** The current versions of the Visa Core Rules & Visa Product and Service Rules, the Mastercard Chargeback Guide/Security Rules and Procedures, and the reason-code set must reference the latest editions as of the submission date.
> - **[TODO — Local scheme]** Dispute handling for domestic cards/local rails (e.g., NITMX/ITMX, PromptCard) and PromptPay/QR (ThaiQR) must be confirmed separately — their chargeback mechanics differ from international cards.
> - **[TODO — Capital / reserves]** The merchant-reserve structure (rolling reserve %, upfront reserve, capped reserve) and the capital backing chargeback losses must align with `02-financial-projections-capital.md` and the ≥ 75% capital-maintenance rule.
> - **[TODO — 3DS / ACS provider]** The EMV 3-D Secure provider (3DS Server/DS/ACS routing) determines the liability shift actually achieved.
> - **[TODO — Company/officer names]** `[บริษัท / Company]`, registered address, and the responsible officer (Head of Risk & Dispute Ops) must be filled in with real values before submission.

---

## 1. Introduction, purpose, and scope

`[บริษัท / Company]` (the "Company") operates a card-acquiring payment gateway (Full Acquiring) and is therefore responsible for managing **refunds**, **reversals/voids**, and **disputes/chargebacks** in accordance with card scheme rules, the sponsor-bank agreement, and Thai law.

**Purpose of this policy:**
1. Define clear timeframes for each stage of refund, void, and chargeback.
2. Define reversal procedures and the corresponding transitions in the system state machine.
3. Define liability allocation among cardholder, issuer, merchant, the Company, and the sponsor bank.
4. Define the reserve and collateral mechanisms that absorb chargeback/merchant-insolvency risk.

**Scope:** covers credit/debit card transactions (Visa, Mastercard, and supported brands), 3DS and card-not-present (CNP)/card-present (CP) transactions, and the linkage to local-rail disputes (QR/PromptPay) where the mechanism permits (see `QR-PAYMENT.md`).

### 1.1 Definitions
| Term | Meaning |
|------|---------|
| **Refund** | Voluntary return of funds by the merchant to the cardholder after a transaction is captured. |
| **Void / Reversal** | Cancellation of an authorization before capture/settlement — funds never actually moved. |
| **Chargeback** | Cardholder disputes a transaction via the issuer, which pulls funds back from the acquirer. |
| **Representment** | The acquirer/merchant contests the chargeback with evidence that the transaction was valid. |
| **Pre-arbitration / Arbitration** | Escalation stages once a representment is contested, up to a card-scheme ruling. |
| **Liability shift** | Transfer of liability from acquirer/merchant to issuer when EMV 3DS authentication succeeds. |
| **Reserve** | Funds the Company withholds from a merchant's settlement to cover future chargebacks/refunds. |
| **Excessive chargeback merchant** | A merchant whose chargeback rate exceeds the scheme threshold (VDMP/VFMP, Mastercard ECM/ECP). |

---

## 2. Refund policy

### 2.1 Conditions and principles
- A refund is allowed only on transactions in state **`captured`** or **`partial_refunded`** (see the state machine in `ARCHITECTURE.md` §5.4).
- Cumulative refunds **must not exceed** the captured amount (`Σ refunds ≤ captured_amount`) — enforced as a service-layer invariant (ARCHITECTURE §5.2).
- A refund is a **money-moving operation** and therefore requires an `Idempotency-Key` on every call to prevent duplicate refunds on retry (ARCHITECTURE §2, Idempotency principle).
- Refunds always return to the original card used to pay — **never** cash or a different card (AML control, see `05-aml-kyc-cdd-policy.md`).
- Every refund writes a `ledger_entries(refund)` record (append-only) and an `audit_log` entry.

### 2.2 Refund timeframes
| Stage | Target SLA | Note |
|-------|-----------|------|
| Merchant requests refund → Company submits to acquirer | Within **1 hour** (automated) / same business day if review needed | Via `POST /v1/payments/{id}/refund` |
| Refund appears in the acquirer's settlement | Within **1–2 business days** | Depends on the acquirer's cut-off |
| Funds credited to the cardholder's account/limit | Typically **5–15 business days** | Set by the issuer's statement cycle — outside the Company's control; disclose realistically |
| `payment.refunded` webhook to the merchant | Within **minutes** of confirmation (at-least-once + retry) | Per ARCHITECTURE §4 Webhook/Notifier |

> **Consumer-protection note:** the Company must clearly disclose approximate refund timeframes and complaint channels to cardholders/merchants, consistent with BOT financial-consumer-protection guidance.

### 2.3 Refund vs. void — when to use which
| Condition | Use |
|-----------|-----|
| Not yet captured (still `authorized`) | **Void** — faster, no fee, no cardholder statement impact |
| Already captured / settled | **Refund** |

**Principle:** if cancelling before settlement, always void — it reduces cost and cardholder confusion.

---

## 3. Reversal / void procedures

### 3.1 Authorization reversal (void)
- Allowed only in state **`authorized`** (before capture) → resulting state **`voided`**.
- Endpoint `POST /v1/payments/{id}/void` → sends a reversal message to the acquirer → writes `ledger_entries(void)`.
- A full auth reversal must be sent within the scheme-mandated window to release the cardholder's open-to-buy immediately. **[TODO — confirm the window with sponsor bank/scheme]**

### 3.2 The "fail closed" principle for reversals
Per ARCHITECTURE §2 principle 7 — when the transaction state is uncertain (e.g., acquirer timeout) the system must **treat it as not successful** and enter reconciliation:
- If an authorize times out → record as pending → reconcile with the acquirer → if the auth actually succeeded but no response was received → send an **auto-reversal**.
- This prevents double charges and dangling authorizations that block the cardholder's limit.

### 3.3 Relevant state machine
```
requires_action ──3DS ok──▶ authorized ──capture──▶ captured ──refund──▶ partial_refunded ──▶ refunded
       │                        │                        │
       └── fail ──▶ failed      └── void ──▶ voided       └── chargeback ──▶ disputed ──▶ (won/lost)
```
> The `disputed`, `dispute_won`, and `dispute_lost` states extend the ARCHITECTURE §5.4 state machine for chargeback handling (Phase 3: chargeback/dispute workflow).

---

## 4. Chargeback lifecycle

### 4.1 Chargeback cycle (dual-message, Visa/Mastercard)
```
Cardholder disputes → Issuer opens a chargeback (reason code)
        ▼
Acquirer/Company notified → alerts merchant (webhook: dispute.opened) → collects evidence
        ▼
[optional] Representment (contest with evidence)  ──── or ──── Accept (accept liability)
        ▼
Issuer reviews → if not accepted → Pre-arbitration
        ▼
Arbitration (card scheme rules) → binding outcome + arbitration fees
```

### 4.2 Chargeback timeframes
> The values below are **indicative of general scheme practice** — **[TODO — confirm actual figures with the sponsor bank per the latest Visa Core Rules / Mastercard Chargeback Guide]**

| Stage | Actor | Timeframe (indicative) | Company internal SLA |
|-------|-------|------------------------|----------------------|
| Issuer opens chargeback | Issuer | Generally within **120 days** of the transaction/expected delivery date (varies by reason code) | Received via acquirer, recorded immediately |
| Company notifies merchant + opens case | Company | — | Within **1 business day** of receipt (webhook `dispute.opened`) |
| Merchant submits evidence to the Company | Merchant | — | Within **7 calendar days** (buffer before scheme deadline) |
| Company files representment | Company → acquirer | Generally within **30 days** of the chargeback | Must file at least **3 business days** before the scheme deadline |
| Pre-arbitration | Issuer/Acquirer | Generally within **30 days** | Dispute Ops decides escalate/accept |
| Arbitration (scheme ruling) | Card scheme | Per scheme rules | Requires economic rationale (value > arbitration fee) |

> **Deadline principle:** the Company always sets an **internal deadline earlier than the scheme deadline** (buffer ≥ 3 business days) to reduce the risk of losing on late filing. The system sets automated reminders at T-7, T-3, and T-1 days.

### 4.3 Reason-code categories (examples)
| Category | Example (Visa / Mastercard) | Cause | Rebuttal evidence needed |
|----------|-----------------------------|-------|--------------------------|
| **Fraud** | Visa 10.4 / MC 4837 | Unauthorized transaction (CNP fraud) | 3DS/AVS/CVV results, IP/device, customer history |
| **Authorization** | Visa 11.x / MC 4808 | Missing/invalid authorization | auth_code, authorization request logs |
| **Processing error** | Visa 12.x / MC 4834 | Duplicate/wrong amount/wrong currency | ledger_entries, receipt, any refund already issued |
| **Consumer dispute** | Visa 13.x / MC 4853 | Goods not received/not as described/cancelled | proof of delivery, T&Cs, refund policy, customer contact |

### 4.4 Evidence collection and filing (compelling evidence)
1. The system creates a dispute case linked to `payment_id` and auto-pulls evidence from ledger/audit_log/3DS results.
2. The merchant uploads additional evidence via the merchant portal (invoice, proof of delivery, customer communications).
3. Dispute Ops assembles a representment package → files via the acquirer.
4. **PDPA:** evidence may contain personal data — restrict access on least privilege and retain per `09-pdpa-privacy-policy.md` and `11-data-retention-deletion.md`.
5. **PCI-DSS:** evidence must never contain full PAN/SAD — use `card_last4` only (ARCHITECTURE §6, `19-pci-dss-roadmap.md`).

---

## 5. Liability allocation

### 5.1 Liability shift with EMV 3-D Secure
A CNP transaction that passes **EMV 3DS 2.x authentication successfully** generally **shifts fraud-chargeback liability to the issuer** (the "liability shift"), so the Company/merchant does not bear fraud cases (reason-code category 10.x).

| Situation | Liable party (fraud chargeback) |
|-----------|---------------------------------|
| 3DS passed (fully authenticated) | **Issuer** (liability shift) |
| Issuer does not support 3DS / attempted | Generally **Issuer** (attempted liability shift) — confirm per scheme |
| Merchant did not submit to 3DS (CNP) | **Company/merchant** |
| Non-fraud chargeback (not as described/not received) | **Merchant** (not covered by the liability shift) |

> **[TODO]** The liability shift actually achieved depends on the 3DS provider and the latest scheme rules — must be confirmed.

### 5.2 Liability matrix
| Dispute type | Initial liability | Final financial bearer (if the case is lost) |
|--------------|-------------------|----------------------------------------------|
| Fraud + 3DS passed | Issuer | Issuer |
| Fraud + no 3DS | Acquirer → merchant | Merchant (Company deducts from settlement/reserve) |
| Goods not received/not as described | Merchant | Merchant |
| Company processing error | Company | Company |
| Merchant insolvency/closure + open chargebacks | Company (to scheme/acquirer) | **Company** (offset from reserve; the shortfall is the Company's risk) |

> **Key principle:** to the card scheme and the sponsor bank, the **Company is the party of last resort (acquirer-of-record risk)** for the chargebacks of merchants in its network. If a merchant cannot make good, the Company bears the loss — this is the primary rationale for **reserves** and merchant underwriting.

### 5.3 Linkage to the merchant agreement
- The merchant agreement must contain a **chargeback-liability & indemnification** clause granting the Company the right to **offset** chargebacks/refunds/fees against the merchant's settlement and reserve.
- The agreement must specify a **right of set-off**, the ability to debit for balances exceeding reserve, and termination conditions for excessive chargebacks.

---

## 6. Reserves & merchant risk management

### 6.1 Reserve types
| Type | Definition | Used when |
|------|-----------|-----------|
| **Rolling reserve** | Withhold a % of sales for a rolling period (e.g., 5–10% for 90–180 days), then release progressively | Medium–high-risk merchants, new businesses |
| **Upfront / capped reserve** | Deposit a lump sum up to a defined cap | High-risk merchants / chargeback history |
| **Ad-hoc / event reserve** | Temporary additional hold on risk signals (chargeback spike, negative news) | Case-specific measure |

> **[TODO — sponsor bank/underwriting]** The %, duration, and cap actually used are set jointly with the acquirer/sponsor bank and the Company's underwriting policy.

### 6.2 Reserve thresholds by merchant risk tier (example)
| Risk tier | Example criteria | Rolling reserve (indicative) | Duration |
|-----------|------------------|------------------------------|----------|
| Low | Low-risk MCC, chargebacks < 0.3% | 0% | — |
| Medium | New business/volatile volume | 5% | 90 days |
| High | High-risk MCC (travel, subscription, gaming), chargebacks 0.6–0.9% | 10% | 180 days |
| Very high | Exceeds scheme thresholds/has history | Negotiated + upfront | Negotiated |

Merchant risk tiering aligns with `08-customer-risk-categorization.md` and CDD per `05-aml-kyc-cdd-policy.md`.

### 6.3 Excessive-chargeback thresholds and actions
| Monthly chargeback rate | Action |
|-------------------------|--------|
| < 0.5% | Normal — monitor |
| 0.5% – 0.9% | Warn + remediation plan + possibly increase reserve |
| ≥ 0.9% or enters VDMP/Mastercard ECM | Enter remediation program, increase reserve, cap volume |
| Sustained breach | Suspend/offboard the merchant |

> Thresholds reference the Visa Dispute Monitoring Program (VDMP) and Mastercard Excessive Chargeback Program (ECM/ECP) — **[TODO — confirm the latest thresholds]**

### 6.4 Accounting and ledger treatment
- Reserves are recorded as a segregated liability in the ledger (not commingled with Company capital) — they are merchant funds held on the merchant's behalf.
- Every reserve hold/release writes `ledger_entries` and `audit_log`.
- Reserves must be reconciled against settlement during reconciliation (ARCHITECTURE §4, Phase 3).

---

## 7. Roles & governance

### 7.1 RACI (summary)
| Role | Primary responsibility |
|------|------------------------|
| **Head of Risk & Dispute Ops** | Owns this policy; decides escalate/accept; sets reserves |
| **Dispute Operations Analyst** | Handles daily chargeback cases, prepares representments, tracks deadlines |
| **Head of Settlement** | Offsets chargebacks/refunds from settlement; manages reserves |
| **Merchant Underwriting / Risk** | Risk tiering, initial reserve, monitoring |
| **Compliance / AML** | Links abnormal refunds/chargebacks to `07-sar-str-procedure.md` (AMLO) |
| **Legal** | Merchant agreement, liability terms, sponsor-bank linkage |
| **Sponsor Bank / Acquirer (external)** | Channel to file chargebacks/representments to the scheme; sets reserve requirements |

### 7.2 Linkage to Thai regulators
- **BOT (ธปท.)** — refunds/chargebacks/reserves are part of operational-risk management and consumer protection; must be ready for on-site inspection and periodic reporting.
- **AMLO (ปปง.)** — abnormal refund patterns (e.g., refunds to a card other than the paying card, cyclical refunds) can be laundering signals → link to SAR/STR (`07-sar-str-procedure.md`).
- **PDPC (PDPA)** — dispute evidence containing personal data must be handled on a lawful basis and retained per `09-` and `11-`.

### 7.3 Record-keeping and retention
- Retain all dispute/refund/reserve records (append-only audit trail) for the periods required by law and scheme rules — consistent with `11-data-retention-deletion.md` (generally ≥ the periods set by BOT/AMLO and ≥ the chargeback lifecycle).
- Never store full PAN/SAD in these records (PCI-DSS v4.0 Req 3).

---

## 8. KPIs and review
| Metric | Target |
|--------|--------|
| Chargeback rate (per merchant / portfolio) | < 0.5% per month |
| Representment win rate | ≥ 40% (adjusted by case type) |
| Refund SLA (submitted to acquirer) | ≥ 99% same business day |
| Cases filed before deadline | 100% (no losses from late filing) |

> Review this policy at least **annually**, or when (a) scheme rules change, (b) the sponsor bank changes terms, or (c) the chargeback rate breaches thresholds.

---

> **Standards referenced:** PCI-DSS v4.0 (PCI SSC), EMV 3-D Secure (EMV 3DS) 2.x, Visa Core Rules/VDMP, Mastercard Chargeback Guide/ECM, ISO 8583.
> **Thai regulations referenced:** Payment Systems Act B.E. 2560 (BOT/ธปท.), PDPA B.E. 2562 (PDPC), AMLO.
> This document must be reviewed by legal counsel and the sponsor bank before enforcement and before submission.
