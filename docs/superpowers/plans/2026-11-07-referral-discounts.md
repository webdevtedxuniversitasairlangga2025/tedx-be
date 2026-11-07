# Referral Discounts Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement admin-managed fixed-amount referral discounts for all ticket orders, with safe quota holds, per-ticket-holder redemption limits, and checkout preview.

**Architecture:** Add backend referral code and claim persistence, admin endpoints, preview and transactional order integration. Store order discount snapshots and hold referral quota with ticket quota; redeem on payment approval and release on failed/expired orders. Update the frontend IdentifyStepper checkout and add admin code management UI.

**Tech Stack:** Go 1.25, Gin, GORM, PostgreSQL, existing `samber/do` DI and middleware; React 18, Vite, existing Axios service layer and admin routing.

**Spec:** `docs/superpowers/specs/2026-11-07-referral-discounts-design.md`

## Global Constraints

- Codes apply to all ticket types and tiers sold through the ticket-order checkout, including Main Event and Pre-Event; they do not apply to merchandise or bundles.
- Admin manually enters a unique code, fixed discount amount per ticket, and total quota; after creation, only activation/deactivation is allowed.
- Code strings are unique and are not reused, including after deactivation.
- Discount per ticket must be strictly less than the selected tier's unit price.
- A normalized ticket-holder email can redeem a given code at most once, including across orders; repeated email within one order receives at most one discount.
- If order request omits attendee details, use the same effective ticket-holder email(s) existing order creation uses to create attendee tickets.
- Allocate discounts from the top of the attendee list down, skipping ineligible or already-claimed emails.
- Preview does not hold quota. Order creation revalidates and atomically holds ticket and referral quotas.
- Held referral claims become permanent on paid approval and are released on cancellation, rejection, or expiry; release is idempotent.
- An inactive, exhausted, or inapplicable code must allow order creation at normal price after informing the customer.
- Backend-calculated prices are authoritative; frontend must not set prices or discount amounts.
- Existing checkout without a referral code must remain unchanged.

---

## File Map

### Backend repo: `/mnt/c/Dev/tedx-be`

- Create `database/entities/referral_code_entity.go`: persisted referral configuration and held quota.
- Create `database/entities/referral_claim_entity.go`: one held/redeemed claim per discounted attendee email.
- Modify `database/migrations.go`: AutoMigrate new entities.
- Create `modules/referral/dto/referral_dto.go`: create/toggle/preview request and response DTOs plus service errors.
- Create `modules/referral/repository/referral_repository.go`: database operations, locking, uniqueness, claim lifecycle.
- Create `modules/referral/service/referral_service.go`: admin management and preview/business calculation.
- Create `modules/referral/handler/referral_handler.go`: HTTP binding and response mapping.
- Create `modules/referral/routes.go`: authenticated admin routes and authenticated preview route.
- Modify `providers/core.go`, `cmd/main.go`: dependency registration and route registration.
- Modify `modules/order/dto/order_dto.go`, `modules/order/service/order_service.go`, `modules/order/repository/order_repository.go`: accept code, transactionally allocate/hold, snapshots, paid/release transitions.
- Add focused tests under `modules/referral/{service,repository,handler}` and extend `modules/order/service/order_service_test.go`.
- Create `API_Test/Referral/` Bruno requests for admin CRUD/toggle and preview.

### Frontend repo: `/mnt/c/Dev/tedxua-frontend`

- Create `src/services/referralService.js`: preview and admin referral API calls.
- Modify `src/services/orderService.js`: include optional `referral_code` in `createOrder`.
- Modify `src/features/tickets/components/IdentifyStepper.jsx`: code entry, preview, authoritative final allocation display and continue behavior.
- Add admin referral-management page under `src/features/admin/pages/` and register route/menu using current admin conventions.
- Add/update tests using the existing frontend test setup; if no test runner exists, do not add a new dependency without approval—verify via existing lint/build scripts.

## Task 1: Add referral persistence and migration

