# โครงสร้างผู้ถือหุ้น คณะกรรมการ และ fit & proper (ไทย)

> เอกสารประกอบคำขอรับใบอนุญาต **การให้บริการรับชำระเงินด้วยวิธีการทางอิเล็กทรอนิกส์ (Acquiring — Full Acquiring Gateway)**
> ภายใต้ **พ.ร.บ. ระบบการชำระเงิน พ.ศ. 2560** กำกับโดย **ธนาคารแห่งประเทศไทย (ธปท.)** · ทุนจดทะเบียนชำระแล้ว 50 ล้านบาท
> เอกสารชุด: `03-shareholder-board-fit-proper.md` — เชื่อมโยงกับ `COMPLIANCE-TH.md`, `ARCHITECTURE.md`, `ROADMAP.md`
> **ข้อสงวน:** เอกสารนี้เป็นแม่แบบเชิงปฏิบัติ ไม่ใช่คำแนะนำทางกฎหมาย ต้องให้ที่ปรึกษากฎหมายด้านใบอนุญาต ธปท. ทวนสอบก่อนยื่นจริง

---

> **⚠️ สมมติฐาน / รายการค้าง (TODO) ที่ต้องปิดก่อนยื่นจริง**
>
> รายการต่อไปนี้เป็น *ตัวเลข/คู่ค้าจริง* ที่ยังไม่ผูกในเอกสารนี้ ห้ามกรอกด้วยข้อมูลสมมติเมื่อยื่น ธปท.
>
> - **[TODO-CAP]** ยอดทุนจดทะเบียนชำระแล้วจริง และหลักฐานการชำระ (บัญชีเงินฝาก/หนังสือรับรองธนาคาร) — ต้องไม่ต่ำกว่า **50,000,000 บาท** และคงไว้ **≥ 75%** ตลอดการดำเนินงาน
> - **[TODO-SPONSOR]** ธนาคารผู้สนับสนุน (sponsor bank) / คู่สัญญา card scheme (Visa/Mastercard) — ยังไม่สรุป มีผลต่อโครงสร้างผู้ถือหุ้นเชิงกลยุทธ์และเงื่อนไข scheme
> - **[TODO-QSA]** ผู้ประเมิน PCI-DSS (QSA vendor) ที่จะออก RoC — ยังไม่เลือก (ดู `ARCHITECTURE.md` §7)
> - **[TODO-UBO]** รายชื่อ UBO จริง สัดส่วนการถือหุ้นจริง และสายการถือหุ้น (ownership chain) จนถึงบุคคลธรรมดา — ต้องยืนยันด้วยบัญชีรายชื่อผู้ถือหุ้น (บอจ.5) ณ วันยื่น
> - **[TODO-FBL]** กรณีผู้ถือหุ้นต่างชาติเกินเกณฑ์ ต้องประเมินความจำเป็นของ Foreign Business License (FBL) ตาม พ.ร.บ. การประกอบธุรกิจของคนต่างด้าว พ.ศ. 2542
> - **[TODO-BOARD]** รายชื่อกรรมการ/ผู้บริหารระดับสูงจริง พร้อมประวัติและเอกสาร fit & proper ครบชุด

---

## 1. วัตถุประสงค์และขอบเขต

เอกสารนี้แสดงต่อ ธปท. ว่า **[บริษัท / Company]** มี (1) โครงสร้างผู้ถือหุ้นที่โปร่งใสและตรวจสอบได้ถึงผู้รับประโยชน์ที่แท้จริง (UBO) (2) คณะกรรมการและผู้บริหารที่มีองค์ประกอบเหมาะสม มีความเป็นอิสระ และมีความรู้ความสามารถ (3) กระบวนการประเมินคุณสมบัติและลักษณะต้องห้าม (fit & proper) ที่เป็นระบบ และ (4) การควบคุมการเปลี่ยนแปลงผู้ถือหุ้น/กรรมการอย่างต่อเนื่องหลังได้รับใบอนุญาต

