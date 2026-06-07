# System Prompt — CloudMart Backend Additions & Bug Fixes (Go + Fiber)

## Context

The CloudMart backend (Go + Fiber) follows a layered architecture: `handler → service → repository`. All response bodies use the `pkg/response.WebResponse` wrapper:

```json
{ "code": 200, "status": "OK", "data": ..., "errors": null }
```

This prompt covers **one confirmed bug fix** and **four missing endpoints** identified from auditing `doc_v1.json` against MVP requirements. Do not modify any existing endpoints. Only touch the files listed per task.

---

## 🔴 PRIORITY 1 — Bug Fix: Voucher Claim URL Decode

### Problem

`POST /api/vouchers/{code}/claim` fails with 404 when the voucher code contains spaces or any characters that get percent-encoded in a URL.

**Reproduction:**
```bash
# Voucher code in DB: "PROMO AKHIR BULAN"
curl -X POST http://localhost:8080/api/vouchers/PROMO%20AKHIR%20BULAN/claim \
  -H "Authorization: Bearer <token>"
# Response: 404 "voucher tidak ditemukan"
```

**Root cause:** Fiber v2's `c.Params()` returns the raw URL path segment without decoding percent-encoding. The GORM query receives the literal string `PROMO%20AKHIR%20BULAN` and finds no matching row because the actual DB value is `PROMO AKHIR BULAN`.

```sql
-- What's actually running (wrong):
SELECT * FROM "vouchers" WHERE code = 'PROMO%20AKHIR%20BULAN'

-- What should run:
SELECT * FROM "vouchers" WHERE code = 'PROMO AKHIR BULAN'
```

### Fix

**File:** `internal/handler/voucher_handler.go`

```go
import (
    "net/url"
    // ... rest of existing imports unchanged
)

func (h *VoucherHandler) ClaimVoucher(c *fiber.Ctx) error {
    rawCode := c.Params("code")

    // Fiber v2 does not auto-decode path params — decode manually
    code, err := url.PathUnescape(rawCode)
    if err != nil || code == "" {
        code = rawCode // safe fallback
    }

    userID := c.Locals("user_id").(uint) // adjust to match your auth middleware locals key
    // ... rest of existing handler logic, replacing c.Params("code") with `code`
}
```

**Scope check:** Search the entire codebase for `c.Params("code")`. If it appears in other handlers, apply the same fix to all of them.

**Test after fix:**
```bash
# Should return 200 or 409 (already claimed), never 404
curl -X POST http://localhost:8080/api/vouchers/PROMO%20AKHIR%20BULAN/claim \
  -H "Authorization: Bearer <token>"
```

**No route changes needed** — the route definition `voucher.Post("/:code/claim", ...)` is correct as-is.

---

## 🟡 PRIORITY 2 — Missing: `GET /api/seller/products/:id`

### Why

`/api/seller/products/{id}` only exposes `PUT` and `DELETE`. There is no `GET` with seller ownership verification, which is needed for the product edit form to safely prefetch data.

### Middleware

BearerAuth + SellerMiddleware

### Handler

**File:** `internal/handler/seller_catalog_handler.go`

Add method `GetProductByID`:

```go
func (h *SellerCatalogHandler) GetProductByID(c *fiber.Ctx) error {
    productID, err := strconv.Atoi(c.Params("id"))
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(response.Error(400, "invalid product ID"))
    }

    sellerUserID := c.Locals("user_id").(uint) // adjust to match your auth locals key

    product, err := h.sellerCatalogService.GetProductByID(uint(productID), sellerUserID)
    if err != nil {
        // service should return sentinel errors for not-found vs forbidden
        return c.Status(fiber.StatusNotFound).JSON(response.Error(404, err.Error()))
    }

    return c.Status(fiber.StatusOK).JSON(response.Success(200, "OK", product))
}
```

### Service

**File:** `internal/service/seller_catalog_service.go`

```go
func (s *SellerCatalogService) GetProductByID(productID, sellerUserID uint) (*entity.Product, error) {
    product, err := s.productRepository.GetByIDForSeller(productID, sellerUserID)
    if err != nil {
        return nil, err
    }
    return product, nil
}
```

### Repository

**File:** `internal/repository/product_repository.go`

```go
func (r *ProductRepository) GetByIDForSeller(productID, sellerUserID uint) (*entity.Product, error) {
    var product entity.Product
    result := r.db.
        Joins("JOIN stores ON stores.id = products.store_id").
        Where("products.id = ? AND stores.user_id = ?", productID, sellerUserID).
        Preload("Variants").
        Preload("Category").
        Preload("Store").
        First(&product)

    if result.Error != nil {
        return nil, result.Error
    }
    return &product, nil
}
```

