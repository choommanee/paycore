# ขั้นตอนการระงับข้อพิพาท (ไทย)

> เอกสารประกอบการยื่นขอใบอนุญาต **การให้บริการรับชำระเงินด้วยวิธีการทางอิเล็กทรอนิกส์ (Acquiring Service)**
> ภายใต้ **พ.ร.บ. ระบบการชำระเงิน พ.ศ. 2560** กำกับโดย **ธนาคารแห่งประเทศไทย (ธปท.)** ทุนจดทะเบียนชำระแล้ว **50 ล้านบาท**
> ควบคู่กับมาตรฐาน **PCI-DSS v4.0 Level 1** และการรับรอง **EMV 3-D Secure (3DS) 2.x** สำหรับกระบวนการ authentication/chargeback
>
> รหัสเอกสาร: `COMP-29` · เวอร์ชัน 0.1 · จัดทำ 2026-07-22 · ทบทวนทุก 12 เดือน (และเมื่อ card scheme ปรับ dispute rules)
> เจ้าของเอกสาร: **Chief Compliance Officer (CCO)** ร่วมกับ **Head of Merchant Operations / Disputes Manager**
> เอกสารที่เกี่ยวข้อง: `../COMPLIANCE-TH.md`, `../ARCHITECTURE.md`, `../ROADMAP.md`, `./04-org-chart-governance.md`, `./05-aml-kyc-cdd-policy.md`, `./07-sar-str-procedure.md`, `./09-pdpa-privacy-policy.md`, `./16-incident-response-breach.md`
>
> **ข้อจำกัดความรับผิด:** เอกสารนี้เป็นข้อมูลอ้างอิงเชิงโครงสร้าง/ปฏิบัติการ ไม่ใช่คำแนะนำทางกฎหมาย
> ต้องผ่านการทบทวนโดยที่ปรึกษากฎหมายด้านใบอนุญาต ธปท. และคู่มือ dispute ของ card scheme ฉบับล่าสุดก่อนยื่นจริง

---

> [!IMPORTANT]
> **สมมติฐานและรายการที่ยังไม่สรุป (Assumptions / TODO)** — ต้องเติมค่าจริงหรือยืนยันก่อนยื่น ธปท. และก่อน go-live
>
> | # | รายการ | สถานะ | ผู้รับผิดชอบ |
> |---|--------|-------|-------------|
> | A1 | **ชื่อบริษัทจริง** — ใช้ placeholder `[บริษัท / Company]` ทั้งเอกสาร | TODO | Corporate Secretary |
> | A2 | **Sponsor bank / Acquiring bank** — ยังไม่ลงนาม (เส้นทาง B ตาม ROADMAP) กำหนด SLA การส่ง representment, การหักบัญชี chargeback และ dispute portal ที่ใช้จริง | ยังไม่สรุป | CEO / Head of Partnerships |
> | A3 | **Card scheme dispute rules ฉบับที่บังคับใช้** — Visa VCR (Visa Claims Resolution) และ Mastercard Dispute Resolution (MDR/Mastercom) ต้องยึดฉบับ ณ วันขึ้นระบบ (เลขข้อ/timeline อาจปรับ) | TODO | Disputes Manager |
> | A4 | **ITMX / local scheme rules** สำหรับบัตรในประเทศและ PromptPay/QR — ขั้นตอน dispute ของ local rails | ยังไม่สรุป | Head of Merchant Ops |
> | A5 | **ศูนย์คุ้มครองผู้ใช้บริการทางการเงิน (ศคง.) สายด่วน ธปท. 1213** — ยืนยันช่องทาง/แบบฟอร์มร้องเรียนและกรอบเวลาตอบกลับที่ ธปท. คาดหวังจากผู้ประกอบธุรกิจ | TODO | CCO / ที่ปรึกษากฎหมาย |
> | A6 | **สถาบันอนุญาโตตุลาการ** — เลือกระหว่าง **THAC (สถาบันอนุญาโตตุลาการ)** หรือ **สำนักงานอนุญาโตตุลาการ สำนักงานศาลยุติธรรม (TAI)** สำหรับ arbitration clause ในสัญญา merchant | ยังไม่สรุป | Legal Counsel |
> | A7 | **วงเงิน chargeback reserve / rolling reserve** ต่อ merchant risk tier — ต้องสอดคล้องเงื่อนไข sponsor bank | ยังไม่สรุป | CFO / Risk |
> | A8 | **ระบบ case management จริง** (dispute workflow tool) และการเชื่อม `webhook_events` เพื่อแจ้ง merchant | TODO | CTO / Disputes Manager |
>
> ห้ามกรอกชื่อผู้ให้บริการ/เลขข้อ scheme/วงเงินสมมติในช่องข้างต้นลงในเอกสารที่ยื่นจริง — ต้องเป็นข้อมูลที่ยืนยันได้เท่านั้น

