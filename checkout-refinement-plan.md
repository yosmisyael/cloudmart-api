# Checkout & Shopping Flow Refinement Plan

## Analysis: Confirmed Gaps Found in the API

After reading every endpoint and every entity/request definition in the API doc, here is a complete and objective audit of every gap in the buyer shopping flow.

---

### GAP 1 — Checkout accepts only a raw address string

**Current:** `CheckoutRequest` has a single `address string` field (min 10 chars). No link to the user's saved `Address` records.

**Problem:** The user has a full address management system at `POST /api/profile/addresses` storing recipient name, phone, city, state, postal code, and country. None of that is usable at checkout. The buyer must retype the full address as a freeform string every time, and the `shipping_address` column on `Order` stores it as an opaque text blob with no structure.

**Fix:** Checkout should accept an `address_id` referencing the user's saved address, OR a freeform string as fallback. The service resolves the address record and formats it into `shipping_address`.

---

### GAP 2 — Checkout accepts no logistic service selection

**Current:** `CheckoutRequest` has no logistic field. `Order.logistic_service` is a `varchar(100)` that is never populated at checkout. `Order.logistic_voucher_id` similarly always null.

**Problem:** The logistic management system exists (sellers create `Logistic` + `LogisticService` records with `base_price`), but there is **no public endpoint for buyers to browse available logistics**, and checkout takes no logistic selection input. Shipping cost is therefore never calculated, making `grand_total` wrong — it only contains item subtotals.

**Fix:** (a) Add a public `GET /api/logistics` endpoint for buyers to browse logistics and their services. (b) Add `logistic_service_id` to `CheckoutRequest`. The service loads the `LogisticService`, extracts its `base_price`, adds it to `grand_total`, and stores `logistic_service` as a formatted string.

---

### GAP 3 — Checkout accepts no voucher

**Current:** `CheckoutRequest` has no voucher field. Sellers create `Voucher` records with types `percentage`, `price`, and `free_shipping`, and there are three important related gaps:

- **No endpoint for buyers to see available vouchers.** There is no `GET /api/vouchers` or any public voucher listing. Vouchers are created by sellers but invisible to buyers.
- **No endpoint for buyers to claim/apply a voucher.** The `Voucher` entity has a `many2many:user_vouchers` join with `User`, but there is no endpoint to add a voucher to a user's account.
- **Discount is never applied to `grand_total`.** Checkout ignores vouchers entirely.

**Fix:** Add three things: (a) `GET /api/vouchers` — browse applicable vouchers by store. (b) `POST /api/vouchers/:code/claim` — buyer claims a voucher (adds to `user_vouchers`). (c) Add optional `voucher_id` to `CheckoutRequest`; service validates ownership, checks expiry, calculates discount, applies it to grand total, and stores `logistic_voucher_id` on the order.

---

### GAP 4 — No order cost preview before checkout

**Current:** The only way to see a total is after checkout is called and the order is created. There is no way for a buyer to preview: "if I pick this logistic and apply this voucher, my total will be X."

**Problem:** On any real shopping app, the checkout screen shows a live breakdown (subtotal + shipping fee − discount = total) before the buyer confirms. Forcing order creation as the only way to see a total is a serious UX flaw and creates orphaned pending orders every time a user abandons.

**Fix:** Add `POST /api/orders/estimate` — accepts the same inputs as checkout (logistic service ID + optional voucher ID) but creates nothing, only returns a cost breakdown object.

---

### GAP 5 — No public logistic browsing endpoint for buyers

**Current:** Logistics are managed by sellers at `GET/POST/PUT/DELETE /api/seller/logistics`. There is zero public-facing logistic endpoint. A buyer has no API call to see what shipping options exist and what they cost.

**Problem:** You cannot build a checkout screen without knowing what shipping options to show. This is a complete dead end for any frontend.

**Fix:** `GET /api/logistics` — public, no auth, returns all logistics with their services and base prices. Already confirmed the entity and data exist; just needs a public route.

---

### GAP 6 — Address management is incomplete