### Route Registration

**File:** `cmd/api/main.go`

```go
// Inside the seller route group, alongside existing seller product routes
seller.Get("/products/:id", sellerCatalogHandler.GetProductByID)
```

### Response

**200 OK:**
```json
{
  "code": 200,
  "status": "OK",
  "data": {
    "id": 1,
    "name": "...",
    "category_id": 2,
    "category": { ... },
    "store_id": 1,
    "store": { ... },
    "variants": [ ... ],
    "image_url": "...",
    "description": "...",
    "created_at": "...",
    "updated_at": "..."
  }
}
```

**Errors:** 400 (invalid ID), 403 (product exists but not owned by this seller), 404 (not found)

---

## 🟡 PRIORITY 3 — Missing: `PUT /api/seller/payment-configs/:id`

### Why

`/api/seller/payment-configs/{id}` only has `DELETE`. There is no `PUT` to rename an existing config group, making the payment method dashboard incomplete CRUD.

### Middleware

BearerAuth + SellerMiddleware

### Handler

**File:** `internal/handler/seller_promo_handler.go`

```go
func (h *SellerPromoHandler) UpdatePaymentConfig(c *fiber.Ctx) error {
    configID, err := strconv.Atoi(c.Params("id"))
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(response.Error(400, "invalid config ID"))
    }

    var req struct {
        Name string `json:"name" validate:"required,min=1,max=100"`
    }
    if err := c.BodyParser(&req); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(response.Error(400, "invalid request body"))
    }
    if err := h.validator.Validate(req); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(response.Error(400, err.Error()))
    }

    sellerUserID := c.Locals("user_id").(uint)

    updated, err := h.sellerPromoService.UpdatePaymentConfig(uint(configID), sellerUserID, req.Name)
    if err != nil {
        return c.Status(fiber.StatusNotFound).JSON(response.Error(404, err.Error()))
    }

    return c.Status(fiber.StatusOK).JSON(response.Success(200, "OK", updated))
}
```

### Service

**File:** `internal/service/seller_promo_service.go`

```go
func (s *SellerPromoService) UpdatePaymentConfig(configID, sellerUserID uint, name string) (*entity.PaymentConfiguration, error) {
    // Verify ownership via store
    store, err := s.storeRepository.GetByUserID(sellerUserID)
    if err != nil {
        return nil, errors.New("store not found")
    }

    config, err := s.paymentRepository.GetConfigByIDAndStoreID(configID, store.ID)
    if err != nil {
        return nil, errors.New("payment config not found or not owned by seller")
    }

    config.Name = name
    if err := s.paymentRepository.UpdateConfig(config); err != nil {
        return nil, err
    }
    return config, nil
}
```

### Repository

**File:** `internal/repository/payment_repository.go`

```go
func (r *PaymentRepository) GetConfigByIDAndStoreID(configID, storeID uint) (*entity.PaymentConfiguration, error) {
    var config entity.PaymentConfiguration
    result := r.db.Where("id = ? AND store_id = ?", configID, storeID).First(&config)
    if result.Error != nil {
        return nil, result.Error
    }
    return &config, nil
}

func (r *PaymentRepository) UpdateConfig(config *entity.PaymentConfiguration) error {
    return r.db.Save(config).Error
}
```

### Route Registration

```go
seller.Put("/payment-configs/:id", sellerPromoHandler.UpdatePaymentConfig)
```

### Response

**200 OK:** Updated `entity.PaymentConfiguration` object.

**Errors:** 400 (validation), 401, 403, 404

---

## 🟡 PRIORITY 4 — Missing: `PUT /api/seller/banks/:id`

### Why

`/api/seller/banks/{id}` only has `DELETE`. No `PUT` to update bank account details.

### Middleware

BearerAuth + SellerMiddleware

### Handler

**File:** `internal/handler/seller_promo_handler.go`

```go
func (h *SellerPromoHandler) UpdateBank(c *fiber.Ctx) error {
    bankID, err := strconv.Atoi(c.Params("id"))
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(response.Error(400, "invalid bank ID"))
    }

    var req struct {
        Name        string `json:"name"         validate:"required"`
        AccountID   string `json:"account_id"   validate:"required"`
        AccountName string `json:"account_name" validate:"required"`
    }
    if err := c.BodyParser(&req); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(response.Error(400, "invalid request body"))
    }
    if err := h.validator.Validate(req); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(response.Error(400, err.Error()))
    }

    sellerUserID := c.Locals("user_id").(uint)

    updated, err := h.sellerPromoService.UpdateBank(uint(bankID), sellerUserID, req.Name, req.AccountID, req.AccountName)
    if err != nil {
        return c.Status(fiber.StatusNotFound).JSON(response.Error(404, err.Error()))
    }

    return c.Status(fiber.StatusOK).JSON(response.Success(200, "OK", updated))
}
```