---

## 1. วัตถุประสงค์และขอบเขต

เอกสารนี้กำหนด **ขั้นตอนการระงับข้อพิพาท (Dispute Resolution Procedure)** ของ [บริษัท / Company] ในฐานะผู้ให้บริการ Full Acquiring Gateway ครอบคลุมข้อพิพาททุกประเภทระหว่างคู่กรณี ได้แก่ **ผู้ถือบัตร/ผู้บริโภค (cardholder)**, **ร้านค้า (merchant)**, **ธนาคารผู้ออกบัตร (issuer)** และ **[บริษัท / Company]** เอง เพื่อให้:

1. มีช่องทางร้องเรียนและกระบวนการยกระดับ (escalation) ที่ชัดเจน เป็นธรรม และมีกรอบเวลา ตามหลัก **การคุ้มครองผู้ใช้บริการทางการเงิน (Market Conduct)** ของ ธปท.
2. จัดการ **chargeback / dispute** ตามกฎ card scheme (Visa VCR, Mastercard MDR) และ local rails (ITMX/PromptPay) อย่างถูกต้องภายในกรอบเวลา
3. รองรับการ **ยกระดับสู่การไกล่เกลี่ยและอนุญาโตตุลาการ (mediation & arbitration)** เมื่อไม่สามารถระงับกันเองได้
4. ปฏิบัติหน้าที่ **รายงานต่อ ธปท. และหน่วยงานคุ้มครองผู้บริโภค** และประสานกับ ปปง./AMLO และ PDPC เมื่อเข้าเงื่อนไข

**ขอบเขต:** ครอบคลุมทุกรายการใน `payments`, `refunds`, `ledger_entries` ตาม `ARCHITECTURE.md` และทุกช่องทาง (บัตร, 3DS, QR/PromptPay) ทั้งผ่านช่องทาง merchant, cardholder และ issuer

> **หลักการที่ยึดตลอดเอกสาร:** *"ไกล่เกลี่ยก่อน — ยกระดับเมื่อจำเป็น — บันทึกทุกขั้นตอนใน audit trail"* ทุกข้อพิพาทได้รับหมายเลข case (`DISP-YYYY-NNNNNN`) และผูกกับ `payment_id` ตั้งแต่วินาทีแรก และทุก state change ลง `audit_log` ตาม PCI-DSS Req 10

---

## 2. ประเภทข้อพิพาทและช่องทางเข้า (Dispute Types & Intake)

| ประเภท | ต้นทาง | ช่องทางเข้า | ผู้รับเรื่องเริ่มต้น |
|--------|--------|-------------|----------------------|
| **ข้อร้องเรียนผู้บริโภค (consumer complaint)** | ผู้ถือบัตร/ผู้ใช้บริการ | อีเมล `disputes@[company]`, สายด่วนลูกค้า, แบบฟอร์มในเว็บ | Customer Support (Tier 1) |
| **Chargeback / dispute (card)** | issuer ผ่าน scheme → sponsor bank | Visa VCR / Mastercom / dispute portal ของ sponsor bank | Disputes Team |
| **ข้อพิพาทธุรกรรม QR/PromptPay** | ผู้จ่าย/ผู้รับผ่าน bank/ITMX | ITMX dispute process | Disputes Team |
| **ข้อพิพาทเชิงพาณิชย์กับ merchant** | merchant (ค่าธรรมเนียม, settlement, reserve, ปิดบัญชี) | Account Manager / อีเมลตามสัญญา | Merchant Operations |
| **ข้อพิพาทที่ส่งผ่าน ธปท. (ศคง. 1213)** | ธปท. ส่งต่อคำร้องเรียน | ช่องทางทางการของ ธปท. | CCO |

**ช่องทาง intake มาตรฐาน:** ทุกช่องทางต้องบันทึกเข้าระบบ case management (รายการ A8) ภายใน **1 วันทำการ** ออกเลข case และส่ง acknowledgement อัตโนมัติให้ผู้ร้อง

---

## 3. บทบาทและความรับผิดชอบ (Roles)