**Current:** `GET /api/profile/addresses` lists addresses. `POST /api/profile/addresses` creates one. There is no `PUT /api/profile/addresses/:id` to update, no `DELETE /api/profile/addresses/:id` to remove, and no way to set a default address.

**Problem:** Users are stuck with addresses they can't update or delete. On mobile apps especially this causes address list pollution over time.

**Fix:** Add `PUT /api/profile/addresses/:id`, `DELETE /api/profile/addresses/:id`, and add an `is_default boolean` field to `Address` with `POST /api/profile/addresses/:id/default` to set the default. Checkout should then also accept just `use_default_address: true` as a shortcut.

---

### GAP 7 — Profile has no update endpoint

**Current:** `GET /api/profile` returns user profile. There is no `PUT /api/profile` to update name or phone. Password change is also absent.

**Problem:** Users cannot update their own name, phone, or password. This is a basic account management gap that blocks any profile settings screen.

**Fix:** Add `PUT /api/profile` (name, phone) and `PUT /api/profile/password` (old_password, new_password).

---

### GAP 8 — Order cancellation by buyer is missing

**Current:** Sellers can update order status to `"cancelled"` via `PUT /api/seller/orders/:id/status`. Buyers have no equivalent. `GET /api/orders` and `GET /api/orders/:id` are read-only.

**Problem:** A buyer cannot cancel their own pending order. This is a core e-commerce requirement, especially because `payment_status = "pending"` orders pile up when a buyer abandons the Midtrans flow without paying.

**Fix:** Add `POST /api/orders/:id/cancel` — buyer-facing, only allowed when `payment_status == "pending"`. Restocks all variants in the same DB transaction.

---

### GAP 9 — Order tracking status is undefined

**Current:** `Order.payment_status` conflates payment state (`pending`, `settlement`, `cancel`, `expire`) with shipping state (`processing`, `shipped`, `delivered`, `cancelled`). These are mixed into one field. There is no dedicated `shipping_status` field.

**Problem:** A buyer looking at their order cannot distinguish "I haven't paid yet" from "I've paid but it hasn't shipped." The seller's status transitions (`processing → shipped → delivered`) write into the same column as Midtrans payment callbacks, meaning a webhook could overwrite a seller's `"shipped"` status with `"settlement"` or vice versa.

**Fix:** Add a separate `shipping_status` field to `Order` (`pending → processing → shipped → delivered`). `payment_status` stays for Midtrans callbacks only. Seller order status update writes to `shipping_status`. Order detail response surfaces both fields clearly.

---

### GAP 10 — Cart has no stock validation at add-to-cart time

**Current:** `POST /api/cart` adds items without checking stock. Stock is only checked at checkout. The cart can silently accumulate quantities that exceed available stock.

**Problem:** A buyer sees 3 items in their cart at a price, proceeds to checkout, and gets a stock error at the last step. This is a bad UX failure that should be caught earlier.

**Fix:** Add stock validation in `AddToCart` — if `requested quantity > variant.stock`, return 409 with a clear message. Also validate that the variant exists at add-to-cart time (currently returns a generic error).

---

### GAP 11 — No voucher code lookup; vouchers have no code field

**Current:** The `Voucher` entity has no code/slug field — it only has `id`, `name`, `type`, `amount`, `max`, `expired_at`. There is no way for a buyer to enter a discount code. The only identifier is a numeric ID.

**Problem:** Every e-commerce flow exposes vouchers as codes the buyer types in. A numeric ID is internal and cannot be shared with customers via marketing campaigns.

**Fix:** Add a `Code string` field to `Voucher` (unique, uppercase slug e.g. `"NIKE10"`). Voucher claim and checkout accept a `voucher_code` string instead of an ID. The seller can set the code when creating a voucher.

---

### GAP 12 — `GET /api/orders` response is too thin

**Current:** `GET /api/orders` returns a list of `Order` entities preloaded with `OrderItems`. But `OrderItem` contains only `variant_id`, `variant_details` (a formatted string), `price`, and `quantity`. There is no product name, no image, no store name in the order history response.

