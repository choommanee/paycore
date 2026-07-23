# แผนธุรกิจ 5 ปี (ไทย)

> เอกสารประกอบคำขอรับใบอนุญาตประกอบธุรกิจ **การให้บริการรับชำระเงินด้วยวิธีการทางอิเล็กทรอนิกส์ (Acquiring)**
> ภายใต้ **พ.ร.บ. ระบบการชำระเงิน พ.ศ. 2560** กำกับโดย **ธนาคารแห่งประเทศไทย (ธปท.)**
> ผู้ขอ: **[บริษัท / Company]** · ทุนจดทะเบียนชำระแล้ว: **50 ล้านบาท** · มาตรฐานเป้าหมาย: **PCI-DSS v4.0 Level 1**
>
> เอกสารนี้เป็นส่วนที่ 01 ของชุดคำขอ (ดู `COMPLIANCE-TH.md`, `ARCHITECTURE.md`, `ROADMAP.md` ประกอบ)
> **ข้อมูลนี้เป็นเอกสารเชิงธุรกิจ/เทคนิค ไม่ใช่คำแนะนำทางกฎหมาย** — ต้องผ่านการตรวจของที่ปรึกษากฎหมายด้านใบอนุญาต ธปท. ก่อนยื่นจริง

---

## 0. บทสรุปผู้บริหาร (Executive Summary)

**[บริษัท / Company]** ขอรับใบอนุญาต Full Acquiring เพื่อให้บริการรับชำระเงินด้วยบัตร (Visa / Mastercard / JCB / UnionPay), Thai QR / PromptPay, และช่องทางชำระเงินท้องถิ่นแก่ผู้ประกอบการในประเทศไทย โดยเป็นผู้รับชำระเงิน (acquirer) โดยตรงผ่านสัญญากับธนาคารผู้สนับสนุน (sponsor bank) และการรับรอง (certification) กับเครือข่ายบัตร

จุดยืนหลัก: **แพลตฟอร์ม acquiring ที่สร้างบนสถาปัตยกรรมสมัยใหม่ (Go/Fiber, double-entry ledger, tokenization vault) ปฏิบัติตาม PCI-DSS v4.0 Level 1 และเกณฑ์ ธปท. ตั้งแต่ออกแบบ (compliance-by-design)** เน้นความโปร่งใสด้านราคา คุณภาพการกระทบยอด (reconciliation) และเวลาการชำระเงินให้ร้านค้า (settlement) ที่คาดการณ์ได้

เป้าหมายการเงิน 5 ปี (สรุป):

| ปี | ปริมาณเงินชำระ (TPV) | รายได้สุทธิ (Net Revenue) | จำนวนร้านค้า active | EBITDA |
|----|----------------------|---------------------------|---------------------|--------|
| ปีที่ 1 (2027) | 1,500 ล้านบาท | 18 ล้านบาท | 400 | ติดลบ |
| ปีที่ 2 (2028) | 6,000 ล้านบาท | 66 ล้านบาท | 1,800 | ใกล้คุ้มทุน |
| ปีที่ 3 (2029) | 15,000 ล้านบาท | 158 ล้านบาท | 5,000 | เป็นบวก |
| ปีที่ 4 (2030) | 30,000 ล้านบาท | 300 ล้านบาท | 10,000 | เป็นบวกแข็งแรง |
| ปีที่ 5 (2031) | 52,000 ล้านบาท | 494 ล้านบาท | 18,000 | มาร์จิ้นสูงขึ้น |

> ตัวเลขข้างต้นเป็นประมาณการเชิงกลยุทธ์ อิงสมมติฐานในข้อ 4 และ 5 (take rate เฉลี่ย ~0.95–1.1% ของ TPV) มิใช่การรับประกันผล

---

## 1. ตลาดและบริบทอุตสาหกรรม (Market)

### 1.1 ภาพรวมตลาด
- ประเทศไทยเป็นตลาดที่ **การชำระเงินดิจิทัลเติบโตสูง** โดยมี PromptPay และ Thai QR Payment เป็นโครงสร้างพื้นฐานที่ ธปท. ผลักดัน และมีการใช้บัตรเดบิต/เครดิตในกลุ่ม e-commerce, ท่องเที่ยว, และค้าปลีกสมัยใหม่อย่างต่อเนื่อง
- แนวโน้มเชิงโครงสร้าง: การเปลี่ยนจากเงินสดสู่ดิจิทัล (cash-to-digital), การเติบโตของ e-commerce ข้ามพรมแดน, การท่องเที่ยวฟื้นตัว (ธุรกรรมบัตรต่างประเทศ), และความต้องการ omni-channel (ออนไลน์ + หน้าร้าน + in-app)