**Files:**
- Create: `database/entities/referral_code_entity.go`
- Create: `database/entities/referral_claim_entity.go`
- Modify: `database/migrations.go`
- Test: existing database migration/test setup (inspect before adding; otherwise entity schema tests under `database/entities/`)

**Interfaces:**
- Code entity fields: `ID`, normalized `Code`, `DiscountPerTicket decimal.Decimal`, `Quota int`, `QuotaHeld int`, `IsActive bool`, timestamps.
- Claim fields: `ID`, `ReferralCodeID`, `OrderID`, normalized `AttendeeEmail`, `AttendeePosition`, `DiscountAmount decimal.Decimal`, `Status`, timestamps.
- Add unique constraints for code and active/held-or-redeemed code+email claim. If released claims are retained, use PostgreSQL partial uniqueness on `status IN ('held','redeemed')`; otherwise delete released held claims.

- [ ] **Step 1: Write schema-focused tests** proving expected fields/tags/indexes and valid status constants.
- [ ] **Step 2: Run the focused test**: `go test ./database/entities/... -run Referral -v`; expected failure until entities/constants exist.
- [ ] **Step 3: Add entities and migration registration**, following UUID, Timestamp and AutoMigrate conventions.
- [ ] **Step 4: Run focused tests and migration compilation**: `go test ./database/entities/...` and `go test ./database/...`.
- [ ] **Step 5: Review migration compatibility** against existing PostgreSQL UUID and naming conventions.

## Task 2: Implement deterministic eligibility/allocation logic

**Files:**
- Create: `modules/referral/service/referral_service.go`
- Create: `modules/referral/service/referral_service_test.go`
- Create: `modules/referral/dto/referral_dto.go`

**Interfaces:**
- `normalizeReferralCode(string) string`: trim and uppercase code.
- `normalizeAttendeeEmail(string) string`: trim and lowercase email.
- `allocateDiscount(attendees []dto.ReferralAttendeeInput, priorClaims map[string]bool, remainingQuota int, discount decimal.Decimal, unitPrice decimal.Decimal) []dto.ReferralAllocation`: walk in input order, skip duplicate/prior-claimed emails and where `discount >= unitPrice`, apply up to remaining quota.
- Allocation result includes attendee position, eligibility/applied state, and amount; never include other users' email/claim data.

- [ ] **Step 1: Write table-driven failing tests** for full/partial quota, attendee order, duplicate email in the order, previous claim, case/space normalization, price equal-to/under discount, zero eligible, and exhausted quota.
- [ ] **Step 2: Run** `go test ./modules/referral/service -run TestAllocateDiscount -v`; expected failures identify missing behavior.
- [ ] **Step 3: Implement minimal pure allocation/normalization functions and DTOs.**
- [ ] **Step 4: Re-run focused tests**, then `go test ./modules/referral/service/...`.

## Task 3: Implement referral repository and transactional claim operations

**Files:**
- Create: `modules/referral/repository/referral_repository.go`
- Create: `modules/referral/repository/referral_repository_test.go`

**Interfaces:**
- Repository supports create/get-by-normalized-code (with `FOR UPDATE` on transaction handle), list, toggle active state, count/read existing claims, create claims, redeem claims by order, and release held claims by order idempotently.
- Any operation participating in order creation/release/approval receives the caller's `*gorm.DB` transaction handle.

- [ ] **Step 1: Write repository tests** for unique code, `FOR UPDATE` lookup, normalized email claim uniqueness, claim status transitions, and idempotent release.
- [ ] **Step 2: Run** `go test ./modules/referral/repository -v`; expected failures until methods are implemented.
- [ ] **Step 3: Implement repository operations** with scoped transaction use and conditional status updates so quota cannot be decremented twice.
- [ ] **Step 4: Verify focused tests** using the existing PostgreSQL integration-test mechanism; if none exists, use current repository-test conventions and document DB-dependent checks.

## Task 4: Build admin management and customer preview API

