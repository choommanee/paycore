# กลยุทธ์การยืนยันตัวตน 3-D Secure 2.x (ไทย)

> เอกสารประกอบการยื่นขอใบอนุญาต **การให้บริการรับชำระเงินด้วยวิธีการทางอิเล็กทรอนิกส์ (Acquiring)**
> ต่อธนาคารแห่งประเทศไทย (ธปท.) ภายใต้ **พ.ร.บ. ระบบการชำระเงิน พ.ศ. 2560** และมาตรฐาน **PCI-DSS v4.0 Level 1**
> จัดทำโดย **[บริษัท / Company]** · เวอร์ชัน 0.1 · วันที่จัดทำ: 2026-07-22
> เอกสารอ้างอิงภายในชุด: `../COMPLIANCE-TH.md`, `../ARCHITECTURE.md`, `../ROADMAP.md`
>
> **ข้อจำกัด:** เอกสารนี้เป็นข้อมูลอ้างอิงเชิงเทคนิค/ปฏิบัติการ ไม่ใช่คำแนะนำทางกฎหมาย
> เกณฑ์ exemption และ liability shift อ้างอิงกฎ scheme (Visa/Mastercard) ซึ่งอาจปรับปรุงได้
> ต้องยืนยันกับ sponsor bank และ QSA ก่อนใช้จริง (ดูกล่อง ASSUMPTIONS / TODO ท้ายเอกสาร)

---

## 1. วัตถุประสงค์และขอบเขต

**3-D Secure 2.x (EMV 3DS)** คือโปรโตคอลยืนยันตัวตนผู้ถือบัตร (cardholder authentication) สำหรับธุรกรรม
Card-Not-Present (CNP) — เว็บ, in-app, และ recurring — เพื่อ (1) ลดการฉ้อโกง (fraud), (2) ย้ายความรับผิดชอบ
กรณี fraud chargeback ไปยัง issuer (**liability shift**), และ (3) รองรับ **frictionless flow** ที่ลด abandonment
โดยใช้ข้อมูลอุปกรณ์/บริบทกว่า 100 ตัวแปรในการประเมินความเสี่ยง (Risk-Based Authentication, RBA)

**ขอบเขตของกลยุทธ์นี้ครอบคลุม:**
- เกณฑ์ว่าเมื่อใด 3DS **บังคับ / แนะนำ / ยกเว้นได้**
- ประเภท exemption และเงื่อนไขการใช้
- Challenge flow (frictionless vs challenge) และ SLA
- กลไก liability shift และผลต่อ dispute/chargeback
- บทบาทหน้าที่ (RACI), ตัวชี้วัด (KPI), และแผน rollout ตาม `ROADMAP.md` (Phase 2 — Security & 3DS)

**สอดคล้องกับสถาปัตยกรรม:** 3DS ทำงานผ่านชั้น `3DS Adapter` (`internal/external` + `internal/pkg/threeds`)
ตาม `ARCHITECTURE.md` ข้อ 4 และ transaction flow ข้อ 5.1 (ขั้นตอน `requires_action` → `next_action_url`)

---

## 2. องค์ประกอบและบทบาทในระบบ 3DS 2.x

| องค์ประกอบ (EMV 3DS) | หน้าที่ | ใครดูแลในบริบทเรา |
|---|---|---|
| **3DS Server (3DSS)** | ริเริ่ม authentication, รวบรวมข้อมูลอุปกรณ์/ธุรกรรม, คุย DS | **[บริษัท / Company]** ผ่าน 3DS Adapter (หรือใช้ผู้ให้บริการ 3DS Server — ดู TODO) |
| **Directory Server (DS)** | สวิตช์กลางของ scheme (Visa, Mastercard) กำหนดเส้นทางไป ACS | Card scheme |
| **Access Control Server (ACS)** | ของ issuer ตัดสิน frictionless/challenge และทำ challenge | Issuing bank |
| **3DS SDK** | สำหรับ in-app (mobile) เก็บ device data | ฝัง SDK ในแอป merchant / SDK ของผู้ให้บริการ |
| **3DS Requestor** | merchant ที่ริเริ่มธุรกรรม | Merchant (ผ่าน API ของเรา) |

**เวอร์ชันโปรโตคอล:** รองรับ **EMV 3DS 2.2.0 เป็นขั้นต่ำ** และเตรียมพร้อม **2.3.x** (2.2.0 รองรับ
decoupled authentication, delegated authentication, และ acquirer exemption flags ที่จำเป็นต่อ RBA/SCA)
**ยุติการรองรับ 3DS 1.0 (3DS1)** ซึ่ง scheme ประกาศ end-of-life แล้ว — ใช้เฉพาะ fallback ชั่วคราวหากจำเป็น

---

## 3. เมื่อใดที่ต้องใช้ 3DS (Mandatory / Recommended / Optional)

> ประเทศไทย (ธปท.) กำหนดกรอบผ่านประกาศด้านความปลอดภัยการชำระเงินและ e-payment โดยอ้างอิงมาตรฐานสากล
> ในทางปฏิบัติ **การบังคับใช้ระดับรายการมาจากกฎ scheme (Visa/Mastercard) และ sponsor bank** ธปท. เน้น
> "มีมาตรการยืนยันตัวตนที่รัดกุมสำหรับธุรกรรม CNP" และการควบคุมความเสี่ยง fraud/AML

