# นโยบาย AML/KYC/CDD (ไทย)

> เอกสารเลขที่ **05** ในชุดเอกสาร Compliance สำหรับการยื่นขอใบอนุญาต **การให้บริการรับชำระเงินด้วยวิธีการทางอิเล็กทรอนิกส์ (Full Acquiring)** ภายใต้ **พ.ร.บ. ระบบการชำระเงิน พ.ศ. 2560** กำกับโดย **ธนาคารแห่งประเทศไทย (ธปท.)** และคู่ขนานกับ **PCI-DSS Level 1**
>
> **สถานะเอกสาร:** ฉบับร่างเพื่อยื่นขออนุมัติ (submission draft) · เวอร์ชัน 0.1 · วันที่ปรับปรุง 2026-07-22
> **เจ้าของเอกสาร:** เจ้าหน้าที่กำกับดูแลการปฏิบัติงานและป้องกันการฟอกเงิน (Compliance & MLRO) ของ **[บริษัท / Company]**
>
> เอกสารนี้เป็นนโยบายภายในและเอกสารประกอบคำขอใบอนุญาต **มิใช่คำแนะนำทางกฎหมาย** — ต้องผ่านการตรวจทานโดยที่ปรึกษากฎหมายและ MLRO ก่อนบังคับใช้จริง เนื่องจากประกาศ/หลักเกณฑ์ของ ปปง. และ ธปท. อาจปรับปรุงได้

---

## บทสรุปสำหรับผู้บริหาร (Executive Summary)

[บริษัท / Company] ในฐานะผู้ให้บริการรับชำระเงินด้วยบัตร (acquiring payment gateway) มีหน้าที่ตามกฎหมายในฐานะ **"สถาบันการเงิน" ตาม พ.ร.บ. ป้องกันและปราบปรามการฟอกเงิน พ.ศ. 2542 และที่แก้ไขเพิ่มเติม** และในฐานะ **ผู้ได้รับใบอนุญาตประกอบธุรกิจบริการการชำระเงินภายใต้การกำกับของ ธปท.** นโยบายฉบับนี้กำหนดกรอบการบริหารความเสี่ยงด้านการฟอกเงิน (ML) และการสนับสนุนทางการเงินแก่การก่อการร้ายและการแพร่ขยายอาวุธที่มีอานุภาพทำลายล้างสูง (TF/PF) ครอบคลุม:

1. การรู้จักลูกค้า (KYC) และการตรวจสอบเพื่อทราบข้อเท็จจริงเกี่ยวกับลูกค้า (CDD) ในการรับผู้ค้า (merchant onboarding)
2. การตรวจสอบเข้มข้น (EDD) สำหรับลูกค้า/ธุรกรรมความเสี่ยงสูง
3. การเฝ้าระวังธุรกรรมต่อเนื่อง (ongoing monitoring) และการรายงานธุรกรรมต่อ ปปง.
4. การเก็บรักษาบันทึกและเอกสาร (record keeping)
5. บทบาทหน้าที่ กำกับดูแล และการฝึกอบรม

> **ขอบเขตลูกค้า:** ลูกค้าหลักของ [บริษัท / Company] คือ **ผู้ค้า (merchants)** ที่รับชำระเงินผ่านเรา ไม่ใช่ผู้ถือบัตรปลายทาง (cardholder) โดยตรง อย่างไรก็ดี เรามีหน้าที่เฝ้าระวังพฤติกรรมธุรกรรมของผู้ถือบัตรที่ไหลผ่านผู้ค้า (transaction laundering / merchant-based money laundering) ด้วย

---

## 1. ฐานกฎหมายและมาตรฐานอ้างอิง