**Problem:** A buyer's order history screen needs to show product thumbnails and names. The current response forces the frontend to make N additional product API calls to display the list.

**Fix:** Extend `OrderItem` preload in `FindByUserID` to include `Variant → Product → Store` so the order history response is self-contained.

---

## Summary Table

| # | Gap | Severity | Fix Type |
|---|-----|----------|----------|
| 1 | Checkout ignores saved addresses | Critical | Extend checkout request + service |
| 2 | No logistic selection at checkout, shipping cost never added | Critical | New public endpoint + extend checkout |
| 3 | No voucher browsing, claiming, or application at checkout | Critical | 3 new endpoints + extend checkout |
| 4 | No cost preview before committing to order | High | New estimate endpoint |
| 5 | No public logistic listing for buyers | Critical | New public endpoint |
| 6 | Address CRUD incomplete (no update/delete/default) | High | 3 new endpoints + entity field |
| 7 | No profile update or password change | Medium | 2 new endpoints |
| 8 | No buyer order cancellation | High | New endpoint + restock logic |
| 9 | `payment_status` conflates payment + shipping state | High | New entity field + service split |
| 10 | No stock check at add-to-cart | Medium | Extend cart service validation |
| 11 | Voucher has no code field; no code-based lookup | High | Entity field + update create/claim |
| 12 | Order history missing product images and names | Medium | Extend preload in repository |

---

## Refinement Plan — What to Build

### Files that need to be created (new)

| File | Purpose |
|------|---------|
| `internal/handler/logistic_handler.go` | Public buyer-facing logistic browsing |

### Files that need to be modified (extended)

| File | Changes |
|------|---------|
| `internal/entity/transaction.go` | Add `ShippingStatus` to `Order`; extend `OrderItem` preload |
| `internal/entity/promo.go` | Add `Code` field to `Voucher` |
| `internal/entity/user.go` | Add `IsDefault` to `Address` |
| `internal/repository/order_repository.go` | Extend `FindByUserID` preload; add `CancelOrder`, `UpdateShippingStatus` |
| `internal/repository/logistic_repository.go` | Add public `FindAll` with eager-loaded services |
| `internal/repository/voucher_repository.go` | Add `FindByCode`, `ClaimVoucher`, `FindClaimedByUser` |
| `internal/repository/user_repository.go` | Add address update, delete, set-default methods |
| `internal/service/order_service.go` | Rewrite `Checkout` with address/logistic/voucher logic; add `Estimate`, `CancelOrder` |
| `internal/service/catalog_service.go` | Add public `GetLogistics` method |
| `internal/service/user_service.go` | Add `UpdateProfile`, `ChangePassword`, address CRUD |
| `internal/handler/order_handler.go` | Add `EstimateOrder`, `CancelOrder` routes |
| `internal/handler/profile_handler.go` | Add address update/delete/default routes + profile update/password |
| `internal/handler/cart_handler.go` | Add stock validation at add-to-cart |
| `internal/handler/seller_promo_handler.go` | Update `VoucherRequest` to include `code` field |
| `cmd/api/main.go` | Register new `LogisticHandler`; pass new repo deps |

---

## Detailed Specs Per Change

### §1 — Entity: `internal/entity/transaction.go`

Add to `Order` struct:
```go
ShippingStatus string `gorm:"type:varchar(20);default:'pending'" json:"shipping_status"`
// values: "pending" | "processing" | "shipped" | "delivered"
```

The existing `PaymentStatus` values remain: `"pending"`, `"settlement"`, `"cancel"`, `"expire"`.

---

### §2 — Entity: `internal/entity/promo.go`

Add to `Voucher` struct:
```go
Code string `gorm:"type:varchar(50);uniqueIndex;not null" json:"code"`
```

---

### §3 — Entity: `internal/entity/user.go`

Add to `Address` struct:
```go
IsDefault bool `gorm:"default:false" json:"is_default"`
```

---

### §4 — New public endpoint: `GET /api/logistics`

**Handler:** `internal/handler/logistic_handler.go` (new file, separate from seller logistic handler)