**นโยบายของ [บริษัท / Company] (policy statement):**
> ธุรกรรมบัตร **Card-Not-Present ทุกรายการต้องผ่าน 3DS 2.x authentication เป็นค่าเริ่มต้น (secure-by-default)**
> ยกเว้นเข้าเงื่อนไข exemption ที่ได้รับอนุมัติในหมวด 4 เท่านั้น การข้าม 3DS โดยไม่มี exemption ที่ถูกต้อง
> ถือเป็นการละเมิดนโยบายและถูก block ที่ระดับ Payment Core

| สถานการณ์ | สถานะ 3DS | เหตุผล |
|---|---|---|
| E-commerce web / in-app CNP (บัตรเครดิต/เดบิต) | **บังคับ (default)** | ลด fraud + ได้ liability shift; ค่าเริ่มต้นของนโยบาย |
| ธุรกรรมข้ามพรมแดน (cross-border) | **บังคับ** | ความเสี่ยง fraud สูง; issuer มักบังคับ challenge |
| ยอดสูงกว่า threshold ความเสี่ยง (ดูหมวด 4.3) | **บังคับ** | เกิน TRA/low-value exemption |
| Merchant/บัตร/อุปกรณ์ในรายการเฝ้าระวัง (watchlist) | **บังคับ (force challenge)** | risk engine flag |
| First-time recurring / initial MIT setup | **บังคับ** | ต้องยืนยันตัวตนครั้งแรกเพื่อผูก MIT |
| Subsequent MIT (recurring, installment ที่ตกลงไว้) | **ยกเว้นได้** | ไม่มี cardholder ณ ขณะนั้น (ดู 4.4) |
| Low-value < เกณฑ์ (LVP) | **ยกเว้นได้** | LVP exemption (4.3) |
| Card-Present (POS, contactless) | **ไม่ใช้ 3DS** | ใช้ EMV chip/PIN แทน (นอกขอบเขต 3DS) |

**Fallback / soft-decline:** หากได้ soft decline `65`/`1A` จาก issuer (SCA required) → ระบบต้อง **retry
ด้วย 3DS challenge อัตโนมัติ** ก่อน decline จริง (ห้าม decline ทันที)

---

## 4. Exemptions — ประเภทและเงื่อนไข

Exemption คือการที่ **acquirer/merchant ร้องขอ (ผ่าน flag ใน authorization) ให้ข้าม challenge** โดยยังคง
ส่งข้อมูลผ่าน 3DS Server เพื่อ RBA ทั้งนี้ **issuer เป็นผู้ตัดสินสุดท้าย** ว่าจะยอมรับ exemption หรือ
บังคับ challenge (soft decline) — และ **การใช้ exemption ส่วนใหญ่ทำให้ liability shift กลับมาที่ acquirer/merchant**

### 4.1 ตาราง Exemption ที่รองรับ

| Exemption | ผู้ร้องขอ | เพดาน/เงื่อนไข (indicative) | Liability | นโยบายของเรา |
|---|---|---|---|---|
| **Low-Value Payment (LVP)** | Acquirer | ≤ ~30 EUR หรือเทียบเท่า; สะสม ≤ 5 รายการ หรือ ≤ ~100 EUR/บัตร | อยู่ที่ merchant/acquirer | ใช้ได้เฉพาะ MCC ความเสี่ยงต่ำ |
| **Transaction Risk Analysis (TRA)** | Acquirer | ต้องอยู่ใต้ fraud-rate threshold ตาม RTS (เช่น ≤ 0.13% ที่ ≤100 EUR, ≤0.06% ที่ ≤250 EUR, ≤0.01% ที่ ≤500 EUR) | อยู่ที่ acquirer | ใช้เมื่อ portfolio fraud rate ต่ำกว่าเกณฑ์เท่านั้น |
| **Trusted Beneficiary / Whitelisting** | Issuer/Cardholder | cardholder เพิ่ม merchant เป็น trusted ที่ ACS | อยู่ที่ issuer | รองรับ แต่ควบคุมโดย issuer |
| **Secure Corporate Payment** | Acquirer | บัตร B2B/lodged card, virtual card, dedicated process | อยู่ที่ acquirer | เฉพาะ merchant องค์กรที่ผ่านการอนุมัติ |
| **Recurring / MIT (subsequent)** | Merchant | ผูก initial CIT ด้วย 3DS แล้ว | อยู่ที่ merchant | บังคับ 3DS เฉพาะรายการแรก |
| **Delegated Authentication (DA)** | Acquirer/3DSS | ยืนยันตัวตนโดย party ที่ได้รับมอบหมาย (เช่น biometrics) | ตามข้อตกลง DA | เป็น TODO (ต้อง cert เพิ่ม) |

> **หมายเหตุ:** เพดานเป็นตัวเลข **indicative จาก PSD2 RTS ของสหภาพยุโรป** ใช้เป็น baseline การออกแบบ
> ระบบ — **เกณฑ์บังคับจริงในไทยกำหนดโดย scheme rule + sponsor bank** และต้อง config ผ่าน parameter table
> ไม่ hard-code (ดู TODO 3)

### 4.2 หลักการควบคุมการใช้ exemption