| กฎหมาย / มาตรฐาน | สาระสำคัญที่เกี่ยวข้องกับนโยบายนี้ |
|---|---|
| **พ.ร.บ. ป้องกันและปราบปรามการฟอกเงิน พ.ศ. 2542** (และที่แก้ไขเพิ่มเติม) | หน้าที่ CDD, การรายงานธุรกรรม, การเก็บรักษาข้อมูล, กำกับโดย **สำนักงาน ปปง. (AMLO)** |
| **พ.ร.บ. ป้องกันและปราบปรามการสนับสนุนทางการเงินแก่การก่อการร้ายและการแพร่ขยายอาวุธฯ พ.ศ. 2559 (CTPF Act)** | การตรวจรายชื่อบุคคลที่ถูกกำหนด (designated persons) และการระงับการดำเนินการกับทรัพย์สินทันที |
| **กฎกระทรวงว่าด้วยการตรวจสอบเพื่อทราบข้อเท็จจริงเกี่ยวกับลูกค้า พ.ศ. 2563** | หลักเกณฑ์ CDD/EDD, การจัดระดับความเสี่ยง, ผู้รับผลประโยชน์ที่แท้จริง (Beneficial Owner) |
| **พ.ร.บ. ระบบการชำระเงิน พ.ศ. 2560** | เงื่อนไขการประกอบธุรกิจภายใต้ ธปท. รวมถึงการมีระบบ AML/KYC ที่ได้มาตรฐานเป็นเงื่อนไขใบอนุญาต |
| **ประกาศ ธปท. ด้านการบริหารความเสี่ยงและ IT/Cyber** | การควบคุมภายใน การกำกับดูแล และความมั่นคงปลอดภัยของระบบที่รองรับกระบวนการ AML |
| **พ.ร.บ. คุ้มครองข้อมูลส่วนบุคคล พ.ศ. 2562 (PDPA)** — กำกับโดย **คณะกรรมการคุ้มครองข้อมูลส่วนบุคคล (PDPC)** | การเก็บ ใช้ เปิดเผยข้อมูล KYC ต้องมีฐานทางกฎหมายและมาตรการคุ้มครองข้อมูล |
| **PCI-DSS v4.0** | การคุ้มครองข้อมูลบัตร (CHD/SAD) ที่ใช้ประกอบการเฝ้าระวังธุรกรรม โดยไม่จัดเก็บเกินความจำเป็น |
| **มาตรฐาน FATF 40 Recommendations** | หลักการ Risk-Based Approach, Travel Rule, sanctions screening (ใช้เป็นแนวปฏิบัติสากล) |
| **EMV 3-D Secure (3DS) 2.x** | ข้อมูลยืนยันตัวตนผู้ถือบัตรที่ใช้เสริมการประเมินความเสี่ยงธุรกรรม |

> **[TODO/ข้อสมมติ]** ต้องยืนยันกับที่ปรึกษากฎหมายว่าประกาศ/แนวปฏิบัติฉบับล่าสุดของ ปปง. และ ธปท. ณ วันยื่นคำขอ มีการแก้ไขเลขที่/ปีใดบ้าง และปรับอ้างอิงให้ตรง

---

## 2. หลักการบริหารความเสี่ยงตามระดับความเสี่ยง (Risk-Based Approach)

[บริษัท / Company] ใช้แนวทาง **Risk-Based Approach (RBA)** จัดสรรทรัพยากรตามระดับความเสี่ยง โดยประเมินตามปัจจัย 4 กลุ่ม:

| ปัจจัยความเสี่ยง | ตัวอย่างตัวชี้วัด |
|---|---|
| **ความเสี่ยงด้านลูกค้า (Customer)** | ประเภทนิติบุคคล/บุคคลธรรมดา, PEP, โครงสร้างการถือหุ้นซับซ้อน, ผู้ถือหุ้นต่างชาติ |
| **ความเสี่ยงด้านผลิตภัณฑ์/ช่องทาง (Product/Channel)** | card-not-present, cross-border, high-ticket, e-commerce ที่ไม่พบหน้า |
| **ความเสี่ยงด้านภูมิศาสตร์ (Geographic)** | ประเทศใน FATF grey/black list, ประเทศที่มีมาตรการคว่ำบาตร |
| **ความเสี่ยงด้านประเภทธุรกิจ (MCC/Industry)** | crypto, gambling, forex, adult, nutraceutical, marketplace |

### 2.1 การจัดระดับความเสี่ยงลูกค้า (Merchant Risk Rating)

| ระดับ | เกณฑ์ (ตัวอย่าง) | มาตรการ CDD | รอบทบทวน |
|---|---|---|---|
| **ต่ำ (Low)** | ธุรกิจในไทย, MCC ความเสี่ยงต่ำ, เจ้าของ/BO ชัดเจน, ปริมาณคาดการณ์ต่ำ | Standard CDD | ทุก 36 เดือน |
| **ปานกลาง (Medium)** | ปริมาณปานกลาง, ขายออนไลน์ทั่วไป, มีปัจจัยเสี่ยงเดี่ยว | Standard CDD + monitoring เข้มขึ้น | ทุก 24 เดือน |
| **สูง (High)** | MCC ความเสี่ยงสูง, cross-border, PEP เกี่ยวข้อง, marketplace/PayFac ย่อย | **EDD** + อนุมัติโดย MLRO | ทุก 12 เดือน |
| **ห้ามรับ (Prohibited)** | ธุรกิจผิดกฎหมาย, อยู่ในบัญชีคว่ำบาตร, ปฏิเสธการเปิดเผย BO | **ไม่รับ / เลิกความสัมพันธ์** | — |

> รายการธุรกิจต้องห้าม (prohibited list) และรายการเฝ้าระวัง (restricted/high-risk MCC list) ให้อ้างอิงภาคผนวก A ซึ่งทบทวนอย่างน้อยปีละครั้งและเมื่อ card scheme/sponsor bank ปรับปรุงเงื่อนไข

---

## 3. การรับลูกค้า / Onboarding (KYC & CDD)

### 3.1 เอกสารและข้อมูลที่ต้องเก็บ (Merchant นิติบุคคล)