### 1.2 กลุ่มตลาดเป้าหมาย (Segments)

| Segment | ลักษณะ | ความต้องการหลัก |
|---------|--------|------------------|
| e-commerce SME | ร้านค้าออนไลน์ ยอดต่อรายการปานกลาง | onboarding เร็ว, ค่าธรรมเนียมโปร่งใส, checkout ที่ conversion ดี |
| Digital / SaaS / Subscription | เก็บเงินซ้ำ (recurring) | tokenization, network token, retry logic, 3DS |
| Travel & Hospitality | ยอดต่อรายการสูง, บัตรต่างชาติ | multi-currency, auth-then-capture, refund/dispute ที่ดี |
| Marketplace / Platform | จ่ายเงินต่อผู้ขายหลายราย (split) | split settlement, sub-merchant, รายงานแยกร้าน |
| Enterprise Retail (omni-channel) | ปริมาณสูง | SLA, การกระทบยอดระดับสาขา, ราคาต่อรายการต่ำ (IC++) |

### 1.3 หน่วยงานกำกับที่เกี่ยวข้อง

| หน่วยงาน / มาตรฐาน | บทบาทต่อธุรกิจนี้ |
|--------------------|-------------------|
| **ธนาคารแห่งประเทศไทย (ธปท.)** | กำกับใบอนุญาต Acquiring, IT risk, cyber resilience, BCP, outsourcing, การรายงานเป็นงวด |
| **สำนักงาน ปปง. (AMLO)** | KYC/CDD, การรายงานธุรกรรม (STR/CTR), sanction/PEP screening ตามกฎหมายฟอกเงิน |
| **สำนักงาน คปช. / PDPC** | การคุ้มครองข้อมูลส่วนบุคคลตาม PDPA (พ.ร.บ. คุ้มครองข้อมูลส่วนบุคคล พ.ศ. 2562) |
| **PCI SSC (ผ่าน QSA/ASV)** | PCI-DSS v4.0 Level 1, EMV 3DS (3-D Secure 2.x) |
| **เครือข่ายบัตร (Visa / Mastercard)** | scheme rules, certification, interchange, chargeback/dispute |

---

## 2. ลูกค้าเป้าหมาย (Target Merchants)

### 2.1 เกณฑ์การรับสมัคร (Merchant Eligibility & Risk Tiers)
ทุกร้านค้าต้องผ่าน **KYC/KYB + CDD** ตามนโยบาย AML (สอดคล้องกฎหมาย ปปง.) ก่อน onboarding:

| Risk Tier | ตัวอย่างประเภทธุรกิจ (MCC) | ระดับการตรวจสอบ | เพดานเบื้องต้น |
|-----------|----------------------------|------------------|-----------------|
| Low | ค้าปลีกทั่วไป, ร้านอาหาร, SaaS | KYB มาตรฐาน + sanction screen | ตามยอดขายจริง |
| Medium | ท่องเที่ยว, ตั๋วเครื่องบิน, marketplace | KYB เข้ม + delayed settlement/rolling reserve | มี rolling reserve 5–10% |
| High / ต้องพิจารณาพิเศษ | high-chargeback, digital goods, ธุรกรรมข้ามพรมแดนสูง | EDD + คณะกรรมการรับความเสี่ยง | อนุมัติรายกรณี |
| **ต้องห้าม (Prohibited)** | สินค้าผิดกฎหมาย, การพนันที่ไม่ได้รับอนุญาต, สินค้าละเมิดสิทธิ์, ธุรกิจในบัญชี sanction | ปฏิเสธ | — |

### 2.2 กระบวนการ onboarding ร้านค้า (สรุป)
1. สมัคร → เก็บเอกสารนิติบุคคล/บุคคลธรรมดา, กรรมการ, ผู้ถือหุ้นที่แท้จริง (UBO)
2. **KYB + sanction/PEP screening** (ปปง.) → ประเมิน risk tier
3. Underwriting (ประเมินความเสี่ยงเครดิต/chargeback, กำหนด reserve)
4. ลงนามข้อตกลงร้านค้า + ค่าธรรมเนียม → ออก API key / merchant ID
5. Technical integration + go-live (ทดสอบใน sandbox → production)