| บทบาท | หน้าที่หลัก |
|-------|------------|
| **Customer Support (Tier 1)** | รับเรื่อง, ยืนยันตัวตน, แก้เคสง่าย (สอบถามยอด, refund ตามนโยบาย), ออก acknowledgement |
| **Disputes Team (Tier 2)** | จัดการ chargeback/representment, รวบรวมหลักฐาน (compelling evidence), คุมกรอบเวลา scheme |
| **Disputes Manager** | เจ้าของกระบวนการ, อนุมัติ representment/accept liability, รายงาน chargeback ratio |
| **Merchant Operations / Account Manager** | ข้อพิพาทเชิงพาณิชย์กับ merchant, การเจรจา reserve/settlement |
| **CCO (Chief Compliance Officer)** | จุดติดต่อ **ธปท./ศคง. 1213** อย่างเป็นทางการ, กำกับ market conduct, รายงานตามใบอนุญาต |
| **MLRO** | ประเมินสัญญาณฟอกเงิน/ทุจริต; หากพบให้เริ่ม STR ตาม `07-sar-str-procedure.md` (ห้าม tipping-off) |
| **DPO** | คุ้มครองข้อมูลส่วนบุคคลในเอกสารข้อพิพาท (PDPA), จำกัดการเปิดเผยข้อมูลผู้ถือบัตร |
| **Legal Counsel** | ไกล่เกลี่ย, อนุญาโตตุลาการ, คดีความ, ทบทวนสัญญา/clause |
| **Dispute Resolution Committee (DRC)** | คณะกรรมการภายในพิจารณาเคสยกระดับ (ดูข้อ 4), มติเป็นลายลักษณ์อักษร |

> **หลักการแบ่งแยกหน้าที่:** มีเพียง **CCO** ที่ติดต่อ ธปท. อย่างเป็นทางการ, มีเพียง **Legal** ที่ผูกพันบริษัทในการไกล่เกลี่ย/อนุญาโตตุลาการ, และผู้ที่ตัดสินใจ refund/chargeback ต้องไม่ใช่คนเดียวกับผู้บันทึก ledger (segregation of duties ตาม `18-segregation-of-duties.md`)

---

## 4. ขั้นตอนการยกระดับ (Escalation Ladder)

ข้อพิพาทดำเนินตามลำดับชั้น 4 ระดับ โดยยกระดับเมื่อไม่สามารถระงับได้ในระดับก่อนหน้าภายในกรอบเวลา:

| ระดับ | ผู้รับผิดชอบ | ขอบเขตอำนาจ | SLA ตอบกลับ | SLA ปิดเคส |
|-------|-------------|-------------|-------------|-------------|
| **L1 — Frontline** | Customer Support | ยืนยันข้อมูล, refund ≤ ตามนโยบาย, ตอบข้อสงสัย | รับทราบภายใน **1 วันทำการ** | **5 วันทำการ** |
| **L2 — Disputes/Ops** | Disputes Team / Account Manager | chargeback/representment, ข้อพิพาทเชิงพาณิชย์ | **2 วันทำการ** | **15 วันทำการ** |
| **L3 — Committee (DRC)** | Dispute Resolution Committee (CCO + Legal + Disputes Manager + Risk) | เคสมูลค่าสูง/ซับซ้อน/มีข้อกฎหมาย | **5 วันทำการ** | **30 วันทำการ** |
| **L4 — External** | Legal Counsel | ไกล่เกลี่ยภายนอก / อนุญาโตตุลาการ / ศาล / ธปท. | ตามกฎของสถาบัน | ตามกระบวนการ |

**เกณฑ์ยกระดับ (escalation triggers):**
- ผู้ร้องไม่พอใจผลและขอทบทวน; หรือเกิน SLA ปิดเคสของระดับนั้น
- มูลค่าข้อพิพาท **≥ 100,000 บาท** หรือกระทบผู้ร้องหลายราย (systemic) → ขึ้น L3 อัตโนมัติ
- มีประเด็นข้อกฎหมาย/ความรับผิด/ความเสี่ยงชื่อเสียง → ขึ้น L3
- ธปท./ศคง. ส่งเรื่องต่อ → CCO รับตรงและถือเป็น **priority** (ดูข้อ 8)
- พบสัญญาณฉ้อโกง/ฟอกเงิน → แจ้ง MLRO ทันที (คู่ขนาน ไม่หยุดกระบวนการ dispute)

> **หมายเหตุ:** กรอบเวลาภายในข้างต้นเป็น **นโยบายภายในของบริษัท** ต้องไม่ขัดและไม่เกินกว่ากรอบเวลาบังคับของ card scheme (ข้อ 5) และแนวปฏิบัติ market conduct ของ ธปท. (ข้อ 8) — เมื่อขัดกันให้ยึดกรอบที่สั้นกว่า/เข้มกว่า

---

## 5. Chargeback และ Representment (Card Scheme Dispute)

กระบวนการ chargeback ยึดตามวงจรของ card scheme ผ่าน sponsor bank (รายการ A2, A3):