| รายการ | รายละเอียด |
|---|---|
| หนังสือรับรองบริษัท (ไม่เกิน 3 เดือน) | เลขทะเบียนนิติบุคคล, กรรมการผู้มีอำนาจ, วัตถุประสงค์ |
| บัญชีรายชื่อผู้ถือหุ้น (บอจ.5) | ระบุสัดส่วนการถือหุ้น |
| บัตรประชาชน/หนังสือเดินทางของกรรมการผู้มีอำนาจและ **ผู้รับผลประโยชน์ที่แท้จริง (BO)** | BO = ผู้ถือหุ้น/ควบคุมตั้งแต่ **25%** ขึ้นไป หรือผู้มีอำนาจควบคุมที่แท้จริง |
| เอกสารที่อยู่สถานประกอบการ | สัญญาเช่า/ทะเบียนบ้าน/บิลสาธารณูปโภค |
| ข้อมูลบัญชีธนาคารเพื่อ settlement | ต้องเป็นบัญชีในชื่อนิติบุคคลเดียวกัน |
| เว็บไซต์/URL, ตัวอย่างสินค้า, นโยบายคืนเงิน | เพื่อประเมิน MCC และความเสี่ยง transaction laundering |
| ประมาณการปริมาณธุรกรรม (expected volume/ticket size) | ใช้ตั้ง baseline สำหรับ ongoing monitoring |

**บุคคลธรรมดา (sole proprietor):** บัตรประชาชน, การพิสูจน์ตัวตน (identity verification), ทะเบียนพาณิชย์ (ถ้ามี), ข้อมูลบัญชี settlement

### 3.2 กระบวนการพิสูจน์และยืนยันตัวตน

1. **Identification** — เก็บข้อมูลระบุตัวตนตามข้อ 3.1
2. **Verification** — ตรวจสอบกับแหล่งข้อมูลที่น่าเชื่อถือ (กรมพัฒนาธุรกิจการค้า/DBD, ระบบ e-KYC, การตรวจสอบเอกสาร, liveness check สำหรับบุคคล)
3. **Beneficial Owner** — ระบุและตรวจสอบ BO ทุกรายที่ถือ/ควบคุม ≥ 25% หรือมีอำนาจควบคุมที่แท้จริง จนถึงบุคคลธรรมดา
4. **Sanctions & PEP screening** — คัดกรองชื่อกับบัญชีรายชื่อ (ดูข้อ 4)
5. **Risk rating** — ให้คะแนนและจัดระดับตามข้อ 2.1
6. **Approval** — ระดับต่ำ/ปานกลางอนุมัติโดยทีม Onboarding; ระดับสูงต้องอนุมัติโดย **MLRO/Compliance Committee**

> **หลักการ:** ห้ามเปิดใช้งานรับชำระเงินจริง (go-live) ก่อน CDD และ screening ผ่านครบถ้วน (no onboarding without completed CDD)

### 3.3 กรอบเวลา (SLA) การ Onboarding

| ระดับความเสี่ยง | SLA เป้าหมายสำหรับ CDD | ผู้อนุมัติ |
|---|---|---|
| ต่ำ | 1-2 วันทำการ | Onboarding Officer |
| ปานกลาง | 3-5 วันทำการ | Onboarding Lead |
| สูง (EDD) | 7-15 วันทำการ | MLRO |

---

## 4. การคัดกรองบัญชีรายชื่อ (Sanctions, PEP, Watchlist Screening)

- คัดกรองลูกค้า, กรรมการ, BO และผู้เกี่ยวข้อง กับบัญชีรายชื่อต่อไปนี้ **ณ ตอนรับลูกค้า และแบบต่อเนื่อง (rescreening)**:
  - **บัญชีรายชื่อบุคคลที่ถูกกำหนด (Designated Persons List)** ตาม CTPF Act ที่เผยแพร่โดย ปปง.
  - **UN Security Council Consolidated List**
  - **OFAC (SDN), EU, UK (OFSI/HMT)** และ sanctions lists สากลอื่นที่เกี่ยวข้อง
  - บัญชี **PEP** (ผู้ดำรงตำแหน่งทางการเมือง) ทั้งในและต่างประเทศ
- **ความถี่:** ทันทีที่รับลูกค้า และ **rescreen อย่างน้อยทุก 24 ชั่วโมง** เมื่อบัญชีรายชื่อมีการปรับปรุง (delta screening)
- **การจับคู่ (match handling):** ทุก potential match ต้องเข้าสู่ **case review** โดย analyst → ตัดสิน false/true positive → บันทึกเหตุผล
- **True match ตาม CTPF Act:** ต้อง **ระงับการดำเนินการกับทรัพย์สินทันที (asset freezing)** และรายงานต่อ ปปง. โดยไม่ชักช้า และห้ามแจ้งลูกค้า (no tipping-off)