```go
func NewLogisticHandler(router fiber.Router, svc service.CatalogService)
// Route: router.Get("/api/logistics", h.GetLogistics)
```

**Service:** Add to `CatalogService` interface and `catalogService` struct:
```go
GetLogistics() ([]entity.Logistic, error)
```
Implementation calls `logisticRepo.FindAll()`.

**Constructor change:** `NewCatalogService` needs `logisticRepo repository.LogisticRepository` added as a dependency.

**Response shape:**
```json
{
  "code": 200,
  "status": "OK",
  "data": [
    {
      "id": 1,
      "name": "JNE",
      "services": [
        { "id": 1, "logistic_id": 1, "name": "REG", "base_price": 15000 },
        { "id": 2, "logistic_id": 1, "name": "YES", "base_price": 35000 }
      ]
    }
  ]
}
```

Swagger:
```go
// @Summary     Get available logistics
// @Description Retrieve all logistics providers and their shipping service options with pricing
// @Tags        Catalog
// @Produce     json
// @Success     200 {object} response.WebResponse{data=[]entity.Logistic} "Logistic list"
// @Failure     500 {object} response.WebResponse "Internal server error"
// @Router      /api/logistics [get]
```

---

### §5 — New public endpoints: Voucher browsing and claiming

**Routes (add to a new `VoucherHandler` or extend existing public catalog handler):**

```
GET  /api/vouchers?store_id=        → GetAvailableVouchers (Protected)
POST /api/vouchers/:code/claim      → ClaimVoucher (Protected)
GET  /api/vouchers/mine             → GetMyVouchers (Protected)
```

`GetAvailableVouchers`: returns non-expired vouchers for a given store (query param `store_id`). Buyer uses this when browsing a store before checkout.

`ClaimVoucher`: finds voucher by `code`, checks expiry, checks not already claimed, adds to `user_vouchers` join table.

`GetMyVouchers`: returns all vouchers the authenticated user has claimed and not yet used (used = applied to a settled order — track via boolean `IsUsed` on the join table, or just return all claimed and let client filter by expiry).

**New repository methods on `VoucherRepository`:**
```
FindByCode(code string) (*entity.Voucher, error)
FindByStoreID(storeID uint) ([]entity.Voucher, error)       // already exists — reuse
ClaimVoucher(userID, voucherID uint) error                   // INSERT into user_vouchers
FindClaimedByUserID(userID uint) ([]entity.Voucher, error)
IsClaimedByUser(userID, voucherID uint) (bool, error)
```

**New service:** `internal/service/voucher_service.go`

```
VoucherService
  GetAvailableVouchers(storeID uint) ([]entity.Voucher, error)
  ClaimVoucher(userID uint, code string) error
  GetMyVouchers(userID uint) ([]entity.Voucher, error)
```

Swagger blocks:
```go
// @Summary     Browse available vouchers
// @Description Retrieve all active, non-expired vouchers for a specific store
// @Tags        Voucher
// @Produce     json
// @Security    BearerAuth
// @Param       store_id query int true "Store ID"
// @Success     200 {object} response.WebResponse{data=[]entity.Voucher} "Voucher list"
// @Failure     400 {object} response.WebResponse "Missing store_id"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Router      /api/vouchers [get]

// @Summary     Claim a voucher
// @Description Add a voucher to the authenticated user's account by voucher code
// @Tags        Voucher
// @Produce     json
// @Security    BearerAuth
// @Param       code path string true "Voucher code"
// @Success     200 {object} response.WebResponse "Voucher claimed"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Failure     404 {object} response.WebResponse "Voucher not found or expired"
// @Failure     409 {object} response.WebResponse "Already claimed"
// @Router      /api/vouchers/{code}/claim [post]

// @Summary     Get my vouchers
// @Description Retrieve all vouchers claimed by the authenticated user
// @Tags        Voucher
// @Produce     json
// @Security    BearerAuth
// @Success     200 {object} response.WebResponse{data=[]entity.Voucher} "My voucher list"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Router      /api/vouchers/mine [get]
```

---