ครอบคลุมสอดคล้องกับ:
- **พ.ร.บ. ระบบการชำระเงิน พ.ศ. 2560** และประกาศ ธปท. ที่เกี่ยวข้อง (ธรรมาภิบาล, ผู้มีอำนาจควบคุม, การรายงานการเปลี่ยนแปลง)
- **พ.ร.บ. ป้องกันและปราบปรามการฟอกเงิน** — สำนักงาน ปปง. (AMLO): การระบุ UBO ตามหลัก CDD
- **พ.ร.บ. คุ้มครองข้อมูลส่วนบุคคล พ.ศ. 2562 (PDPA)** — สำนักงาน PDPC: การเก็บ/ใช้ข้อมูลส่วนบุคคลของกรรมการ/ผู้ถือหุ้น
- **PCI-DSS v4.0** — ในส่วนบุคลากรที่เข้าถึงสภาพแวดล้อมข้อมูลบัตร (CDE) ต้องผ่านการตรวจประวัติ (Req. 12.x, personnel screening)

---

## 2. โครงสร้างผู้ถือหุ้น (Shareholder Structure)

### 2.1 หลักการ

- ทุนจดทะเบียนชำระแล้ว **≥ 50,000,000 บาท** (เกณฑ์บริการ Acquiring ตาม `COMPLIANCE-TH.md` §2) และคงไว้ **≥ 75%** ตลอดเวลา — มีการติดตามอัตราส่วนความเพียงพอของทุนทุกไตรมาส
- ผู้ถือหุ้นทุกทอดตรวจสอบได้จนถึง **บุคคลธรรมดา** ที่เป็น UBO
- กรรมการอย่างน้อย **1 คนเป็นสัญชาติไทยและมีถิ่นที่อยู่ในไทย** (ตาม §3 ของ `COMPLIANCE-TH.md`)
- โครงสร้างถือหุ้นแบบชั้นซ้อน (holding layer) ที่ทำให้ตรวจสอบ UBO ได้ยาก — **ไม่อนุญาต** เว้นแต่มีเหตุผลทางธุรกิจที่อธิบายได้และเปิดเผยครบ

### 2.2 ตารางผู้ถือหุ้น (แม่แบบ — เติมข้อมูลจริงจาก บอจ.5)

| ลำดับ | ผู้ถือหุ้น | ประเภท | สัญชาติ | จำนวนหุ้น | % ถือหุ้น | สิทธิออกเสียง % | หมายเหตุ |
|------|-----------|--------|---------|-----------|-----------|-----------------|----------|
| 1 | `[ผู้ก่อตั้ง/บุคคล A]` | บุคคลธรรมดา | ไทย | `[TODO-UBO]` | `[TODO-UBO]` | `[TODO-UBO]` | UBO |
| 2 | `[บริษัทโฮลดิ้ง B]` | นิติบุคคล | ไทย | `[TODO-UBO]` | `[TODO-UBO]` | `[TODO-UBO]` | ต้องเปิด ownership chain ต่อ |
| 3 | `[นักลงทุน C]` | นิติบุคคล/ต่างชาติ | `[TODO]` | `[TODO-UBO]` | `[TODO-UBO]` | `[TODO-UBO]` | ตรวจเกณฑ์ FBL `[TODO-FBL]` |
| — | **รวม** | | | | **100%** | **100%** | |

> สัดส่วนต่างชาติรวมทุกทอด: `[TODO-FBL]` — หากรวมแล้ว **≥ 50%** ต้องประเมินความจำเป็นของ FBL และผลต่อเงื่อนไข ธปท.

### 2.3 เกณฑ์สัดส่วนที่ต้องรายงาน/ขออนุมัติ

