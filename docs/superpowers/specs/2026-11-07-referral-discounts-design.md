# Referral Discounts Design

**Status:** Approved in conversation; awaiting written-spec review  
**Scope:** Backend and frontend changes for fixed-amount referral discounts on all ticket orders.

## Goal

Allow admins to create customer-facing referral codes that give a fixed discount per eligible ticket, with a capped number of discounted tickets. A referral's ticket quota must be reserved consistently with the ticket quota when an order is created, and released if the order does not become paid.

## Agreed business rules

- Admin manually enters a unique referral code, discount amount per ticket, and maximum number of discounted tickets (the code's quota).
- Codes apply to all ticket types and tiers sold through the ticket-order checkout, including Main Event and Pre-Event. They do not apply to merchandise or bundles.
- Admin can only activate/deactivate a code after creation. To change its discount amount or quota, create a new code.
- Code strings are unique and are not reused, including after deactivation.
- There are no start/end dates; availability is controlled by the active flag and remaining quota.
- Discount is a fixed currency amount per ticket. It must be strictly less than the selected tier's unit price. If it is not, that ticket is ineligible for the code.
- A ticket-holder email can redeem a given code at most once, including across orders. Normalize email before comparison (trim whitespace and case-fold consistently).
- When an order contains repeated ticket-holder emails, at most one ticket for that email receives the discount.
- If the order request omits attendee details, use the same effective ticket-holder email(s) that the existing order flow uses to create attendee tickets; repeated fallback emails still receive at most one discount per code.
- Allocate discounts from the top of the submitted attendee list down. Skip an attendee if their email already redeemed the code or their tier is ineligible; continue until the code quota is exhausted or no eligible attendees remain.
- Partial application is allowed. For example, with one discounted ticket remaining and two eligible attendees, only the first eligible attendee receives the discount.
- A referral quota is held at order creation only for the number of tickets actually discounted. Preview does not hold quota.
- Referral holds follow the order lifecycle: held while awaiting payment approval; permanently redeemed when an admin approves payment and the order becomes paid; released on cancellation, rejection, or expiry before payment approval.
- Deactivating a code prevents new orders from using it, but does not change discounts/holds on existing orders.
- If a code is inactive, exhausted, or otherwise inapplicable (including tier price restriction), the customer may continue without a discount after being informed. The backend's order-creation response is authoritative.
- The frontend provides code entry on the ticket-holder details page, before order creation, and presents the discount and total before payment.

## Approaches considered

1. Store only a discount summary on the order. This is small but cannot safely represent per-email claims or support releasing claims and quota on failed orders.
2. Store referral-code configuration and per-ticket redemption records, plus a discount snapshot on each order. **Chosen.** Supports concurrency-safe quota holds, per-email uniqueness, deterministic partial allocation, release on failed orders, and stable historical prices.
3. Add a preview endpoint and reserve on submit. This improves price visibility, but preview is only advisory because another order may consume quota before submission. Reservation and final calculation still happen atomically during order creation.

## Proposed data model

Exact names/types should follow existing entity and migration conventions during implementation.

### Referral code

Store at least: ID, normalized unique code, fixed discount amount per ticket, total ticket quota, number of currently held tickets, active flag, and timestamps. Preserve code uniqueness for the lifetime of the record; do not recycle deactivated code strings. Do not expose edit operations for code, discount amount, or quota.

### Referral redemption / claim

Store one row per discounted attendee ticket, including referral code ID, order ID, normalized ticket-holder email, attendee position within the order, discount amount snapshot, lifecycle status (`held` or `redeemed`), and timestamps. Enforce no duplicate email claim for a code across held and redeemed claims. A held claim is released/deleted (implementation may retain audit history with a released state) when its order is cancelled, rejected, or expires. Redeemed claims remain permanent.

### Order snapshot

Persist enough information to explain the final amount independently of future code changes: applied referral code (or code text snapshot), discounted attendee count, and total discount. Keep existing unit-price and total-amount semantics consistent with current order behavior. The per-attendee claim rows retain which attendees received the discount.

## Components and data flow

### Admin management

- Add authenticated/admin-authorized operations to create a code and activate/deactivate it.
- Creation validates code uniqueness, positive discount, and positive quota; eligibility is checked against the selected tier price at redemption time.
- No update or reuse operation is provided.

### Customer preview

- On the ticket-holder details page, the customer can enter a code and request an advisory price/eligibility preview.
- Preview reads current code status/quota and attendee data but creates no hold or redemption claim.
- The response should explain the proposed discount allocation, including partial application and ineligible/previously redeemed emails as appropriate. Avoid exposing sensitive information beyond what the customer submitted.
- If preview cannot apply the code, show the reason and let the customer continue at normal price.

### Order creation and reservation

- Submit the optional referral code with the order and attendee list.
- Within the same database transaction as existing tier quota reservation and order/attendee creation:
  1. Lock/read the chosen tier and referral code in a consistent lock order.
  2. Revalidate code active status and remaining quota; recompute eligible attendees from the submitted list in order.
  3. Skip attendee emails already claimed by the code (including duplicate emails earlier in this order) and tickets whose discount is not strictly below unit price.
  4. Apply discount to eligible attendees from the top of the list until quota is exhausted.
  5. Create the order using the server-calculated final total; persist order snapshot and held per-ticket claims; increment/record the held referral quota.
- Any failure in this transaction must roll back ticket and referral holds together. Concurrent attempts for the last referral slot must not overbook it. Existing PostgreSQL locking/transaction conventions should be followed.
- If the submitted code is unavailable/inapplicable at create time, create the order at normal price and return a clear notice. Do not let client-supplied prices or discount totals override server calculations.
- The response includes final applied count, discount amount/total, and per-attendee allocation sufficient for the UI to show the final price before continuing to payment.

### Order lifecycle

- On payment approval / transition to `paid`, transition that order's held claims to redeemed; the quota remains consumed.
- On cancellation, rejection, or expiry while awaiting approval, release that order's held claims and referral quota, so the same emails may use the code again.
- Release must be idempotent and run with the corresponding order status transition. Repeated expiry/rejection processing must not decrement quota twice.
- Code deactivation does not affect existing holds or redeemed claims.

### Frontend

- Add code entry and preview to the ticket-holder details step before order creation.
- Show subtotal, each eligible/discounted attendee (allocation follows list order), total discount, and final total. Clearly disclose partial discount or ineligibility.
- On order submission, treat backend response as the final source of truth. If the preview and final allocation differ because quota changed, show the returned final amount/discount before proceeding to payment.
- Keep the current no-code checkout flow working unchanged.

## Failure handling and consistency

- Use database constraints as the final guard for unique normalized codes and per-code/email claims, together with transactional row locking for finite quota.
- Consider races in which a preview is stale, another order reserves the last slot, or the same email submits concurrent orders. Exactly one transaction may hold a given last slot or email claim.
- A failed/cancelled/expired order releases both ticket quota and referral quota/claims. A paid order retains both as consumed/sold.
- A transaction failure must not leave either quota held independently of the order.
- If an order's payment is approved after its hold has already been released due to expiry/cancellation, existing order status rules must reject that invalid transition; implementation must not silently consume released referral quota.

## Testing requirements

- Unit/service tests: full and partial allocation; ordered allocation; skip repeated email; skip previously redeemed email; case/whitespace-normalized email; skip tier whose price is less than or equal to discount; code inactive/exhausted; zero eligible attendees; no-code behavior.
- Lifecycle tests: held claims become redeemed on payment approval; cancel/reject/expiry release claims and quota; release is idempotent; code deactivation leaves existing order snapshots/holds untouched.
- Database/integration tests: transaction rollback releases both quota holds; concurrent attempts for one remaining referral slot do not overbook; concurrent same-code/same-email attempts produce at most one active claim.
- API/UI tests: preview does not reserve; create response is authoritative when preview becomes stale; user can proceed at normal price when code is not applicable; final price and partial allocation are displayed before payment.
- Regression tests: existing ticket quota behavior, order creation without a referral code, payment approval, cancellation, and expiration continue to work.

## Out of scope

- Referrer accounts, commissions, analytics/reporting, code attribution, campaign start/end dates, percentage discounts, stacking multiple codes, changing a code after creation, or reusing code text.
- Discounting merchandise/bundles or changing payment provider behavior.

## Open implementation details

These are intentionally left to codebase-convention review during the implementation plan, not unresolved business decisions: exact endpoint paths/HTTP methods, entity/column naming, whether released claims are deleted or retained with a `released` status, and how order API responses represent per-attendee allocation.
