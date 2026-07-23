# ประมาณการทางการเงินและโครงสร้างทุน 50 ล้านบาท (ไทย)

> เอกสารประกอบคำขอรับใบอนุญาตประกอบธุรกิจ **บริการรับชำระเงินด้วยวิธีการทางอิเล็กทรอนิกส์ (Acquiring Service)**
> ภายใต้ **พ.ร.บ. ระบบการชำระเงิน พ.ศ. 2560** กำกับโดย **ธนาคารแห่งประเทศไทย (ธปท.)**
> ผู้ขอ: **[บริษัท / Company]** · เวอร์ชัน 0.1 · วันที่จัดทำ: 2026-07-22
> เอกสารที่เกี่ยวข้อง: `../COMPLIANCE-TH.md`, `../ARCHITECTURE.md`, `../ROADMAP.md`
>
> **ข้อจำกัดความรับผิด:** เอกสารนี้เป็นเอกสารประกอบเชิงเทคนิค/การเงินสำหรับการยื่นคำขอ ไม่ใช่คำแนะนำทางกฎหมายหรือการลงทุน
> ตัวเลขประมาณการต้องได้รับการรับรองจากผู้สอบบัญชีรับอนุญาต (CPA) และที่ปรึกษากฎหมายด้านใบอนุญาต ธปท. ก่อนยื่นจริง

---

## 1. บทสรุปผู้บริหาร (Executive Summary)

**[บริษัท / Company]** ยื่นขอใบอนุญาต Acquiring Service ซึ่งกำหนด **ทุนจดทะเบียนที่ชำระแล้วขั้นต่ำ 50 ล้านบาท**
โดยบริษัทจะดำรงทุนดังกล่าวเป็นทุนที่ "มีและคงไว้" (maintained capital) ไม่ใช่ค่าใช้จ่ายหมุนเวียน และจะรักษาระดับทุน
**ไม่ต่ำกว่า 75% (37.5 ล้านบาท)** ของทุนขั้นต่ำตลอดการดำเนินงานตามหลักเกณฑ์ ธปท.

หลักการทางการเงินหลัก:

1. **แยกเงินลูกค้าออกจากเงินบริษัท (client money segregation)** — เงินที่รับแทน merchant ในระหว่างรอ settlement
   จะเก็บใน **บัญชีแยกต่างหาก (segregated / trust-style account)** กับสถาบันการเงิน ไม่ปะปนกับทุนดำเนินงาน
2. **ทุนขั้นต่ำเป็นเงินสด/สินทรัพย์สภาพคล่องสูง** — ถือเป็น buffer รองรับ operational risk, chargeback, settlement timing
3. **ประมาณการ 5 ปี** สร้างบน 3 สถานการณ์ (Base / Conservative / Aggressive) พร้อมการทดสอบภาวะวิกฤต (stress test)
4. **Liquidity & solvency** ติดตามด้วยตัวชี้วัดรายเดือน (LCR-style, current ratio, capital adequacy) และรายงานต่อ ธปท. ตามงวด

---

## 2. โครงสร้างทุนจดทะเบียน 50 ล้านบาท

### 2.1 องค์ประกอบของทุนและการจัดสรร

ทุนจดทะเบียนที่ชำระแล้ว 50,000,000 บาท แบ่งการจัดสรรเชิงบริหาร (ทุนยังคงเป็นของบริษัท เพียงจัดสรรวัตถุประสงค์การใช้):

| องค์ประกอบ | จำนวน (ล้านบาท) | สัดส่วน | รูปแบบสินทรัพย์ | วัตถุประสงค์ |
|-----------|----------------|--------|----------------|-------------|
| **Regulatory capital buffer (คงไว้ ≥ 75%)** | 37.50 | 75% | เงินฝากประจำ/พันธบัตรรัฐบาล/T-bill | ทุนขั้นต่ำตามเกณฑ์ ธปท. ห้ามนำไปใช้หมุนเวียน |
| **Operational working capital** | 7.50 | 15% | เงินฝากกระแสรายวัน/ออมทรัพย์ | เงินเดือน, ค่าโครงสร้างพื้นฐาน, PCI/QSA, ค่าธรรมเนียม scheme |
| **Settlement / chargeback reserve** | 3.00 | 6% | เงินฝากสภาพคล่องสูง | รองรับ timing gap ของ settlement และ chargeback liability |
| **Contingency / stress reserve** | 2.00 | 4% | เงินฝากสภาพคล่องสูง | รองรับเหตุการณ์ไม่คาดคิด / fraud loss |
| **รวม** | **50.00** | **100%** | | |

> **หมายเหตุ:** Settlement reserve และ chargeback reserve ข้างต้นเป็น buffer ที่มาจากทุนบริษัท **แยกต่างหาก**
> จากเงินของ merchant ที่อยู่ในบัญชี segregated (ดูข้อ 5) — ไม่นับซ้ำ

### 2.2 โครงสร้างผู้ถือหุ้น (Cap Table) — โครงร่าง