1. **Exemption engine อยู่ใน Risk/Fraud Engine** (`internal/service`) — คำนวณ risk score ก่อนตัดสินใจ
   ขอ exemption; ทุกการขอ exemption ต้องบันทึกลง `audit_log` พร้อมเหตุผลและ parameter ที่ใช้
2. **Fail closed** (ตาม `ARCHITECTURE.md` ข้อ 2.7) — หากไม่มั่นใจว่า exemption ใช้ได้ → **บังคับ challenge**
3. **Soft-decline handling** — หาก issuer ปฏิเสธ exemption (soft decline) → retry เป็น challenge อัตโนมัติ
4. **TRA fraud-rate monitoring** — ต้องติดตาม fraud rate ของ portfolio ต่อเนื่อง; หากเกิน threshold ต้อง
   **ปิด TRA exemption อัตโนมัติ** และรายงานต่อ Compliance
5. ห้ามใช้ exemption กับธุรกรรมที่ risk engine flag เป็น high-risk แม้ยอดจะเข้าเกณฑ์ LVP/TRA

### 4.3 Threshold configuration (parameterized)

| พารามิเตอร์ | ค่าตั้งต้น (design baseline) | เจ้าของ | หมายเหตุ |
|---|---|---|---|
| `lvp.amount_cap` | เทียบเท่า ~30 EUR (config เป็น THB) | Compliance + Sponsor bank | ปรับตาม scheme rule |
| `lvp.count_cap` / `lvp.cumulative_cap` | 5 รายการ / ~100 EUR | Compliance | ต่อบัตรต่อ issuer |
| `tra.fraud_rate.tier` | 0.13% / 0.06% / 0.01% | Risk | ผูกกับเพดานยอด |
| `challenge.force_above` | THB threshold (TODO) | Risk + Sponsor bank | บังคับ challenge เหนือยอดนี้ |
| `high_risk_mcc[]` | รายการ MCC เสี่ยง | Compliance | ห้าม exemption |

### 4.4 Recurring / MIT

- **Initial (CIT — Customer Initiated):** บังคับ 3DS challenge เพื่อยืนยันตัวตนและได้ liability shift
  พร้อมเก็บ `3DS transaction ID` / network transaction reference สำหรับผูก MIT
- **Subsequent (MIT — Merchant Initiated):** ไม่มี cardholder present จึงยกเว้น 3DS แต่ต้องส่ง
  MIT indicator + reference ที่ถูกต้อง มิฉะนั้นเสี่ยงถูก decline/chargeback

---

## 5. Challenge Flow — Frictionless และ Challenge

### 5.1 ภาพรวมลำดับ (sequence)

```
Cardholder ─┐
Merchant ── POST /v1/payments (payment_token, browser/device data)
            ▼
     [บริษัท/Company] Payment Core
            │  detokenize (PCI scope) → Risk scoring → ตัดสิน exemption?
            ▼  AReq (Authentication Request) ผ่าน 3DS Server
        Directory Server (scheme)
            ▼
        ACS (issuer)  ── ประเมิน RBA ──►  ┌─ Frictionless: ARes(status=Y) ─► authorize
                                          └─ Challenge:    ARes(status=C) ─► CReq/CRes
                                                             ▼
                              คืน next_action_url ให้ cardholder ทำ challenge (OTP/biometric)
                                                             ▼
                              RReq/RRes → ผล authentication (Y/N/A/U/R) → authorize หรือ fail
```

**Mapping กับ state machine** (`ARCHITECTURE.md` ข้อ 5.1/5.4):
- ต้อง challenge → คืน `requires_action` + `next_action_url`
- authentication สำเร็จ (Y/A) → `authorized` → capture
- authentication fail (N/R) → `failed`

### 5.2 ผลลัพธ์ authentication (transStatus)

| สถานะ | ความหมาย | การกระทำ | Liability |
|---|---|---|---|
| **Y** | Authenticated (frictionless หรือ challenge สำเร็จ) | ดำเนินการ authorize | Shift → issuer |
| **A** | Attempted (issuer/ACS ไม่พร้อม แต่บันทึก attempt) | ดำเนินการ authorize | Shift → issuer (โดยทั่วไป) |
| **C** | Challenge required | คืน `requires_action` | รอผล |
| **D** | Decoupled challenge (async) | รอ callback ตาม SLA | รอผล |
| **N** | Not authenticated | **decline** (ไม่ authorize) | อยู่ที่ merchant |
| **U** | Unable / technical | ตามนโยบายความเสี่ยง (ดู 5.4) | ไม่ shift |
| **R** | Rejected โดย issuer | decline | อยู่ที่ merchant |

### 5.3 SLA และ timeout

| ขั้นตอน | เป้าหมาย | Timeout / นโยบาย |
|---|---|---|
| AReq → ARes (frictionless) | < 3 วินาที (p95) | timeout 10 วินาที → treat as U |
| Challenge (cardholder ทำ OTP) | ผู้ใช้ 3-5 นาที | ACS timeout มาตรฐาน; หมดเวลา → fail closed |
| Decoupled (status D) | callback ภายในหน้าต่าง scheme | มี worker ตาม `next_action` + expiry |
| รวม end-to-end auth latency | สอดคล้อง p99 < 800 ms (ส่วน authorize) | 3DS challenge ไม่นับใน auth latency เพราะ async ฝั่งผู้ใช้ |