| ขั้น | ผู้กระทำ | สาระ | กรอบเวลาบังคับ (โดยประมาณ — ต้องยืนยันตาม A3) |
|------|---------|------|-----------------------------------------------|
| **1. Retrieval / Inquiry** (ถ้ามี) | issuer | ขอสำเนาหลักฐานธุรกรรม | ตอบภายในกรอบ scheme |
| **2. First Chargeback** | issuer → acquirer | issuer หักเงินคืน cardholder อ้าง reason code | โดยทั่วไป cardholder โต้แย้งได้ภายใน **120 วัน** นับจากวันธุรกรรม/คาดว่าจะได้รับสินค้า |
| **3. Representment** | acquirer → issuer | [บริษัท / Company] ส่ง **compelling evidence** โต้กลับแทน merchant | โดยทั่วไป **≤ 30 วัน** นับจากได้รับ chargeback |
| **4. Pre-Arbitration** | issuer | issuer ยืนกรานพร้อมหลักฐานเพิ่ม | ตาม scheme |
| **5. Arbitration (scheme)** | Visa/Mastercard ชี้ขาด | scheme ตัดสินความรับผิดและปรับค่าธรรมเนียม | คำชี้ขาดผูกพันทั้งสองฝ่าย |

**หลักฐาน compelling evidence ที่จัดเก็บ:** `auth_code`, ผล **3DS 2.x** (liability shift เมื่อ authenticated), หลักฐานการส่งมอบ/ใช้บริการ, `card_last4` + `card_brand` (ห้ามเก็บ PAN/CVV ตาม `ARCHITECTURE.md` ข้อ 6), IP/device, ประวัติการติดต่อ merchant

**การควบคุมอัตรา chargeback:**
- Monitor **chargeback ratio** รายเดือนต่อ merchant; เกณฑ์เตือนภายในที่ **> 0.9%** หรือ **> 100 รายการ/เดือน** → เข้าโปรแกรม remediation
- Merchant ที่เข้าข่าย scheme monitoring program (เช่น Visa VDMP / Mastercard ECM) → เพิ่ม reserve และทบทวนการต่อสัญญา
- **3DS 2.x บังคับ** สำหรับ MCC เสี่ยงสูง เพื่อย้าย liability ไปที่ issuer

> **หมายเหตุ:** ตัวเลข reason code, จำนวนวัน และเกณฑ์ program เป็นค่าประมาณตามฉบับที่เผยแพร่ทั่วไป — **ต้องยึดคู่มือ Visa VCR / Mastercard MDR ฉบับที่บังคับ ณ วันขึ้นระบบ** (รายการ A3) และเงื่อนไข sponsor bank

---

## 6. การไกล่เกลี่ย (Mediation)

ก่อนเข้าสู่อนุญาโตตุลาการ [บริษัท / Company] เสนอ **การไกล่เกลี่ยโดยสมัครใจ** เป็นทางเลือก:

1. **การไกล่เกลี่ยภายใน** — โดย Dispute Resolution Committee (DRC) เชิญคู่กรณีชี้แจง มติเป็นลายลักษณ์อักษร ไม่ผูกมัดหากไม่ยอมรับ
2. **การไกล่เกลี่ยผ่าน ธปท.** — สำหรับข้อร้องเรียนผู้บริโภคที่ส่งผ่าน ศคง. 1213 บริษัทให้ความร่วมมือเต็มที่ตามแนวทาง market conduct
3. **การไกล่เกลี่ยภายนอก** — ผ่านสถาบันที่คู่กรณีตกลง (เช่น THAC) กรณีข้อพิพาทเชิงพาณิชย์กับ merchant

การไกล่เกลี่ยไม่ตัดสิทธิคู่กรณีในการดำเนินการตาม scheme rules หรือกฎหมาย

---

## 7. อนุญาโตตุลาการ (Arbitration)

สัญญา merchant ของ [บริษัท / Company] มี **ข้อสัญญาอนุญาโตตุลาการ (arbitration clause)** สำหรับข้อพิพาทเชิงพาณิชย์ที่ไม่สามารถระงับด้วยการไกล่เกลี่ย ตาม **พ.ร.บ. อนุญาโตตุลาการ พ.ศ. 2545**:

| องค์ประกอบ | ข้อกำหนด |
|-----------|---------|
| **สถาบัน** | THAC หรือ TAI (รายการ A6 — เลือกก่อนยื่น) |
| **สถานที่ / ภาษา** | กรุงเทพมหานคร / ภาษาไทย (คำแปลภาษาอังกฤษเมื่อจำเป็น) |
| **จำนวนอนุญาโตตุลาการ** | คนเดียว หากทุนทรัพย์ต่ำ; คณะ 3 คนหากทุนทรัพย์สูง/ซับซ้อน |
| **กฎหมายที่ใช้บังคับ** | กฎหมายไทย |
| **ผลของคำชี้ขาด** | ผูกพันและบังคับได้ตามกฎหมาย |
| **ข้อยกเว้น** | ไม่ตัด (ก) chargeback/arbitration ของ card scheme, (ข) การร้องเรียนต่อ ธปท./ศคง., (ค) มาตรการเร่งด่วน/คุ้มครองชั่วคราวจากศาล |

