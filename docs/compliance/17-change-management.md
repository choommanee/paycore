# นโยบายการบริหารการเปลี่ยนแปลง (Change Management) (ไทย)

> เอกสารประกอบการยื่นขอใบอนุญาต **บริการรับชำระเงินด้วยวิธีการทางอิเล็กทรอนิกส์ (Acquiring)**
> ภายใต้ พ.ร.บ. ระบบการชำระเงิน พ.ศ. 2560 · ทุนจดทะเบียนชำระแล้ว 50 ล้านบาท · PCI-DSS v4.0 Level 1
> เอกสารเลขที่ **17 — นโยบายการบริหารการเปลี่ยนแปลงและการปล่อยซอฟต์แวร์ (Change Management & Release Policy — PCI Req 6)**
> เจ้าของเอกสาร: **[บริษัท / Company]** · เวอร์ชัน 1.0 · จัดทำ 2026-07-22 · ทบทวนทุก 12 เดือน
>
> **หมายเหตุ:** เอกสารนี้เป็นเอกสารเชิงเทคนิค/ปฏิบัติการ ไม่ใช่คำแนะนำทางกฎหมาย ต้องผ่านการตรวจ
> โดยที่ปรึกษากฎหมายและ QSA ก่อนยื่นจริง เอกสารอ้างอิงร่วม: `COMPLIANCE-TH.md`, `ARCHITECTURE.md`, `ROADMAP.md`

---

## 1. วัตถุประสงค์และขอบเขต

นโยบายนี้กำหนดวิธีที่ [บริษัท / Company] **ควบคุมการเปลี่ยนแปลง (change)** ทุกชนิดต่อระบบที่อยู่ในและติดกับ
**Cardholder Data Environment (CDE)** ตั้งแต่การแก้โค้ด, การเปลี่ยน config, การเปลี่ยน infrastructure/network,
การ patch, ไปจนถึงการเปลี่ยนแปลงฉุกเฉิน โดยยึดหลัก **ทุกการเปลี่ยนแปลงต้องมีการอนุมัติ ทดสอบ และย้อนกลับได้
(approved, tested, reversible)** เพื่อรักษาความมั่นคงปลอดภัยและความต่อเนื่องของระบบชำระเงิน

นโยบายนี้ตอบข้อกำหนดหลักดังนี้:

- **PCI-DSS v4.0 Requirement 6** — พัฒนาและดูแลระบบ/ซอฟต์แวร์อย่างปลอดภัย โดยเฉพาะ 6.1–6.5
  (secure SDLC, การจัดการช่องโหว่, secure coding, change control และ **การแยกสภาพแวดล้อม dev/prod**)
- **PCI-DSS v4.0 Requirement 6.5.1–6.5.6** — ควบคุมการเปลี่ยนแปลง: เอกสารผลกระทบ, อนุมัติจากผู้มีอำนาจ,
  ทดสอบว่าไม่กระทบความปลอดภัย, มีขั้นตอน rollback, และแยกหน้าที่ dev/test/prod
- **PCI-DSS v4.0 Requirement 6.4.1–6.4.3** — ตรวจสอบ payment page scripts และ change ที่กระทบ CDE
- **ประกาศ ธปท.** ด้าน IT risk / cyber resilience — การควบคุมการเปลี่ยนแปลงระบบสารสนเทศและการแยกหน้าที่
- **PDPA / PDPC** — การเปลี่ยนแปลงที่กระทบข้อมูลส่วนบุคคลต้องผ่าน privacy review (DPIA เมื่อเข้าเกณฑ์)

> **ขอบเขต CDE ตาม `ARCHITECTURE.md`:** ทุกการเปลี่ยนแปลงต่อ Payment Core, Tokenization Vault, Acquirer/3DS
> Adapter, Ledger, network segmentation, HSM/KMS และ pipeline CI/CD ถือเป็น **in-scope PCI** ต้องผ่านนโยบายนี้ครบทุกขั้น

---

## 2. บทบาทและความรับผิดชอบ (Roles & Responsibilities)