### §6 — Rewrite Checkout: `POST /api/orders/checkout`

**New `CheckoutRequest`:**
```go
type CheckoutRequest struct {
    AddressID         *uint  `json:"address_id"`          // optional — use saved address
    Address           string `json:"address"`              // optional — freeform fallback
    LogisticServiceID uint   `json:"logistic_service_id" validate:"required"`
    VoucherCode       string `json:"voucher_code"`         // optional
}
```

Validation: either `address_id` OR `address` must be present (validate in service, not struct tag).

**Updated `order_service.Checkout` logic:**
1. Resolve shipping address: if `address_id` provided, load `Address` record (assert `address.UserID == userID`), format as structured string. Else use freeform `address` string (min 10 chars).
2. Load `LogisticService` by ID, get `base_price` as `shippingFee`.
3. Calculate items subtotal (existing logic).
4. If `voucher_code` provided:
   a. Find voucher by code via `voucherRepo.FindByCode`
   b. Check `voucher.ExpiredAt.After(time.Now())`
   c. Check `voucherRepo.IsClaimedByUser(userID, voucher.ID)`
   d. Apply discount:
      - `"percentage"`: `discount = subtotal * (voucher.Amount / 100)`, capped at `voucher.Max` if `Max > 0`
      - `"price"`: `discount = voucher.Amount`
      - `"free_shipping"`: `discount = shippingFee` (shipping becomes 0)
5. `grand_total = subtotal + shippingFee - discount`
6. Store `logistic_service = logisticService.Name`, `logistic_voucher_id = &voucher.ID` on order.
7. Create order (existing transaction logic).

**New repo methods needed:**
```
// on LogisticRepository:
FindServiceByID(id uint) (*entity.LogisticService, error)   // may already exist — verify
```

**Updated Swagger for checkout:**
```go
// @Summary     Checkout order
// @Description Create an order from cart with shipping selection, optional voucher, and full cost calculation
// @Tags        Order
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       request body CheckoutRequest true "Checkout payload"
// @Success     201 {object} response.WebResponse{data=entity.Order} "Order created"
// @Failure     400 {object} response.WebResponse "Validation error or invalid logistic"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Failure     404 {object} response.WebResponse "Address or logistic not found"
// @Failure     409 {object} response.WebResponse "Stock insufficient or voucher invalid"
// @Router      /api/orders/checkout [post]
```

---

### §7 — New endpoint: `POST /api/orders/estimate`

**Handler method on `OrderHandler`:** `EstimateOrder`

**Route:** `POST /api/orders/estimate` (Protected)

**Request:** same struct as `CheckoutRequest` — reuse it.

**Service method on `OrderService`:**
```
EstimateOrder(userID uint, logisticServiceID uint, voucherCode string) (*OrderEstimate, error)
```

**Response struct (defined in service):**
```go
type OrderEstimate struct {
    Subtotal     float64 `json:"subtotal"`
    ShippingFee  float64 `json:"shipping_fee"`
    Discount     float64 `json:"discount"`
    VoucherName  string  `json:"voucher_name,omitempty"`
    GrandTotal   float64 `json:"grand_total"`
}
```

Logic: same as Checkout steps 1–5 but creates nothing — only returns the `OrderEstimate`.

Swagger:
```go
// @Summary     Estimate order cost
// @Description Preview the full cost breakdown for cart contents with a selected logistic and optional voucher, without creating an order
// @Tags        Order
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       request body CheckoutRequest true "Estimate payload"
// @Success     200 {object} response.WebResponse{data=service.OrderEstimate} "Cost estimate"
// @Failure     400 {object} response.WebResponse "Validation error"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Failure     404 {object} response.WebResponse "Logistic or voucher not found"
// @Router      /api/orders/estimate [post]
```

---

### §8 — New endpoint: `POST /api/orders/:id/cancel`

**Handler method on `OrderHandler`:** `CancelOrder`

**Route:** `POST /api/orders/:id/cancel` (Protected)

**Service method on `OrderService`:**
```
CancelOrder(userID, orderID uint) error
```