### 5.4 นโยบาย status U (Unable) และ fallback

- **U จากปัญหาเทคนิค:** ใช้นโยบาย **fail closed** — ธุรกรรมยอดสูง/เสี่ยงสูง → decline; ยอดต่ำ/เสี่ยงต่ำ
  อาจ authorize ตาม risk policy แต่ **ไม่ได้ liability shift** (merchant รับความเสี่ยง)
- ห้าม fallback ไป 3DS1 เป็น default (EOL) — ใช้เฉพาะกรณีที่ scheme/sponsor ยังอนุญาตชั่วคราวเท่านั้น
- ทุกกรณี U/A/fallback ต้องบันทึก `audit_log` เพื่อการตรวจสอบและ dispute

---

## 6. Liability Shift — กลไกและผลต่อ Chargeback

**หลักการ:** เมื่อธุรกรรมผ่าน 3DS สำเร็จ (transStatus = Y หรือ A) **ความรับผิดชอบต่อ fraud chargeback
(reason code กลุ่ม fraud/unauthorized) ย้ายจาก acquirer/merchant ไปยัง issuer** ทำให้ [บริษัท / Company]
และ merchant ได้รับการคุ้มครองจาก dispute ประเภทนี้

| กรณี | Liability | หมายเหตุ |
|---|---|---|
| 3DS สำเร็จ (Y) | **Issuer** | คุ้มครอง fraud chargeback |
| 3DS attempted (A) | **Issuer** (โดยทั่วไป) | ขึ้นกับ scheme rule |
| ใช้ exemption (TRA/LVP) โดย acquirer/merchant | **Acquirer/Merchant** | แลกกับ frictionless แต่รับความเสี่ยง |
| ไม่ทำ 3DS ทั้งที่ควรทำ | **Merchant** | เต็มจำนวน |
| status U / N / R | **Merchant** | ไม่ shift |

**ผลต่อ dispute/chargeback workflow** (สอดคล้อง `ROADMAP.md` Phase 3):
1. เมื่อเกิด chargeback ต้อง lookup **3DS authentication data (CAVV/AAV, ECI, DS transaction ID,
   transStatus)** จาก `audit_log` เพื่อประกอบ representment
2. **ECI (E-Commerce Indicator)** บันทึกทุกธุรกรรม: Visa ECI 05 / Mastercard ECI 02 = fully authenticated;
   ECI 06/01 = attempted; ECI 07/00 = ไม่ authenticated
3. หากมี liability shift แต่ยังโดน chargeback → ยื่น **pre-arbitration/representment** พร้อมหลักฐาน 3DS
4. เก็บ authentication artifacts ตาม retention policy (ดูหมวด 8)

> **สำคัญ:** liability shift **ไม่คุ้มครอง** dispute ประเภท non-fraud (สินค้าไม่ตรง, ไม่ได้รับของ,
> processing error) — ต้องจัดการผ่าน dispute workflow ปกติ

---

## 7. บทบาทหน้าที่ (RACI)

| กิจกรรม | Backend/Eng | Risk/DevSecOps | Compliance/Legal | Sponsor Bank | QSA |
|---|---|---|---|---|---|
| ออกแบบ/พัฒนา 3DS Adapter | **R/A** | C | I | C | I |
| กำหนด exemption threshold | C | R | **A** | C | I |
| RBA / risk scoring engine | R | **R/A** | C | I | I |
| Fraud-rate monitoring (TRA) | C | **R** | A | I | I |
| Liability/chargeback workflow | R | C | **A** | C | I |
| PCI scope ของ 3DS data | R | R | C | I | **A/V** |
| ยื่นเอกสารต่อ ธปท. | I | I | **R/A** | I | I |
| Scheme certification (3DS) | R | C | I | **A** | I |

(R=Responsible, A=Accountable, C=Consulted, I=Informed, V=Verify)

---

## 8. ข้อมูล 3DS, PCI-DSS v4.0 และ PDPA

- **PCI-DSS v4.0:** ข้อมูล authentication (CAVV/AAV, ECI, DS transaction ID) **ไม่ใช่ full PAN** แต่ถือเป็น
  sensitive — เก็บใน scope ที่ควบคุม, เข้ารหัส at-rest, ห้าม log ในระบบทั่วไป (สอดคล้อง `ARCHITECTURE.md` ข้อ 7)
- **ห้ามเก็บ** full PAN/CVV/PIN ในระบบ operational — 3DS ทำงานบน payment token/detokenize ใน PCI scope เท่านั้น
- **PCI v4.0 e-commerce controls:** Req 6.4.3 และ 11.6.1 (การจัดการ payment page scripts / change detection)
  บังคับตั้งแต่ 31 มี.ค. 2025 — ต้องรวมในการออกแบบหน้ารับชำระ/3DS challenge redirect
- **PDPA (พ.ร.บ. คุ้มครองข้อมูลส่วนบุคคล พ.ศ. 2562):** device data / browser fingerprint ที่เก็บเพื่อ RBA
  เป็นข้อมูลส่วนบุคคล — ต้องระบุฐานการประมวลผล (legitimate interest / fraud prevention), แจ้ง privacy notice,
  จำกัดวัตถุประสงค์, และ retention ตามความจำเป็น