> **[TODO/ข้อสมมติ — ผู้ให้บริการภายนอก]** ยังไม่ได้เลือกผู้ให้บริการ screening (เช่น ระบบ sanctions/PEP/adverse media ของ vendor รายใด) — ต้องระบุชื่อ vendor, ข้อตกลงระดับบริการ (SLA) และความถี่การปรับปรุงฐานข้อมูลก่อนยื่นจริง

---

## 5. การตรวจสอบเข้มข้น (Enhanced Due Diligence — EDD)

ใช้ EDD กับลูกค้า/ธุรกรรมที่จัดอยู่ในระดับ **สูง** หรือเข้าเงื่อนไขต่อไปนี้:

- เกี่ยวข้องกับ **PEP** หรือครอบครัว/ผู้ใกล้ชิดของ PEP
- ธุรกิจใน **high-risk MCC** (crypto, gambling, forex/CFD, adult, marketplace ที่มีผู้ขายย่อย)
- ธุรกรรม **cross-border** หรือเกี่ยวข้องกับประเทศความเสี่ยงสูง (FATF grey/black list)
- โครงสร้างการถือหุ้นซับซ้อน/ไม่โปร่งใส หรือปฏิเสธการเปิดเผย BO
- ปริมาณธุรกรรมจริงเบี่ยงเบนจาก baseline อย่างมีนัยสำคัญ

**มาตรการ EDD เพิ่มเติม:**

1. ขอเอกสารพิสูจน์แหล่งที่มาของเงินทุน/รายได้ (Source of Funds / Source of Wealth)
2. ตรวจ **adverse media / negative news** ของลูกค้าและ BO
3. เยี่ยมสถานประกอบการ (site visit) หรือ virtual verification สำหรับ high-risk
4. อนุมัติโดย **MLRO หรือคณะกรรมการกำกับการปฏิบัติงาน** ก่อน go-live
5. เพิ่มความถี่การเฝ้าระวังและลด threshold การแจ้งเตือน
6. ทบทวนความสัมพันธ์ทุก 12 เดือน (หรือถี่กว่าเมื่อความเสี่ยงเปลี่ยน)

---

## 6. การเฝ้าระวังธุรกรรมต่อเนื่อง (Ongoing Monitoring) และการรายงาน

### 6.1 การเฝ้าระวังธุรกรรม

- ระบบเฝ้าระวังธุรกรรมทำงานแบบ **near-real-time** + **batch** ตรวจจับรูปแบบผิดปกติเทียบกับ baseline ของลูกค้าแต่ละราย
- ตัวอย่างสถานการณ์แจ้งเตือน (monitoring scenarios):
  - ปริมาณ/มูลค่าธุรกรรมพุ่งสูงผิดปกติเทียบ baseline (volume spike)
  - อัตรา chargeback/refund สูงผิดปกติ (บ่งชี้ fraud/transaction laundering)
  - ธุรกรรมกระจายเป็นก้อนย่อยเพื่อเลี่ยง threshold (structuring / smurfing)
  - ธุรกรรม card testing (BIN attack) — ยอดเล็กจำนวนมากในเวลาสั้น
  - ธุรกรรมจากประเทศความเสี่ยงสูง หรือ IP/geolocation ไม่สอดคล้อง
  - ธุรกรรม round-amount ซ้ำ ๆ, การใช้บัตรจำนวนมากผิดปกติต่อผู้ค้า
- สัญญาณจาก **3DS 2.x** และ risk scoring ที่ authorization ใช้เสริมการประเมิน

### 6.2 หน้าที่การรายงานต่อ ปปง. (AMLO Reporting)

| ประเภทรายงาน | เกณฑ์ (ตามกฎหมายไทย) | กรอบเวลา |
|---|---|---|
| **ธุรกรรมเงินสด (Cash Transaction Report)** | ตั้งแต่ **2,000,000 บาท** ขึ้นไป | ตามกำหนดของ ปปง. |
| **ธุรกรรมเกี่ยวกับทรัพย์สิน** | ตั้งแต่ **5,000,000 บาท** ขึ้นไป | ตามกำหนดของ ปปง. |
| **ธุรกรรมที่มีเหตุอันควรสงสัย (Suspicious Transaction Report — STR)** | เมื่อมีเหตุอันควรสงสัย ไม่มีขั้นต่ำจำนวนเงิน | **โดยไม่ชักช้า** หลังพบเหตุ |

> **หมายเหตุ:** โมเดลธุรกิจ acquiring เป็นธุรกรรมทางอิเล็กทรอนิกส์ (ไม่ใช่เงินสด) เป็นหลัก จึงเน้นที่ **STR** และเกณฑ์ธุรกรรมทรัพย์สินเป็นสำคัญ — ต้องยืนยันเกณฑ์/แบบฟอร์มที่ใช้จริงกับ ปปง.
>
> **[TODO/ข้อสมมติ]** ยืนยันเกณฑ์มูลค่าและช่องทางรายงานอิเล็กทรอนิกส์ (AERS ของ ปปง.) รวมถึงกำหนดเวลารายงานที่บังคับใช้ ณ วันยื่นคำขอ กับที่ปรึกษากฎหมาย/MLRO