| บทบาท | ความรับผิดชอบด้าน change management |
|-------|--------------------------------------|
| **Change Requester (ผู้ขอ)** | สร้าง Change Request (CR), เขียน scope/impact/rollback, แนบผลทดสอบ |
| **Engineering Lead / Peer Reviewer** | review โค้ด(≥1 คนที่ไม่ใช่ผู้เขียน), ตรวจ secure coding, อนุมัติ merge |
| **QA Lead** | ยืนยันผลทดสอบ, regression, security test ผ่านก่อนปล่อย |
| **Change Advisory Board (CAB)** | อนุมัติ change ระดับ significant/high ตาม threshold ในข้อ 5 |
| **Release Manager** | จัดตารางปล่อย, คุม change window, สั่ง go/no-go, ยืนยัน rollback readiness |
| **CISO / DevSecOps** | เจ้าของนโยบาย, อนุมัติ change ที่กระทบความปลอดภัย/CDE, คุม secrets/HSM/KMS |
| **SRE / Infra** | ดำเนินการ deploy สู่ prod, monitor, ทำ rollback |
| **DPO** | privacy review / DPIA สำหรับ change ที่กระทบข้อมูลส่วนบุคคล (PDPA) |
| **Internal Audit** | ตรวจสอบว่าทุก prod change มี CR + อนุมัติ + ผลทดสอบ ครบตาม PCI Req 6 |

> **หลัก Separation of Duties (SoD):** ผู้เขียนโค้ด **ห้าม** อนุมัติ change ของตัวเอง และ **ห้าม** เป็นผู้ deploy
> สู่ production ด้วยตนเอง การ deploy สู่ prod ทำผ่าน pipeline อัตโนมัติที่ต้องมีผู้มีสิทธิ์ (approver) กดอนุมัติแยกจากผู้เขียน

---

## 3. การจัดชั้นการเปลี่ยนแปลง (Change Classification)

| ประเภท | นิยาม | ตัวอย่าง | ผู้อนุมัติ | Lead time ขั้นต่ำ |
|--------|-------|---------|-----------|-------------------|
| **Standard (pre-approved)** | change ซ้ำ ๆ ความเสี่ยงต่ำ มีขั้นตอนสำเร็จรูป | อัปเดต dependency minor ที่ผ่าน scan, config ที่ไม่กระทบ CDE | Engineering Lead (อนุมัติล่วงหน้าตาม runbook) | 1 วันทำการ |
| **Normal — Low/Medium** | change ปกติ กระทบ non-CDE หรือ CDE เล็กน้อย | feature ใหม่ใน API non-CDE, แก้ bug | Engineering Lead + QA Lead | 2 วันทำการ |
| **Significant / High** | กระทบ CDE, network segmentation, crypto/key, acquirer/3DS, schema เงิน | เปลี่ยน vault, HSM rotation, เปลี่ยน acquirer adapter, schema `payments`/`ledger_entries` | **CAB** + CISO | 5 วันทำการ |
| **Emergency** | แก้ปัญหาที่กระทบบริการ/ความปลอดภัยเร่งด่วน | hotfix ช่องโหว่ critical, incident production | Emergency approver (CISO หรือผู้แทน) + retro CAB | ทันที (ดูข้อ 7) |

การจัดชั้นบันทึกไว้ใน CR และเป็นตัวกำหนดว่าต้องผ่านกี่ระดับอนุมัติ

---

## 4. วงจรการเปลี่ยนแปลง (Change Lifecycle — SDLC ตาม PCI 6.2–6.5)

```
ขอ (CR)  ─▶  ออกแบบ/พัฒนา  ─▶  Peer review + SAST/SCA  ─▶  ทดสอบใน dev/staging
   │                                                              │
   │                                                              ▼
   └──────────────  อนุมัติ (ตาม threshold)  ◀────  QA sign-off + security test
                             │
                             ▼
                จัดตารางปล่อย (change window)  ─▶  Deploy prod (pipeline + approver)
                             │
                             ▼
             Post-deploy verification (smoke)  ─▶  ปิด CR / ถ้า fail ─▶ Rollback (ข้อ 8)
```

**ข้อบังคับต่อทุก CR ก่อนขึ้น production (PCI 6.5.1):**