- **AMLO (ปปง.):** ผล 3DS เป็น input หนึ่งของ fraud/AML monitoring; ธุรกรรมที่ผิดปกติเข้าสู่ suspicious
  transaction workflow
- **Data residency:** เก็บข้อมูลในไทยตามข้อกำหนด ธปท./PDPA (`ARCHITECTURE.md` ข้อ 8)
- **Retention:** เก็บ 3DS authentication artifacts อย่างน้อยตามหน้าต่าง dispute/chargeback ของ scheme
  (baseline **18 เดือน**, TODO ยืนยันกับ sponsor) เพื่อรองรับ representment และ audit ธปท.

---

## 9. KPI และการติดตาม

| KPI | เป้าหมาย | ความถี่ |
|---|---|---|
| Frictionless rate | ≥ 90% ของธุรกรรมที่ผ่าน 3DS | รายเดือน |
| Challenge abandonment | ≤ 10% | รายเดือน |
| Fraud rate (post-3DS) | ต่ำกว่า TRA threshold ที่ใช้ | รายสัปดาห์ |
| Authentication success (Y+A) | ≥ 95% | รายสัปดาห์ |
| Soft-decline retry success | ติดตามและปรับ parameter | รายเดือน |
| Exemption acceptance rate | ติดตามต่อ issuer | รายเดือน |

---

## 10. แผน Rollout (สอดคล้อง ROADMAP Phase 2)

1. **สัปดาห์ 8-10:** เชื่อม 3DS Server sandbox, implement AReq/ARes, frictensionless + challenge flow, mapping state machine
2. **สัปดาห์ 10-12:** exemption engine + risk scoring, parameter table, soft-decline retry, audit logging
3. **สัปดาห์ 12-14:** liability/ECI capture, chargeback data plumbing, PDPA/PCI review, เตรียม scheme certification
4. **Phase 4:** 3DS scheme certification (Visa/Mastercard) ผ่าน sponsor bank + รวมใน RoC ของ QSA

---

## 11. ASSUMPTIONS / TODO (ยังไม่ resolve — ต้องยืนยันก่อนใช้จริง)

> กล่องนี้ระบุ external dependency ที่ยังไม่ resolved อย่างชัดเจน — **ห้ามถือว่าเป็นข้อเท็จจริงจนกว่าจะยืนยัน**

- **[TODO-1] Sponsor bank / acquirer:** ยังไม่ล็อก partner (ตาม `ROADMAP.md` critical path) — เกณฑ์
  exemption จริง, scheme certification, และ liability rule ต้องยืนยันกับ sponsor
- **[TODO-2] 3DS Server:** ตัดสินใจ **build เอง vs ใช้ผู้ให้บริการ 3DS Server ที่ผ่าน EMVCo** (กระทบ PCI scope + timeline)
- **[TODO-3] Threshold เป็น THB:** ค่า LVP/TRA/force-challenge ในเอกสารเป็น baseline PSD2 RTS (EU) —
  ต้องแปลง/ยืนยันเป็น THB ตาม scheme rule ไทยและ config ใน parameter table (ห้าม hard-code)
- **[TODO-4] Delegated Authentication:** ต้อง cert เพิ่มและตกลง liability กับ scheme — ยังไม่รองรับใน MVP
- **[TODO-5] QSA vendor:** ยังไม่เลือก QSA — การกำหนด scope ของ 3DS authentication data ใน RoC ต้องรอ QSA ยืนยัน
- **[TODO-6] ทุนจดทะเบียนจริง:** เอกสารอ้าง 50 ล้านบาท (เกณฑ์ Acquiring) — ยืนยันทุนชำระแล้วจริงและการคงไว้ ≥75%
- **[ASSUMPTION] EMVCo 3DS 2.2.0** เป็นเวอร์ชันขั้นต่ำ และ scheme ยัง honor liability shift สำหรับ status Y/A
  ตามกฎปัจจุบัน — ต้อง re-validate เมื่อ scheme ปรับ mandate

---
---

# 3DS 2.x authentication strategy: when mandatory, exemptions, challenge flow, liability shift (English)

> Supporting document for the **Acquiring (electronic payment acceptance) license** application to the
> Bank of Thailand (BOT / ธปท.) under the **Payment Systems Act B.E. 2560** and **PCI-DSS v4.0 Level 1**.
> Prepared by **[บริษัท / Company]** · Version 0.1 · Date: 2026-07-22.
> Companion docs: `../COMPLIANCE-TH.md`, `../ARCHITECTURE.md`, `../ROADMAP.md`.
>
> **Disclaimer:** Technical/operational reference, not legal advice. Exemption thresholds and liability-shift
> rules derive from scheme rules (Visa/Mastercard) and may change; confirm with sponsor bank and QSA before
> production use (see ASSUMPTIONS / TODO callout at the end).

---

## 1. Purpose and scope

**3-D Secure 2.x (EMV 3DS)** is a cardholder-authentication protocol for Card-Not-Present (CNP) transactions —
web, in-app, and recurring — intended to (1) reduce fraud, (2) shift fraud-chargeback liability to the issuer
(**liability shift**), and (3) enable a low-friction **frictionless flow** using 100+ device/context data
elements for Risk-Based Authentication (RBA).