### 6.3 กระบวนการจัดการ Alert และ STR

1. ระบบสร้าง alert → เข้า **case management queue**
2. Analyst สืบสวน รวบรวมหลักฐาน บันทึกเหตุผลการตัดสิน
3. หากเข้าข่ายสงสัย → ยกระดับให้ **MLRO**
4. MLRO ตัดสินยื่น **STR** ต่อ ปปง. และพิจารณา **freeze / offboard**
5. ห้าม tipping-off ลูกค้า
6. เก็บ audit trail ทุกขั้นตอนแบบ append-only

---

## 7. การเก็บรักษาบันทึกและเอกสาร (Record Keeping)

| ประเภทข้อมูล | ระยะเวลาเก็บขั้นต่ำ | หมายเหตุ |
|---|---|---|
| เอกสาร CDD/KYC และเอกสารระบุตัวตน | **≥ 10 ปี** นับแต่วันสิ้นสุดความสัมพันธ์ | ตาม พ.ร.บ. ปปง. |
| บันทึกธุรกรรมและ ledger | **≥ 10 ปี** นับแต่วันทำธุรกรรม | สอดคล้องกับ append-only ledger ในสถาปัตยกรรม |
| เอกสารการรายงาน (STR/CTR) และหลักฐานประกอบ | **≥ 10 ปี** | เก็บแยกและเข้าถึงจำกัด |
| Audit log ของกระบวนการ AML | **≥ 10 ปี** (สอดคล้อง PCI ขั้นต่ำ 12 เดือน online) | append-only `audit_log` |

**หลักการเก็บรักษา:**

- เก็บในรูปแบบที่ค้นคืนได้และส่งมอบต่อ ปปง./ธปท. ได้ภายในกรอบเวลาที่กำหนด
- **Data residency:** เก็บข้อมูลในไทยตามข้อกำหนด ธปท./PDPA
- ข้อมูลบัตร (PAN/CVV/track) **ไม่จัดเก็บเกินความจำเป็น** และปฏิบัติตาม **PCI-DSS v4.0** — ระบบหลักเห็นเพียง token + `card_last4` (สอดคล้องกับ ARCHITECTURE.md)
- การเข้าถึงข้อมูล KYC ต้องมีฐานทางกฎหมายและมาตรการคุ้มครองตาม **PDPA (กำกับโดย PDPC)** พร้อม RBAC, encryption at rest และ access log

---

## 8. บทบาทหน้าที่และการกำกับดูแล (Roles & Governance)

| บทบาท | หน้าที่หลัก |
|---|---|
| **คณะกรรมการบริษัท (Board)** | อนุมัตินโยบาย AML, กำกับดูแลระดับสูง, จัดสรรทรัพยากร |
| **MLRO (Money Laundering Reporting Officer)** | เจ้าหน้าที่รับผิดชอบสูงสุดด้าน AML, ตัดสินยื่น STR, เป็นจุดติดต่อ ปปง. |
| **Compliance Officer / ทีม Compliance** | ดูแลนโยบาย, การฝึกอบรม, การทดสอบการปฏิบัติตาม |
| **Onboarding / KYC Team** | ทำ CDD/EDD, จัดระดับความเสี่ยง |
| **Transaction Monitoring Analysts** | สืบสวน alert, จัดทำ case |
| **Internal Audit** | ตรวจสอบอิสระ (independent audit) ประจำปี |
| **DPO (Data Protection Officer)** | กำกับการปฏิบัติตาม PDPA สำหรับข้อมูล KYC |

- **ความเป็นอิสระ:** MLRO รายงานตรงต่อ Board/คณะกรรมการตรวจสอบ ไม่ขึ้นกับฝ่ายขาย
- **การฝึกอบรม:** พนักงานทุกคนที่เกี่ยวข้องต้องอบรม AML แรกเข้าและ **ทบทวนอย่างน้อยปีละครั้ง** เก็บหลักฐานการอบรม
- **การทดสอบอิสระ:** Internal Audit และ/หรือผู้ตรวจภายนอกประเมินประสิทธิผลของโปรแกรม AML อย่างน้อยปีละครั้ง

> **[TODO/ข้อสมมติ]** ยังต้องแต่งตั้ง MLRO และ DPO อย่างเป็นทางการ พร้อมเอกสาร fit & proper สำหรับยื่น ธปท. — ระบุชื่อ/คุณสมบัติก่อนยื่นจริง

---

## 9. การเลิกความสัมพันธ์ (Offboarding / De-risking)