### Service

**File:** `internal/service/seller_promo_service.go`

```go
func (s *SellerPromoService) UpdateBank(bankID, sellerUserID uint, name, accountID, accountName string) (*entity.PaymentBank, error) {
    store, err := s.storeRepository.GetByUserID(sellerUserID)
    if err != nil {
        return nil, errors.New("store not found")
    }

    bank, err := s.paymentRepository.GetBankByIDAndStoreID(bankID, store.ID)
    if err != nil {
        return nil, errors.New("bank not found or not owned by seller")
    }

    bank.Name        = name
    bank.AccountID   = accountID
    bank.AccountName = accountName

    if err := s.paymentRepository.UpdateBank(bank); err != nil {
        return nil, err
    }
    return bank, nil
}
```

### Repository

**File:** `internal/repository/payment_repository.go`

```go
// GetBankByIDAndStoreID — join through PaymentConfiguration to verify store ownership
func (r *PaymentRepository) GetBankByIDAndStoreID(bankID, storeID uint) (*entity.PaymentBank, error) {
    var bank entity.PaymentBank
    result := r.db.
        Joins("JOIN payment_configurations ON payment_configurations.id = payment_banks.payment_configuration_id").
        Where("payment_banks.id = ? AND payment_configurations.store_id = ?", bankID, storeID).
        First(&bank)
    if result.Error != nil {
        return nil, result.Error
    }
    return &bank, nil
}

func (r *PaymentRepository) UpdateBank(bank *entity.PaymentBank) error {
    return r.db.Save(bank).Error
}
```

### Route Registration

```go
seller.Put("/banks/:id", sellerPromoHandler.UpdateBank)
```

### Response

**200 OK:** Updated `entity.PaymentBank` object.

**Errors:** 400 (validation), 401, 403, 404

---

## 🟢 PRIORITY 5 — Missing: `POST /api/orders/:id/confirm`

### Why

There is no endpoint for the buyer to manually confirm they received their order. `settlement` status is currently only set by the Midtrans webhook. This endpoint provides a fallback so users can unlock the review feature if the webhook is delayed.

### Middleware

BearerAuth only (no seller middleware — this is a buyer action)

### Handler

**File:** `internal/handler/order_handler.go`

```go
func (h *OrderHandler) ConfirmOrder(c *fiber.Ctx) error {
    orderID, err := strconv.Atoi(c.Params("id"))
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(response.Error(400, "invalid order ID"))
    }

    userID := c.Locals("user_id").(uint)

    if err := h.orderService.ConfirmOrder(uint(orderID), userID); err != nil {
        return c.Status(fiber.StatusForbidden).JSON(response.Error(403, err.Error()))
    }

    return c.Status(fiber.StatusOK).JSON(response.Success(200, "OK", "Order confirmed as received"))
}
```

### Service

**File:** `internal/service/order_service.go`

```go
func (s *OrderService) ConfirmOrder(orderID, userID uint) error {
    order, err := s.orderRepository.GetByIDAndUserID(orderID, userID)
    if err != nil {
        return errors.New("order not found")
    }
    if order.ShippingStatus != "delivered" {
        return errors.New("order has not been delivered yet")
    }
    if order.PaymentStatus == "settlement" {
        return nil // idempotent — already confirmed, no-op
    }
    return s.orderRepository.UpdatePaymentStatus(orderID, "settlement")
}
```

### Repository

**File:** `internal/repository/order_repository.go`

```go
// Add if not already present:
func (r *OrderRepository) GetByIDAndUserID(orderID, userID uint) (*entity.Order, error) {
    var order entity.Order
    result := r.db.Where("id = ? AND user_id = ?", orderID, userID).First(&order)
    if result.Error != nil {
        return nil, result.Error
    }
    return &order, nil
}

func (r *OrderRepository) UpdatePaymentStatus(orderID uint, status string) error {
    return r.db.Model(&entity.Order{}).Where("id = ?", orderID).Update("payment_status", status).Error
}
```

### Route Registration

```go
// In the user orders route group
orders.Post("/:id/confirm", orderHandler.ConfirmOrder)
```

### Response

**200 OK:** `"Order confirmed as received"`

**Errors:** 400 (invalid ID), 403 (not delivered yet or not owner), 404 (not found)

---