**This strategy covers:**
- When 3DS is **mandatory / recommended / exemptible**
- Exemption types and conditions
- Challenge flow (frictionless vs challenge) and SLAs
- Liability-shift mechanics and dispute/chargeback impact
- Roles (RACI), KPIs, and rollout aligned with `ROADMAP.md` (Phase 2 — Security & 3DS)

**Architecture alignment:** 3DS runs through the `3DS Adapter` layer (`internal/external` +
`internal/pkg/threeds`) per `ARCHITECTURE.md` §4 and the transaction flow in §5.1 (`requires_action` →
`next_action_url`).

---

## 2. Components and roles in 3DS 2.x

| Component (EMV 3DS) | Function | Owner in our context |
|---|---|---|
| **3DS Server (3DSS)** | Initiates auth, gathers device/txn data, talks to DS | **[บริษัท / Company]** via 3DS Adapter (or third-party 3DS Server — see TODO) |
| **Directory Server (DS)** | Scheme switch (Visa, Mastercard) routing to ACS | Card scheme |
| **Access Control Server (ACS)** | Issuer engine deciding frictionless/challenge | Issuing bank |
| **3DS SDK** | In-app (mobile) device-data capture | Embedded in merchant app / provider SDK |
| **3DS Requestor** | Merchant initiating the transaction | Merchant (via our API) |

**Protocol version:** support **EMV 3DS 2.2.0 as a minimum** and prepare for **2.3.x** (2.2.0 adds decoupled
authentication, delegated authentication, and acquirer exemption flags needed for RBA/SCA). **3DS 1.0 is
deprecated** (scheme EOL) — used only as a temporary fallback if unavoidable.

---

## 3. When 3DS is required (Mandatory / Recommended / Optional)

> In Thailand, BOT (ธปท.) sets the framework via payment-security and e-payment notifications referencing
> international standards. In practice, **per-transaction enforcement comes from scheme rules (Visa/Mastercard)
> and the sponsor bank**; BOT emphasizes "strong authentication for CNP transactions" and fraud/AML risk control.

**[บริษัท / Company] policy statement:**
> **Every Card-Not-Present card transaction must undergo 3DS 2.x authentication by default (secure-by-default)**,
> except where an approved exemption in §4 applies. Skipping 3DS without a valid exemption is a policy violation
> and is blocked at the Payment Core level.

| Scenario | 3DS status | Rationale |
|---|---|---|
| E-commerce web / in-app CNP (credit/debit) | **Mandatory (default)** | Fraud reduction + liability shift; policy default |
| Cross-border transactions | **Mandatory** | High fraud risk; issuers often force challenge |
| Amount above risk threshold (see §4.3) | **Mandatory** | Exceeds TRA/low-value exemption |
| Merchant/card/device on watchlist | **Mandatory (force challenge)** | Risk-engine flag |
| First-time recurring / initial MIT setup | **Mandatory** | Must authenticate initial CIT to anchor MIT |
| Subsequent MIT (recurring, agreed installment) | **Exemptible** | No cardholder present (see §4.4) |
| Low-value < threshold (LVP) | **Exemptible** | LVP exemption (§4.3) |
| Card-Present (POS, contactless) | **No 3DS** | Uses EMV chip/PIN instead (out of 3DS scope) |

**Fallback / soft-decline:** on issuer soft decline `65`/`1A` (SCA required), the system must **automatically
retry with a 3DS challenge** before a hard decline (never decline immediately).

---

## 4. Exemptions — types and conditions

An exemption is a request (via an authorization flag) by the acquirer/merchant to **skip the challenge** while
still passing data through the 3DS Server for RBA. The **issuer makes the final decision** to honor the
exemption or force a challenge (soft decline) — and **most exemptions move liability back to the
acquirer/merchant**.

### 4.1 Supported exemptions

| Exemption | Requestor | Cap/conditions (indicative) | Liability | Our policy |
|---|---|---|---|---|
| **Low-Value Payment (LVP)** | Acquirer | ≤ ~EUR 30 equiv.; cumulative ≤ 5 txns or ≤ ~EUR 100/card | Merchant/acquirer | Low-risk MCCs only |
| **Transaction Risk Analysis (TRA)** | Acquirer | Under RTS fraud-rate threshold (e.g. ≤0.13% @ ≤100 EUR, ≤0.06% @ ≤250 EUR, ≤0.01% @ ≤500 EUR) | Acquirer | Only while portfolio fraud rate is below threshold |
| **Trusted Beneficiary / Whitelisting** | Issuer/Cardholder | Cardholder marks merchant trusted at ACS | Issuer | Supported, issuer-controlled |
| **Secure Corporate Payment** | Acquirer | B2B/lodged/virtual cards, dedicated process | Acquirer | Approved corporate merchants only |
| **Recurring / MIT (subsequent)** | Merchant | Initial CIT anchored with 3DS | Merchant | 3DS on first txn only |
| **Delegated Authentication (DA)** | Acquirer/3DSS | Auth delegated (e.g. biometrics) | Per DA agreement | TODO (extra cert required) |

> **Note:** Caps are **indicative from the EU PSD2 RTS**, used as a design baseline — **actual binding
> thresholds in Thailand are set by scheme rules + sponsor bank** and must be configured via a parameter table,
> not hard-coded (see TODO 3).

### 4.2 Exemption-usage controls

1. **Exemption logic lives in the Risk/Fraud Engine** (`internal/service`) — compute a risk score before
   requesting an exemption; every exemption request is logged to `audit_log` with reason and parameters used.