| เหตุการณ์ | เกณฑ์ | การดำเนินการ |
|-----------|-------|--------------|
| ผู้ถือหุ้นรายใหญ่ (major shareholder) | ถือ **≥ 10%** ของทุนหรือสิทธิออกเสียง | ต้องผ่าน fit & proper และแจ้ง ธปท. |
| การเปลี่ยนผู้มีอำนาจควบคุม | ได้มา/เปลี่ยนอำนาจควบคุมกิจการ | **แจ้ง/ขออนุมัติ ธปท. ล่วงหน้า** ตามประกาศที่เกี่ยวข้อง |
| การเปลี่ยนสัดส่วน ≥ 5% | เพิ่ม/ลดข้ามขั้น 5%, 10%, 25%, 50% | บันทึกในทะเบียน และประเมินผลต่อเงื่อนไขใบอนุญาต |
| การลดทุน | ทำให้ทุนต่ำกว่าเกณฑ์หรือต่ำกว่า 75% | **ห้าม** เว้นแต่ได้รับอนุมัติ ธปท. ก่อน |

---

## 3. ผู้รับประโยชน์ที่แท้จริง (Ultimate Beneficial Ownership — UBO)

### 3.1 นิยามและเกณฑ์การระบุ (สอดคล้อง ปปง./AMLO)

**UBO** = บุคคลธรรมดาที่ (ก) ถือหุ้นหรือมีสิทธิออกเสียงทั้งทางตรง/ทางอ้อม **≥ 25%** หรือ (ข) ใช้อำนาจควบคุมกิจการโดยวิธีอื่น (เช่น สิทธิแต่งตั้งกรรมการข้างมาก ข้อตกลงผู้ถือหุ้น) หรือ (ค) กรณีไม่พบตาม (ก)/(ข) ให้ระบุ **ผู้บริหารระดับสูงสุด** เป็น UBO โดยอนุโลม

### 3.2 ขั้นตอนพิสูจน์ UBO

1. รวบรวมบัญชีรายชื่อผู้ถือหุ้น (บอจ.5) และหนังสือรับรองบริษัททุกทอด
2. สร้าง **ownership chain diagram** จากนิติบุคคลจนถึงบุคคลธรรมดา คำนวณสัดส่วนแบบคูณต่อทอด (multiplication of holdings)
3. ระบุ UBO ทุกรายที่ผ่านเกณฑ์ 25% หรือมีอำนาจควบคุม
4. คัดกรอง UBO กับ **sanction list** (UN, OFAC, EU) และ **PEP** (Politically Exposed Persons) ก่อนยื่นและทบทวนอย่างน้อยปีละครั้ง
5. จัดเก็บเอกสารระบุตัวตน (บัตรประชาชน/หนังสือเดินทาง) ภายใต้ PDPA (ดู §7)

### 3.3 ตาราง UBO (แม่แบบ)

| UBO | สัญชาติ | เส้นทางการถือหุ้น | สัดส่วนทางอ้อมรวม % | PEP? | ผล sanction/PEP screening | วันที่คัดกรองล่าสุด |
|-----|---------|-------------------|--------------------|------|---------------------------|---------------------|
| `[บุคคล A]` | ไทย | ถือตรง | `[TODO-UBO]` | ไม่/ใช่ | Clear/`[TODO]` | `[TODO]` |
| `[บุคคล X]` | `[TODO]` | ผ่าน `[โฮลดิ้ง B]` | `[TODO-UBO]` | ไม่/ใช่ | Clear/`[TODO]` | `[TODO]` |

---

## 4. องค์ประกอบคณะกรรมการ (Board Composition)

### 4.1 หลักการธรรมาภิบาล

- ขนาดคณะกรรมการ: **5–9 คน** เพื่อความคล่องตัวและถ่วงดุล
- กรรมการอิสระ **≥ 1 ใน 3** ของจำนวนกรรมการทั้งหมด
- แยกบทบาท **ประธานกรรมการ** ออกจาก **ประธานเจ้าหน้าที่บริหาร (CEO)**
- กรรมการ **≥ 1 คนเป็นสัญชาติไทยและมีถิ่นที่อยู่ในไทย** (ข้อกำหนดคุณสมบัติผู้ขอ)
- มีคณะกรรมการชุดย่อยด้าน **ตรวจสอบ (Audit)** และ **บริหารความเสี่ยง (Risk)** เป็นอย่างน้อย

### 4.2 องค์ประกอบและความรับผิดชอบ (แม่แบบ)