## 🟢 PRIORITY 6 — Nice-to-Have: `PATCH /api/seller/variants/:id/stock`

### Why

`PUT /api/seller/variants/{id}` requires all fields (`color`, `price`, `size`, `sku`) — all marked `required` in `VariantRequest`. Doing a quick stock update forces the client to send the full payload. A dedicated PATCH reduces the payload and makes stock management UX cleaner.

### Handler

**File:** `internal/handler/seller_catalog_handler.go`

```go
func (h *SellerCatalogHandler) UpdateVariantStock(c *fiber.Ctx) error {
    variantID, err := strconv.Atoi(c.Params("id"))
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(response.Error(400, "invalid variant ID"))
    }

    var req struct {
        Stock int `json:"stock" validate:"min=0"`
    }
    if err := c.BodyParser(&req); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(response.Error(400, "invalid request body"))
    }
    if err := h.validator.Validate(req); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(response.Error(400, err.Error()))
    }

    sellerUserID := c.Locals("user_id").(uint)

    updated, err := h.sellerCatalogService.UpdateVariantStock(uint(variantID), sellerUserID, req.Stock)
    if err != nil {
        return c.Status(fiber.StatusForbidden).JSON(response.Error(403, err.Error()))
    }

    return c.Status(fiber.StatusOK).JSON(response.Success(200, "OK", updated))
}
```

### Service

**File:** `internal/service/seller_catalog_service.go`

```go
func (s *SellerCatalogService) UpdateVariantStock(variantID, sellerUserID uint, stock int) (*entity.ProductVariant, error) {
    // Verify ownership: variant → product → store → user
    variant, err := s.productRepository.GetVariantByIDForSeller(variantID, sellerUserID)
    if err != nil {
        return nil, errors.New("variant not found or not owned by seller")
    }
    variant.Stock = stock
    if err := s.productRepository.UpdateVariantStock(variant); err != nil {
        return nil, err
    }
    return variant, nil
}
```

### Repository

**File:** `internal/repository/product_repository.go`

```go
func (r *ProductRepository) GetVariantByIDForSeller(variantID, sellerUserID uint) (*entity.ProductVariant, error) {
    var variant entity.ProductVariant
    result := r.db.
        Joins("JOIN products ON products.id = product_variants.product_id").
        Joins("JOIN stores ON stores.id = products.store_id").
        Where("product_variants.id = ? AND stores.user_id = ?", variantID, sellerUserID).
        First(&variant)
    if result.Error != nil {
        return nil, result.Error
    }
    return &variant, nil
}

func (r *ProductRepository) UpdateVariantStock(variant *entity.ProductVariant) error {
    return r.db.Model(variant).Update("stock", variant.Stock).Error
}
```

### Route Registration

```go
seller.Patch("/variants/:id/stock", sellerCatalogHandler.UpdateVariantStock)
```

---

## Files Modified Summary

```
internal/handler/voucher_handler.go         ← Bug fix: URL-decode code param
internal/handler/seller_catalog_handler.go  ← Add GetProductByID, UpdateVariantStock
internal/handler/seller_promo_handler.go    ← Add UpdatePaymentConfig, UpdateBank
internal/handler/order_handler.go           ← Add ConfirmOrder

internal/service/seller_catalog_service.go  ← Add GetProductByID, UpdateVariantStock
internal/service/seller_promo_service.go    ← Add UpdatePaymentConfig, UpdateBank
internal/service/order_service.go           ← Add ConfirmOrder

internal/repository/product_repository.go   ← Add GetByIDForSeller, GetVariantByIDForSeller, UpdateVariantStock
internal/repository/payment_repository.go   ← Add GetConfigByIDAndStoreID, UpdateConfig, GetBankByIDAndStoreID, UpdateBank
internal/repository/order_repository.go     ← Add GetByIDAndUserID, UpdatePaymentStatus

cmd/api/main.go                             ← Register all new routes
```

---

## Route Registration Checklist

Add these lines to `cmd/api/main.go` in the appropriate route groups:

```go
// Seller group
seller.Get("/products/:id",        sellerCatalogHandler.GetProductByID)    // Priority 2
seller.Put("/payment-configs/:id", sellerPromoHandler.UpdatePaymentConfig) // Priority 3
seller.Put("/banks/:id",           sellerPromoHandler.UpdateBank)           // Priority 4
seller.Patch("/variants/:id/stock", sellerCatalogHandler.UpdateVariantStock) // Priority 6

// User orders group
orders.Post("/:id/confirm", orderHandler.ConfirmOrder) // Priority 5
```

The voucher bug fix (Priority 1) requires **no route changes** — only the handler body.
