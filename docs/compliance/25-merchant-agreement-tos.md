# สัญญาและข้อกำหนดการให้บริการร้านค้า (ไทย)

> เอกสารประกอบการยื่นขอใบอนุญาต **การให้บริการรับชำระเงินด้วยวิธีการทางอิเล็กทรอนิกส์ (Acquiring)**
> ต่อธนาคารแห่งประเทศไทย (ธปท.) ภายใต้ พ.ร.บ. ระบบการชำระเงิน พ.ศ. 2560 · ทุนจดทะเบียนชำระแล้ว 50 ล้านบาท
> **เอกสารนี้เป็นแบบร่างเชิงเทคนิค/ธุรกิจ ไม่ใช่คำแนะนำทางกฎหมาย** — ต้องให้ที่ปรึกษากฎหมายตรวจทานก่อนใช้จริง
>
> เวอร์ชัน 0.1 · เจ้าของเอกสาร: Compliance & Legal · ทบทวนล่าสุด: 2026-07-22 · รอบทบทวน: ทุก 12 เดือน หรือเมื่อเกณฑ์ ธปท./PCI-DSS เปลี่ยน

---

> [!IMPORTANT]
> **สมมติฐาน / TODO ที่ต้องปิดก่อนใช้จริง (External Dependencies ที่ยังไม่ระบุแน่นอน)**
>
> รายการต่อไปนี้ยังไม่มีข้อมูลจริง — ระบุไว้เป็นสมมติฐานอย่างชัดเจน ห้ามนำตัวเลข/ชื่อสมมติไปใช้ในสัญญาที่มีผลผูกพัน
> จนกว่าจะยืนยัน:
>
> - **[TODO] ชื่อและเลขทะเบียนนิติบุคคล** — ใช้ตัวยึด `[บริษัท / Company]` ตลอดเอกสาร; ต้องแทนด้วยชื่อจริง เลขทะเบียน และที่อยู่จดทะเบียน
> - **[TODO] Sponsor bank / Acquiring partner** — ยังไม่ลงนาม; เงื่อนไข scheme (Visa/Mastercard), settlement cut-off, ค่าธรรมเนียม interchange และ reserve จะปรับตามสัญญากับธนาคารผู้รับเชื่อม
> - **[TODO] ผู้ประเมิน QSA (Qualified Security Assessor)** สำหรับ PCI-DSS v4.0 Level 1 — ยังไม่เลือก; วันออก RoC/AoC ยังไม่กำหนด
> - **[TODO] เลขที่ใบอนุญาต ธปท.** — จะเติมเมื่อได้รับใบอนุญาตจากรัฐมนตรีว่าการกระทรวงการคลังตามคำแนะนำของ ธปท.
> - **[TODO] อัตราค่าธรรมเนียมจริง (MDR/ค่าธรรมเนียมรายรายการ/ค่าธรรมเนียม chargeback)** — ตัวเลขในตารางที่ 5 เป็น *ช่วงตัวอย่างเชิงพาณิชย์* ต้องแทนด้วยตารางราคาที่ตกลงกับร้านค้าและสอดคล้องกับต้นทุน scheme จริง
> - **[TODO] วัน go-live และวงเงินเริ่มต้น (processing limit)** — ขึ้นกับผล scheme certification ตาม `ROADMAP.md` Phase 4

---

## 1. บทนิยามและคู่สัญญา (Definitions & Parties)

**คู่สัญญา:** สัญญานี้ทำขึ้นระหว่าง **[บริษัท / Company]** ("ผู้ให้บริการ" / "Gateway") ในฐานะผู้รับใบอนุญาต Acquiring จาก ธปท. กับ **ร้านค้า** ("Merchant") ที่สมัครและได้รับอนุมัติให้ใช้บริการ

