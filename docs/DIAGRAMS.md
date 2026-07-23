# Diagrams (Mermaid)

เปิดไฟล์นี้ในเครื่องมือที่ render mermaid ได้ (VS Code + Mermaid, GitHub, mermaid.live)

## 1. System Context

```mermaid
flowchart TD
    M["Merchant<br/>(web / app / POS)"] -->|HTTPS + API key + Idempotency-Key| GW

    subgraph GW["Payment Gateway (Go/Fiber)"]
        EDGE["API Edge / Middleware"]
        CORE["Payment Core (Service)"]
        VAULT["Tokenization Vault (PCI)"]
        RISK["Risk / Fraud Engine"]
        LEDGER["Ledger (append-only)"]
        HOOK["Webhook / Notifier"]
        EDGE --> CORE
        CORE --> VAULT
        CORE --> RISK
        CORE --> LEDGER
        CORE --> HOOK
    end

    CORE -->|ISO 8583 / API| ACQ["Acquirer / Card Switch"]
    CORE -->|3DS 2.x| TDS["3-D Secure (ACS/DS)"]
    ACQ --> NET["Card Networks<br/>Visa / Mastercard"]
    NET --> ISS["Issuing Banks"]
    HOOK -->|signed webhook| M
    CORE -.settlement.-> SET["Settlement / Recon<br/>(sponsor bank / ITMX)"]
```

## 2. Component / Layers (Clean Architecture)

```mermaid
flowchart LR
    subgraph API["internal/handler"]
        H1["PaymentHandler"]
        H2["HealthHandler"]
    end
    subgraph SVC["internal/service"]
        S1["PaymentService<br/>state machine"]
    end
    subgraph EXT["internal/external"]
        E1["Acquirer (iface)"]
        E2["ThreeDS (iface)"]
    end
    subgraph REPO["internal/repository (sqlc)"]
        R1["Querier"]
    end
    DB[("PostgreSQL")]
    H1 --> S1
    S1 --> E1
    S1 --> E2
    S1 --> R1
    R1 --> DB
```

## 3. Sequence — Authorize + Capture (with 3DS)

```mermaid
sequenceDiagram
    autonumber
    participant Mer as Merchant
    participant GW as Gateway
    participant V as Vault
    participant TDS as 3DS
    participant ACQ as Acquirer
    participant DB as DB/Ledger

    Mer->>GW: POST /v1/payments (token, Idempotency-Key)
    GW->>DB: check idempotency key
    GW->>V: detokenize -> PAN (PCI scope)
    GW->>GW: risk scoring
    GW->>TDS: authenticate (amount, returnURL)
    TDS-->>GW: challenge required
    GW-->>Mer: 200 requires_action + next_action_url
    Mer->>TDS: complete 3DS challenge
    TDS-->>GW: POST /3ds/return (success + cryptogram)
    GW->>ACQ: authorize (PAN, amount, cryptogram)
    ACQ-->>GW: approved (auth_code, acquirer_ref)
    GW->>DB: insert payment + ledger(authorize) [tx]
    GW->>ACQ: capture
    ACQ-->>GW: captured
    GW->>DB: ledger(capture), status=captured
    GW-->>Mer: webhook payment.captured
```

## 4. Payment State Machine

```mermaid
stateDiagram-v2
    [*] --> requires_action: 3DS needed
    [*] --> authorized: no challenge
    requires_action --> authorized: 3DS ok
    requires_action --> failed: 3DS fail
    authorized --> captured: capture
    authorized --> voided: void
    captured --> partial_refunded: partial refund
    partial_refunded --> refunded: remaining refund
    captured --> refunded: full refund
    failed --> [*]
    voided --> [*]
    refunded --> [*]
```

## 5. Data Model (ERD)

```mermaid
erDiagram
    merchants ||--o{ payments : has
    payments ||--o{ ledger_entries : records
    payments ||--o{ refunds : has
    merchants ||--o{ webhook_events : receives

    merchants {
        uuid id PK
        text name
        text status
        text api_key_hash
        text settlement_currency
    }
    payments {
        uuid id PK
        uuid merchant_id FK
        bigint amount_minor
        bigint captured_amount_minor
        bigint refunded_amount_minor
        text currency
        text status
        char card_last4
        text idempotency_key
    }
    ledger_entries {
        uuid id PK
        uuid payment_id FK
        text entry_type
        bigint amount_minor
        text currency
    }
    refunds {
        uuid id PK
        uuid payment_id FK
        bigint amount_minor
        text status
    }
    webhook_events {
        uuid id PK
        uuid merchant_id FK
        text event_type
        jsonb payload
        bool delivered
    }
```