| ผู้ถือหุ้น | ประเภท | สัดส่วน | หมายเหตุ |
|-----------|--------|--------|---------|
| ผู้ก่อตั้ง / กรรมการไทย | สามัญ | [TODO %] | ต้องมีกรรมการอย่างน้อย 1 คนสัญชาติไทยและมีถิ่นที่อยู่ในไทย |
| นักลงทุนเชิงกลยุทธ์ | สามัญ/บุริมสิทธิ | [TODO %] | หากต่างชาติถือข้างมากอาจต้องมี Foreign Business License (FBL) |
| ESOP pool | — | [TODO %] | ตามนโยบายบริษัท |

> 🔲 **ASSUMPTION / TODO — โครงสร้างผู้ถือหุ้นจริง:** สัดส่วนการถือหุ้นและสัญชาติผู้ถือหุ้นข้างมากยังไม่สรุป
> ต้องยืนยันก่อนยื่น เพราะกระทบเงื่อนไข FBL และ fit & proper ของกรรมการ/ผู้บริหาร

### 2.3 ที่มาของทุน (Source of Funds)

| แหล่งที่มา | จำนวน (ล้านบาท) | สถานะ | เอกสารหลักฐาน |
|-----------|----------------|-------|---------------|
| ทุนจากผู้ก่อตั้ง (paid-in) | [TODO] | [TODO] | หนังสือรับรอง, statement ธนาคาร |
| Equity round (นักลงทุน) | [TODO] | [TODO] | Share subscription agreement |
| **รวม paid-up capital** | **50.00** | ต้องชำระเต็มก่อนยื่น | ใบสำคัญรับชำระค่าหุ้น, บอจ.5 |

> 🔲 **ASSUMPTION / TODO — การระดมทุนจริง:** ยอดและกำหนดเวลาการชำระค่าหุ้นจริงยังไม่ยืนยัน ที่มาของเงินทุน
> ต้องผ่าน **AML/CDD source-of-fund verification** และรายงานตามเกณฑ์ **ปปง. (AMLO)** ก่อนรับเข้าเป็นทุน

---

## 3. สมมติฐานของประมาณการ (Key Assumptions)

| พารามิเตอร์ | สมมติฐาน Base | หมายเหตุ |
|-------------|---------------|---------|
| Merchant discount rate (MDR) เฉลี่ย | 1.85% ของ TPV | รายได้หลัก; หัก interchange + scheme fee |
| Interchange + scheme fee (ต้นทุนขาย) | ~1.20% ของ TPV | จ่ายให้ issuer/scheme (net MDR margin ~0.65%) |
| ค่าธรรมเนียม gateway ต่อรายการ | 1.50 บาท/รายการ | สำหรับ authorization + 3DS |
| Chargeback rate | 0.10% ของจำนวนรายการ | ตั้งเป้าต่ำกว่าเกณฑ์ scheme (Visa/MC 0.9%–1.0%) |
| Fraud loss (net) | 0.03% ของ TPV | หลังหัก 3DS liability shift |
| Settlement cycle | T+1 / T+2 | ตามข้อตกลง sponsor bank |
| อัตราภาษีเงินได้นิติบุคคล | 20% | ตามประมวลรัษฎากร |
| อัตราดอกเบิ้ยรับจากทุน buffer | 2.0%/ปี | เงินฝากประจำ/พันธบัตรระยะสั้น |

> 🔲 **ASSUMPTION / TODO — เศรษฐศาสตร์ต่อรายการ (unit economics):** MDR, interchange และ scheme fee จริง
> ขึ้นกับสัญญากับ **sponsor bank** และ **card scheme (Visa/Mastercard)** ที่ยังไม่สรุป ตัวเลขข้างต้นเป็นค่าอ้างอิงตลาด

### 3.1 สมมติฐานปริมาณธุรกรรม (TPV — Total Payment Volume)

| ปี | TPV (ล้านบาท) | จำนวนรายการ (ล้านรายการ) | Merchant active |
|----|---------------|--------------------------|-----------------|
| ปีที่ 1 | 1,200 | 4.0 | 300 |
| ปีที่ 2 | 4,500 | 15.0 | 1,200 |
| ปีที่ 3 | 12,000 | 40.0 | 3,500 |
| ปีที่ 4 | 24,000 | 80.0 | 7,000 |
| ปีที่ 5 | 40,000 | 130.0 | 12,000 |

> จำนวนรายการปีที่ 2 เป็นต้นไป > 6 ล้านรายการ/ปี → **ยืนยันความจำเป็นของ PCI-DSS Level 1** (สอดคล้อง `ARCHITECTURE.md` §1, §7)

---

## 4. ประมาณการทางการเงิน 5 ปี (Base Case)

### 4.1 งบกำไรขาดทุนโดยสรุป (P&L, หน่วย: ล้านบาท)