> **สำคัญ:** สำหรับ **ผู้บริโภค (cardholder)** การบังคับใช้ arbitration clause อาจถูกจำกัดตาม **พ.ร.บ. วิธีพิจารณาคดีผู้บริโภค** และกฎหมายคุ้มครองผู้บริโภค — ผู้บริโภคยังคงมีสิทธิร้องเรียนต่อ ธปท./สคบ. และฟ้องคดีผู้บริโภคได้เสมอ ต้องให้ Legal ทบทวน clause นี้ก่อนใช้จริง (รายการ A6)

---

## 8. การรายงานต่อ ธปท. และการคุ้มครองผู้บริโภค (BOT / Consumer Protection Reporting)

| กรณี | หน่วยงาน | ช่องทาง | กรอบเวลา | ผู้รับผิดชอบ |
|------|----------|---------|----------|-------------|
| ข้อร้องเรียนผู้บริโภคที่ ธปท. ส่งต่อ | **ธปท. / ศคง. 1213** | ช่องทางทางการ ธปท. (ยืนยัน A5) | ตอบกลับตามกรอบที่ ธปท. กำหนด (โดยทั่วไปเร่งด่วน) | CCO |
| รายงานสถิติข้อร้องเรียน/ข้อพิพาทเป็นงวด | **ธปท.** | รายงานตามเงื่อนไขใบอนุญาต | รายไตรมาส/รายปี (ตามที่กำหนด) | CCO |
| เหตุการณ์กระทบผู้ใช้บริการวงกว้าง (systemic) | **ธปท.** | แจ้งเหตุตามประกาศ IT risk/market conduct | ทันทีเมื่อเข้าเงื่อนไข | CCO / CISO |
| พบธุรกรรมน่าสงสัย (ฟอกเงิน/ทุจริต) | **ปปง. (AMLO)** | STR/SAR ตาม `07-sar-str-procedure.md` | ตามกำหนดของ ปปง. (ห้าม tipping-off) | MLRO |
| ข้อมูลส่วนบุคคลรั่วไหลจากข้อพิพาท | **PDPC** | ตาม `16-incident-response-breach.md` | แจ้ง PDPC ภายใน **72 ชม.** ตาม PDPA ม.37(4) | DPO |
| ข้อร้องเรียนผู้บริโภคทั่วไปเรื่องสินค้า/บริการ | **สคบ.** | เมื่อเกี่ยวข้อง | ตามกระบวนการ สคบ. | Legal / CCO |

**หลักการรายงาน:**
- [บริษัท / Company] เก็บ **register กลางของข้อร้องเรียน/ข้อพิพาท** ทั้งหมด (case log) พร้อม root cause, ผล, ระยะเวลา และมาตรการแก้ไข เพื่อพร้อมให้ ธปท. ตรวจสอบ (on-site/off-site)
- วิเคราะห์แนวโน้ม (trend/root-cause analysis) รายไตรมาสเสนอ DRC และผู้บริหาร เพื่อปรับปรุงกระบวนการเชิงระบบ
- ข้อมูลผู้ถือบัตรในเอกสารรายงานต้องผ่านการ minimize/mask ตาม PDPA และ PCI (ไม่มี PAN/CVV)

---

## 9. Audit Trail และการเก็บรักษาเอกสาร (Recordkeeping)

- ทุกข้อพิพาทมี case record แบบ append-only เชื่อมกับ `payment_id`, `ledger_entries` และ `audit_log`
- บันทึกครบ: ผู้ร้อง, วันเวลา, ช่องทาง, ระดับ escalation, ผู้ตัดสินใจ, หลักฐาน, ผล, การจ่ายเงินคืน/ปรับ ledger
- **ระยะเก็บรักษา:** อย่างน้อย **5 ปี** สอดคล้อง พ.ร.บ. ป้องกันและปราบปรามการฟอกเงิน และเงื่อนไขใบอนุญาต ธปท. (ยืนยันตาม `11-data-retention-deletion.md`); ทำลายตามนโยบายเมื่อครบกำหนดและไม่มีคดี
- แจ้งความคืบหน้าและผลให้ merchant ผ่าน `webhook_events` (เหตุการณ์ `dispute.opened`, `dispute.evidence_required`, `dispute.won`, `dispute.lost`)

---

## 10. ตัวชี้วัด (KPIs)