| ตำแหน่ง | ประเภท | คุณสมบัติหลัก | ความรับผิดชอบด้านกำกับ |
|---------|--------|---------------|------------------------|
| ประธานกรรมการ | Non-executive | ประสบการณ์บริหารสถาบันการเงิน/payment | กำกับทิศทาง, ธรรมาภิบาล |
| กรรมการอิสระ 1 | Independent | ผู้เชี่ยวชาญบัญชี/การเงิน | ประธานคณะกรรมการตรวจสอบ |
| กรรมการอิสระ 2 | Independent | ผู้เชี่ยวชาญกฎหมาย/compliance | กำกับ AML/PDPA |
| CEO | Executive | ประสบการณ์ payment/fintech | รับผิดชอบผลการดำเนินงานรวม |
| กรรมการ (ไทย, residency) | Executive/Non-exec | สัญชาติไทย มีถิ่นที่อยู่ในไทย | เงื่อนไขคุณสมบัติผู้ขอ |
| กรรมการด้านเทคโนโลยี | Executive | IT security/PCI/architecture | กำกับ cyber resilience, PCI-DSS v4.0 |

> **[TODO-BOARD]** เติมรายชื่อจริง พร้อมประวัติ (CV), หนังสือรับรองการเป็นกรรมการ และแบบ fit & proper รายบุคคล

### 4.3 บทบาทผู้บริหารสำคัญที่ต้องผ่าน fit & proper

CEO, CFO, **Chief Compliance Officer / AMLCO** (เจ้าหน้าที่กำกับ AML ตาม ปปง.), **CISO** (ผู้รับผิดชอบ PCI-DSS/cyber), **DPO** (Data Protection Officer ตาม PDPA), Head of Internal Audit, Head of Risk

---

## 5. คุณสมบัติและลักษณะต้องห้าม (Fit & Proper Criteria)

การประเมิน fit & proper ใช้กับ **กรรมการ ผู้มีอำนาจในการจัดการ ผู้ถือหุ้นรายใหญ่ (≥10%) และ UBO** ครอบคลุม 4 มิติ:

### 5.1 ความซื่อสัตย์และชื่อเสียง (Honesty & Integrity)
- ไม่เคยต้องคำพิพากษาถึงที่สุดในความผิดเกี่ยวกับทรัพย์ ทุจริต ฉ้อโกง ฟอกเงิน
- ไม่เป็นบุคคลล้มละลาย/พิทักษ์ทรัพย์
- ไม่เคยถูกเพิกถอน/พักใบอนุญาตด้านการเงินโดยหน่วยงานกำกับ
- ไม่อยู่ในบัญชี sanction หรือรายชื่อผู้กระทำผิดของ ปปง.

### 5.2 ความรู้ ความสามารถ และประสบการณ์ (Competence & Capability)
- มีคุณวุฒิ/ประสบการณ์เหมาะสมกับบทบาท (การเงิน, payment, IT security, compliance)
- สามารถอุทิศเวลาให้กับหน้าที่ได้เพียงพอ

### 5.3 ฐานะการเงิน (Financial Soundness)
- ไม่มีประวัติผิดนัดชำระหนี้ร้ายแรง/ถูกฟ้องล้มละลาย

### 5.4 ความขัดแย้งทางผลประโยชน์ (Conflict of Interest & Independence)
- เปิดเผยการถือหุ้น/ตำแหน่งในกิจการที่เกี่ยวโยงกัน
- กรรมการอิสระต้องไม่มีความสัมพันธ์ที่กระทบความเป็นอิสระ

### 5.5 ตารางลักษณะต้องห้าม (Disqualification Matrix)

| ลักษณะต้องห้าม | ผล |
|----------------|-----|
| เคยต้องโทษคดีทุจริต/ฟอกเงิน (คำพิพากษาถึงที่สุด) | ขาดคุณสมบัติ |
| เป็นบุคคลล้มละลาย/ไร้ความสามารถ | ขาดคุณสมบัติ |
| ถูกหน่วยงานกำกับเพิกถอน/พักใบอนุญาต | ขาดคุณสมบัติ (ตามระยะเวลาที่กำหนด) |
| อยู่ใน sanction list (UN/OFAC/EU) | ขาดคุณสมบัติ |
| ปกปิดข้อมูลสำคัญในแบบ fit & proper | ขาดคุณสมบัติ + ทบทวนความน่าเชื่อถือ |