| รายการ | ปีที่ 1 | ปีที่ 2 | ปีที่ 3 | ปีที่ 4 | ปีที่ 5 |
|--------|--------|--------|--------|--------|--------|
| รายได้ MDR (net margin ~0.65% ของ TPV) | 7.80 | 29.25 | 78.00 | 156.00 | 260.00 |
| รายได้ค่าธรรมเนียมรายการ + อื่น ๆ | 6.00 | 22.50 | 60.00 | 120.00 | 195.00 |
| **รายได้รวม** | **13.80** | **51.75** | **138.00** | **276.00** | **455.00** |
| ต้นทุนขาย (processing, 3DS, scheme pass-through) | (5.20) | (18.00) | (46.00) | (88.00) | (140.00) |
| **กำไรขั้นต้น** | **8.60** | **33.75** | **92.00** | **188.00** | **315.00** |
| ค่าใช้จ่ายบุคลากร (ทีมตาม ROADMAP §3) | (18.00) | (26.00) | (38.00) | (52.00) | (66.00) |
| PCI-DSS / QSA / ASV / pentest | (4.50) | (5.00) | (5.50) | (6.00) | (6.50) |
| HSM/KMS + Infra (HA + DR) | (5.00) | (6.00) | (7.50) | (9.00) | (11.00) |
| ที่ปรึกษากฎหมาย / ใบอนุญาต / compliance | (2.50) | (1.50) | (1.50) | (1.80) | (2.00) |
| Chargeback + fraud loss | (0.50) | (1.80) | (4.80) | (9.60) | (16.00) |
| ค่าใช้จ่ายดำเนินงานอื่น | (3.00) | (4.50) | (6.50) | (9.00) | (12.00) |
| **EBITDA** | **(24.90)** | **(11.05)** | **27.70** | **90.80** | **201.50** |
| ค่าเสื่อม/ตัดจำหน่าย | (1.50) | (2.00) | (2.50) | (3.00) | (3.50) |
| **EBIT** | **(26.40)** | **(13.05)** | **25.20** | **87.80** | **198.00** |
| รายได้ดอกเบี้ยจากทุน buffer | 0.75 | 0.75 | 0.75 | 0.90 | 1.00 |
| ภาษีเงินได้ (20%, เมื่อมีกำไรและใช้ผลขาดทุนสะสมหมด) | 0.00 | 0.00 | 0.00 | (14.40) | (39.80) |
| **กำไร(ขาดทุน)สุทธิ** | **(25.65)** | **(12.30)** | **25.95** | **74.30** | **159.20** |
| ขาดทุนสะสม (ยกไป) | (25.65) | (37.95) | (12.00) | 62.30 | 221.50 |

**จุดคุ้มทุน (break-even):** คาดในปีที่ 3 (EBITDA เป็นบวก) และกำไรสุทธิสะสมเป็นบวกในปีที่ 4

### 4.2 กระแสเงินสด (Cash Flow โดยสรุป, ล้านบาท)

| รายการ | ปีที่ 1 | ปีที่ 2 | ปีที่ 3 | ปีที่ 4 | ปีที่ 5 |
|--------|--------|--------|--------|--------|--------|
| เงินสดจากดำเนินงาน | (24.90) | (11.05) | 27.70 | 90.80 | 201.50 |
| เงินลงทุน (capex: HSM, infra) | (3.00) | (2.50) | (3.00) | (3.50) | (4.00) |
| เงินสดต้นงวด | 50.00 | 22.10 | 8.55 | 33.25 | 120.55 |
| **เงินสดปลายงวด** | **22.10** | **8.55** | **33.25** | **120.55** | **318.05** |

> **จุดวิกฤต liquidity:** ปลายปีที่ 2 เงินสดเหลือ ~8.55 ล้านบาท ต่ำกว่า operational working capital ที่จัดสรร
> → **ต้องมีแผนระดมทุนรอบเพิ่ม (bridge/Series) ปลายปีที่ 1 หรือกลางปีที่ 2** ก่อนถึงจุดนี้ (ดู runway ข้อ 6.3)

> 🔲 **ASSUMPTION / TODO — Runway และการเพิ่มทุน:** ประมาณการนี้ชี้ว่าลำพังทุน 50 ล้านไม่พอครอบคลุมช่วงขาดทุนสะสม
> ก่อน break-even บริษัทต้องวางแผนเงินทุนหมุนเวียนเพิ่ม โดย **ห้ามแตะ regulatory buffer 37.5 ล้านบาท**

### 4.3 งบดุลโดยสรุป (Balance Sheet, ปลายปี, ล้านบาท)

| รายการ | ปีที่ 1 | ปีที่ 2 | ปีที่ 3 | ปีที่ 4 | ปีที่ 5 |
|--------|--------|--------|--------|--------|--------|
| สินทรัพย์สภาพคล่อง (ไม่รวมเงิน merchant) | 22.10 | 8.55 | 33.25 | 120.55 | 318.05 |
| สินทรัพย์ถาวรสุทธิ | 4.50 | 5.00 | 5.50 | 6.00 | 6.50 |
| เงิน merchant ในบัญชี segregated (off-B/S ในเชิงบริหาร) | ~3.3 | ~12.3 | ~33 | ~66 | ~110 |
| หนี้สินหมุนเวียน (settlement payable ต่อ merchant) | ~3.3 | ~12.3 | ~33 | ~66 | ~110 |
| ส่วนของผู้ถือหุ้น (ทุน + กำไรสะสม) | 24.35 | 12.05 | 38.00 | 112.30 | 271.50 |