| คำ | ความหมาย |
|----|----------|
| **บริการ (Services)** | การรับ-ประมวลผลการชำระเงินด้วยบัตร/ช่องทางอิเล็กทรอนิกส์, authorization, capture, refund, void, settlement, การออก report และ dashboard |
| **ข้อมูลผู้ถือบัตร (Cardholder Data / CHD)** | PAN, ชื่อผู้ถือบัตร, วันหมดอายุ, service code ตามนิยาม PCI-DSS v4.0 |
| **ข้อมูลยืนยันตัวตนที่อ่อนไหว (SAD)** | full track, CVV/CVV2/CVC2, PIN/PIN block — **ห้ามร้านค้าจัดเก็บโดยเด็ดขาด** |
| **ธุรกรรม (Transaction)** | การขอ authorize/capture/refund/void หนึ่งครั้งผ่าน API หรือช่องทางที่ Gateway กำหนด |
| **Chargeback** | การโต้แย้งรายการโดยผู้ถือบัตร/ธนาคารผู้ออกบัตรผ่าน card scheme |
| **Settlement** | การโอนยอดสุทธิ (ยอดขาย − refund − ค่าธรรมเนียม − reserve) เข้าบัญชีร้านค้า |
| **Reserve** | เงินกันสำรองเพื่อความเสี่ยง chargeback/refund |
| **สกุลเงินขั้นต่ำ (Minor units)** | จำนวนเงินเป็นจำนวนเต็มหน่วยย่อย (สตางค์) ตาม `ARCHITECTURE.md` |

---

## 2. หน้าที่ของร้านค้า (Merchant Obligations)

### 2.1 การขึ้นทะเบียนและ KYC/CDD (ตาม พ.ร.บ. ป้องกันและปราบปรามการฟอกเงิน / ปปง.-AMLO)

ร้านค้าต้องส่งและรับรองความถูกต้องของเอกสาร:

| ประเภทร้านค้า | เอกสาร CDD ขั้นต่ำ |
|---------------|--------------------|
| นิติบุคคล | หนังสือรับรอง (ไม่เกิน 90 วัน), บัญชีรายชื่อผู้ถือหุ้น (บอจ.5), บัตร ปชช. กรรมการผู้มีอำนาจ, ผู้รับผลประโยชน์ที่แท้จริง (UBO) ตั้งแต่ 25% ขึ้นไป, บัญชีธนาคารเพื่อ settlement |
| บุคคลธรรมดา | บัตรประชาชน, ทะเบียนพาณิชย์ (ถ้ามี), บัญชีธนาคาร |

- Gateway ทำ **sanction/PEP/adverse-media screening** ต่อ AMLO, UN, OFAC, EU list ก่อนอนุมัติและทำซ้ำ (rescreening) ทุก 24 ชม. หรือเมื่อ list เปลี่ยน
- ร้านค้าต้องแจ้ง **การเปลี่ยนแปลงสาระสำคัญ** (ผู้ถือหุ้น, กรรมการ, ประเภทสินค้า/บริการ, URL) ภายใน **7 วันทำการ**
- ร้านค้าต้องยินยอมให้ Gateway ทำ re-KYC เป็นระยะตามระดับความเสี่ยง (Enhanced Due Diligence สำหรับ high-risk MCC)

### 2.2 ประเภทธุรกิจต้องห้าม (Prohibited Business)

ห้ามใช้บริการกับ: สินค้า/บริการผิดกฎหมาย, การพนันที่ไม่ได้รับอนุญาต, ยาเสพติด, อาวุธ, สินค้าละเมิดทรัพย์สินทางปัญญา, สื่อลามกที่ผิดกฎหมาย, แชร์ลูกโซ่/Ponzi, สินทรัพย์ดิจิทัลที่ไม่ได้รับอนุญาตจาก ก.ล.ต., และธุรกิจที่ card scheme หรือ ธปท. ห้าม

### 2.3 หน้าที่ด้านความปลอดภัยของข้อมูล

