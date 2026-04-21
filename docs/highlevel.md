# HOF Backend — High-Level Architecture

## Purpose

The HOF Backend serves as the data and business-logic layer for the Heritage of Faith
Church mobile application. It manages:

- **Users** — registration, login, profile, roles, device tracking
- **Audio content** — sermon messages and series, meditations, homepage curation
- **Subscriptions** — plan catalogue, Paystack payment processing, access control
- **File storage** — sermon artwork and audio hosted on AWS S3

---

## Bounded contexts

The system is divided into three main domains:

```
┌──────────────────────────────────────────────────────────────────┐
│                        HOF Platform                               │
│                                                                  │
│  ┌────────────┐    ┌─────────────────┐    ┌──────────────────┐  │
│  │   User &   │    │ Audio Content   │    │  Subscription &  │  │
│  │   Roles    │───▶│ (Messages,      │◀───│  Payment         │  │
│  │  Context   │    │  Series,        │    │  Context         │  │
│  │            │    │  Meditations)   │    │                  │  │
│  └────────────┘    └─────────────────┘    └──────────────────┘  │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

### User & Roles context

Handles user identity, authentication, and role-based authorization.

**Key concepts:**
- A `User` is the aggregate root.
- A user may hold one or more `Role`s simultaneously (1:M via `user_roles` join table).
- Roles: `member` (default), `steward`, `church_friend`, `team_lead`, `church_admin`.
- Passwords are bcrypt-hashed; legacy MD5 accounts upgrade on first login.

### Audio Content context

Manages all audio-based content items served to users.

**Key concepts:**
- `AudioMessage` — a single sermon or teaching recording.
  - `is_free`: accessible without a subscription.
  - `allow_steward`: accessible to steward-role users without a subscription.
- `AudioSeries` — a named playlist of related messages.
- `Meditation` — a short guided meditation clip.

### Subscription & Payment context

Controls access tiers and processes payments via Paystack.

**Key concepts:**
- `Plan` — a purchasable tier (e.g. Monthly Premium).
- `Offering` — a named feature included in a plan.
- `Subscription` — a user's current entitlement record.
- `GlobalParameters` — app-wide feature flags (e.g. `activate_subscription`).

---

## Layered architecture (Domain-Driven Design)

```
┌────────────────────────────────────────────────────┐
│               Interfaces / HTTP                     │  ← chi router, handlers, middleware
├────────────────────────────────────────────────────┤
│               Application Services                  │  ← use cases, DTOs, orchestration
├────────────────────────────────────────────────────┤
│               Domain                                │  ← entities, repo interfaces, errors
├────────────────────────────────────────────────────┤
│               Infrastructure                        │  ← GORM, S3, Paystack, SMTP, config
└────────────────────────────────────────────────────┘
```

Each layer **only depends on the layer below it**. Domain code has zero external imports;
the application layer depends on domain interfaces only; infrastructure implements those
interfaces and knows about the ORM, cloud SDKs, etc.

---

## External integrations

| System       | Purpose                          | Direction   |
|--------------|----------------------------------|-------------|
| PostgreSQL   | Primary data store               | Outbound    |
| AWS S3       | Audio/image file hosting         | Outbound    |
| Paystack     | Payment processing, subscriptions| Outbound    |
| SendinBlue (SMTP) | Transactional email         | Outbound    |

---

## Security model

1. **JWT Bearer tokens** — short-lived access token (48 h) + long-lived refresh (30 d).
2. **bcrypt** passwords (cost 12) — transparent upgrade from legacy MD5 on login.
3. **CORS** — allow-listed HTTP methods and headers via go-chi/cors.
4. **Role-based access** — enforced at the application-service level based on claims
   stored in the JWT.
5. **Webhook verification** — Paystack webhook payloads are validated before processing.

---

## Scalability considerations

- The service is **stateless** — horizontal scaling is straightforward behind a load balancer.
- JWT removes the need for session storage; any instance can validate any token.
- S3 pre-signed URLs or CDN fronting can be added for audio delivery without touching
  the API layer.
- The `GlobalParameters` table allows feature-flagging (e.g. disabling subscriptions
  during maintenance) without a deployment.