---

## 6. กระบวนการและวงจร fit & proper

### 6.1 ขั้นตอน (Onboarding)

1. ผู้ได้รับการเสนอชื่อกรอก **แบบประเมิน fit & proper** + ให้ความยินยอมตรวจประวัติ (PDPA consent)
2. **Compliance** ตรวจ: sanction/PEP screening, ประวัติอาชญากรรม (หนังสือรับรองความประพฤติ), เครดิตบูโร (กรณีที่เกี่ยวข้อง), ทะเบียนล้มละลาย, ฐานข้อมูลหน่วยกำกับ
3. คณะกรรมการสรรหา/Nomination พิจารณาและเสนอคณะกรรมการอนุมัติ
4. **แจ้ง/ยื่นต่อ ธปท.** ตามที่ประกาศกำหนดก่อนเข้าดำรงตำแหน่ง (สำหรับกรรมการ/ผู้มีอำนาจจัดการ/ผู้ถือหุ้นรายใหญ่)
5. จัดเก็บเอกสารในทะเบียน fit & proper แบบควบคุมการเข้าถึง

### 6.2 การทบทวนต่อเนื่อง (Ongoing)

| กิจกรรม | ความถี่ | ผู้รับผิดชอบ |
|---------|---------|-------------|
| ทบทวน fit & proper รายบุคคล | อย่างน้อยปีละครั้ง | Compliance |
| Sanction/PEP re-screening (กรรมการ, UBO, ผู้ถือหุ้น ≥10%) | อย่างน้อยปีละครั้ง + เมื่อ list เปลี่ยน | Compliance/AMLCO |
| ปรับปรุงทะเบียนผู้ถือหุ้น/UBO | เมื่อมีการเปลี่ยนแปลง (ภายใน **15 วันทำการ**) | Corporate Secretary |
| รายงานการเปลี่ยนกรรมการ/ผู้ถือหุ้นรายใหญ่ต่อ ธปท. | ตามกำหนดในประกาศ (ปกติ **ล่วงหน้า/ทันทีที่เกิดเหตุ**) | Compliance |
| ทบทวนความเป็นอิสระของกรรมการอิสระ | ปีละครั้ง | Nomination Committee |

### 6.3 การรายงานเหตุการณ์ (Event-driven)

หากพบว่าบุคคลใด **ขาดคุณสมบัติ/มีลักษณะต้องห้ามภายหลัง** → ระงับอำนาจ, รายงานคณะกรรมการและ ธปท. โดยไม่ชักช้า, จัดหาผู้ทดแทนที่ผ่าน fit & proper

---

## 7. การคุ้มครองข้อมูลส่วนบุคคล (PDPA) ในกระบวนการนี้

- เก็บข้อมูลกรรมการ/ผู้ถือหุ้น/UBO **เท่าที่จำเป็น** ตามฐานทางกฎหมาย (การปฏิบัติตามกฎหมาย + ความยินยอม)
- แจ้ง **privacy notice** ก่อนตรวจประวัติ และขอความยินยอมสำหรับข้อมูลอ่อนไหว (เช่น ประวัติอาชญากรรม)
- เก็บรักษาแบบเข้ารหัส จำกัดการเข้าถึง (RBAC) และกำหนดระยะเวลาเก็บ (retention) ตามข้อกำหนด ธปท./PDPA
- **DPO** กำกับดูแลการประมวลผลข้อมูลนี้ และรองรับสิทธิของเจ้าของข้อมูลตาม PDPA

---

## 8. เอกสารประกอบที่ต้องแนบกับคำขอ (Checklist)