- **ห้ามจัดเก็บ SAD** (CVV, full track, PIN) หลัง authorization ทุกกรณี
- ต้องใช้ **tokenization/hosted fields ของ Gateway** เพื่อให้ PAN ไม่ผ่านระบบร้านค้า (ลด PCI scope ของร้านค้าให้เหลือ SAQ-A ที่เหมาะสม)
- รักษาความลับ **API key/secret**; ต้อง rotate เมื่อสงสัยว่ารั่วไหลและแจ้ง Gateway ภายใน **24 ชม.**
- ปฏิบัติตามส่วนที่เกี่ยวข้องของ **PCI-DSS v4.0** ตามช่องทางการรับชำระของตน

### 2.4 หน้าที่ด้านคุณภาพธุรกรรมและการเปิดเผยต่อผู้บริโภค

- แสดง **ชื่อร้านค้า (descriptor)**, ราคา, สกุลเงิน, นโยบายคืนเงิน/คืนสินค้า, ช่องทางติดต่อ อย่างชัดเจนก่อนชำระ
- ส่งมอบสินค้า/บริการตามที่โฆษณา; ดำเนินการ refund ตามนโยบายที่ประกาศ
- รักษา **อัตรา chargeback ไม่เกิน 0.9%** และ **fraud ไม่เกิน 0.9%** ของปริมาณรายการ (สอดคล้องเกณฑ์ Visa VDMP/VFMP และ Mastercard ECP; หากเกินอาจถูกจัดโปรแกรมติดตามและปรับค่าธรรมเนียม)

### 2.5 PDPA (พ.ร.บ. คุ้มครองข้อมูลส่วนบุคคล พ.ศ. 2562 / PDPC)

- ในการประมวลผลข้อมูลลูกค้าร่วมกัน คู่สัญญาตกลงบทบาทตามภาคผนวก DPA: โดยทั่วไป Gateway เป็น **ผู้ควบคุมข้อมูลร่วม/ผู้ประมวลผล** เฉพาะเพื่อการชำระเงินและป้องกันการฉ้อโกงตามฐาน "สัญญา" และ "ประโยชน์โดยชอบด้วยกฎหมาย/หน้าที่ตามกฎหมาย"
- ร้านค้าต้องมีฐานการประมวลผลที่ชอบด้วยกฎหมายของตนเอง และแจ้งเหตุละเมิดข้อมูลที่กระทบข้อมูลชำระเงินภายใน **72 ชม.**

---

## 3. หน้าที่และความรับผิดของ Gateway (Gateway Obligations & Liability)

### 3.1 หน้าที่ของ Gateway
- ให้บริการ authorize/capture/refund/void, settlement, dashboard และ webhook แบบ at-least-once ตาม `ARCHITECTURE.md`
- รักษา **Availability ≥ 99.95%** ของ payment core; auth latency p99 < 800 ms
- ดำรง **PCI-DSS v4.0 Level 1** (RoC โดย QSA รายปี + ASV scan รายไตรมาส + pentest รายปี) และรายงานสถานะต่อร้านค้าเมื่อร้องขอ (AoC)
- ดำรงทุนจดทะเบียนชำระแล้วและปฏิบัติตามเงื่อนไขใบอนุญาต ธปท. (คงทุน **ไม่น้อยกว่า 75%** ตลอดการดำเนินงาน)

### 3.2 การจำกัดความรับผิด (Limitation of Liability)

| กรณี | ผู้รับผิด |
|------|----------|
| Chargeback, การฉ้อโกงจากลูกค้าร้านค้า, สินค้าไม่ส่งมอบ/ไม่ตรงปก | **ร้านค้า** (Gateway หักคืนได้จาก settlement/reserve) |
| ค่าปรับจาก scheme/ธปท. อันเกิดจากการละเมิดของร้านค้า | **ร้านค้า** |
| ระบบ Gateway ขัดข้องเกิน SLA จนเกิดความเสียหายพิสูจน์ได้โดยตรง | **Gateway** ภายในเพดานที่กำหนด |
| การละเมิดข้อมูลที่เกิดในระบบ/ความรับผิดชอบของ Gateway | **Gateway** |