- เลิกความสัมพันธ์เมื่อ: พบว่าลูกค้าอยู่ในบัญชีคว่ำบาตร, ให้ข้อมูลเท็จ, ปฏิเสธ CDD, มีพฤติกรรมฟอกเงิน/ฉ้อโกงชัดเจน
- กระบวนการต้องบันทึกเหตุผล, ได้รับอนุมัติจาก MLRO, พิจารณายื่น STR ควบคู่ และปฏิบัติตาม no tipping-off
- คงการเก็บรักษาเอกสารตามข้อ 7 แม้เลิกความสัมพันธ์แล้ว

---

## 10. สรุปข้อสมมติและ TODO ที่ต้องปิดก่อนยื่น (Open Items)

> **กรอบข้อสมมติ (ต้องยืนยันข้อเท็จจริงจริงก่อนยื่น ธปท./ปปง.):**
>
> - **Sponsor bank / card scheme:** ยังไม่ลงนาม — เงื่อนไข AML ที่ scheme/sponsor กำหนดเพิ่มเติมอาจกระทบ prohibited list และ EDD
> - **QSA vendor (PCI-DSS L1):** ยังไม่เลือก — การควบคุมข้อมูลบัตรที่ใช้ประกอบ monitoring ต้องสอดคล้อง RoC
> - **ผู้ให้บริการ screening / transaction monitoring:** ยังไม่เลือก vendor (ข้อ 4)
> - **ทุนจดทะเบียนชำระแล้ว 50 ล้านบาท:** ต้องยืนยันการชำระจริงและคงไว้ ≥ 75% ตลอดการดำเนินงาน (ตาม COMPLIANCE-TH.md)
> - **การแต่งตั้ง MLRO/DPO:** ต้องมีเอกสาร fit & proper
> - **เกณฑ์/แบบฟอร์มรายงาน ปปง. ล่าสุด:** ยืนยันเลขที่ประกาศและช่องทาง (AERS) ณ วันยื่น

---
---

# AML/KYC/CDD policy aligned to AMLO + BOT: onboarding, EDD, ongoing monitoring, record keeping (English)

> Document **05** in the Compliance set supporting the license application for **Electronic Payment Acceptance (Full Acquiring)** service under the **Payment Systems Act B.E. 2560 (2017)**, supervised by the **Bank of Thailand (BOT / ธปท.)**, in parallel with **PCI-DSS Level 1**.
>
> **Status:** Submission draft · Version 0.1 · Updated 2026-07-22
> **Owner:** Compliance & MLRO function of **[บริษัท / Company]**
>
> This is an internal policy and license-application supporting document — **not legal advice**. It must be reviewed by legal counsel and the MLRO before it takes effect, as AMLO and BOT notifications may change.

---

## Executive Summary

As a card **acquiring payment gateway**, [บริษัท / Company] is a **"financial institution" under the Anti-Money Laundering Act B.E. 2542 (as amended)** and a **BOT-licensed designated payment service provider**. This policy sets the framework for managing money-laundering (ML) and terrorism/proliferation-financing (TF/PF) risk, covering:

1. Know-Your-Customer (KYC) and Customer Due Diligence (CDD) at merchant onboarding
2. Enhanced Due Diligence (EDD) for high-risk customers/transactions
3. Ongoing transaction monitoring and reporting to AMLO
4. Record keeping
5. Roles, governance, and training

> **Customer scope:** Our direct customers are **merchants**, not end cardholders. We nonetheless monitor cardholder transaction behavior flowing through merchants (transaction laundering / merchant-based money laundering).

---

## 1. Legal Basis and Reference Standards

| Law / Standard | Relevance to this policy |
|---|---|
| **Anti-Money Laundering Act B.E. 2542** (as amended) | CDD duties, transaction reporting, record keeping; supervised by the **Anti-Money Laundering Office (AMLO / ปปง.)** |
| **Counter-Terrorism and Proliferation of Weapons Financing Act B.E. 2559 (CTPF Act)** | Designated-persons screening and immediate asset-freezing obligations |
| **Ministerial Regulation on CDD B.E. 2563 (2020)** | CDD/EDD rules, risk rating, Beneficial Owner identification |
| **Payment Systems Act B.E. 2560** | Licensing conditions under BOT, including a compliant AML/KYC system as a license prerequisite |
| **BOT notifications on risk management and IT/Cyber** | Internal control, governance, and security of systems supporting AML |
| **Personal Data Protection Act B.E. 2562 (PDPA)** — regulated by the **PDPC** | Lawful basis and safeguards for KYC data processing |
| **PCI-DSS v4.0** | Protection of card data (CHD/SAD) used in monitoring, without over-retention |
| **FATF 40 Recommendations** | Risk-Based Approach, Travel Rule, sanctions screening (international guidance) |
| **EMV 3-D Secure (3DS) 2.x** | Cardholder authentication signals enriching transaction risk assessment |

> **[TODO/Assumption]** Confirm with legal counsel the exact numbers/years of the latest AMLO and BOT notifications in force at the filing date and align citations.