1. **เอกสารผลกระทบ (impact analysis)** — ระบุระบบที่กระทบ, ผล PCI scope, ผลต่อ PDPA
2. **การอนุมัติจากผู้มีอำนาจ** — ตาม threshold ในข้อ 5 (ต้องเป็นคนละคนกับผู้เขียน)
3. **การทดสอบว่า change ไม่ลดทอนความปลอดภัย** — functional + security regression
4. **ขั้นตอน rollback** — เขียนไว้ล่วงหน้าและทดสอบได้ (ดูข้อ 8)

---

## 5. เกณฑ์การอนุมัติ (Approval Matrix)

| ระดับ change | Peer review (โค้ด) | QA sign-off | อนุมัติเพื่อปล่อย | จำนวนผู้อนุมัติ |
|--------------|---------------------|-------------|-------------------|-----------------|
| Standard | 1 | อัตโนมัติ (CI) | Engineering Lead (pre-approved) | 1 |
| Normal Low/Medium | ≥1 (ไม่ใช่ผู้เขียน) | required | Engineering Lead + Release Manager | 2 |
| Significant / High (CDE) | ≥2 รวม security reviewer | required + pentest ถ้าเข้าเกณฑ์ | **CAB** (Eng Lead + Release Mgr + CISO) | ≥3 |
| Emergency | ≥1 (post หรือ pre) | smoke test | Emergency approver, retro ที่ CAB ครั้งถัดไป | 1 (+ retro) |

- ทุกการอนุมัติต้องบันทึกใน ticketing/Git (ผู้อนุมัติ, เวลา, เหตุผล) และ **ห้ามอนุมัติ CR ของตนเอง**
- Merge เข้า branch ที่ deploy ได้ ต้องผ่าน **branch protection**: ผ่าน CI (build, test, SAST, SCA `govulncheck`),
  peer review ที่กำหนด, และไม่มี finding ระดับ high/critical ค้าง

---

## 6. การแยกสภาพแวดล้อม dev / test / prod (Segregation — PCI 6.5.3, 6.5.4)

| มิติ | Development | Staging / Test | Production (CDE) |
|------|-------------|----------------|-----------------|
| **เครือข่าย** | segment แยก, ไม่มี route ตรงสู่ prod | segment แยก | segment CDE, WAF/IDS, firewall (Req 1) |
| **ข้อมูล** | **ห้ามใช้ข้อมูลจริง/PAN จริง** ใช้ข้อมูลสังเคราะห์ | ข้อมูลสังเคราะห์ / masked เท่านั้น | ข้อมูลจริง เข้ารหัส |
| **สิทธิ์เข้าถึง** | นักพัฒนา | QA + นักพัฒนา (จำกัด) | SRE เท่านั้น ผ่าน MFA + bastion, least privilege (Req 7-8) |
| **ผู้ deploy** | นักพัฒนา | pipeline | **pipeline + approver แยกจากผู้เขียน** |
| **Secrets/Key** | คีย์ทดสอบ | คีย์ทดสอบ | HSM/KMS จริง, dual control (Req 3) |

**ข้อบังคับ (PCI 6.5.3 / 6.5.4):**

- แยก **หน้าที่ (roles)** และ **สิทธิ์ (access)** ระหว่างสภาพแวดล้อม dev/test และ prod อย่างชัดเจน
- **ห้าม** ใช้ **PAN จริง / cardholder data จริง** ในสภาพแวดล้อม dev/test เด็ดขาด (สอดคล้อง `ARCHITECTURE.md` ข้อ 6
  และ PCI 6.5.5) — ใช้ข้อมูลสังเคราะห์หรือ token/mask เท่านั้น
- **ห้าม** ทิ้ง test data, account ทดสอบ, hardcoded credential ไว้ในโค้ดก่อนขึ้น prod (PCI 6.5.6) — ตรวจด้วย SAST + secret scanning
- นักพัฒนา **ไม่มี** สิทธิ์ standing access เข้า production; หากต้องเข้าใช้ **break-glass** (ข้อ 7) มี MFA + บันทึก `audit_log`

---