---

## 5. การบริหารเงินลูกค้าและ Settlement (Client Money & Settlement)

### 5.1 หลักการแยกเงิน (Segregation)

- เงินที่ได้รับจาก card network เพื่อ settle ให้ merchant **ไม่ถือเป็นรายได้/สินทรัพย์ของบริษัท** เป็นเพียง fiduciary liability
- เก็บใน **บัญชีแยกต่างหาก (segregated account)** ที่สถาบันการเงินที่ ธปท. กำกับ แยกจากบัญชีดำเนินงานของบริษัท
- **ห้ามนำเงิน merchant ไปใช้เป็นทุนหมุนเวียน** หรือลงทุนใด ๆ (no commingling)
- กระทบยอด (reconciliation) เงินในบัญชี segregated กับยอด `settlement payable` ใน ledger **ทุกวันทำการ**
  โดยอ้างอิง `ledger_entries(settlement)` (double-entry, append-only) ตาม `ARCHITECTURE.md` §6

### 5.2 บทบาทและอำนาจอนุมัติ (Roles & Authority)

| บทบาท | หน้าที่ด้านการเงิน | อำนาจ |
|-------|--------------------|-------|
| CFO / Head of Finance | นโยบายเงินทุน, กำกับ solvency/liquidity, รายงาน ธปท. | อนุมัติสูงสุด, เจ้าของ capital plan |
| Treasury Manager | บริหาร buffer, บัญชี segregated, reconciliation รายวัน | ดำเนินการภายใต้ limit ที่ CFO กำหนด |
| Compliance Officer | AML/CDD source-of-fund, รายงาน ปปง., เกณฑ์ ธปท. | veto ธุรกรรมที่มีความเสี่ยง AML |
| Internal Audit | ตรวจสอบการแยกเงิน, reconciliation, capital adequacy | รายงานตรงต่อคณะกรรมการตรวจสอบ |
| External Auditor (CPA) | รับรองงบการเงินและประมาณการ | อิสระ |

### 5.3 Chargeback & Reserve

- ตั้งสำรอง chargeback แบบ dynamic ตาม risk profile ของ merchant (rolling reserve สำหรับ merchant ความเสี่ยงสูง)
- Chargeback liability เชื่อมกับ dispute/chargeback workflow (`ROADMAP.md` Phase 3) และ 3DS 2.x liability shift (EMV 3DS)
- ตั้งเป้า chargeback rate < 0.10% ต่ำกว่าเกณฑ์ card scheme (Visa VDMP / Mastercard ECM) อย่างมีนัยสำคัญ

---

## 6. สภาพคล่องและความสามารถในการชำระหนี้ (Liquidity & Solvency)

### 6.1 ตัวชี้วัดและเกณฑ์เตือน (Thresholds)

| ตัวชี้วัด | สูตร | เกณฑ์เป้าหมาย | ระดับเตือน (Amber) | ระดับวิกฤต (Red) |
|----------|------|---------------|--------------------|------------------|
| **Capital adequacy** | ทุนที่คงไว้ / ทุนขั้นต่ำ 50 ล้าน | ≥ 100% | < 90% | < 75% (ผิดเกณฑ์ ธปท.) |
| **Current ratio** | สินทรัพย์หมุนเวียน / หนี้สินหมุนเวียน | ≥ 1.5 | < 1.2 | < 1.0 |
| **Liquidity coverage (LCR-style)** | สินทรัพย์สภาพคล่องสูง / กระแสจ่ายสุทธิ 30 วัน | ≥ 1.2 | < 1.1 | < 1.0 |
| **Segregation ratio** | ยอดบัญชี segregated / settlement payable | = 100% | < 100% | < 100% (ต้องเติมทันที) |
| **Cash runway** | เงินสดดำเนินงาน / เงินเผาต่อเดือน | ≥ 12 เดือน | < 9 เดือน | < 6 เดือน |

### 6.2 การทดสอบภาวะวิกฤต (Stress Testing)

ทดสอบราย 6 เดือน อย่างน้อยตามสถานการณ์:

1. **Settlement delay** — sponsor bank ชะลอ settlement 5 วันทำการ → ทดสอบว่า buffer + segregated account รองรับได้
2. **Chargeback spike** — chargeback พุ่งเป็น 3 เท่า (0.30%) ต่อเนื่อง 3 เดือน → ทดสอบ reserve
3. **Revenue shortfall** — TPV ต่ำกว่า Base 40% → ทดสอบ runway และ capital adequacy
4. **Fraud event** — เหตุ fraud ครั้งเดียวมูลค่า 5 ล้านบาท → ทดสอบ contingency reserve
5. **Operational outage** — downtime กระทบรายได้ + ค่าปรับ SLA (อ้างอิง availability ≥ 99.95%, `ARCHITECTURE.md` §8)