---

## 2. Risk-Based Approach (RBA)

[บริษัท / Company] applies an **RBA**, allocating resources by risk across four factor groups:

| Risk factor | Example indicators |
|---|---|
| **Customer** | entity vs. individual, PEP, complex ownership, foreign shareholders |
| **Product/Channel** | card-not-present, cross-border, high-ticket, unattended e-commerce |
| **Geographic** | FATF grey/black-list jurisdictions, sanctioned countries |
| **Business type (MCC/Industry)** | crypto, gambling, forex, adult, nutraceutical, marketplace |

### 2.1 Merchant Risk Rating

| Level | Criteria (example) | CDD measures | Review cycle |
|---|---|---|---|
| **Low** | Thai business, low-risk MCC, clear BO, low projected volume | Standard CDD | Every 36 months |
| **Medium** | Moderate volume, general e-commerce, single risk factor | Standard CDD + tighter monitoring | Every 24 months |
| **High** | High-risk MCC, cross-border, PEP nexus, sub-merchant marketplace/PayFac | **EDD** + MLRO approval | Every 12 months |
| **Prohibited** | Illegal business, sanctioned, refuses BO disclosure | **Reject / terminate** | — |

> The prohibited list and restricted/high-risk MCC list are maintained in Appendix A, reviewed at least annually and whenever the card scheme/sponsor bank updates conditions.

---

## 3. Onboarding (KYC & CDD)

### 3.1 Required documents/information (corporate merchant)

| Item | Detail |
|---|---|
| Company affidavit (≤ 3 months old) | Registration number, authorized directors, objectives |
| Shareholder list (Bor.Or.Jor.5) | Ownership percentages |
| ID/passport of authorized directors and **Beneficial Owners (BO)** | BO = holds/controls **≥ 25%** or exercises effective control |
| Proof of business address | Lease/house registration/utility bill |
| Settlement bank account details | Must be in the same legal entity's name |
| Website/URL, product samples, refund policy | To assess MCC and transaction-laundering risk |
| Expected volume/ticket size | Baseline for ongoing monitoring |

**Sole proprietor:** national ID, identity verification, commercial registration (if any), settlement account details.

### 3.2 Identification & verification process

1. **Identification** — collect data per 3.1
2. **Verification** — validate against reliable sources (DBD registry, e-KYC, document checks, liveness for individuals)
3. **Beneficial Owner** — identify/verify every BO holding or controlling ≥ 25% or exercising effective control, down to natural persons
4. **Sanctions & PEP screening** — see Section 4
5. **Risk rating** — score and classify per 2.1
6. **Approval** — Low/Medium approved by Onboarding; High requires **MLRO/Compliance Committee** approval

> **Principle:** No live payment acceptance (go-live) before CDD and screening are fully completed.

### 3.3 Onboarding SLAs

| Risk level | Target CDD SLA | Approver |
|---|---|---|
| Low | 1-2 business days | Onboarding Officer |
| Medium | 3-5 business days | Onboarding Lead |
| High (EDD) | 7-15 business days | MLRO |

---

## 4. Sanctions, PEP, and Watchlist Screening

- Screen customers, directors, BOs and related parties against the following lists **at onboarding and continuously (rescreening)**:
  - **Designated Persons List** under the CTPF Act, published by AMLO
  - **UN Security Council Consolidated List**
  - **OFAC (SDN), EU, UK (OFSI/HMT)** and other relevant international sanctions lists
  - **PEP** lists (domestic and foreign)
- **Frequency:** immediately at onboarding, and **rescreen at least every 24 hours** on list updates (delta screening)
- **Match handling:** every potential match enters **case review** by an analyst → decide false/true positive → record rationale
- **True match under the CTPF Act:** **freeze assets immediately**, report to AMLO without delay, and **do not tip off** the customer

> **[TODO/Assumption — vendor]** Screening provider (sanctions/PEP/adverse-media vendor) not yet selected — specify vendor, SLA, and list-refresh frequency before filing.

---

## 5. Enhanced Due Diligence (EDD)

Apply EDD to **High**-rated customers/transactions or where:

- A **PEP** (or family/close associate) is involved
- The business is a **high-risk MCC** (crypto, gambling, forex/CFD, adult, sub-merchant marketplaces)
- Transactions are **cross-border** or involve high-risk jurisdictions (FATF grey/black list)
- Ownership is complex/opaque or BO disclosure is refused
- Actual volume deviates materially from baseline

**Additional EDD measures:**

1. Obtain Source of Funds / Source of Wealth evidence
2. Conduct **adverse media / negative news** checks on customer and BOs
3. Site visit (or virtual verification) for high-risk cases
4. Approval by **MLRO or the Compliance Committee** before go-live
5. Increase monitoring frequency and lower alert thresholds
6. Review the relationship every 12 months (or sooner if risk changes)

---

## 6. Ongoing Monitoring and Reporting