> รายละเอียด KYC/AML และ PDPA อยู่ในเอกสารส่วนแยกของชุดคำขอ (นโยบาย AML/KYC/CDD และนโยบายคุ้มครองข้อมูลส่วนบุคคล)

---

## 3. โครงสร้างราคาและค่าธรรมเนียม (Pricing)

### 3.1 โมเดลราคา
เสนอ 2 โครงสร้างตาม segment:

| โครงสร้าง | เหมาะกับ | รูปแบบ |
|-----------|----------|--------|
| **Blended (flat)** | SME / e-commerce | อัตราเดียวต่อรายการ เช่น **2.65% + 5 บาท** (บัตรในประเทศ) — เข้าใจง่าย โปร่งใส |
| **Interchange++ (IC++)** | Enterprise / ปริมาณสูง | interchange (ของ issuer) + scheme fee + **markup ของเรา** แยกบรรทัด — โปร่งใสต่อรายการ |

ตัวอย่างอัตราอ้างอิง (บัตรในประเทศ ยังไม่รวม VAT):

| ช่องทาง | อัตราตัวอย่าง |
|---------|----------------|
| บัตรเครดิต/เดบิตในประเทศ (blended) | 2.4% – 2.9% |
| บัตรต่างประเทศ / cross-border | +1.0% – 1.5% |
| Thai QR / PromptPay | 0.5% – 1.0% (หรือคงที่ต่อรายการ) |
| Refund / chargeback handling fee | คงที่ต่อรายการ (เช่น 300–600 บาท ต่อ chargeback) |
| Payout / settlement fee | คงที่ต่อรอบ (option) |

### 3.2 หมายเหตุด้านต้นทุน
- ต้นทุนหลักคือ **interchange (จ่ายให้ issuer)** + **scheme fee (Visa/MC)** + ค่าธรรมเนียม sponsor bank
- **take rate สุทธิของเรา** (net revenue / TPV) หลังหักต้นทุนต่อรายการ ประมาณ **0.9% – 1.1%** เป็นฐานประมาณการรายได้

> **[ASSUMPTION / TODO]** อัตรา interchange และ scheme fee ที่แน่นอน รวมถึง sponsor bank fee เป็นตัวเลขที่ต้อง **ยืนยันหลังเจรจากับ sponsor bank และ scheme** (ดูข้อ 7 สมมติฐาน)

---

## 4. โมเดลรายได้ (Revenue Model)

### 4.1 แหล่งรายได้
1. **Merchant Discount Rate (MDR)** — รายได้หลัก คิดเป็น % ของ TPV
2. **Per-transaction fee** — คงที่ต่อรายการ (เสริมมาร์จิ้นในรายการยอดต่ำ)
3. **Value-added services** — tokenization/network token, สมัครสมาชิก (recurring), fraud/risk tools, รายงาน/แดชบอร์ด, FX markup
4. **Chargeback / dispute handling fees**
5. **Settlement / payout options** (จ่ายเร็ว next-day เป็นบริการเสริม)

### 4.2 หน่วยเศรษฐศาสตร์ (Unit Economics — ตัวอย่าง)

| รายการ | ค่า (ตัวอย่าง) |
|--------|----------------|
| MDR ที่เก็บจากร้านค้า | 2.65% |
| หัก: interchange + scheme + sponsor | ~1.6% – 1.75% |
| **Net take rate** | **~0.9% – 1.05%** |
| ต้นทุน fraud/chargeback loss (เฉลี่ย) | 0.05% – 0.15% ของ TPV |

### 4.3 ประมาณการรายได้ 5 ปี
(สอดคล้องตารางบทสรุปข้อ 0; net take rate เฉลี่ยแบบถ่วงน้ำหนัก ~0.95–1.1%)

| ปี | TPV (ล้านบาท) | net take rate | Net Revenue (ล้านบาท) |
|----|---------------|----------------|------------------------|
| 1 | 1,500 | 1.20% | 18 |
| 2 | 6,000 | 1.10% | 66 |
| 3 | 15,000 | 1.05% | 158 |
| 4 | 30,000 | 1.00% | 300 |
| 5 | 52,000 | 0.95% | 494 |

> take rate ค่อย ๆ ลดลงตามปริมาณ (สัดส่วนลูกค้า enterprise/IC++ และ QR เพิ่มขึ้น) แต่รายได้รวมโตจาก TPV