## 7. การเปลี่ยนแปลงฉุกเฉิน (Emergency / Break-glass Change)

1. ประกาศเป็น **Emergency CR** พร้อมเหตุผลและระดับผลกระทบ
2. **Emergency approver** (CISO หรือผู้แทนที่กำหนด) อนุมัติวาจา/แชท แล้วบันทึกเป็นลายลักษณ์ภายใน **1 ชั่วโมง**
3. ดำเนินการโดยมี **ผู้ทำ + ผู้สังเกตการณ์ (2 คน)** และบันทึกทุกขั้นลง `audit_log`
4. เมื่อเสถียร ให้ทำ **post-implementation review** และนำเข้า **CAB ครั้งถัดไปภายใน 5 วันทำการ** เพื่อ ratify ย้อนหลัง
5. Break-glass credential ต้องถูก **rotate ทันที** หลังใช้ และตรวจสอบ log การใช้งาน

---

## 8. การย้อนกลับ (Rollback / Back-out)

**ทุก production change ต้องมีแผน rollback ที่เขียนไว้ล่วงหน้าและทดสอบได้** ก่อนได้รับอนุมัติปล่อย

| กลไก | วิธีการ |
|------|--------|
| **Application** | deploy แบบ immutable image (distroless ตาม `ARCHITECTURE.md`); rollback = redeploy image เวอร์ชันก่อนหน้า (blue-green / canary) |
| **Database schema** | migration แบบ **backward-compatible / expand-contract**; ทุก `up` ต้องมี `down` ที่ทดสอบใน staging; หลีกเลี่ยง destructive migration ในรอบเดียว |
| **Config / feature flag** | เปลี่ยนผ่าน flag ที่ปิดได้ทันทีโดยไม่ต้อง redeploy |
| **Ledger (append-only)** | **ห้ามลบ/แก้ย้อนหลัง** — แก้ด้วย compensating entry เท่านั้น (fail closed ตามหลักสถาปัตยกรรม) |

**เกณฑ์สั่ง rollback (rollback triggers):**

- Smoke/health check หลัง deploy ไม่ผ่านภายใน **15 นาที**
- Error rate เพิ่ม > เกณฑ์ที่ตั้ง หรือ authorization success rate ลดผิดปกติ
- พบผลกระทบต่อความปลอดภัย/ข้อมูลบัตร → rollback ทันทีและเปิด incident

Release Manager หรือ SRE on-call มีอำนาจสั่ง rollback ทันที; ต้องบันทึกใน CR และแจ้ง CAB

---

## 9. การจัดการช่องโหว่และ patch (PCI 6.2, 6.3)

| ระดับความรุนแรง | ระยะเวลา patch สูงสุด |
|------------------|----------------------|
| Critical / High (CVSS ≥ 7.0 หรือกระทบ CDE) | **ภายใน 30 วัน** (เร่งด่วนกว่าถ้ามี exploit จริง) |
| Medium | ภายใน 90 วัน |
| Low | ตามรอบ maintenance ปกติ |

- สแกน dependency ทุก build ด้วย `govulncheck` + SCA; SAST/DAST ใน CI (ตาม `ARCHITECTURE.md` ข้อ 7)
- **Quarterly ASV scan** และ **annual penetration test** ตามที่ระบุใน `COMPLIANCE-TH.md` ข้อ 6
- Change ที่กระทบ **payment page scripts** ต้องผ่าน inventory + authorization ตาม PCI 6.4.3

---

## 10. บันทึกและการตรวจสอบ (Records & Audit)

- **CR ทุกใบ** เก็บ: ผู้ขอ, ผู้อนุมัติ, ผลทดสอบ, แผน rollback, ผลลัพธ์ — เก็บ **≥ 12 เดือน** (สอดคล้อง PCI Req 10
  และ retention ในเอกสารเลขที่ 11)
- ทุก prod change ต้อง traceable: CR ↔ Git commit/PR ↔ pipeline run ↔ approver
- Internal Audit สุ่มตรวจรายไตรมาสว่าไม่มี prod change ที่ไม่มี CR/อนุมัติ (unauthorized change = ข้อบกพร่องร้ายแรง)