- **เพดานความรับผิดรวมของ Gateway** ต่อร้านค้าในรอบ 12 เดือน จำกัดไม่เกิน **ค่าธรรมเนียมสุทธิที่ร้านค้าชำระให้ Gateway ในรอบ 3 เดือนก่อนเหตุ** เว้นแต่กรณีจงใจ/ประมาทเลินเล่ออย่างร้ายแรง หรือความรับผิดที่กฎหมายห้ามจำกัด
- ไม่รับผิดในความเสียหายทางอ้อม/ผลสืบเนื่อง (indirect/consequential) รวมถึงการสูญเสียกำไร ยกเว้นที่กฎหมายกำหนด

---

## 4. การจัดการข้อพิพาท Chargeback และการกันสำรอง (Reserve)

| ขั้นตอน | ผู้รับผิดชอบ | Timeline |
|---------|-------------|----------|
| แจ้งเตือน chargeback ผ่าน dashboard/webhook | Gateway | ภายใน 1 วันทำการหลังได้รับจาก scheme |
| ส่งหลักฐานโต้แย้ง (representment) | ร้านค้า | ภายใน **7 วันทำการ** (ต้องไม่เกินกรอบ scheme) |
| ยื่นต่อ scheme และแจ้งผล | Gateway | ตามกรอบ Visa/Mastercard |

- **Rolling reserve (สมมติฐาน):** เริ่มต้น **5–10%** ของยอดธุรกรรม กันไว้ **90–180 วัน** สำหรับร้านค้าใหม่/high-risk — [TODO] อัตราจริงกำหนดตามผลประเมินความเสี่ยงและเงื่อนไข sponsor bank
- Gateway มีสิทธิปรับ reserve, ระงับ settlement บางส่วน หรือชะลอการจ่ายเมื่อพบความเสี่ยงผิดปกติ (velocity, fraud spike) โดยแจ้งร้านค้า

---

## 5. ค่าธรรมเนียมและการชำระเงินคืน (Fees & Settlement)

> [!NOTE]
> ตัวเลขด้านล่างเป็น **ช่วงตัวอย่างเชิงพาณิชย์** เพื่อประกอบการยื่น — ต้องแทนด้วยอัตราจริงที่ตกลงกับร้านค้าและสอดคล้องกับต้นทุน interchange/scheme ของ sponsor bank ([TODO])

| รายการค่าธรรมเนียม | ฐานคิด | ช่วงตัวอย่าง |
|--------------------|--------|--------------|
| MDR (Merchant Discount Rate) บัตรในประเทศ | % ต่อรายการสำเร็จ | 1.8–2.5% |
| MDR บัตรต่างประเทศ / cross-border | % ต่อรายการ | 2.9–3.9% |
| ค่าธรรมเนียมรายรายการ (per-transaction) | ต่อ authorization | 1–3 บาท |
| Refund fee | ต่อรายการ refund | 0–5 บาท |
| Chargeback fee | ต่อ chargeback | 300–800 บาท |
| ค่าธรรมเนียม settlement/payout | ต่อรอบโอน | ตามข้อตกลง |

- **รอบ settlement (สมมติฐาน):** T+1 ถึง T+2 วันทำการหลัง capture โดยหักค่าธรรมเนียมและ reserve — [TODO] cut-off จริงตาม sponsor bank
- เงินคิดเป็น **จำนวนเต็มหน่วยย่อย (สตางค์)** และคำนวณด้วย `decimal` (ไม่ใช้ float) ตาม `ARCHITECTURE.md`
- Gateway ออกใบแจ้ง/รายงานค่าธรรมเนียมและ VAT ตามกฎหมายภาษีไทย; แจ้งการเปลี่ยนอัตราล่วงหน้าไม่น้อยกว่า **30 วัน**

---

## 6. การบอกเลิกและการระงับบริการ (Termination & Suspension)