---

## 5. จุดยืนเชิงแข่งขันและการเติบโต (Competitive Positioning & Growth)

### 5.1 คู่แข่งและการวางตำแหน่ง
- คู่แข่ง: PSP/gateway ในไทย (ทั้ง facilitator ที่ต่อ acquirer เดิม และ acquirer ธนาคาร) และผู้เล่นสากลที่ให้บริการในไทย
- **ความแตกต่างของเรา:**
  1. **Full acquiring** → คุมต้นทุนต่อรายการได้ดีกว่า facilitator, เสนอ IC++ ได้จริง
  2. **Compliance-by-design** → PCI-DSS v4.0 L1 + เกณฑ์ ธปท. + PDPA ฝังในสถาปัตยกรรม (double-entry ledger, audit trail, tokenization vault)
  3. **Reconciliation & settlement ที่โปร่งใส** → ledger append-only เป็น source of truth, รายงานกระทบยอดต่อร้าน
  4. **Developer experience** → API สะอาด, idempotency, webhook ลงลายเซ็น, sandbox

### 5.2 กลยุทธ์การเติบโต (แผน 5 ปี)

| ช่วง | โฟกัส |
|------|-------|
| **ปีที่ 1 (2027)** | เปิดตัวหลังได้ใบอนุญาต + PCI L1; จับกลุ่ม e-commerce SME; onboarding ~400 ร้าน; พิสูจน์ reconciliation/settlement |
| **ปีที่ 2 (2028)** | ขยาย recurring/SaaS + travel; เพิ่ม value-added (tokenization, fraud tools); พันธมิตร platform/marketplace |
| **ปีที่ 3 (2029)** | เข้ากลุ่ม enterprise (IC++, SLA); cross-border/multi-currency; EBITDA เป็นบวก |
| **ปีที่ 4 (2030)** | omni-channel (หน้าร้าน + ออนไลน์); ขยายช่องทางท้องถิ่นเพิ่ม; scale ปริมาณ |
| **ปีที่ 5 (2031)** | ครองส่วนแบ่งกลุ่มเป้าหมาย; พิจารณาบริการเสริม (เช่น payout, ผลิตภัณฑ์ต่อยอด); เตรียมขยายภูมิภาค |

### 5.3 ตัวชี้วัดหลัก (KPI) ที่ติดตามและรายงานต่อ ธปท./ภายใน
TPV, จำนวนร้าน active, authorization success rate, **chargeback rate (เป้าหมาย < 0.9% ตาม scheme threshold)**, fraud loss rate, uptime (เป้า ≥ 99.95%), settlement accuracy, จำนวน STR ที่รายงาน ปปง.

### 5.4 ความเสี่ยงเชิงกลยุทธ์
เวลา/ต้นทุนใบอนุญาต + PCI, การพึ่งพา sponsor bank, การแข่งขันด้านราคา, fraud/chargeback, และ compliance ต่อเนื่อง (ดู `ARCHITECTURE.md` ข้อ 9 และ `ROADMAP.md` ข้อ 5)

---

## 6. บทบาทและองค์กร (Roles & Governance)

| บทบาท | ความรับผิดชอบหลัก |
|-------|--------------------|
| CEO / กรรมการผู้จัดการ | กลยุทธ์, ความสัมพันธ์ ธปท./sponsor bank, มีกรรมการสัญชาติไทยถิ่นที่อยู่ในไทยอย่างน้อย 1 คน (เกณฑ์ ธปท.) |
| Compliance Officer / MLRO | AML/KYC, รายงาน ปปง., เกณฑ์ ธปท., ประสาน QSA |
| DPO (Data Protection Officer) | PDPA, สิทธิเจ้าของข้อมูล, การประมวลผลข้อมูล |
| CISO / DevSecOps Lead | PCI-DSS v4.0, HSM/KMS, network segmentation, pentest/ASV |
| Head of Risk / Underwriting | merchant risk tier, reserve, chargeback/fraud |
| Head of Engineering / SRE | availability, ledger, reconciliation, DR |
| Head of Finance | ทุนจดทะเบียน (คงไว้ ≥75%), settlement, ประมาณการการเงิน |

---

## 7. สมมติฐานและรายการที่ยังไม่ยืนยัน (Assumptions / TODO)