- [ ] บัญชีรายชื่อผู้ถือหุ้น (บอจ.5) ล่าสุด + หนังสือรับรองบริษัท (วัตถุประสงค์ครอบคลุมธุรกิจ payment)
- [ ] หลักฐานทุนจดทะเบียนชำระแล้ว ≥ 50 ล้านบาท `[TODO-CAP]`
- [ ] แผนภาพโครงสร้างผู้ถือหุ้นและ ownership chain ถึง UBO `[TODO-UBO]`
- [ ] ตาราง UBO + ผลคัดกรอง sanction/PEP
- [ ] รายชื่อคณะกรรมการ + CV + หนังสือรับรองการดำรงตำแหน่ง `[TODO-BOARD]`
- [ ] แบบ fit & proper รายบุคคลครบทุกคน + เอกสารตรวจประวัติ
- [ ] กฎบัตรคณะกรรมการและคณะกรรมการชุดย่อย (Audit, Risk)
- [ ] ผลประเมิน FBL (กรณีต่างชาติ) `[TODO-FBL]`

---

# Shareholder structure, board composition, fit-and-proper declarations, ultimate beneficial ownership (English)

> Supporting document for the license application for **Electronic Payment Acquiring Service (Full Acquiring Gateway)**
> under the **Payment Systems Act B.E. 2560 (2017)**, regulated by the **Bank of Thailand (BOT)** · paid-up capital THB 50 million.
> Document set: `03-shareholder-board-fit-proper.md` — cross-references `COMPLIANCE-TH.md`, `ARCHITECTURE.md`, `ROADMAP.md`.
> **Disclaimer:** This is a practical template, not legal advice. It must be reviewed by BOT-licensing legal counsel before submission.

---

> **⚠️ ASSUMPTIONS / OPEN TODOs to close before actual submission**
>
> The following are *real figures/counterparties* not yet bound in this document. Do NOT populate them with invented data when filing with the BOT.
>
> - **[TODO-CAP]** Actual paid-up capital and proof of payment (bank statement / bank confirmation letter) — must be **≥ THB 50,000,000** and maintained at **≥ 75%** throughout operations.
> - **[TODO-SPONSOR]** Sponsor bank / card scheme counterparty (Visa/Mastercard) — not yet finalized; affects strategic shareholding and scheme conditions.
> - **[TODO-QSA]** PCI-DSS assessor (QSA vendor) issuing the RoC — not yet selected (see `ARCHITECTURE.md` §7).
> - **[TODO-UBO]** Actual UBO names, real shareholding percentages, and the full ownership chain down to natural persons — must be verified against the shareholder register (BorOorJor.5 / "บอจ.5") as of the filing date.
> - **[TODO-FBL]** If foreign shareholding exceeds the threshold, assess whether a Foreign Business License (FBL) is required under the Foreign Business Act B.E. 2542.
> - **[TODO-BOARD]** Actual directors / senior executives with CVs and complete fit & proper documentation.

---

## 1. Purpose and scope

This document demonstrates to the BOT that **[บริษัท / Company]** has (1) a transparent shareholder structure traceable to its ultimate beneficial owners (UBOs), (2) a board and management with appropriate composition, independence, and competence, (3) a systematic fit-and-proper assessment process, and (4) ongoing controls over shareholder/director changes after licensing.

It is aligned with:
- **Payment Systems Act B.E. 2560** and related BOT notifications (governance, controlling persons, change reporting).
- **Anti-Money Laundering Act** — Anti-Money Laundering Office (AMLO): UBO identification per CDD principles.
- **Personal Data Protection Act B.E. 2562 (PDPA)** — PDPC: processing of directors'/shareholders' personal data.
- **PCI-DSS v4.0** — personnel accessing the cardholder data environment (CDE) must undergo screening (Req. 12.x personnel screening).

---

## 2. Shareholder structure

### 2.1 Principles

- Paid-up capital **≥ THB 50,000,000** (Acquiring threshold per `COMPLIANCE-TH.md` §2), maintained at **≥ 75%** at all times — capital adequacy monitored quarterly.
- Every tier of ownership is traceable to the **natural persons** who are UBOs.
- At least **1 director must be a Thai national resident in Thailand** (per §3 of `COMPLIANCE-TH.md`).
- Nested holding layers that obscure UBO traceability are **not permitted** unless a defensible business rationale is fully disclosed.

### 2.2 Shareholder table (template — populate from BorOorJor.5)