### 6.1 การระงับทันที (Immediate Suspension)
Gateway ระงับบริการได้ทันทีเมื่อ: ต้องสงสัยฟอกเงิน/สนับสนุนการก่อการร้าย, ติด sanction list, fraud/chargeback เกินเกณฑ์อย่างมีนัยสำคัญ, ทำธุรกิจต้องห้าม, ละเมิดข้อมูลบัตร หรือมีคำสั่งจาก ธปท./ปปง./ศาล

### 6.2 การบอกเลิกโดยคู่สัญญา
- **ร้านค้า:** บอกเลิกได้โดยแจ้งล่วงหน้า **30 วัน** และชำระค่าธรรมเนียมค้างจ่ายให้ครบ
- **Gateway:** บอกเลิกโดยแจ้งล่วงหน้า **30 วัน** (หรือทันทีในกรณีตาม 6.1)

### 6.3 ผลของการเลิกสัญญา
- ระงับการรับรายการใหม่ทันที; รายการที่ authorize แล้วดำเนินให้เสร็จ
- **คืน reserve** หลังพ้นระยะความเสี่ยง chargeback (สมมติฐาน **90–180 วัน**)
- **เก็บรักษาข้อมูลธุรกรรม/audit log** ตามที่กฎหมายกำหนด: ธุรกรรม/AML **อย่างน้อย 5–10 ปี** ตาม พ.ร.บ. AMLO; audit log ตามเกณฑ์ ธปท./PCI-DSS — จากนั้นลบ/ทำให้ไม่ระบุตัวตนตาม PDPA
- ห้ามร้านค้าเก็บ CHD ที่ได้จากบริการหลังเลิกสัญญา

---

## 7. ข้อกำหนดทั่วไป (General)
- **กฎหมายที่ใช้บังคับ:** กฎหมายไทย; ข้อพิพาทขึ้นศาลไทย (หรืออนุญาโตตุลาการตามที่ตกลง)
- **การแก้ไข:** Gateway แก้ไขข้อกำหนดได้โดยแจ้งล่วงหน้าไม่น้อยกว่า 30 วัน เว้นแต่จำเป็นตามกฎหมาย/ธปท./scheme ให้มีผลทันที
- **การโอนสิทธิ:** ร้านค้าโอนสิทธิไม่ได้หากไม่ได้รับความยินยอมเป็นลายลักษณ์อักษร
- **ภาคผนวก:** (A) ตารางค่าธรรมเนียม, (B) DPA ตาม PDPA, (C) ข้อกำหนด scheme (Visa/Mastercard), (D) SLA

---

# Merchant agreement / Terms of Service: obligations, liability, compliance, termination, fees (English)

> Supporting document for the application for an **Electronic Payment Acquiring Service license**
> to the Bank of Thailand (BOT) under the Payment Systems Act B.E. 2560 · paid-up capital THB 50M.
> **This is a technical/business draft, not legal advice** — must be reviewed by qualified counsel before use.
>
> Version 0.1 · Owner: Compliance & Legal · Last review: 2026-07-22 · Review cycle: every 12 months or upon BOT/PCI-DSS change.

---

> [!IMPORTANT]
> **Assumptions / TODOs to close before execution (unresolved external dependencies)**
>
> The following items have no confirmed data — flagged explicitly as assumptions. Do NOT put placeholder figures/names
> into a binding contract until confirmed:
>
> - **[TODO] Legal entity name & registration number** — `[บริษัท / Company]` placeholder used throughout; replace with real name, registration number, registered address.
> - **[TODO] Sponsor bank / acquiring partner** — not signed; scheme terms (Visa/Mastercard), settlement cut-off, interchange fees and reserve depend on the sponsor bank contract.
> - **[TODO] QSA (Qualified Security Assessor)** for PCI-DSS v4.0 Level 1 — not selected; RoC/AoC issuance date TBD.
> - **[TODO] BOT license number** — to be filled once granted by the Minister of Finance on BOT's recommendation.
> - **[TODO] Actual fee rates (MDR / per-transaction / chargeback fees)** — figures in Table 5 are *illustrative commercial ranges*; replace with the agreed schedule aligned to real scheme costs.
> - **[TODO] Go-live date & initial processing limit** — depend on scheme certification per `ROADMAP.md` Phase 4.