### 6.3 สถานการณ์จำลอง (Scenario Analysis)

| สถานการณ์ | TPV vs Base | EBITDA break-even | เงินสดต่ำสุด | ความเห็น |
|-----------|-------------|-------------------|--------------|---------|
| **Conservative** | -40% | ปีที่ 4 | ~2–3 ล้าน (ปีที่ 2–3) | ต้องเพิ่มทุน ~15–20 ล้านก่อนปีที่ 2 |
| **Base** | — | ปีที่ 3 | ~8.55 ล้าน (ปีที่ 2) | ต้องเตรียม bridge ปลายปีที่ 1 |
| **Aggressive** | +50% | ปีที่ 2 | ~18 ล้าน (ปีที่ 2) | อาจไม่ต้องเพิ่มทุน |

> **สรุปด้านสภาพคล่อง:** ในทุกสถานการณ์ **regulatory buffer 37.5 ล้านบาทต้องคงอยู่ครบ** และไม่ถูกนำมาชดเชยการขาดทุน
> การขาดสภาพคล่องเชิงดำเนินงานต้องแก้ด้วยการเพิ่มทุน/หนี้ระยะยาว ไม่ใช่การลดทุนขั้นต่ำ

---

## 7. การกำกับ การรายงาน และการควบคุมภายใน (Governance & Reporting)

| รายงาน | ความถี่ | ผู้รับ | อ้างอิงเกณฑ์ |
|--------|--------|-------|--------------|
| การดำรงทุน (capital maintenance ≥ 75%) | รายเดือน + รายไตรมาส | ธปท. | พ.ร.บ. ระบบการชำระเงิน 2560 |
| งบการเงินที่ผู้สอบบัญชีรับรอง | รายปี | ธปท. / กรมพัฒนาธุรกิจการค้า | มาตรฐานบัญชีไทย |
| Reconciliation บัญชี segregated | รายวัน | Treasury / Internal Audit | เกณฑ์การแยกเงิน |
| รายงานธุรกรรมน่าสงสัย (STR/CTR) | ตามเหตุการณ์ | ปปง. (AMLO) | พ.ร.บ. ป้องกันและปราบปรามการฟอกเงิน |
| เหตุการณ์ด้าน IT/ไซเบอร์ | ตามเหตุการณ์ + งวด | ธปท. | ประกาศ ธปท. ด้าน IT/cyber resilience |
| การละเมิดข้อมูลส่วนบุคคล | ภายใน 72 ชม. | PDPC | PDPA พ.ศ. 2562 |

**การควบคุมภายในหลัก:** dual control และ split knowledge สำหรับการโยกย้ายเงิน buffer/reserve (สอดคล้องหลัก key management PCI Req 3),
maker-checker สำหรับ payout, และ segregation of duties ระหว่าง Treasury / Compliance / Internal Audit

---

## 8. รายการสมมติฐานและสิ่งที่ต้องยืนยัน (Assumptions & TODO — สรุป)

| # | รายการ | ผู้รับผิดชอบ | Deadline (อ้าง ROADMAP) |
|---|--------|--------------|--------------------------|
| A1 | โครงสร้างผู้ถือหุ้น + สัญชาติ (FBL?) | Legal / Founders | ก่อนยื่น ธปท. (เตรียมนิติบุคคล 45 วัน) |
| A2 | ยอด/กำหนดชำระค่าหุ้นจริง + source of funds (AML) | Finance / Compliance | ก่อนยื่น |
| A3 | สัญญา sponsor bank + MDR/interchange จริง | Product / BD | Critical path เส้นทาง B |
| A4 | QSA vendor + ขอบเขต/ราคา PCI-DSS L1 | Security / DevSecOps | Phase 4 (Q4 2026) |
| A5 | แผนเพิ่มทุน/bridge ก่อน break-even | CFO / Board | ปลายปีที่ 1 |
| A6 | ธนาคารที่เปิดบัญชี segregated + เงื่อนไข trust | Treasury / Legal | ก่อน go-live |

---

# 5-year financial projections + registered capital structure of 50M THB for Acquiring service, liquidity/solvency (English)

> Supporting document for the license application for **Acquiring Service (electronic payment acceptance)**
> under the **Payment Systems Act B.E. 2560**, supervised by the **Bank of Thailand (BOT / ธปท.)**.
> Applicant: **[บริษัท / Company]** · Version 0.1 · Date: 2026-07-22
> Related docs: `../COMPLIANCE-TH.md`, `../ARCHITECTURE.md`, `../ROADMAP.md`
>
> **Disclaimer:** This is a technical/financial supporting document, not legal or investment advice.
> All projected figures must be certified by a licensed CPA and reviewed by BOT-licensing counsel before submission.