Logic:
1. Load order via `orderRepo.FindByID(orderID, userID)` — asserts buyer ownership
2. Assert `order.PaymentStatus == "pending"` — only pending orders can be cancelled
3. Call `orderRepo.CancelOrder(orderID)` — sets `payment_status = "cancel"`, restocks all variants in a single DB transaction

**New repo method on `OrderRepository`:**
```
CancelOrder(orderID uint) error
```
Implementation (in a transaction):
```go
// 1. Load order items
// 2. For each item: UPDATE product_variants SET stock = stock + item.Quantity WHERE id = item.VariantID
// 3. UPDATE orders SET payment_status = 'cancel' WHERE id = orderID
```

Swagger:
```go
// @Summary     Cancel order
// @Description Cancel a pending order and restock all variants atomically
// @Tags        Order
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "Order ID"
// @Success     200 {object} response.WebResponse "Order cancelled"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Failure     403 {object} response.WebResponse "Order cannot be cancelled — not pending"
// @Failure     404 {object} response.WebResponse "Order not found"
// @Router      /api/orders/{id}/cancel [post]
```

---

### §9 — Seller order status → writes to `shipping_status`

**Change in `internal/handler/seller_order_handler.go`:** `UpdateOrderStatus` handler now calls a service method that writes to `shipping_status`, not `payment_status`.

**Change in `internal/repository/seller_order_repository.go`:**
Rename/add:
```
UpdateShippingStatus(orderID uint, status string) error
// UPDATE orders SET shipping_status = ? WHERE id = ?
```

Remove (or keep for backward compat but stop calling it from seller handler):
- The seller handler must NOT write to `payment_status`. Only the Midtrans webhook writes to `payment_status`.

**Valid `shipping_status` values:** `"pending"`, `"processing"`, `"shipped"`, `"delivered"`

**`UpdateOrderStatusRequest` enum changes** (in seller_order_handler.go):
```go
// validate:"required,oneof=processing shipped delivered"
// (remove "cancelled" — buyers cancel via the buyer endpoint; sellers ship)
```

---

### §10 — Address CRUD completion: `internal/handler/profile_handler.go`

New routes (all Protected):
```
PUT    /api/profile/addresses/:id         → UpdateAddress
DELETE /api/profile/addresses/:id         → DeleteAddress
POST   /api/profile/addresses/:id/default → SetDefaultAddress
```

**New `UpdateAddressRequest`:** same fields as `CreateAddressRequest` (all fields updatable).

**New repo methods on `UserRepository`:**
```
UpdateAddress(address *entity.Address) error
DeleteAddress(addressID, userID uint) error
SetDefaultAddress(addressID, userID uint) error
// SetDefaultAddress: in a transaction, set all user's addresses is_default=false, then set target is_default=true
```

Swagger blocks:
```go
// @Summary     Update address
// @Description Update a saved shipping address for the authenticated user
// @Tags        Profile
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id      path int                  true "Address ID"
// @Param       request body UpdateAddressRequest true "Address payload"
// @Success     200 {object} response.WebResponse{data=entity.Address} "Address updated"
// @Failure     400 {object} response.WebResponse "Validation error"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Failure     403 {object} response.WebResponse "Not your address"
// @Failure     404 {object} response.WebResponse "Address not found"
// @Router      /api/profile/addresses/{id} [put]

// @Summary     Delete address
// @Description Delete a saved shipping address for the authenticated user
// @Tags        Profile
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "Address ID"
// @Success     200 {object} response.WebResponse "Address deleted"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Failure     403 {object} response.WebResponse "Not your address"
// @Failure     404 {object} response.WebResponse "Address not found"
// @Router      /api/profile/addresses/{id} [delete]

// @Summary     Set default address
// @Description Mark an address as the user's default shipping address
// @Tags        Profile
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "Address ID"
// @Success     200 {object} response.WebResponse "Default address set"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Failure     403 {object} response.WebResponse "Not your address"
// @Failure     404 {object} response.WebResponse "Address not found"
// @Router      /api/profile/addresses/{id}/default [post]
```

---