---

## 1. Definitions & Parties

**Parties:** This agreement is between **[บริษัท / Company]** ("Gateway"), holder of a BOT Acquiring license, and the **Merchant** who applies and is approved to use the Services.

| Term | Meaning |
|------|---------|
| **Services** | Card/electronic payment acceptance and processing: authorization, capture, refund, void, settlement, reporting and dashboard. |
| **Cardholder Data (CHD)** | PAN, cardholder name, expiry, service code per PCI-DSS v4.0. |
| **Sensitive Authentication Data (SAD)** | Full track, CVV/CVV2/CVC2, PIN/PIN block — **Merchant must never store these.** |
| **Transaction** | One authorize/capture/refund/void request via the Gateway API or approved channel. |
| **Chargeback** | A dispute raised by the cardholder/issuer through the card scheme. |
| **Settlement** | Transfer of net proceeds (sales − refunds − fees − reserve) to the Merchant. |
| **Reserve** | Funds held back to cover chargeback/refund risk. |
| **Minor units** | Money as integer minor units (satang) per `ARCHITECTURE.md`. |

---

## 2. Merchant Obligations

### 2.1 Onboarding & KYC/CDD (Anti-Money Laundering Act / AMLO)

The Merchant must submit and certify:

| Merchant type | Minimum CDD documents |
|---------------|-----------------------|
| Juristic person | Company affidavit (≤90 days), shareholder list (BorOrJor.5), authorized director's ID, Ultimate Beneficial Owners (UBO) ≥25%, settlement bank account. |
| Individual | National ID, commercial registration (if any), bank account. |

- Gateway performs **sanction/PEP/adverse-media screening** against AMLO, UN, OFAC, EU lists before approval, re-screened every 24h or on list change.
- Merchant must report **material changes** (shareholders, directors, product type, URL) within **7 business days**.
- Merchant consents to periodic re-KYC by risk tier (Enhanced Due Diligence for high-risk MCCs).

### 2.2 Prohibited Business

No illegal goods/services, unlicensed gambling, narcotics, weapons, IP-infringing goods, illegal adult content, Ponzi/pyramid schemes, digital assets not licensed by the SEC, and any business prohibited by the card schemes or BOT.

### 2.3 Data Security Obligations

- **Never store SAD** (CVV, full track, PIN) after authorization.
- Use Gateway **tokenization/hosted fields** so PAN never traverses Merchant systems (reducing Merchant PCI scope to the appropriate SAQ-A).
- Protect **API keys/secrets**; rotate on suspected compromise and notify Gateway within **24h**.
- Comply with the applicable parts of **PCI-DSS v4.0** for the Merchant's acceptance channels.

### 2.4 Transaction Quality & Consumer Disclosure

- Clearly display **merchant descriptor**, price, currency, refund/return policy, and contact before payment.
- Deliver goods/services as advertised; process refunds per published policy.
- Keep **chargeback ratio ≤ 0.9%** and **fraud ratio ≤ 0.9%** of volume (aligned with Visa VDMP/VFMP and Mastercard ECP; breach may trigger monitoring programs and fee adjustments).

### 2.5 PDPA (Personal Data Protection Act B.E. 2562 / PDPC)

- Per the DPA annex, Gateway acts as **joint controller/processor** solely for payment and fraud-prevention purposes under the "contract" and "legitimate interest / legal obligation" bases.
- Merchant must maintain its own lawful basis and report any payment-data breach within **72h**.

---

## 3. Gateway Obligations & Liability

### 3.1 Gateway Obligations
- Provide authorize/capture/refund/void, settlement, dashboard and at-least-once webhooks per `ARCHITECTURE.md`.
- Maintain **≥ 99.95%** availability of payment core; auth latency p99 < 800 ms.
- Maintain **PCI-DSS v4.0 Level 1** (annual QSA RoC + quarterly ASV scans + annual pentest) and provide the AoC on request.
- Maintain paid-up capital and BOT license conditions (keeping capital **≥ 75%** at all times).