---

## 11. ข้อสมมติและสิ่งที่ต้องดำเนินการ (Assumptions & TODO)

> **⚠️ ต้องยืนยันก่อนยื่น — ข้อมูลภายนอกที่ยังไม่สรุป (ห้ามสมมติเป็นข้อเท็จจริง):**
>
> - **[TODO — Sponsor bank / Acquirer]** ยังไม่สรุปคู่สัญญา; ข้อกำหนด change/certification window ของ scheme
>   (Visa/Mastercard) และ acquirer จะเพิ่มเงื่อนไขต่อ release ที่กระทบ authorization/settlement — ต้องผนวกเมื่อสรุป
> - **[TODO — QSA vendor]** ยังไม่ลงนามผู้ประเมิน; ต้องให้ QSA ทบทวนขั้นตอนนี้เทียบ PCI-DSS v4.0 Req 6 ก่อน RoC
> - **[TODO — ทุนจดทะเบียนที่ชำระจริง]** เอกสารอ้างอิงเกณฑ์ 50 ล้านบาท (Acquiring); ยอดที่ชำระจริงต้องยืนยันด้วย
>   หลักฐานจากกรมพัฒนาธุรกิจการค้า/งบการเงิน (ดูเอกสารเลขที่ 02)
> - **[TODO — เครื่องมือ CI/CD & ticketing จริง]** ชื่อผลิตภัณฑ์ (เช่น Git host, pipeline, ticketing) ให้ระบุเมื่อจัดซื้อจริง

---

# Change management & release policy (PCI Req 6): approvals, testing, segregation of dev/prod, rollback (English)

> Supporting document for the **Acquiring (electronic payment acceptance) service** license application
> under the Payment Systems Act B.E. 2560 · Paid-up capital THB 50M · PCI-DSS v4.0 Level 1
> Document **17 — Change Management & Release Policy (PCI Req 6)**
> Owner: **[บริษัท / Company]** · Version 1.0 · Issued 2026-07-22 · Reviewed every 12 months
>
> **Note:** This is a technical/operational document, not legal advice. It must be reviewed by legal counsel
> and the QSA before submission. Related documents: `COMPLIANCE-TH.md`, `ARCHITECTURE.md`, `ROADMAP.md`

---

## 1. Purpose & Scope

This policy defines how [บริษัท / Company] **controls all changes** to systems within and connected to the
**Cardholder Data Environment (CDE)** — code, configuration, infrastructure/network, patches, and emergency
changes — under the principle that **every change must be approved, tested, and reversible**, in order to
preserve the security and continuity of the payment system.

It satisfies the following key requirements:

- **PCI-DSS v4.0 Requirement 6** — develop and maintain secure systems/software, in particular 6.1–6.5
  (secure SDLC, vulnerability management, secure coding, change control, and **dev/prod segregation**)
- **PCI-DSS v4.0 Req 6.5.1–6.5.6** — change control: documented impact, authorized approval, testing that
  security is not reduced, rollback procedures, and separation of dev/test/prod duties and environments
- **PCI-DSS v4.0 Req 6.4.1–6.4.3** — review/authorize payment page scripts and CDE-affecting changes
- **Bank of Thailand (ธปท.) notifications** on IT risk / cyber resilience — IT change control and SoD
- **PDPA / PDPC** — changes affecting personal data require a privacy review (DPIA where triggered)

> **CDE scope per `ARCHITECTURE.md`:** any change to the Payment Core, Tokenization Vault, Acquirer/3DS
> adapters, Ledger, network segmentation, HSM/KMS, and the CI/CD pipeline is **PCI in-scope** and must pass
> every stage of this policy.

---

## 2. Roles & Responsibilities