**Files:**
- Create: `modules/referral/dto/referral_dto.go` (extend)
- Create: `modules/referral/service/referral_service.go` (extend)
- Create: `modules/referral/handler/referral_handler.go`
- Create: `modules/referral/routes.go`
- Modify: `providers/core.go`
- Modify: `cmd/main.go`
- Tests: `modules/referral/service/*_test.go`, `modules/referral/handler/*_test.go`

**Interfaces:**
- Admin service: `Create(ctx, request)`, `List(ctx)`, `SetActive(ctx, codeID, active)`.
- Customer service: `Preview(ctx, tierID, code, attendeeEmails)`, which reads but never reserves quota.
- HTTP operations: admin-authorized create/list/activate/deactivate; authenticated preview.
- Create validation rejects empty/invalid/duplicate code, non-positive discount or quota. Tier price eligibility remains per selected order, not code creation.

- [ ] **Step 1: Add failing service tests** for create validation, no edit/reuse, active toggle and preview outcomes (applicable, inactive, exhausted, inapplicable tier/price, prior email, partial allocation).
- [ ] **Step 2: Run focused service tests** and confirm failure.
- [ ] **Step 3: Implement service methods** delegating persistence to repository and calculator to Task 2.
- [ ] **Step 4: Add handler tests** for auth context, validation and response envelope; implement handler/routes.
- [ ] **Step 5: Wire DI and routes** following existing module pattern; verify `go test ./modules/referral/... ./providers/... ./cmd/...`.
- [ ] **Step 6: Add Bruno examples** for admin create/list/toggle and customer preview.

## Task 5: Integrate referral allocation into order create and lifecycle

**Files:**
- Modify: `modules/order/dto/order_dto.go`
- Modify: `modules/order/service/order_service.go`
- Modify: `modules/order/repository/order_repository.go`
- Modify: `modules/order/service/order_service_test.go`
- Tests: referral repository integration tests as needed

**Interfaces:**
- Add optional `ReferralCode *string` to `OrderCreateRequest`.
- Extend `OrderResponse` with referral code snapshot, discounted attendee count, total discount, and per-attendee allocation fields required by FE.
- Order service consumes referral repository/calculator through DI or an injected interface; avoid global DB access.

- [ ] **Step 1: Write failing order service tests** for no-code regression, ticket-tier price eligibility, ticket/referral quota partial application, server-side final amount, claim snapshots, and stale/inactive/exhausted code proceeding at normal price.
- [ ] **Step 2: Add failing lifecycle tests**: approve redeems held claims; cancel/reject/expiry releases claims and quota; repeated release is idempotent; invalid approval after released expiry does not redeem.
- [ ] **Step 3: Run focused tests**: `go test ./modules/order/service -run 'Referral|Create|Approve|Reject|Cancel|Expired' -v`.
- [ ] **Step 4: Integrate inside existing order transaction** using a stable lock order (tier then referral code), calculate allocation against current held/redeemed claims, write discount snapshots/claims, and update total from server prices.
- [ ] **Step 5: Integrate claim transitions into approve/reject/cancel/expiry** in the same transaction as order status and tier quota transitions.
- [ ] **Step 6: Run** `go test ./modules/order/... ./modules/referral/...` and `go test ./...`.
- [ ] **Step 7: Add a PostgreSQL concurrency test** with two simultaneous create calls competing for one referral slot and assert only one is discounted/held; test rollback leaves both tier and referral quotas unchanged.

## Task 6: Add frontend referral API and checkout preview

**Files:**
- Create: `/mnt/c/Dev/tedxua-frontend/src/services/referralService.js`
- Modify: `/mnt/c/Dev/tedxua-frontend/src/services/orderService.js`
- Modify: `/mnt/c/Dev/tedxua-frontend/src/features/tickets/components/IdentifyStepper.jsx`
- Tests: existing IdentifyStepper/order service test locations (inspect package scripts and current test conventions first)