> **[ASSUMPTION / TODO — ต้องยืนยันก่อนยื่น/ก่อน go-live]**
>
> 1. **Sponsor bank / scheme certification** — ยังไม่ได้เลือกธนาคารผู้สนับสนุนและยังไม่เริ่ม certification กับ Visa/Mastercard เป็นทางการ ตัวเลข fee และ timeline certification เป็นประมาณการ ต้องยืนยันหลังลงนาม (ดู `ROADMAP.md`)
> 2. **QSA vendor** — ยังไม่ได้เลือก QSA สำหรับ PCI-DSS v4.0 L1 (RoC) และ ASV สำหรับ quarterly scan ต้นทุน PCI เป็นกรอบกว้าง (`ROADMAP.md` ข้อ 4)
> 3. **ทุนจดทะเบียนชำระแล้ว** — แผนอิงเกณฑ์ 50 ล้านบาท (Acquiring) และต้องคงไว้ ≥ 75% ตลอดการดำเนินงาน ต้องยืนยันการชำระทุนจริงและวัตถุประสงค์บริษัทครอบคลุมธุรกิจนี้
> 4. **อัตรา interchange / scheme fee / MDR** — ตัวเลขในข้อ 3–4 เป็นตัวอย่างอ้างอิง ต้องแทนที่ด้วยอัตราจริงหลังเจรจา
> 5. **ประมาณการ TPV/รายได้** — เป็นสมมติฐานเชิงกลยุทธ์เพื่อวางแผน มิใช่การรับประกันผล จะปรับตามผลจริงและ pipeline

---
---

# 5-year business plan: market, target merchants, pricing, revenue model, competitive positioning, growth (English)

> Supporting document for the license application to operate an **Electronic Payment Acquiring Service (Acquiring)**
> under the **Payment Systems Act B.E. 2560 (2017)**, supervised by the **Bank of Thailand (BOT / ธปท.)**.
> Applicant: **[บริษัท / Company]** · Paid-up capital: **THB 50 million** · Target standard: **PCI-DSS v4.0 Level 1**.
>
> This is document 01 of the application set (read together with `COMPLIANCE-TH.md`, `ARCHITECTURE.md`, `ROADMAP.md`).
> **This is a business/technical document, not legal advice** — must be reviewed by BOT-licensing counsel before submission.

---

## 0. Executive Summary

**[บริษัท / Company]** seeks a Full Acquiring licence to provide card acceptance (Visa / Mastercard / JCB / UnionPay), Thai QR / PromptPay, and local payment methods to merchants in Thailand, acting as a direct acquirer via a sponsor-bank agreement and card-network certification.

Core positioning: **a modern acquiring platform (Go/Fiber, double-entry ledger, tokenization vault) that is compliant-by-design with PCI-DSS v4.0 Level 1 and BOT requirements**, emphasising pricing transparency, reconciliation quality, and predictable merchant settlement.

Five-year financial targets (summary):

| Year | Total Payment Volume (TPV) | Net Revenue | Active merchants | EBITDA |
|------|-----------------------------|-------------|------------------|--------|
| Y1 (2027) | THB 1,500M | THB 18M | 400 | Negative |
| Y2 (2028) | THB 6,000M | THB 66M | 1,800 | Near break-even |
| Y3 (2029) | THB 15,000M | THB 158M | 5,000 | Positive |
| Y4 (2030) | THB 30,000M | THB 300M | 10,000 | Strongly positive |
| Y5 (2031) | THB 52,000M | THB 494M | 18,000 | Expanding margin |

> Figures are strategic projections based on the assumptions in Sections 4–5 (average take rate ~0.95–1.1% of TPV), not guarantees.

---

## 1. Market

### 1.1 Market overview
- Thailand has a **fast-growing digital-payments market**, anchored by PromptPay and Thai QR Payment infrastructure promoted by the BOT, with sustained debit/credit card usage in e-commerce, travel, and modern retail.
- Structural trends: cash-to-digital migration, cross-border e-commerce, tourism recovery (foreign-card volume), and demand for omni-channel (online + in-store + in-app).

### 1.2 Target segments

| Segment | Profile | Key needs |
|---------|---------|-----------|
| e-commerce SME | Online shops, mid-size ticket | Fast onboarding, transparent fees, high-conversion checkout |
| Digital / SaaS / Subscription | Recurring billing | Tokenization, network token, retry logic, 3DS |
| Travel & Hospitality | High ticket, foreign cards | Multi-currency, auth-then-capture, strong refund/dispute |
| Marketplace / Platform | Split payouts to many sellers | Split settlement, sub-merchant, per-store reporting |
| Enterprise Retail (omni-channel) | High volume | SLA, branch-level reconciliation, low per-txn cost (IC++) |