---

## 1. Executive Summary

**[บริษัท / Company]** applies for an Acquiring Service license, which mandates a **minimum paid-up registered
capital of THB 50 million**. This capital is treated as **maintained capital** (held, not consumed), and the
company will keep it **at no less than 75% (THB 37.5M)** of the minimum threshold at all times, per BOT rules.

Core financial principles:

1. **Client money segregation** — funds received on behalf of merchants pending settlement are held in a
   **segregated / trust-style account** at a supervised financial institution, never commingled with operating capital.
2. **Minimum capital held as cash / highly liquid assets** — serving as a buffer for operational risk, chargebacks and settlement timing.
3. **5-year projection** built on three scenarios (Base / Conservative / Aggressive) plus stress testing.
4. **Liquidity & solvency** tracked via monthly metrics (LCR-style, current ratio, capital adequacy) and reported to BOT on the required cadence.

---

## 2. Registered Capital Structure — THB 50 Million

### 2.1 Capital Components & Allocation

Paid-up capital of THB 50,000,000, with a managerial (purpose) allocation — the capital remains the company's:

| Component | Amount (THB M) | Share | Asset form | Purpose |
|-----------|----------------|-------|------------|---------|
| **Regulatory capital buffer (keep ≥ 75%)** | 37.50 | 75% | Fixed deposit / govt bonds / T-bills | BOT minimum capital; not to be used as working capital |
| **Operational working capital** | 7.50 | 15% | Current / savings deposits | Payroll, infra, PCI/QSA, scheme fees |
| **Settlement / chargeback reserve** | 3.00 | 6% | Highly liquid deposits | Absorbs settlement timing gaps and chargeback liability |
| **Contingency / stress reserve** | 2.00 | 4% | Highly liquid deposits | Unexpected events / fraud loss |
| **Total** | **50.00** | **100%** | | |

> **Note:** Settlement and chargeback reserves above are company-funded buffers, **separate** from merchant funds
> held in the segregated account (see §5) — not double-counted.

### 2.2 Cap Table — Outline

| Shareholder | Type | Share | Note |
|-------------|------|-------|------|
| Founders / Thai director(s) | Ordinary | [TODO %] | At least one director must be a Thai national resident in Thailand |
| Strategic investor(s) | Ordinary / Preferred | [TODO %] | Foreign majority ownership may require a Foreign Business License (FBL) |
| ESOP pool | — | [TODO %] | Per company policy |

> 🔲 **ASSUMPTION / TODO — actual shareholding:** Ownership percentages and majority-shareholder nationality are
> not finalized. Must be confirmed before submission as they affect FBL requirements and directors'/executives' fit & proper.

### 2.3 Source of Funds

| Source | Amount (THB M) | Status | Evidence |
|--------|----------------|--------|----------|
| Founder paid-in capital | [TODO] | [TODO] | Certificates, bank statements |
| Equity round (investors) | [TODO] | [TODO] | Share subscription agreement |
| **Total paid-up capital** | **50.00** | Must be fully paid before filing | Share payment receipts, BOJ.5 |

> 🔲 **ASSUMPTION / TODO — actual fundraising:** Amounts and payment timing are unconfirmed. Source of funds must
> pass **AML/CDD source-of-fund verification** and be reported per **AMLO (ปปง.)** rules before being accepted as capital.

---

## 3. Key Assumptions

| Parameter | Base assumption | Note |
|-----------|-----------------|------|
| Average Merchant Discount Rate (MDR) | 1.85% of TPV | Primary revenue; net of interchange + scheme fees |
| Interchange + scheme fee (COGS) | ~1.20% of TPV | Paid to issuer/scheme (net MDR margin ~0.65%) |
| Gateway fee per transaction | THB 1.50 / txn | For authorization + 3DS |
| Chargeback rate | 0.10% of transactions | Target well below scheme thresholds (Visa/MC 0.9%–1.0%) |
| Net fraud loss | 0.03% of TPV | After 3DS liability shift |
| Settlement cycle | T+1 / T+2 | Per sponsor-bank agreement |
| Corporate income tax | 20% | Thai Revenue Code |
| Interest yield on buffer | 2.0% p.a. | Fixed deposit / short-term bonds |

> 🔲 **ASSUMPTION / TODO — unit economics:** Actual MDR, interchange and scheme fees depend on unresolved contracts
> with the **sponsor bank** and **card schemes (Visa/Mastercard)**. Figures above are market reference points.

### 3.1 Transaction Volume Assumptions (TPV — Total Payment Volume)

| Year | TPV (THB M) | Transactions (M) | Active merchants |
|------|-------------|------------------|------------------|
| Y1 | 1,200 | 4.0 | 300 |
| Y2 | 4,500 | 15.0 | 1,200 |
| Y3 | 12,000 | 40.0 | 3,500 |
| Y4 | 24,000 | 80.0 | 7,000 |
| Y5 | 40,000 | 130.0 | 12,000 |