**Interfaces:**
- `previewReferral({ ticket_tier_id, referral_code, attendees })` calls the authenticated backend preview endpoint and returns the API `data` payload.
- `createOrder` accepts optional `referral_code` and includes it in POST body only when provided.
- Preview response and order response share a presentation shape for attendee allocation, discount amount, total discount and final total.

- [ ] **Step 1: Inspect frontend `package.json`, tests, IdentifyStepper current order reuse, and API error envelope**; use existing tooling only.
- [ ] **Step 2: Write failing service/component tests** for preview, optional create payload, partial allocation display, preview-not-reserved messaging, and final submit result differing from preview.
- [ ] **Step 3: Run the relevant test(s)** and verify they fail for missing referral flow.
- [ ] **Step 4: Implement referral API calls** and include optional code in `createOrder`.
- [ ] **Step 5: Add code field and preview state** in IdentifyStepper; debounce or explicit button consistent with existing UX; surface unavailable/inactive/exhausted reason while permitting normal-price continuation.
- [ ] **Step 6: On submit, display authoritative order result** (including changed allocation/price) and only continue to payment after the user has seen the final amount.
- [ ] **Step 7: Run frontend tests and lint/build scripts** from `package.json`.

## Task 7: Add admin referral management UI

**Files:**
- Create: `/mnt/c/Dev/tedxua-frontend/src/features/admin/pages/ReferralCodesPage.jsx`
- Modify: frontend admin route/menu file(s) identified in Task 6 inspection
- Modify: `/mnt/c/Dev/tedxua-frontend/src/services/referralService.js`
- Tests: admin page tests if supported

**Interfaces:**
- Page can list code, discount per ticket, total/held/remaining quota, and active state.
- Admin can create code with manual text, discount, and quota; can toggle active state only. No edit or delete/reuse UI.

- [ ] **Step 1: Add failing page tests** for listing, create validation, toggle, and no edit controls.
- [ ] **Step 2: Run page tests** and verify failure.
- [ ] **Step 3: Implement page and API calls** following current admin patterns, design tokens and role protection.
- [ ] **Step 4: Run frontend test/lint/build** and verify admin route is protected.

## Task 8: End-to-end verification and graph update

**Files:**
- Modify: `API_Test/Referral/` only for any missing API examples
- Modify: `docs/database-schema.md` if its entity table list/schema documentation is maintained manually

- [ ] Run `go test ./...` from `/mnt/c/Dev/tedx-be`.
- [ ] Run frontend test/lint/build from `/mnt/c/Dev/tedxua-frontend`.
- [ ] Exercise migration + Bruno flows against local PostgreSQL: create code, preview without reservation, create order with partial allocation, approve and verify permanent claims, then cancel/expire another order and verify quota/email release.
- [ ] Run `graphify update .` from `/mnt/c/Dev/tedx-be` after code changes.
- [ ] Review `git status` and diff in both repositories; ensure only feature files/docs changed.

## Self-review

- **Spec coverage:** Admin create/toggle/no reuse → Tasks 4 and 7; all-ticket tier/email eligibility, ordering and partial discounts → Tasks 2, 4 and 5; advisory preview → Tasks 4 and 6; atomic ticket/referral hold and concurrency → Task 5; order snapshot → Task 5; paid/release lifecycle → Task 5; FE checkout/admin display → Tasks 6 and 7; tests and regressions → all tasks and Task 8.
- **Placeholders:** No unresolved business rules; “inspect existing test setup/route/menu convention” is a required discovery step because FE repo test scripts and admin wiring were not yet inspected in this planning session.
- **Type consistency:** Service method names, optional `ReferralCode`, and response data are declared in the tasks where introduced; exact endpoint paths, entity names, and released-claim retention policy remain implementation details as explicitly allowed by the spec.
- **Scope:** BE referral module, order lifecycle integration, customer checkout and admin management are separate deliverables but coupled by the spec's shared API and persisted claims. The implementation plan groups them into independently testable tasks while keeping this single cohesive feature.