### 1.3 Relevant regulators & standards

| Body / standard | Role for this business |
|-----------------|------------------------|
| **Bank of Thailand (ธปท.)** | Acquiring licence, IT risk, cyber resilience, BCP, outsourcing, periodic reporting |
| **AMLO (ปปง.)** | KYC/CDD, transaction reporting (STR/CTR), sanction/PEP screening under AML law |
| **PDPC (คปช.)** | Personal-data protection under PDPA (B.E. 2562) |
| **PCI SSC (via QSA/ASV)** | PCI-DSS v4.0 Level 1, EMV 3DS (3-D Secure 2.x) |
| **Card networks (Visa / Mastercard)** | Scheme rules, certification, interchange, chargeback/dispute |

---

## 2. Target Merchants

### 2.1 Merchant eligibility & risk tiers
Every merchant passes **KYC/KYB + CDD** per AML policy (aligned with AMLO law) before onboarding:

| Risk Tier | Example business types (MCC) | Diligence | Initial controls |
|-----------|-------------------------------|-----------|-------------------|
| Low | General retail, restaurants, SaaS | Standard KYB + sanction screen | Per actual volume |
| Medium | Travel, airline tickets, marketplace | Enhanced KYB + delayed settlement / rolling reserve | 5–10% rolling reserve |
| High / special review | High-chargeback, digital goods, high cross-border | EDD + risk committee | Case-by-case approval |
| **Prohibited** | Illegal goods, unlicensed gambling, IP-infringing goods, sanctioned parties | Reject | — |

### 2.2 Merchant onboarding process (summary)
1. Apply → collect entity/individual docs, directors, ultimate beneficial owners (UBO).
2. **KYB + sanction/PEP screening** (AMLO) → assign risk tier.
3. Underwriting (credit/chargeback risk, set reserve).
4. Sign merchant agreement + fees → issue API key / merchant ID.
5. Technical integration + go-live (sandbox → production).

> Detailed KYC/AML and PDPA controls live in separate documents of the application set (AML/KYC/CDD policy and Personal-Data Protection policy).

---

## 3. Pricing

### 3.1 Pricing model
Two structures by segment:

| Structure | Best for | Form |
|-----------|----------|------|
| **Blended (flat)** | SME / e-commerce | Single rate per txn, e.g. **2.65% + THB 5** (domestic) — simple, transparent |
| **Interchange++ (IC++)** | Enterprise / high volume | interchange (issuer) + scheme fee + **our markup**, itemised — per-txn transparency |

Indicative rates (domestic cards, excl. VAT):

| Channel | Indicative rate |
|---------|------------------|
| Domestic credit/debit (blended) | 2.4% – 2.9% |
| Foreign / cross-border card | +1.0% – 1.5% |
| Thai QR / PromptPay | 0.5% – 1.0% (or fixed per txn) |
| Refund / chargeback handling fee | Fixed per txn (e.g. THB 300–600 per chargeback) |
| Payout / settlement fee | Fixed per cycle (optional) |

### 3.2 Cost notes
- Main costs are **interchange (paid to issuer)** + **scheme fee (Visa/MC)** + sponsor-bank fees.
- **Net take rate** (net revenue / TPV) after per-txn costs is ~**0.9% – 1.1%**, the basis for revenue projections.

> **[ASSUMPTION / TODO]** Exact interchange, scheme fees, and sponsor-bank fees must be **confirmed after negotiating with the sponsor bank and scheme** (see Section 7).

---

## 4. Revenue Model

### 4.1 Revenue sources
1. **Merchant Discount Rate (MDR)** — primary revenue, % of TPV.
2. **Per-transaction fee** — fixed per txn (protects margin on low-value txns).
3. **Value-added services** — tokenization/network token, recurring billing, fraud/risk tools, reporting/dashboard, FX markup.
4. **Chargeback / dispute handling fees.**
5. **Settlement / payout options** (faster next-day payout as an add-on).

### 4.2 Unit economics (example)

| Item | Value (example) |
|------|-----------------|
| MDR charged to merchant | 2.65% |
| Less: interchange + scheme + sponsor | ~1.6% – 1.75% |
| **Net take rate** | **~0.9% – 1.05%** |
| Fraud/chargeback loss (avg) | 0.05% – 0.15% of TPV |