| # | Shareholder | Type | Nationality | Shares | Ownership % | Voting % | Notes |
|---|-------------|------|-------------|--------|-------------|----------|-------|
| 1 | `[Founder/Person A]` | Individual | Thai | `[TODO-UBO]` | `[TODO-UBO]` | `[TODO-UBO]` | UBO |
| 2 | `[Holding Co. B]` | Legal entity | Thai | `[TODO-UBO]` | `[TODO-UBO]` | `[TODO-UBO]` | Expand ownership chain |
| 3 | `[Investor C]` | Legal/foreign | `[TODO]` | `[TODO-UBO]` | `[TODO-UBO]` | `[TODO-UBO]` | Check FBL `[TODO-FBL]` |
| — | **Total** | | | | **100%** | **100%** | |

> Aggregate foreign ownership across all tiers: `[TODO-FBL]` — if **≥ 50%**, assess FBL necessity and impact on BOT conditions.

### 2.3 Thresholds requiring reporting/approval

| Event | Threshold | Action |
|-------|-----------|--------|
| Major shareholder | Holds **≥ 10%** of capital or voting rights | Must pass fit & proper and be reported to BOT |
| Change of controlling person | Acquisition/change of control of the company | **Prior BOT notification/approval** per relevant notification |
| Ownership change ≥ 5% | Crossing 5%, 10%, 25%, 50% bands | Record in register; assess impact on license conditions |
| Capital reduction | Below threshold or below 75% | **Prohibited** unless BOT approval obtained in advance |

---

## 3. Ultimate Beneficial Ownership (UBO)

### 3.1 Definition and identification (aligned with AMLO)

**UBO** = a natural person who (a) directly/indirectly holds shares or voting rights **≥ 25%**, or (b) exercises control by other means (e.g., right to appoint the majority of directors, shareholder agreements), or (c) where none can be identified under (a)/(b), the **most senior managing official** is designated as UBO.

### 3.2 UBO verification procedure

1. Collect shareholder registers (BorOorJor.5) and company affidavits at every tier.
2. Build an **ownership chain diagram** from legal entities down to natural persons; compute indirect stakes by multiplication of holdings across tiers.
3. Identify all UBOs meeting the 25% threshold or exercising control.
4. Screen UBOs against **sanction lists** (UN, OFAC, EU) and **PEP** lists before filing and review at least annually.
5. Store identity documents (national ID/passport) under PDPA controls (see §7).

### 3.3 UBO table (template)

| UBO | Nationality | Ownership path | Aggregate indirect % | PEP? | Sanction/PEP screening result | Last screened |
|-----|-------------|----------------|----------------------|------|-------------------------------|---------------|
| `[Person A]` | Thai | Direct | `[TODO-UBO]` | No/Yes | Clear/`[TODO]` | `[TODO]` |
| `[Person X]` | `[TODO]` | via `[Holding B]` | `[TODO-UBO]` | No/Yes | Clear/`[TODO]` | `[TODO]` |

---

## 4. Board composition

### 4.1 Governance principles

- Board size: **5–9 members** for agility and balance.
- Independent directors **≥ one-third** of the board.
- Separate the roles of **Chairman** and **Chief Executive Officer (CEO)**.
- At least **1 director is a Thai national resident in Thailand** (applicant eligibility requirement).
- At minimum, an **Audit Committee** and a **Risk Committee** as board subcommittees.

### 4.2 Composition and responsibilities (template)

| Position | Type | Key qualifications | Governance responsibility |
|----------|------|--------------------|---------------------------|
| Chairman | Non-executive | Financial institution / payments management experience | Strategic oversight, governance |
| Independent Director 1 | Independent | Accounting/finance expert | Chair of Audit Committee |
| Independent Director 2 | Independent | Legal/compliance expert | AML/PDPA oversight |
| CEO | Executive | Payments/fintech experience | Overall performance |
| Director (Thai, resident) | Executive/Non-exec | Thai national resident in Thailand | Applicant eligibility condition |
| Technology Director | Executive | IT security/PCI/architecture | Cyber resilience, PCI-DSS v4.0 |