### 3.2 Limitation of Liability

| Case | Liable party |
|------|--------------|
| Chargebacks, fraud by Merchant's customers, non-delivery / not-as-described | **Merchant** (Gateway may recover from settlement/reserve) |
| Scheme/BOT fines arising from Merchant breach | **Merchant** |
| Gateway outage beyond SLA causing proven direct loss | **Gateway** within the cap |
| Data breach within Gateway's systems/responsibility | **Gateway** |

- **Gateway's aggregate liability** per 12 months is capped at **net fees paid by the Merchant in the 3 months before the event**, except for willful misconduct/gross negligence or liability that law forbids limiting.
- No liability for indirect/consequential damages including lost profits, save as required by law.

---

## 4. Chargeback Handling & Reserve

| Step | Owner | Timeline |
|------|-------|----------|
| Notify chargeback via dashboard/webhook | Gateway | within 1 business day of scheme receipt |
| Submit representment evidence | Merchant | within **7 business days** (within scheme window) |
| File to scheme and report outcome | Gateway | per Visa/Mastercard windows |

- **Rolling reserve (assumption):** initial **5–10%** of volume held **90–180 days** for new/high-risk merchants — [TODO] actual rate per risk assessment and sponsor-bank terms.
- Gateway may adjust reserve, partially hold settlement, or delay payout on abnormal risk (velocity, fraud spike) with notice.

---

## 5. Fees & Settlement

> [!NOTE]
> Figures below are **illustrative commercial ranges** for submission — replace with the agreed rates aligned to sponsor-bank interchange/scheme costs ([TODO]).

| Fee | Basis | Illustrative range |
|-----|-------|--------------------|
| MDR — domestic cards | % per successful txn | 1.8–2.5% |
| MDR — cross-border | % per txn | 2.9–3.9% |
| Per-transaction fee | per authorization | THB 1–3 |
| Refund fee | per refund | THB 0–5 |
| Chargeback fee | per chargeback | THB 300–800 |
| Settlement/payout fee | per payout cycle | as agreed |

- **Settlement cycle (assumption):** T+1 to T+2 business days after capture, net of fees and reserve — [TODO] real cut-off per sponsor bank.
- Money is held as **integer minor units (satang)** and computed with `decimal` (no float) per `ARCHITECTURE.md`.
- Gateway issues fee statements and VAT per Thai tax law; changes to rates notified at least **30 days** in advance.

---

## 6. Termination & Suspension

### 6.1 Immediate Suspension
Gateway may suspend immediately on: suspected money laundering/terrorist financing, sanction-list hit, material fraud/chargeback breach, prohibited business, cardholder-data breach, or order from BOT/AMLO/court.

### 6.2 Termination by Either Party
- **Merchant:** on **30 days'** notice, settling all outstanding fees.
- **Gateway:** on **30 days'** notice (or immediately under 6.1).

### 6.3 Effects of Termination
- Stop accepting new transactions immediately; already-authorized transactions complete.
- **Release reserve** after the chargeback risk window (assumption **90–180 days**).
- **Retain transaction/audit records** as required by law: transaction/AML records **at least 5–10 years** per the AMLO Act; audit logs per BOT/PCI-DSS — then delete/anonymize per PDPA.
- Merchant must not retain CHD obtained via the Services after termination.

---

## 7. General
- **Governing law:** Thai law; disputes before Thai courts (or arbitration if agreed).
- **Amendments:** Gateway may amend on ≥30 days' notice, except where law/BOT/scheme requires immediate effect.
- **Assignment:** Merchant may not assign without prior written consent.
- **Annexes:** (A) Fee schedule, (B) DPA per PDPA, (C) Scheme rules (Visa/Mastercard), (D) SLA.