| ตัวชี้วัด | เป้าหมาย |
|---------|---------|
| Acknowledgement ผู้ร้อง | ≤ 1 วันทำการ (≥ 98%) |
| ปิดเคส L1 ภายใน SLA | ≥ 90% ภายใน 5 วันทำการ |
| ปิดเคส L2 ภายใน SLA | ≥ 90% ภายใน 15 วันทำการ |
| Representment win rate | ติดตามและปรับปรุงต่อเนื่อง |
| Chargeback ratio (portfolio) | < 0.9% |
| เคสจาก ธปท./ศคง. ตอบกลับตรงเวลา | 100% |

---

# Dispute resolution procedure: escalation, arbitration, BOT/consumer-protection reporting (English)

> Supporting document for the **Acquiring Service license application** under the **Payment Systems Act B.E. 2560 (2017)**, supervised by the **Bank of Thailand (BOT)**, with paid-up registered capital of **THB 50 million**.
> Aligned with **PCI-DSS v4.0 Level 1** and **EMV 3-D Secure (3DS) 2.x** for authentication/chargeback processes.
>
> Document ID: `COMP-29` · Version 0.1 · Prepared 2026-07-22 · Reviewed every 12 months (and whenever card-scheme dispute rules change)
> Document owner: **Chief Compliance Officer (CCO)** together with **Head of Merchant Operations / Disputes Manager**
> Related documents: `../COMPLIANCE-TH.md`, `../ARCHITECTURE.md`, `../ROADMAP.md`, `./04-org-chart-governance.md`, `./05-aml-kyc-cdd-policy.md`, `./07-sar-str-procedure.md`, `./09-pdpa-privacy-policy.md`, `./16-incident-response-breach.md`
>
> **Disclaimer:** This document is operational/structural reference material, not legal advice. It must be reviewed against BOT licensing counsel and the latest card-scheme dispute manuals before submission.

---

> [!IMPORTANT]
> **Assumptions / TODO** — values to confirm before BOT submission and go-live
>
> | # | Item | Status | Owner |
> |---|------|--------|-------|
> | A1 | **Legal company name** — placeholder `[บริษัท / Company]` used throughout | TODO | Corporate Secretary |
> | A2 | **Sponsor / acquiring bank** — not yet signed (Path B per ROADMAP); defines representment SLA, chargeback debit and the actual dispute portal | Open | CEO / Head of Partnerships |
> | A3 | **Applicable card-scheme dispute rules** — Visa VCR (Visa Claims Resolution) and Mastercard Dispute Resolution (MDR/Mastercom); must follow the edition in force at go-live (reason codes/timelines may change) | TODO | Disputes Manager |
> | A4 | **ITMX / local scheme rules** for domestic cards and PromptPay/QR dispute handling | Open | Head of Merchant Ops |
> | A5 | **BOT Financial Consumer Protection Center (hotline 1213)** — confirm official channel/form and the response window BOT expects from operators | TODO | CCO / Legal counsel |
> | A6 | **Arbitration institution** — choose **THAC** or the **Thai Arbitration Institute (TAI)** for the merchant-agreement arbitration clause | Open | Legal Counsel |
> | A7 | **Chargeback / rolling reserve amounts** per merchant risk tier — must align with sponsor-bank terms | Open | CFO / Risk |
> | A8 | **Actual case-management system** (dispute workflow tool) and `webhook_events` integration for merchant notification | TODO | CTO / Disputes Manager |
>
> Do not enter assumed vendor names / scheme article numbers / amounts into the submitted document — only verifiable data.

---

## 1. Purpose & Scope

This document defines the **Dispute Resolution Procedure** of [บริษัท / Company] as a Full Acquiring Gateway, covering disputes among all parties — **cardholders/consumers**, **merchants**, **issuers**, and [บริษัท / Company] itself — in order to:

1. Provide clear, fair, time-bound complaint and escalation channels consistent with BOT **Market Conduct / financial consumer protection** principles.
2. Handle **chargebacks / disputes** per card-scheme rules (Visa VCR, Mastercard MDR) and local rails (ITMX/PromptPay) within mandatory timelines.
3. Support **escalation to mediation and arbitration** where parties cannot resolve directly.
4. Fulfil **reporting duties to BOT and consumer-protection bodies** and coordinate with AMLO and PDPC when triggered.

**Scope:** all records in `payments`, `refunds`, `ledger_entries` per `ARCHITECTURE.md`, across all channels (card, 3DS, QR/PromptPay) and all originators (merchant, cardholder, issuer).

> **Guiding principle:** *"Mediate first — escalate when needed — record every step in the audit trail."* Every dispute is assigned a case number (`DISP-YYYY-NNNNNN`) linked to its `payment_id` from the outset, and every state change is written to `audit_log` per PCI-DSS Req 10.

---

## 2. Dispute Types & Intake