### 4.3 Five-year revenue projection
(Consistent with Section 0; weighted-average net take rate ~0.95–1.1%.)

| Year | TPV (THB M) | Net take rate | Net Revenue (THB M) |
|------|-------------|----------------|----------------------|
| 1 | 1,500 | 1.20% | 18 |
| 2 | 6,000 | 1.10% | 66 |
| 3 | 15,000 | 1.05% | 158 |
| 4 | 30,000 | 1.00% | 300 |
| 5 | 52,000 | 0.95% | 494 |

> Take rate declines with volume (rising enterprise/IC++ and QR mix) while total revenue grows with TPV.

---

## 5. Competitive Positioning & Growth

### 5.1 Competitors & positioning
- Competitors: Thai PSPs/gateways (facilitators riding existing acquirers, and bank acquirers) plus international players serving Thailand.
- **Our differentiation:**
  1. **Full acquiring** → better per-txn cost control than facilitators; genuine IC++ offering.
  2. **Compliance-by-design** → PCI-DSS v4.0 L1 + BOT requirements + PDPA embedded in architecture (double-entry ledger, audit trail, tokenization vault).
  3. **Transparent reconciliation & settlement** → append-only ledger as source of truth, per-merchant reconciliation reporting.
  4. **Developer experience** → clean API, idempotency, signed webhooks, sandbox.

### 5.2 Growth strategy (5-year)

| Period | Focus |
|--------|-------|
| **Y1 (2027)** | Launch post-licence + PCI L1; capture e-commerce SME; onboard ~400 merchants; prove reconciliation/settlement |
| **Y2 (2028)** | Expand recurring/SaaS + travel; add value-added (tokenization, fraud tools); platform/marketplace partnerships |
| **Y3 (2029)** | Enter enterprise (IC++, SLA); cross-border/multi-currency; EBITDA positive |
| **Y4 (2030)** | Omni-channel (in-store + online); add local payment methods; scale volume |
| **Y5 (2031)** | Lead target segments; consider adjacent services (e.g. payout); prepare regional expansion |

### 5.3 Key KPIs reported to BOT/internally
TPV, active merchants, authorization success rate, **chargeback rate (target < 0.9% per scheme threshold)**, fraud loss rate, uptime (target ≥ 99.95%), settlement accuracy, number of STRs filed to AMLO.

### 5.4 Strategic risks
Licence/PCI time and cost, sponsor-bank dependency, price competition, fraud/chargeback, and ongoing compliance (see `ARCHITECTURE.md` §9 and `ROADMAP.md` §5).

---

## 6. Roles & Governance

| Role | Key responsibilities |
|------|----------------------|
| CEO / Managing Director | Strategy, BOT/sponsor-bank relations; at least one Thai-national director resident in Thailand (BOT rule) |
| Compliance Officer / MLRO | AML/KYC, AMLO reporting, BOT requirements, QSA liaison |
| DPO (Data Protection Officer) | PDPA, data-subject rights, processing governance |
| CISO / DevSecOps Lead | PCI-DSS v4.0, HSM/KMS, network segmentation, pentest/ASV |
| Head of Risk / Underwriting | Merchant risk tiering, reserves, chargeback/fraud |
| Head of Engineering / SRE | Availability, ledger, reconciliation, DR |
| Head of Finance | Paid-up capital (maintain ≥75%), settlement, financial projections |

---

## 7. Assumptions / TODO

> **[ASSUMPTION / TODO — confirm before submission / go-live]**
>
> 1. **Sponsor bank / scheme certification** — sponsor bank not yet selected and Visa/Mastercard certification not formally started; fee figures and certification timeline are estimates to be confirmed after signing (see `ROADMAP.md`).
> 2. **QSA vendor** — QSA for PCI-DSS v4.0 L1 (RoC) and ASV for quarterly scans not yet selected; PCI costs are broad ranges (`ROADMAP.md` §4).
> 3. **Paid-up capital** — plan assumes the THB 50M Acquiring threshold, maintained at ≥75% throughout operations; actual capital payment and that company objectives cover this business must be confirmed.
> 4. **Interchange / scheme fee / MDR rates** — figures in Sections 3–4 are indicative and must be replaced with actual negotiated rates.
> 5. **TPV/revenue projections** — strategic planning assumptions, not guarantees; to be revised against actuals and pipeline.