> From Y2, transactions exceed 6M/year → **confirms the need for PCI-DSS Level 1** (aligns with `ARCHITECTURE.md` §1, §7).

---

## 4. 5-Year Financial Projections (Base Case)

### 4.1 Summary P&L (THB M)

| Item | Y1 | Y2 | Y3 | Y4 | Y5 |
|------|----|----|----|----|----|
| MDR revenue (net margin ~0.65% of TPV) | 7.80 | 29.25 | 78.00 | 156.00 | 260.00 |
| Per-transaction & other fees | 6.00 | 22.50 | 60.00 | 120.00 | 195.00 |
| **Total revenue** | **13.80** | **51.75** | **138.00** | **276.00** | **455.00** |
| COGS (processing, 3DS, scheme pass-through) | (5.20) | (18.00) | (46.00) | (88.00) | (140.00) |
| **Gross profit** | **8.60** | **33.75** | **92.00** | **188.00** | **315.00** |
| Personnel (team per ROADMAP §3) | (18.00) | (26.00) | (38.00) | (52.00) | (66.00) |
| PCI-DSS / QSA / ASV / pentest | (4.50) | (5.00) | (5.50) | (6.00) | (6.50) |
| HSM/KMS + Infra (HA + DR) | (5.00) | (6.00) | (7.50) | (9.00) | (11.00) |
| Legal / licensing / compliance | (2.50) | (1.50) | (1.50) | (1.80) | (2.00) |
| Chargeback + fraud loss | (0.50) | (1.80) | (4.80) | (9.60) | (16.00) |
| Other operating expenses | (3.00) | (4.50) | (6.50) | (9.00) | (12.00) |
| **EBITDA** | **(24.90)** | **(11.05)** | **27.70** | **90.80** | **201.50** |
| Depreciation & amortization | (1.50) | (2.00) | (2.50) | (3.00) | (3.50) |
| **EBIT** | **(26.40)** | **(13.05)** | **25.20** | **87.80** | **198.00** |
| Interest income on buffer | 0.75 | 0.75 | 0.75 | 0.90 | 1.00 |
| Income tax (20%, after loss carryforward used) | 0.00 | 0.00 | 0.00 | (14.40) | (39.80) |
| **Net profit (loss)** | **(25.65)** | **(12.30)** | **25.95** | **74.30** | **159.20** |
| Accumulated P&L (carried) | (25.65) | (37.95) | (12.00) | 62.30 | 221.50 |

**Break-even:** EBITDA turns positive in Y3; cumulative net profit turns positive in Y4.

### 4.2 Summary Cash Flow (THB M)

| Item | Y1 | Y2 | Y3 | Y4 | Y5 |
|------|----|----|----|----|----|
| Operating cash flow | (24.90) | (11.05) | 27.70 | 90.80 | 201.50 |
| Investing (capex: HSM, infra) | (3.00) | (2.50) | (3.00) | (3.50) | (4.00) |
| Opening cash | 50.00 | 22.10 | 8.55 | 33.25 | 120.55 |
| **Closing cash** | **22.10** | **8.55** | **33.25** | **120.55** | **318.05** |

> **Liquidity pinch point:** End of Y2 leaves ~THB 8.55M, below allocated operational working capital
> → **a bridge/Series raise is required by end of Y1 or mid-Y2**, before reaching this point (see runway §6.3).

> 🔲 **ASSUMPTION / TODO — runway & top-up:** This projection shows THB 50M alone is insufficient to cover the
> cumulative loss period before break-even. Additional working capital must be planned, **without touching the THB 37.5M regulatory buffer.**

### 4.3 Summary Balance Sheet (year-end, THB M)

| Item | Y1 | Y2 | Y3 | Y4 | Y5 |
|------|----|----|----|----|----|
| Liquid assets (excl. merchant funds) | 22.10 | 8.55 | 33.25 | 120.55 | 318.05 |
| Net fixed assets | 4.50 | 5.00 | 5.50 | 6.00 | 6.50 |
| Merchant funds in segregated account (managerially off-B/S) | ~3.3 | ~12.3 | ~33 | ~66 | ~110 |
| Current liabilities (settlement payable to merchants) | ~3.3 | ~12.3 | ~33 | ~66 | ~110 |
| Shareholders' equity (capital + retained earnings) | 24.35 | 12.05 | 38.00 | 112.30 | 271.50 |

---

## 5. Client Money & Settlement

### 5.1 Segregation Principle

- Funds received from card networks to settle merchants are **not company revenue/assets** — a fiduciary liability only.
- Held in a **segregated account** at a BOT-supervised institution, separate from the company's operating account.
- **Merchant funds must never be used as working capital** or invested (no commingling).
- The segregated account is reconciled against `settlement payable` in the ledger **every business day**, referencing
  `ledger_entries(settlement)` (double-entry, append-only) per `ARCHITECTURE.md` §6.

### 5.2 Roles & Authority