| Type | Origin | Intake channel | First responder |
|------|--------|----------------|-----------------|
| **Consumer complaint** | Cardholder/user | `disputes@[company]`, support hotline, web form | Customer Support (Tier 1) |
| **Chargeback / dispute (card)** | Issuer via scheme → sponsor bank | Visa VCR / Mastercom / sponsor-bank portal | Disputes Team |
| **QR/PromptPay transaction dispute** | Payer/payee via bank/ITMX | ITMX dispute process | Disputes Team |
| **Commercial dispute with merchant** | Merchant (fees, settlement, reserve, offboarding) | Account Manager / contractual email | Merchant Operations |
| **BOT-referred complaint (hotline 1213)** | BOT forwards complaint | BOT official channel | CCO |

**Standard intake:** every channel must be logged into the case-management system (A8) within **1 business day**, assigned a case number, with an automatic acknowledgement to the complainant.

---

## 3. Roles & Responsibilities

| Role | Primary responsibility |
|------|------------------------|
| **Customer Support (Tier 1)** | Intake, identity verification, simple resolutions (balance queries, policy refunds), acknowledgements |
| **Disputes Team (Tier 2)** | Chargeback/representment handling, compelling-evidence gathering, scheme timeline control |
| **Disputes Manager** | Process owner; approves representment / liability acceptance; reports chargeback ratio |
| **Merchant Operations / Account Manager** | Commercial disputes with merchants; reserve/settlement negotiation |
| **CCO** | Official point of contact for **BOT / hotline 1213**, market-conduct oversight, license-based reporting |
| **MLRO** | Assesses ML/fraud signals; if present, initiates STR per `07-sar-str-procedure.md` (no tipping-off) |
| **DPO** | Protects personal data within dispute files (PDPA); limits cardholder-data disclosure |
| **Legal Counsel** | Mediation, arbitration, litigation, contract/clause review |
| **Dispute Resolution Committee (DRC)** | Internal committee for escalated cases (see §4); written resolutions |

> **Segregation of duties:** only the **CCO** contacts BOT officially, only **Legal** binds the company in mediation/arbitration, and whoever decides a refund/chargeback must not be the person posting the ledger entry (per `18-segregation-of-duties.md`).

---

## 4. Escalation Ladder

Disputes follow four levels, escalating when unresolved within the level's timeframe:

| Level | Owner | Authority | Response SLA | Resolution SLA |
|-------|-------|-----------|--------------|----------------|
| **L1 — Frontline** | Customer Support | Verify data, policy refunds, answer queries | Ack within **1 business day** | **5 business days** |
| **L2 — Disputes/Ops** | Disputes Team / Account Manager | Chargeback/representment, commercial disputes | **2 business days** | **15 business days** |
| **L3 — Committee (DRC)** | DRC (CCO + Legal + Disputes Manager + Risk) | High-value/complex/legal cases | **5 business days** | **30 business days** |
| **L4 — External** | Legal Counsel | External mediation / arbitration / court / BOT | Per institution rules | Per process |

**Escalation triggers:**
- Complainant dissatisfied and requests review; or the level's resolution SLA is exceeded.
- Dispute value **≥ THB 100,000** or multi-party (systemic) impact → auto-escalate to L3.
- Legal/liability/reputational issues → L3.
- BOT / hotline 1213 referral → CCO handles directly as **priority** (see §8).
- Fraud/ML indicators → notify MLRO immediately (in parallel; dispute process continues).

> **Note:** the internal timeframes above are **company policy** and must never conflict with or exceed mandatory card-scheme timelines (§5) or BOT market-conduct expectations (§8). Where they conflict, the shorter/stricter timeframe prevails.

---

## 5. Chargeback & Representment (Card-Scheme Dispute)

The chargeback lifecycle follows the card-scheme cycle via the sponsor bank (A2, A3):

| Stage | Actor | Substance | Mandatory timeframe (indicative — confirm per A3) |
|-------|-------|-----------|----------------------------------------------------|
| **1. Retrieval / Inquiry** (if used) | Issuer | Request transaction evidence | Respond within scheme window |
| **2. First Chargeback** | Issuer → acquirer | Issuer debits, cites a reason code | Cardholder typically disputes within **120 days** of transaction/expected delivery |
| **3. Representment** | Acquirer → issuer | [บริษัท / Company] submits **compelling evidence** on the merchant's behalf | Typically **≤ 30 days** of receiving the chargeback |
| **4. Pre-Arbitration** | Issuer | Issuer maintains with further evidence | Per scheme |
| **5. Arbitration (scheme)** | Visa/Mastercard rules | Scheme rules on liability and fees | Ruling binding on both parties |

**Compelling evidence retained:** `auth_code`, **3DS 2.x** result (liability shift when authenticated), proof of delivery/service, `card_last4` + `card_brand` (never PAN/CVV per `ARCHITECTURE.md` §6), IP/device, merchant contact history.