### §11 — Profile update and password change: `internal/handler/profile_handler.go`

New routes (all Protected):
```
PUT /api/profile          → UpdateProfile
PUT /api/profile/password → ChangePassword
```

**Request structs:**
```go
type UpdateProfileRequest struct {
    Name  string `json:"name" validate:"required,min=2"`
    Phone string `json:"phone" validate:"required"`
}
type ChangePasswordRequest struct {
    OldPassword string `json:"old_password" validate:"required"`
    NewPassword string `json:"new_password" validate:"required,min=6"`
}
```

**New service methods on `UserService`:**
```
UpdateProfile(userID uint, name, phone string) error
ChangePassword(userID uint, oldPassword, newPassword string) error
```

`ChangePassword` logic: load user via `authRepo.FindByID`, verify `bcrypt.CompareHashAndPassword`, hash new password, call `userRepo.UpdatePassword`.

**New repo methods on `UserRepository`:**
```
UpdateProfile(userID uint, name, phone string) error
UpdatePassword(userID uint, hashedPassword string) error
```

Swagger blocks:
```go
// @Summary     Update profile
// @Description Update the authenticated user's name and phone number
// @Tags        Profile
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       request body UpdateProfileRequest true "Profile payload"
// @Success     200 {object} response.WebResponse{data=entity.User} "Profile updated"
// @Failure     400 {object} response.WebResponse "Validation error"
// @Failure     401 {object} response.WebResponse "Unauthorized"
// @Router      /api/profile [put]

// @Summary     Change password
// @Description Change the authenticated user's password
// @Tags        Profile
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       request body ChangePasswordRequest true "Password payload"
// @Success     200 {object} response.WebResponse "Password changed"
// @Failure     400 {object} response.WebResponse "Validation error"
// @Failure     401 {object} response.WebResponse "Wrong current password"
// @Failure     404 {object} response.WebResponse "User not found"
// @Router      /api/profile/password [put]
```

---

### §12 — Cart stock validation: `internal/handler/cart_handler.go` + `internal/service/cart_service.go`

**In `CartService.AddToCart`:**
After loading the variant, before inserting/updating the cart:
```go
if variant.Stock < req.Quantity {
    return errors.New(fmt.Sprintf("stok %s tidak mencukupi, tersedia %d", variant.SKU, variant.Stock))
}
```
Return 409 from handler when this error is received.

---

### §13 — Order history enriched preload

**In `internal/repository/order_repository.go`, `FindByUserID`:**

Change:
```go
Preload("OrderItems")
```
To:
```go
Preload("OrderItems").
Preload("OrderItems.Variant").
Preload("OrderItems.Variant.Product")
```

This makes order history self-contained. No new fields needed — `Variant` and `Product` already exist as associations on `OrderItem` via `ProductVariant` (which has `Product Product`).

Note: `OrderItem` entity already has `VariantID` — the `Variant` association just needs to be declared on the struct. Add to `OrderItem`:
```go
Variant ProductVariant `gorm:"foreignKey:VariantID" json:"variant,omitempty"`
```

---

### §14 — Seller voucher create: add `code` field

**In `internal/handler/seller_promo_handler.go`:**

Update `VoucherRequest` to add:
```go
Code string `json:"code" validate:"required,min=3,max=50"`
```

**In `internal/service/seller_promo_service.go`, `CreateVoucher`:**
Pass `req.Code` when constructing `entity.Voucher`.

**In `internal/service/seller_promo_service.go`, `UpdateVoucher`:**
Allow updating `Code` field.

---

### §15 — AutoMigrate additions: `cmd/api/main.go`

The new fields (`ShippingStatus` on `Order`, `Code` on `Voucher`, `IsDefault` on `Address`, `Variant` association on `OrderItem`) are all on existing entities — GORM `AutoMigrate` will add the new columns automatically. No new entities.

No new table needed for the `user_vouchers` join — it is already declared via `many2many:user_vouchers` on `Voucher.Users`. GORM auto-creates join tables.