2. **Fail closed** (`ARCHITECTURE.md` §2.7) — if unsure an exemption applies, **force a challenge**.
3. **Soft-decline handling** — if the issuer rejects the exemption (soft decline), automatically retry as a challenge.
4. **TRA fraud-rate monitoring** — continuously track portfolio fraud rate; if it exceeds the threshold,
   **automatically disable TRA** and report to Compliance.
5. Never use an exemption for transactions flagged high-risk by the risk engine, even if within LVP/TRA caps.

### 4.3 Threshold configuration (parameterized)

| Parameter | Default (design baseline) | Owner | Note |
|---|---|---|---|
| `lvp.amount_cap` | ~EUR 30 equiv. (configured in THB) | Compliance + Sponsor bank | Per scheme rule |
| `lvp.count_cap` / `lvp.cumulative_cap` | 5 txns / ~EUR 100 | Compliance | Per card per issuer |
| `tra.fraud_rate.tier` | 0.13% / 0.06% / 0.01% | Risk | Tied to amount tiers |
| `challenge.force_above` | THB threshold (TODO) | Risk + Sponsor bank | Force challenge above this |
| `high_risk_mcc[]` | Risky MCC list | Compliance | No exemption allowed |

### 4.4 Recurring / MIT

- **Initial (CIT — Customer Initiated):** force a 3DS challenge to authenticate and obtain liability shift; store
  the `3DS transaction ID` / network transaction reference to anchor future MIT.
- **Subsequent (MIT — Merchant Initiated):** no cardholder present, so 3DS is exempted, but a correct MIT
  indicator + reference must be sent, otherwise risk of decline/chargeback.

---

## 5. Challenge flow — frictionless and challenge

### 5.1 Sequence overview

```
Cardholder ─┐
Merchant ── POST /v1/payments (payment_token, browser/device data)
            ▼
     [บริษัท/Company] Payment Core
            │  detokenize (PCI scope) → risk scoring → exemption decision?
            ▼  AReq (Authentication Request) via 3DS Server
        Directory Server (scheme)
            ▼
        ACS (issuer)  ── RBA ──►  ┌─ Frictionless: ARes(status=Y) ─► authorize
                                  └─ Challenge:    ARes(status=C) ─► CReq/CRes
                                                     ▼
                       return next_action_url; cardholder completes challenge (OTP/biometric)
                                                     ▼
                       RReq/RRes → auth result (Y/N/A/U/R) → authorize or fail
```

**State-machine mapping** (`ARCHITECTURE.md` §5.1/§5.4):
- Challenge required → return `requires_action` + `next_action_url`
- Auth success (Y/A) → `authorized` → capture
- Auth failure (N/R) → `failed`

### 5.2 Authentication outcomes (transStatus)

| Status | Meaning | Action | Liability |
|---|---|---|---|
| **Y** | Authenticated (frictionless or challenge success) | Proceed to authorize | Shifts → issuer |
| **A** | Attempted (ACS/issuer unavailable but attempt recorded) | Proceed to authorize | Shifts → issuer (generally) |
| **C** | Challenge required | Return `requires_action` | Pending |
| **D** | Decoupled challenge (async) | Await callback per SLA | Pending |
| **N** | Not authenticated | **Decline** (no authorize) | Merchant |
| **U** | Unable / technical | Per risk policy (see §5.4) | No shift |
| **R** | Rejected by issuer | Decline | Merchant |

### 5.3 SLA and timeouts

| Step | Target | Timeout / policy |
|---|---|---|
| AReq → ARes (frictionless) | < 3 s (p95) | 10 s timeout → treat as U |
| Challenge (cardholder OTP) | User 3–5 min | Standard ACS timeout; expiry → fail closed |
| Decoupled (status D) | Callback within scheme window | Worker on `next_action` + expiry |
| End-to-end auth latency | Consistent with authorize p99 < 800 ms | 3DS challenge excluded (async, user-side) |

### 5.4 Status-U (Unable) and fallback policy

- **U from technical issues:** apply **fail closed** — high-value/high-risk → decline; low-value/low-risk may
  authorize per risk policy but **without liability shift** (merchant bears risk).
- Never default to 3DS1 fallback (EOL) — only where scheme/sponsor still temporarily permit it.
- Every U/A/fallback case must be recorded in `audit_log` for audit and dispute.

---

## 6. Liability shift — mechanics and chargeback impact

**Principle:** when a transaction is successfully authenticated via 3DS (transStatus = Y or A), **liability for
fraud chargebacks (fraud/unauthorized reason-code groups) shifts from acquirer/merchant to the issuer**,
protecting [บริษัท / Company] and the merchant against those disputes.

| Case | Liability | Note |
|---|---|---|
| 3DS success (Y) | **Issuer** | Fraud chargebacks covered |
| 3DS attempted (A) | **Issuer** (generally) | Depends on scheme rule |
| Exemption used (TRA/LVP) by acquirer/merchant | **Acquirer/Merchant** | Frictionless traded for risk |
| No 3DS where it should apply | **Merchant** | Full amount |
| Status U / N / R | **Merchant** | No shift |