| Role | Change-management responsibility |
|------|----------------------------------|
| **Change Requester** | Raises the Change Request (CR); documents scope/impact/rollback; attaches test evidence |
| **Engineering Lead / Peer Reviewer** | Reviews code (≥1 non-author), checks secure coding, approves merge |
| **QA Lead** | Confirms functional + security regression pass before release |
| **Change Advisory Board (CAB)** | Approves significant/high-risk changes per §5 thresholds |
| **Release Manager** | Schedules releases, owns the change window, gives go/no-go, confirms rollback readiness |
| **CISO / DevSecOps** | Policy owner; approves security/CDE-affecting changes; controls secrets/HSM/KMS |
| **SRE / Infra** | Executes production deploys, monitors, performs rollback |
| **DPO** | Privacy review / DPIA for changes affecting personal data (PDPA) |
| **Internal Audit** | Verifies every prod change has a CR + approval + test evidence per PCI Req 6 |

> **Separation of Duties (SoD):** an author **must not** approve their own change and **must not** deploy it to
> production themselves. Production deploys run through an automated pipeline requiring a separate authorized approver.

---

## 3. Change Classification

| Type | Definition | Examples | Approver | Min lead time |
|------|-----------|----------|----------|---------------|
| **Standard (pre-approved)** | Low-risk, repeatable, runbook-driven | Minor dependency bump passing scans; non-CDE config | Engineering Lead (pre-approved via runbook) | 1 business day |
| **Normal — Low/Medium** | Ordinary change; non-CDE or minor CDE impact | New non-CDE API feature; bug fix | Engineering Lead + QA Lead | 2 business days |
| **Significant / High** | Affects CDE, segmentation, crypto/keys, acquirer/3DS, money schema | Vault change, HSM rotation, acquirer adapter, `payments`/`ledger_entries` schema | **CAB** + CISO | 5 business days |
| **Emergency** | Urgent fix for a service/security incident | Critical-vuln hotfix; production incident | Emergency approver (CISO or delegate) + retro CAB | Immediate (§7) |

The classification is recorded in the CR and drives how many approval tiers apply.

---

## 4. Change Lifecycle (SDLC per PCI 6.2–6.5)

```
Request (CR) ─▶ Design/Develop ─▶ Peer review + SAST/SCA ─▶ Test in dev/staging
   │                                                             │
   │                                                             ▼
   └──────────────  Approval (per threshold)  ◀────  QA sign-off + security test
                            │
                            ▼
               Schedule release (change window) ─▶ Deploy prod (pipeline + approver)
                            │
                            ▼
            Post-deploy verification (smoke) ─▶ Close CR / on fail ─▶ Rollback (§8)
```

**Mandatory for every CR before production (PCI 6.5.1):**

1. **Documented impact analysis** — affected systems, PCI-scope impact, PDPA impact
2. **Authorized approval** — per §5 thresholds (must differ from the author)
3. **Testing that security is not reduced** — functional + security regression
4. **Rollback procedure** — pre-written and tested (see §8)

---

## 5. Approval Matrix

| Change level | Code peer review | QA sign-off | Release approval | # approvers |
|--------------|------------------|-------------|------------------|-------------|
| Standard | 1 | Automated (CI) | Engineering Lead (pre-approved) | 1 |
| Normal Low/Medium | ≥1 (non-author) | Required | Engineering Lead + Release Manager | 2 |
| Significant / High (CDE) | ≥2 incl. security reviewer | Required + pentest if triggered | **CAB** (Eng Lead + Release Mgr + CISO) | ≥3 |
| Emergency | ≥1 (pre or post) | Smoke test | Emergency approver, retro at next CAB | 1 (+ retro) |

- Every approval is logged in ticketing/Git (approver, timestamp, rationale); **self-approval is prohibited**.
- Merging into a deployable branch requires **branch protection**: CI green (build, test, SAST, SCA `govulncheck`),
  required peer review, and no open high/critical findings.

---

## 6. Segregation of dev / test / prod (PCI 6.5.3, 6.5.4)

| Dimension | Development | Staging / Test | Production (CDE) |
|-----------|-------------|----------------|------------------|
| **Network** | Separate segment; no direct route to prod | Separate segment | CDE segment, WAF/IDS, firewall (Req 1) |
| **Data** | **No real data/PAN**; synthetic only | Synthetic / masked only | Real data, encrypted |
| **Access** | Developers | QA + limited devs | SRE only, via MFA + bastion, least privilege (Req 7-8) |
| **Deployer** | Developers | Pipeline | **Pipeline + approver separate from author** |
| **Secrets/Keys** | Test keys | Test keys | Real HSM/KMS, dual control (Req 3) |