Add to wiring in `main.go`:
```go
// New service
voucherService := service.NewVoucherService(voucherRepo, storeRepo)

// New handler
handler.NewLogisticHandler(app, catalogService)
handler.NewVoucherHandler(app, voucherService, cfg)

// Update CatalogService constructor (now needs logisticRepo)
catalogService = service.NewCatalogService(categoryRepo, productRepo, logisticRepo)
```

---

## Implementation Order

Implement in exactly this order to avoid compile errors:

**Entity changes (no deps)**
1. Add `ShippingStatus` to `entity.Order` in `transaction.go`
2. Add `Variant` association to `entity.OrderItem` in `transaction.go`
3. Add `Code` to `entity.Voucher` in `promo.go`
4. Add `IsDefault` to `entity.Address` in `user.go`

**Repository extensions**
5. Add `UpdateShippingStatus` to `seller_order_repository.go`
6. Add `CancelOrder` to `order_repository.go`
7. Extend `order_repository.go` `FindByUserID` preload
8. Add `FindByCode`, `ClaimVoucher`, `FindClaimedByUserID`, `IsClaimedByUser` to `voucher_repository.go`
9. Add `UpdateAddress`, `DeleteAddress`, `SetDefaultAddress` to `user_repository.go`
10. Add `UpdateProfile`, `UpdatePassword` to `user_repository.go`
11. Add `GetLogistics` (FindAll with services) to `logistic_repository.go` — verify if already exists

**New service**
12. Create `internal/service/voucher_service.go`

**Service extensions**
13. Add `GetLogistics` to `catalog_service.go` + update constructor to accept `logisticRepo`
14. Add `EstimateOrder`, `CancelOrder` to `order_service.go`; rewrite `Checkout` with full logic
15. Add `UpdateProfile`, `ChangePassword` to `user_service.go`
16. Extend `seller_order_service.go` to call `UpdateShippingStatus` instead of `UpdatePaymentStatus`
17. Add stock check to `cart_service.go`
18. Update `seller_promo_service.go` to include `Code` in voucher create/update

**Handler changes**
19. Create `internal/handler/logistic_handler.go`
20. Create `internal/handler/voucher_handler.go`
21. Update `internal/handler/order_handler.go` — new `CheckoutRequest`, add `EstimateOrder`, `CancelOrder` routes
22. Update `internal/handler/profile_handler.go` — address update/delete/default + profile update + password change
23. Update `internal/handler/cart_handler.go` — map stock error to 409
24. Update `internal/handler/seller_order_handler.go` — fix enum + call shipping status method
25. Update `internal/handler/seller_promo_handler.go` — add `code` to `VoucherRequest`

**Wire everything**
26. Update `cmd/api/main.go` — new service, updated constructors, new handler registrations

**Docs**
27. Run `swag init -g cmd/api/main.go`
28. Run `go build ./...` and fix errors

---

## New Endpoint Summary

| Method | Path | Auth | Tag |
|--------|------|------|-----|
| GET | `/api/logistics` | Public | Catalog |
| GET | `/api/vouchers` | Protected | Voucher |
| POST | `/api/vouchers/:code/claim` | Protected | Voucher |
| GET | `/api/vouchers/mine` | Protected | Voucher |
| POST | `/api/orders/estimate` | Protected | Order |
| POST | `/api/orders/:id/cancel` | Protected | Order |
| PUT | `/api/profile` | Protected | Profile |
| PUT | `/api/profile/password` | Protected | Profile |
| PUT | `/api/profile/addresses/:id` | Protected | Profile |
| DELETE | `/api/profile/addresses/:id` | Protected | Profile |
| POST | `/api/profile/addresses/:id/default` | Protected | Profile |

---

## Modified Endpoint Summary

| Method | Path | Change |
|--------|------|--------|
| POST | `/api/orders/checkout` | New request body; full cost calculation |
| POST | `/api/cart` | Stock validation before insert |
| PUT | `/api/seller/orders/:id/status` | Writes to `shipping_status` not `payment_status` |
| POST | `/api/seller/vouchers` | `code` field now required |
| PUT | `/api/seller/vouchers/:id` | `code` field updatable |