**Dispute/chargeback workflow impact** (aligns with `ROADMAP.md` Phase 3):
1. On a chargeback, look up **3DS authentication data (CAVV/AAV, ECI, DS transaction ID, transStatus)** from
   `audit_log` to build the representment.
2. **ECI (E-Commerce Indicator)** recorded on every transaction: Visa ECI 05 / Mastercard ECI 02 = fully
   authenticated; ECI 06/01 = attempted; ECI 07/00 = not authenticated.
3. Where liability shift applies but a chargeback still arrives → file **pre-arbitration/representment** with 3DS evidence.
4. Retain authentication artifacts per retention policy (see §8).

> **Important:** liability shift does **not** cover non-fraud disputes (item not as described, not received,
> processing error) — handle those via the normal dispute workflow.

---

## 7. Roles and responsibilities (RACI)

| Activity | Backend/Eng | Risk/DevSecOps | Compliance/Legal | Sponsor Bank | QSA |
|---|---|---|---|---|---|
| Design/build 3DS Adapter | **R/A** | C | I | C | I |
| Set exemption thresholds | C | R | **A** | C | I |
| RBA / risk-scoring engine | R | **R/A** | C | I | I |
| Fraud-rate monitoring (TRA) | C | **R** | A | I | I |
| Liability/chargeback workflow | R | C | **A** | C | I |
| PCI scope of 3DS data | R | R | C | I | **A/V** |
| BOT filing | I | I | **R/A** | I | I |
| Scheme certification (3DS) | R | C | I | **A** | I |

(R=Responsible, A=Accountable, C=Consulted, I=Informed, V=Verify)

---

## 8. 3DS data, PCI-DSS v4.0 and PDPA

- **PCI-DSS v4.0:** authentication data (CAVV/AAV, ECI, DS transaction ID) is **not full PAN** but is sensitive —
  store within controlled scope, encrypt at rest, never log in general systems (`ARCHITECTURE.md` §7).
- **Never store** full PAN/CVV/PIN in operational systems — 3DS operates on payment tokens/detokenized data
  within PCI scope only.
- **PCI v4.0 e-commerce controls:** Req 6.4.3 and 11.6.1 (payment-page script management / change detection),
  mandatory since 31 Mar 2025 — must be part of the checkout/3DS challenge-redirect design.
- **PDPA (Personal Data Protection Act B.E. 2562):** device data / browser fingerprints collected for RBA are
  personal data — define a lawful basis (legitimate interest / fraud prevention), publish a privacy notice,
  limit purpose, and retain only as necessary.
- **AMLO (ปปง.):** 3DS outcomes feed fraud/AML monitoring; anomalous transactions enter the suspicious-transaction workflow.
- **Data residency:** store data in Thailand per BOT/PDPA requirements (`ARCHITECTURE.md` §8).
- **Retention:** retain 3DS authentication artifacts at least through the scheme dispute/chargeback window
  (baseline **18 months**, TODO confirm with sponsor) to support representment and BOT audit.

---

## 9. KPIs and monitoring

| KPI | Target | Frequency |
|---|---|---|
| Frictionless rate | ≥ 90% of 3DS transactions | Monthly |
| Challenge abandonment | ≤ 10% | Monthly |
| Fraud rate (post-3DS) | Below the applied TRA threshold | Weekly |
| Authentication success (Y+A) | ≥ 95% | Weekly |
| Soft-decline retry success | Tracked and tuned | Monthly |
| Exemption acceptance rate | Tracked per issuer | Monthly |

---

## 10. Rollout plan (aligned with ROADMAP Phase 2)

1. **Weeks 8–10:** integrate 3DS Server sandbox, implement AReq/ARes, frictionless + challenge flow, state-machine mapping.
2. **Weeks 10–12:** exemption engine + risk scoring, parameter table, soft-decline retry, audit logging.
3. **Weeks 12–14:** liability/ECI capture, chargeback data plumbing, PDPA/PCI review, prep for scheme certification.
4. **Phase 4:** 3DS scheme certification (Visa/Mastercard) via sponsor bank + inclusion in the QSA's RoC.

---

## 11. ASSUMPTIONS / TODO (unresolved — confirm before production use)

> This callout marks external dependencies that are **not yet resolved** — do not treat as facts until confirmed.

- **[TODO-1] Sponsor bank / acquirer:** partner not yet locked (per `ROADMAP.md` critical path) — actual
  exemption thresholds, scheme certification, and liability rules must be confirmed with the sponsor.
- **[TODO-2] 3DS Server:** decide **build vs use an EMVCo-approved third-party 3DS Server** (affects PCI scope + timeline).
- **[TODO-3] THB thresholds:** LVP/TRA/force-challenge values here are EU PSD2 RTS baselines — convert/confirm to
  THB per Thai scheme rules and configure in the parameter table (never hard-code).
- **[TODO-4] Delegated Authentication:** requires extra certification and liability agreement with the scheme — not in MVP.
- **[TODO-5] QSA vendor:** QSA not yet selected — scoping of 3DS authentication data in the RoC awaits QSA confirmation.
- **[TODO-6] Actual registered capital:** document assumes THB 50M (Acquiring threshold) — confirm actual paid-up
  capital and maintenance of ≥75%.
- **[ASSUMPTION] EMVCo 3DS 2.2.0** is the minimum version and schemes still honor liability shift for status Y/A
  under current rules — re-validate whenever a scheme changes its mandate.