**Mandatory (PCI 6.5.3 / 6.5.4):**

- Clearly separate **roles** and **access** between dev/test and production.
- **Never** use **live PANs / live cardholder data** in dev/test (per `ARCHITECTURE.md` §6 and PCI 6.5.5) —
  synthetic or token/masked data only.
- **No** leftover test data, test accounts, or hardcoded credentials before production (PCI 6.5.6) — enforced by
  SAST + secret scanning.
- Developers hold **no standing production access**; access requires **break-glass** (§7) with MFA and `audit_log` entry.

---

## 7. Emergency / Break-glass Change

1. Raise an **Emergency CR** with rationale and impact level.
2. The **Emergency approver** (CISO or delegate) approves verbally/chat, then records it in writing within **1 hour**.
3. Execute with a **doer + observer (two people)**, logging every step to `audit_log`.
4. Once stable, run a **post-implementation review** and bring it to the **next CAB within 5 business days** to ratify.
5. Any break-glass credential is **rotated immediately** after use, and its usage logs are reviewed.

---

## 8. Rollback / Back-out

**Every production change must have a pre-written, testable rollback plan** before release approval.

| Mechanism | Method |
|-----------|--------|
| **Application** | Immutable image deploys (distroless per `ARCHITECTURE.md`); rollback = redeploy prior image (blue-green / canary) |
| **Database schema** | **Backward-compatible / expand-contract** migrations; every `up` has a `down` tested in staging; avoid single-step destructive migrations |
| **Config / feature flag** | Toggle via flags that disable instantly without redeploy |
| **Ledger (append-only)** | **Never delete/edit retroactively** — correct via compensating entries only (fail-closed per architecture) |

**Rollback triggers:**

- Post-deploy smoke/health check fails within **15 minutes**
- Error rate exceeds threshold, or authorization success rate drops abnormally
- Any security/cardholder-data impact detected → immediate rollback and incident opened

The Release Manager or on-call SRE may trigger rollback immediately; it is logged in the CR and reported to the CAB.

---

## 9. Vulnerability & Patch Management (PCI 6.2, 6.3)

| Severity | Max patch time |
|----------|----------------|
| Critical / High (CVSS ≥ 7.0 or CDE-affecting) | **Within 30 days** (faster if actively exploited) |
| Medium | Within 90 days |
| Low | Next regular maintenance cycle |

- Scan dependencies on every build with `govulncheck` + SCA; SAST/DAST in CI (per `ARCHITECTURE.md` §7).
- **Quarterly ASV scan** and **annual penetration test** per `COMPLIANCE-TH.md` §6.
- Changes touching **payment page scripts** pass inventory + authorization per PCI 6.4.3.

---

## 10. Records & Audit

- **Every CR** retains: requester, approver(s), test evidence, rollback plan, outcome — kept **≥ 12 months**
  (aligned with PCI Req 10 and the retention schedule in document 11).
- Every prod change is traceable: CR ↔ Git commit/PR ↔ pipeline run ↔ approver.
- Internal Audit samples quarterly to confirm no prod change lacks a CR/approval (an unauthorized change is a major finding).

---

## 11. Assumptions & TODO

> **⚠️ Confirm before submission — unresolved external dependencies (do not invent as fact):**
>
> - **[TODO — Sponsor bank / Acquirer]** Counterparty not finalized; the scheme (Visa/Mastercard) and acquirer
>   change/certification windows will add release constraints for authorization/settlement-affecting changes — to be merged once signed.
> - **[TODO — QSA vendor]** Assessor not yet engaged; the QSA must review this procedure against PCI-DSS v4.0 Req 6 before the RoC.
> - **[TODO — Actual paid-up capital]** This document references the THB 50M Acquiring threshold; the actual paid-up
>   amount must be evidenced via DBD registration / financial statements (see document 02).
> - **[TODO — Actual CI/CD & ticketing tooling]** Product names (Git host, pipeline, ticketing) to be filled in once procured.