### 6.1 Transaction monitoring

- Monitoring runs **near-real-time** + **batch**, detecting anomalies against each customer's baseline.
- Example scenarios:
  - Abnormal volume/value spikes vs. baseline
  - Abnormally high chargeback/refund rates (fraud / transaction laundering)
  - Structuring / smurfing to evade thresholds
  - Card testing (BIN attack) — many small transactions in a short window
  - Transactions from high-risk countries or mismatched IP/geolocation
  - Repeated round-amount transactions; abnormal card count per merchant
- Signals from **3DS 2.x** and authorization-time risk scoring enrich the assessment.

### 6.2 AMLO reporting obligations

| Report type | Threshold (Thai law) | Timing |
|---|---|---|
| **Cash Transaction Report** | From **THB 2,000,000** | Per AMLO schedule |
| **Property/asset transaction report** | From **THB 5,000,000** | Per AMLO schedule |
| **Suspicious Transaction Report (STR)** | On reasonable grounds; no minimum amount | **Without delay** after detection |

> **Note:** The acquiring model is predominantly electronic (non-cash), so the emphasis is on **STR** and the property-transaction threshold. Confirm applicable thresholds/forms with AMLO.
>
> **[TODO/Assumption]** Confirm reporting thresholds and the electronic filing channel (AMLO's AERS), plus statutory reporting deadlines in force at the filing date, with counsel/MLRO.

### 6.3 Alert and STR handling

1. System raises an alert → enters the **case-management queue**
2. Analyst investigates, gathers evidence, records the decision rationale
3. If suspicious → escalate to the **MLRO**
4. MLRO decides to file an **STR** with AMLO and considers **freeze / offboard**
5. No tipping-off
6. Full append-only audit trail at every step

---

## 7. Record Keeping

| Data type | Minimum retention | Notes |
|---|---|---|
| CDD/KYC and identity documents | **≥ 10 years** after end of relationship | Per AML Act |
| Transaction records and ledger | **≥ 10 years** from transaction date | Aligns with the append-only ledger in the architecture |
| Reporting records (STR/CTR) and supporting evidence | **≥ 10 years** | Stored separately, access-restricted |
| AML process audit logs | **≥ 10 years** (PCI minimum 12 months online) | Append-only `audit_log` |

**Principles:**

- Retain in a retrievable form deliverable to AMLO/BOT within required timeframes
- **Data residency:** keep data in Thailand per BOT/PDPA requirements
- Card data (PAN/CVV/track) is **not over-retained** and follows **PCI-DSS v4.0** — the core system sees only token + `card_last4` (consistent with ARCHITECTURE.md)
- KYC-data access requires a lawful basis and safeguards under **PDPA (regulated by the PDPC)**, with RBAC, encryption at rest, and access logging

---

## 8. Roles & Governance

| Role | Primary responsibilities |
|---|---|
| **Board of Directors** | Approve the AML policy, high-level oversight, resource allocation |
| **MLRO (Money Laundering Reporting Officer)** | Ultimate AML accountability, decides STR filings, AMLO point of contact |
| **Compliance Officer / team** | Policy, training, compliance testing |
| **Onboarding / KYC Team** | Perform CDD/EDD, risk rating |
| **Transaction Monitoring Analysts** | Investigate alerts, build cases |
| **Internal Audit** | Independent annual audit |
| **DPO (Data Protection Officer)** | PDPA compliance for KYC data |

- **Independence:** the MLRO reports directly to the Board/Audit Committee, independent of sales
- **Training:** all relevant staff complete AML training at onboarding and **at least annually**, with evidence retained
- **Independent testing:** Internal Audit and/or external assessors evaluate AML program effectiveness at least annually

> **[TODO/Assumption]** MLRO and DPO must be formally appointed with fit & proper documentation for the BOT filing — name/qualify before filing.

---

## 9. Offboarding / De-risking

- Terminate when a customer is sanctioned, provided false information, refuses CDD, or shows clear ML/fraud behavior
- Document the rationale, obtain MLRO approval, consider filing an STR in parallel, and observe no tipping-off
- Retain records per Section 7 even after termination

---

## 10. Open Items — Assumptions & TODOs to Close Before Filing

> **Assumptions (confirm real facts before filing to BOT/AMLO):**
>
> - **Sponsor bank / card scheme:** not yet signed — additional scheme/sponsor AML conditions may affect the prohibited list and EDD
> - **QSA vendor (PCI-DSS L1):** not yet selected — card-data controls used in monitoring must align with the RoC
> - **Screening / transaction-monitoring vendor:** not yet selected (Section 4)
> - **Paid-up capital THB 50,000,000:** confirm actual payment and maintenance at ≥ 75% throughout operations (per COMPLIANCE-TH.md)
> - **MLRO/DPO appointment:** fit & proper documentation required
> - **Latest AMLO reporting thresholds/forms:** confirm notification numbers and the AERS channel at filing date