**Chargeback-rate controls:**
- Monitor monthly **chargeback ratio** per merchant; internal warning at **> 0.9%** or **> 100 items/month** → remediation program.
- Merchants entering scheme monitoring programs (e.g., Visa VDMP / Mastercard ECM) → increased reserve and contract review.
- **3DS 2.x mandatory** for high-risk MCCs to shift liability to the issuer.

> **Note:** reason codes, day counts and program thresholds are indicative per commonly published editions — **the Visa VCR / Mastercard MDR manuals in force at go-live govern** (A3), together with sponsor-bank terms.

---

## 6. Mediation

Before arbitration, [บริษัท / Company] offers **voluntary mediation** as an option:

1. **Internal mediation** — via the Dispute Resolution Committee (DRC); parties are heard; written, non-binding unless accepted.
2. **BOT-facilitated mediation** — for consumer complaints routed through hotline 1213; the company cooperates fully under market-conduct guidance.
3. **External mediation** — through a mutually agreed institution (e.g., THAC) for commercial disputes with merchants.

Mediation does not waive any party's rights under scheme rules or law.

---

## 7. Arbitration

The [บริษัท / Company] merchant agreement contains an **arbitration clause** for commercial disputes unresolved by mediation, under the **Arbitration Act B.E. 2545 (2002)**:

| Element | Provision |
|---------|-----------|
| **Institution** | THAC or TAI (A6 — select before submission) |
| **Seat / language** | Bangkok / Thai (English translation where needed) |
| **Number of arbitrators** | Sole arbitrator for low value; panel of 3 for high-value/complex |
| **Governing law** | Thai law |
| **Effect of award** | Binding and enforceable at law |
| **Carve-outs** | Does not bar (a) card-scheme chargeback/arbitration, (b) complaints to BOT/hotline 1213, (c) urgent/interim court relief |

> **Important:** for **consumers (cardholders)**, enforceability of the arbitration clause may be limited under the **Consumer Case Procedure Act** and consumer-protection law — consumers always retain the right to complain to BOT/OCPB and to bring a consumer case. Legal must review this clause before use (A6).

---

## 8. BOT & Consumer-Protection Reporting

| Case | Authority | Channel | Timeframe | Owner |
|------|-----------|---------|-----------|-------|
| BOT-referred consumer complaint | **BOT / hotline 1213** | BOT official channel (confirm A5) | Respond within BOT-set window (usually expedited) | CCO |
| Periodic complaint/dispute statistics | **BOT** | License-condition reporting | Quarterly/annual (as prescribed) | CCO |
| Systemic event affecting many users | **BOT** | IT-risk/market-conduct notification | Immediately when triggered | CCO / CISO |
| Suspicious transaction (ML/fraud) | **AMLO** | STR/SAR per `07-sar-str-procedure.md` | Per AMLO deadlines (no tipping-off) | MLRO |
| Personal-data breach from a dispute | **PDPC** | Per `16-incident-response-breach.md` | Notify PDPC within **72 hrs** per PDPA s.37(4) | DPO |
| General consumer complaint (goods/services) | **OCPB** | When relevant | Per OCPB process | Legal / CCO |

**Reporting principles:**
- [บริษัท / Company] maintains a **central complaint/dispute register** (case log) with root cause, outcome, duration and remediation, ready for BOT inspection (on-site/off-site).
- Quarterly trend / root-cause analysis is presented to the DRC and management to drive systemic improvements.
- Cardholder data in any report is minimized/masked per PDPA and PCI (no PAN/CVV).

---

## 9. Audit Trail & Recordkeeping

- Every dispute has an append-only case record linked to `payment_id`, `ledger_entries` and `audit_log`.
- Records capture: complainant, timestamps, channel, escalation level, decision-maker, evidence, outcome, refund/ledger adjustment.
- **Retention:** at least **5 years**, consistent with the Anti-Money Laundering Act and BOT license conditions (confirm per `11-data-retention-deletion.md`); destroyed per policy once the period lapses and no litigation is pending.
- Progress and outcomes are pushed to merchants via `webhook_events` (`dispute.opened`, `dispute.evidence_required`, `dispute.won`, `dispute.lost`).

---

## 10. KPIs

| Metric | Target |
|--------|--------|
| Complainant acknowledgement | ≤ 1 business day (≥ 98%) |
| L1 closure within SLA | ≥ 90% within 5 business days |
| L2 closure within SLA | ≥ 90% within 15 business days |
| Representment win rate | Tracked and continuously improved |
| Chargeback ratio (portfolio) | < 0.9% |
| BOT/hotline-1213 cases answered on time | 100% |