> **[TODO-BOARD]** Populate with real names plus CVs, director certificates, and individual fit & proper forms.

### 4.3 Key executives requiring fit & proper

CEO, CFO, **Chief Compliance Officer / AMLCO** (AML compliance officer per AMLO), **CISO** (PCI-DSS/cyber owner), **DPO** (Data Protection Officer under PDPA), Head of Internal Audit, Head of Risk.

---

## 5. Fit-and-proper criteria

Fit-and-proper assessment applies to **directors, persons with management authority, major shareholders (≥10%), and UBOs**, across 4 dimensions:

### 5.1 Honesty & integrity
- No final judgment for offenses involving property, dishonesty, fraud, or money laundering.
- Not a bankrupt / under receivership.
- Never had a financial license revoked/suspended by a regulator.
- Not on any sanction list or AMLO offender list.

### 5.2 Competence & capability
- Qualifications/experience appropriate to the role (finance, payments, IT security, compliance).
- Able to dedicate sufficient time to the role.

### 5.3 Financial soundness
- No record of serious debt default / bankruptcy litigation.

### 5.4 Conflict of interest & independence
- Disclose holdings/positions in related businesses.
- Independent directors must have no relationship impairing independence.

### 5.5 Disqualification matrix

| Disqualifying characteristic | Effect |
|------------------------------|--------|
| Prior conviction for fraud/money laundering (final judgment) | Disqualified |
| Bankrupt / legally incapacitated | Disqualified |
| License revoked/suspended by a regulator | Disqualified (for the prescribed period) |
| On a sanction list (UN/OFAC/EU) | Disqualified |
| Concealment of material information on the fit & proper form | Disqualified + credibility review |

---

## 6. Fit-and-proper process and lifecycle

### 6.1 Onboarding steps

1. Nominee completes the **fit & proper questionnaire** and grants background-check consent (PDPA consent).
2. **Compliance** verifies: sanction/PEP screening, criminal record (police clearance), credit bureau (where relevant), bankruptcy register, regulator databases.
3. Nomination Committee reviews and recommends to the Board for approval.
4. **Notify/file with the BOT** as prescribed before assuming office (for directors/persons with management authority/major shareholders).
5. Store documents in an access-controlled fit & proper register.

### 6.2 Ongoing review

| Activity | Frequency | Owner |
|----------|-----------|-------|
| Individual fit & proper review | At least annually | Compliance |
| Sanction/PEP re-screening (directors, UBOs, shareholders ≥10%) | At least annually + on list changes | Compliance/AMLCO |
| Update shareholder/UBO register | On change (within **15 business days**) | Corporate Secretary |
| Report director/major-shareholder changes to BOT | As prescribed (typically **in advance / immediately on occurrence**) | Compliance |
| Review independence of independent directors | Annually | Nomination Committee |

### 6.3 Event-driven reporting

If a person is later found **disqualified / to have a prohibited characteristic** → suspend authority, report to the Board and BOT without delay, and appoint a fit-and-proper replacement.

---

## 7. Personal data protection (PDPA) in this process

- Collect director/shareholder/UBO data **only as necessary** under a lawful basis (legal obligation + consent).
- Provide a **privacy notice** before background checks and obtain consent for sensitive data (e.g., criminal record).
- Store encrypted, restrict access (RBAC), and set retention periods per BOT/PDPA requirements.
- The **DPO** oversees this processing and supports data-subject rights under PDPA.

---

## 8. Application attachments (checklist)

- [ ] Latest shareholder register (BorOorJor.5) + company affidavit (objectives covering payment business)
- [ ] Proof of paid-up capital ≥ THB 50M `[TODO-CAP]`
- [ ] Shareholder structure diagram and ownership chain to UBO `[TODO-UBO]`
- [ ] UBO table + sanction/PEP screening results
- [ ] Board list + CVs + director certificates `[TODO-BOARD]`
- [ ] Individual fit & proper forms for all persons + background-check documents
- [ ] Board and subcommittee charters (Audit, Risk)
- [ ] FBL assessment result (if foreign ownership applies) `[TODO-FBL]`