| Role | Financial responsibility | Authority |
|------|--------------------------|-----------|
| CFO / Head of Finance | Capital policy, solvency/liquidity oversight, BOT reporting | Highest approval; owns capital plan |
| Treasury Manager | Buffer management, segregated account, daily reconciliation | Acts within CFO-defined limits |
| Compliance Officer | AML/CDD source-of-fund, AMLO reporting, BOT rules | Veto over AML-risky transactions |
| Internal Audit | Audits segregation, reconciliation, capital adequacy | Reports to Audit Committee |
| External Auditor (CPA) | Certifies financials & projections | Independent |

### 5.3 Chargeback & Reserve

- Dynamic chargeback reserve per merchant risk profile (rolling reserve for high-risk merchants).
- Chargeback liability linked to the dispute/chargeback workflow (`ROADMAP.md` Phase 3) and 3DS 2.x liability shift (EMV 3DS).
- Target chargeback rate < 0.10%, significantly below card-scheme thresholds (Visa VDMP / Mastercard ECM).

---

## 6. Liquidity & Solvency

### 6.1 Metrics & Thresholds

| Metric | Formula | Target | Amber | Red |
|--------|---------|--------|-------|-----|
| **Capital adequacy** | Maintained capital / THB 50M minimum | ≥ 100% | < 90% | < 75% (BOT breach) |
| **Current ratio** | Current assets / current liabilities | ≥ 1.5 | < 1.2 | < 1.0 |
| **Liquidity coverage (LCR-style)** | HQLA / net 30-day outflow | ≥ 1.2 | < 1.1 | < 1.0 |
| **Segregation ratio** | Segregated balance / settlement payable | = 100% | < 100% | < 100% (top up now) |
| **Cash runway** | Operating cash / monthly burn | ≥ 12 mo | < 9 mo | < 6 mo |

### 6.2 Stress Testing

Run at least semi-annually across:

1. **Settlement delay** — sponsor bank delays settlement by 5 business days → buffer + segregated account coverage.
2. **Chargeback spike** — chargebacks triple (0.30%) for 3 months → reserve adequacy.
3. **Revenue shortfall** — TPV 40% below Base → runway and capital adequacy.
4. **Fraud event** — single fraud loss of THB 5M → contingency reserve.
5. **Operational outage** — downtime impacting revenue + SLA penalties (availability ≥ 99.95%, `ARCHITECTURE.md` §8).

### 6.3 Scenario Analysis

| Scenario | TPV vs Base | EBITDA break-even | Lowest cash | Comment |
|----------|-------------|-------------------|-------------|---------|
| **Conservative** | -40% | Y4 | ~THB 2–3M (Y2–Y3) | Needs ~THB 15–20M raise before Y2 |
| **Base** | — | Y3 | ~THB 8.55M (Y2) | Prepare a bridge by end of Y1 |
| **Aggressive** | +50% | Y2 | ~THB 18M (Y2) | May not need a raise |

> **Liquidity conclusion:** In every scenario the **THB 37.5M regulatory buffer must remain fully intact** and never
> offset operating losses. Operating liquidity gaps must be closed via capital/long-term debt raises, not by reducing minimum capital.

---

## 7. Governance, Reporting & Internal Controls

| Report | Frequency | Recipient | Basis |
|--------|-----------|-----------|-------|
| Capital maintenance (≥ 75%) | Monthly + quarterly | BOT (ธปท.) | Payment Systems Act B.E. 2560 |
| Audited financial statements | Annual | BOT / DBD | Thai accounting standards |
| Segregated-account reconciliation | Daily | Treasury / Internal Audit | Segregation rules |
| Suspicious/cash transaction reports (STR/CTR) | Event-driven | AMLO (ปปง.) | Anti-Money Laundering Act |
| IT / cyber incidents | Event-driven + periodic | BOT | BOT IT/cyber resilience notifications |
| Personal data breach | Within 72 hours | PDPC | PDPA B.E. 2562 |

**Key internal controls:** dual control and split knowledge for moving buffer/reserve funds (consistent with PCI Req 3
key-management principles), maker-checker for payouts, and segregation of duties across Treasury / Compliance / Internal Audit.

---

## 8. Assumptions & TODO — Summary

| # | Item | Owner | Deadline (per ROADMAP) |
|---|------|-------|------------------------|
| A1 | Shareholding structure + nationality (FBL?) | Legal / Founders | Before BOT filing (45-day entity prep) |
| A2 | Actual share payment amounts/timing + source of funds (AML) | Finance / Compliance | Before filing |
| A3 | Sponsor-bank contract + actual MDR/interchange | Product / BD | Critical path (Track B) |
| A4 | QSA vendor + PCI-DSS L1 scope/pricing | Security / DevSecOps | Phase 4 (Q4 2026) |
| A5 | Top-up/bridge plan before break-even | CFO / Board | End of Y1 |
| A6 | Bank for segregated account + trust terms | Treasury / Legal | Before go-live |

---

*End of document — `docs/compliance/02-financial-projections-capital.md`*
