package config

import (
	"log"

	"github.com/yosmisyael/cloudmart-web-service/internal/entity"
	"gorm.io/gorm"
)

const hashedPassword = "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lHuu"

func RunSeeder(db *gorm.DB) {
	var count int64
	db.Model(&entity.User{}).Count(&count)
	if count > 0 {
		log.Println("database not empty, seeding skipped.")
		return
	}

	log.Println("Seeding database...")

	// 1. Categories
	catLari := entity.Category{Name: "Sepatu Lari", IsDefault: true}
	catKasual := entity.Category{Name: "Sepatu Kasual", IsDefault: true}
	catTraining := entity.Category{Name: "Sepatu Training", IsDefault: true}
	catBasket := entity.Category{Name: "Sepatu Basket", IsDefault: true}
	catPakaian := entity.Category{Name: "Pakaian Olahraga", IsDefault: true}
	catSandal := entity.Category{Name: "Sandal & Slides", IsDefault: true}

	db.Create(&catLari)
	db.Create(&catKasual)
	db.Create(&catTraining)
	db.Create(&catBasket)
	db.Create(&catPakaian)
	db.Create(&catSandal)

	// 2. Nike
	nikeUser := entity.User{
		Name:     "Nike",
		Email:    "official@nike.com",
		Phone:    "08001234001",
		Password: hashedPassword,
		Role:     "seller",
	}
	db.Create(&nikeUser)

	nikeStore := entity.Store{
		UserID: nikeUser.ID,
		Name:   "Nike Official Store",
	}
	db.Create(&nikeStore)

	// Nike Product 1
	nikeP1 := entity.Product{
		StoreID:     nikeStore.ID,
		CategoryID:  catKasual.ID,
		Name:        "Nike Air Max 270",
		Description: "Sepatu kasual ikonik dengan unit udara Max di tumit untuk kenyamanan sepanjang hari",
		ImageURL:    "https://images.unsplash.com/photo-1542291026-7eec264c27ff?w=800",
	}
	db.Create(&nikeP1)

	db.Create(&entity.ProductVariant{ProductID: nikeP1.ID, SKU: "NK-AM270-BLK-39", Color: "Black", Size: "39", Price: 1899000, Stock: 15, ImageURL: "https://images.unsplash.com/photo-1542291026-7eec264c27ff?w=800"})
	db.Create(&entity.ProductVariant{ProductID: nikeP1.ID, SKU: "NK-AM270-BLK-40", Color: "Black", Size: "40", Price: 1899000, Stock: 20, ImageURL: "https://images.unsplash.com/photo-1542291026-7eec264c27ff?w=800"})
	db.Create(&entity.ProductVariant{ProductID: nikeP1.ID, SKU: "NK-AM270-BLK-41", Color: "Black", Size: "41", Price: 1899000, Stock: 18, ImageURL: "https://images.unsplash.com/photo-1542291026-7eec264c27ff?w=800"})
	db.Create(&entity.ProductVariant{ProductID: nikeP1.ID, SKU: "NK-AM270-WHT-39", Color: "White", Size: "39", Price: 1899000, Stock: 12, ImageURL: "https://images.unsplash.com/photo-1549298916-b41d501d3772?w=800"})
	db.Create(&entity.ProductVariant{ProductID: nikeP1.ID, SKU: "NK-AM270-WHT-40", Color: "White", Size: "40", Price: 1899000, Stock: 25, ImageURL: "https://images.unsplash.com/photo-1549298916-b41d501d3772?w=800"})
	db.Create(&entity.ProductVariant{ProductID: nikeP1.ID, SKU: "NK-AM270-WHT-41", Color: "White", Size: "41", Price: 1899000, Stock: 10, ImageURL: "https://images.unsplash.com/photo-1549298916-b41d501d3772?w=800"})

	// Nike Product 2
	nikeP2 := entity.Product{
		StoreID:     nikeStore.ID,
		CategoryID:  catLari.ID,
		Name:        "Nike Air Zoom Pegasus 40",
		Description: "Sepatu lari harian dengan bantalan responsif dan upper yang breathable untuk performa optimal",
		ImageURL:    "https://images.unsplash.com/photo-1606107557195-0e29a4b5b4aa?w=800",
	}
	db.Create(&nikeP2)

	db.Create(&entity.ProductVariant{ProductID: nikeP2.ID, SKU: "NK-PEG40-BLU-40", Color: "Blue", Size: "40", Price: 1599000, Stock: 20, ImageURL: "https://images.unsplash.com/photo-1606107557195-0e29a4b5b4aa?w=800"})
	db.Create(&entity.ProductVariant{ProductID: nikeP2.ID, SKU: "NK-PEG40-BLU-41", Color: "Blue", Size: "41", Price: 1599000, Stock: 15, ImageURL: "https://images.unsplash.com/photo-1606107557195-0e29a4b5b4aa?w=800"})
	db.Create(&entity.ProductVariant{ProductID: nikeP2.ID, SKU: "NK-PEG40-BLU-42", Color: "Blue", Size: "42", Price: 1599000, Stock: 10, ImageURL: "https://images.unsplash.com/photo-1606107557195-0e29a4b5b4aa?w=800"})
	db.Create(&entity.ProductVariant{ProductID: nikeP2.ID, SKU: "NK-PEG40-BLK-40", Color: "Black", Size: "40", Price: 1599000, Stock: 18, ImageURL: "https://images.unsplash.com/photo-1542291026-7eec264c27ff?w=800"})
	db.Create(&entity.ProductVariant{ProductID: nikeP2.ID, SKU: "NK-PEG40-BLK-41", Color: "Black", Size: "41", Price: 1599000, Stock: 22, ImageURL: "https://images.unsplash.com/photo-1542291026-7eec264c27ff?w=800"})

	// Nike Product 3
	nikeP3 := entity.Product{
		StoreID:     nikeStore.ID,
		CategoryID:  catPakaian.ID,
		Name:        "Nike Dri-FIT Running Tee",
		Description: "Kaos lari dengan teknologi Dri-FIT yang menyerap keringat dan menjaga tubuh tetap kering",
		ImageURL:    "https://images.unsplash.com/photo-1521572163474-6864f9cf17ab?w=800",
	}
	db.Create(&nikeP3)

	db.Create(&entity.ProductVariant{ProductID: nikeP3.ID, SKU: "NK-DFIT-BLK-S", Color: "Black", Size: "S", Price: 449000, Stock: 30, ImageURL: "https://images.unsplash.com/photo-1521572163474-6864f9cf17ab?w=800"})
	db.Create(&entity.ProductVariant{ProductID: nikeP3.ID, SKU: "NK-DFIT-BLK-M", Color: "Black", Size: "M", Price: 449000, Stock: 35, ImageURL: "https://images.unsplash.com/photo-1521572163474-6864f9cf17ab?w=800"})
	db.Create(&entity.ProductVariant{ProductID: nikeP3.ID, SKU: "NK-DFIT-BLK-L", Color: "Black", Size: "L", Price: 449000, Stock: 25, ImageURL: "https://images.unsplash.com/photo-1521572163474-6864f9cf17ab?w=800"})
	db.Create(&entity.ProductVariant{ProductID: nikeP3.ID, SKU: "NK-DFIT-WHT-S", Color: "White", Size: "S", Price: 449000, Stock: 20, ImageURL: "https://images.unsplash.com/photo-1576566588028-4147f3842f27?w=800"})
	db.Create(&entity.ProductVariant{ProductID: nikeP3.ID, SKU: "NK-DFIT-WHT-M", Color: "White", Size: "M", Price: 449000, Stock: 28, ImageURL: "https://images.unsplash.com/photo-1576566588028-4147f3842f27?w=800"})
	db.Create(&entity.ProductVariant{ProductID: nikeP3.ID, SKU: "NK-DFIT-WHT-L", Color: "White", Size: "L", Price: 449000, Stock: 15, ImageURL: "https://images.unsplash.com/photo-1576566588028-4147f3842f27?w=800"})

	// 3. Adidas
	adidasUser := entity.User{
		Name:     "Adidas",
		Email:    "official@adidas.com",
		Phone:    "08001234002",
		Password: hashedPassword,
		Role:     "seller",
	}
	db.Create(&adidasUser)

	adidasStore := entity.Store{
		UserID: adidasUser.ID,
		Name:   "Adidas Official Store",
	}
	db.Create(&adidasStore)

	// Adidas Product 1
	adidasP1 := entity.Product{
		StoreID:     adidasStore.ID,
		CategoryID:  catLari.ID,
		Name:        "Adidas Ultraboost 22",
		Description: "Sepatu lari premium dengan teknologi Boost untuk pengembalian energi maksimal setiap langkah",
		ImageURL:    "https://images.unsplash.com/photo-1608231387042-66d1773070a5?w=800",
	}
	db.Create(&adidasP1)

	db.Create(&entity.ProductVariant{ProductID: adidasP1.ID, SKU: "AD-UB22-WHT-40", Color: "White", Size: "40", Price: 2199000, Stock: 12, ImageURL: "https://images.unsplash.com/photo-1608231387042-66d1773070a5?w=800"})
	db.Create(&entity.ProductVariant{ProductID: adidasP1.ID, SKU: "AD-UB22-WHT-41", Color: "White", Size: "41", Price: 2199000, Stock: 15, ImageURL: "https://images.unsplash.com/photo-1608231387042-66d1773070a5?w=800"})
	db.Create(&entity.ProductVariant{ProductID: adidasP1.ID, SKU: "AD-UB22-WHT-42", Color: "White", Size: "42", Price: 2199000, Stock: 10, ImageURL: "https://images.unsplash.com/photo-1608231387042-66d1773070a5?w=800"})
	db.Create(&entity.ProductVariant{ProductID: adidasP1.ID, SKU: "AD-UB22-BLK-40", Color: "Black", Size: "40", Price: 2199000, Stock: 18, ImageURL: "https://images.unsplash.com/photo-1542291026-7eec264c27ff?w=800"})
	db.Create(&entity.ProductVariant{ProductID: adidasP1.ID, SKU: "AD-UB22-BLK-41", Color: "Black", Size: "41", Price: 2199000, Stock: 20, ImageURL: "https://images.unsplash.com/photo-1542291026-7eec264c27ff?w=800"})
	db.Create(&entity.ProductVariant{ProductID: adidasP1.ID, SKU: "AD-UB22-BLK-42", Color: "Black", Size: "42", Price: 2199000, Stock: 8, ImageURL: "https://images.unsplash.com/photo-1542291026-7eec264c27ff?w=800"})

	// Adidas Product 2
	adidasP2 := entity.Product{
		StoreID:     adidasStore.ID,
		CategoryID:  catKasual.ID,
		Name:        "Adidas Stan Smith",
		Description: "Sepatu kasual ikonik dengan desain minimalis dan sol karet yang tahan lama",
		ImageURL:    "https://images.unsplash.com/photo-1539185441755-769473a23570?w=800",
	}
	db.Create(&adidasP2)

	db.Create(&entity.ProductVariant{ProductID: adidasP2.ID, SKU: "AD-SS-WHT-38", Color: "White", Size: "38", Price: 1299000, Stock: 20, ImageURL: "https://images.unsplash.com/photo-1539185441755-769473a23570?w=800"})
	db.Create(&entity.ProductVariant{ProductID: adidasP2.ID, SKU: "AD-SS-WHT-39", Color: "White", Size: "39", Price: 1299000, Stock: 25, ImageURL: "https://images.unsplash.com/photo-1539185441755-769473a23570?w=800"})
	db.Create(&entity.ProductVariant{ProductID: adidasP2.ID, SKU: "AD-SS-WHT-40", Color: "White", Size: "40", Price: 1299000, Stock: 30, ImageURL: "https://images.unsplash.com/photo-1539185441755-769473a23570?w=800"})
	db.Create(&entity.ProductVariant{ProductID: adidasP2.ID, SKU: "AD-SS-WHT-41", Color: "White", Size: "41", Price: 1299000, Stock: 15, ImageURL: "https://images.unsplash.com/photo-1539185441755-769473a23570?w=800"})

	// Adidas Product 3
	adidasP3 := entity.Product{
		StoreID:     adidasStore.ID,
		CategoryID:  catPakaian.ID,
		Name:        "Adidas Tiro Training Pants",
		Description: "Celana training dengan teknologi AEROREADY untuk menyerap kelembaban dan desain tapered yang stylish",
		ImageURL:    "https://images.unsplash.com/photo-1506629082955-511b1aa562c8?w=800",
	}
	db.Create(&adidasP3)

	db.Create(&entity.ProductVariant{ProductID: adidasP3.ID, SKU: "AD-TIRO-BLK-S", Color: "Black", Size: "S", Price: 549000, Stock: 25, ImageURL: "https://images.unsplash.com/photo-1506629082955-511b1aa562c8?w=800"})
	db.Create(&entity.ProductVariant{ProductID: adidasP3.ID, SKU: "AD-TIRO-BLK-M", Color: "Black", Size: "M", Price: 549000, Stock: 30, ImageURL: "https://images.unsplash.com/photo-1506629082955-511b1aa562c8?w=800"})
	db.Create(&entity.ProductVariant{ProductID: adidasP3.ID, SKU: "AD-TIRO-BLK-L", Color: "Black", Size: "L", Price: 549000, Stock: 20, ImageURL: "https://images.unsplash.com/photo-1506629082955-511b1aa562c8?w=800"})
	db.Create(&entity.ProductVariant{ProductID: adidasP3.ID, SKU: "AD-TIRO-NVY-S", Color: "Navy", Size: "S", Price: 549000, Stock: 15, ImageURL: "https://images.unsplash.com/photo-1556906781-9a412961a28b?w=800"})
	db.Create(&entity.ProductVariant{ProductID: adidasP3.ID, SKU: "AD-TIRO-NVY-M", Color: "Navy", Size: "M", Price: 549000, Stock: 20, ImageURL: "https://images.unsplash.com/photo-1556906781-9a412961a28b?w=800"})

	// 4. Skechers
	skechersUser := entity.User{
		Name:     "Skechers",
		Email:    "official@skechers.com",
		Phone:    "08001234003",
		Password: hashedPassword,
		Role:     "seller",
	}
	db.Create(&skechersUser)

	skechersStore := entity.Store{
		UserID: skechersUser.ID,
		Name:   "Skechers Official Store",
	}
	db.Create(&skechersStore)

	// Skechers Product 1
	skechersP1 := entity.Product{
		StoreID:     skechersStore.ID,
		CategoryID:  catKasual.ID,
		Name:        "Skechers Go Walk 6",
		Description: "Sepatu walking ultra-nyaman dengan sol ULTRA GO yang ringan dan teknologi 5GEN",
		ImageURL:    "https://images.unsplash.com/photo-1560769629-975ec94e6a86?w=800",
	}
	db.Create(&skechersP1)

	db.Create(&entity.ProductVariant{ProductID: skechersP1.ID, SKU: "SK-GW6-GRY-39", Color: "Gray", Size: "39", Price: 899000, Stock: 20, ImageURL: "https://images.unsplash.com/photo-1560769629-975ec94e6a86?w=800"})
	db.Create(&entity.ProductVariant{ProductID: skechersP1.ID, SKU: "SK-GW6-GRY-40", Color: "Gray", Size: "40", Price: 899000, Stock: 25, ImageURL: "https://images.unsplash.com/photo-1560769629-975ec94e6a86?w=800"})
	db.Create(&entity.ProductVariant{ProductID: skechersP1.ID, SKU: "SK-GW6-GRY-41", Color: "Gray", Size: "41", Price: 899000, Stock: 18, ImageURL: "https://images.unsplash.com/photo-1560769629-975ec94e6a86?w=800"})
	db.Create(&entity.ProductVariant{ProductID: skechersP1.ID, SKU: "SK-GW6-BLK-39", Color: "Black", Size: "39", Price: 899000, Stock: 15, ImageURL: "https://images.unsplash.com/photo-1542291026-7eec264c27ff?w=800"})
	db.Create(&entity.ProductVariant{ProductID: skechersP1.ID, SKU: "SK-GW6-BLK-40", Color: "Black", Size: "40", Price: 899000, Stock: 22, ImageURL: "https://images.unsplash.com/photo-1542291026-7eec264c27ff?w=800"})
	db.Create(&entity.ProductVariant{ProductID: skechersP1.ID, SKU: "SK-GW6-BLK-41", Color: "Black", Size: "41", Price: 899000, Stock: 12, ImageURL: "https://images.unsplash.com/photo-1542291026-7eec264c27ff?w=800"})

	// Skechers Product 2
	skechersP2 := entity.Product{
		StoreID:     skechersStore.ID,
		CategoryID:  catKasual.ID,
		Name:        "Skechers D'Lites",
		Description: "Sepatu chunky retro dengan cushioning tebal dan desain dua warna yang ikonik",
		ImageURL:    "https://images.unsplash.com/photo-1595950653106-6c9ebd614d3a?w=800",
	}
	db.Create(&skechersP2)

	db.Create(&entity.ProductVariant{ProductID: skechersP2.ID, SKU: "SK-DL-WHTBLK-38", Color: "White/Black", Size: "38", Price: 1099000, Stock: 15, ImageURL: "https://images.unsplash.com/photo-1595950653106-6c9ebd614d3a?w=800"})
	db.Create(&entity.ProductVariant{ProductID: skechersP2.ID, SKU: "SK-DL-WHTBLK-39", Color: "White/Black", Size: "39", Price: 1099000, Stock: 20, ImageURL: "https://images.unsplash.com/photo-1595950653106-6c9ebd614d3a?w=800"})
	db.Create(&entity.ProductVariant{ProductID: skechersP2.ID, SKU: "SK-DL-WHTBLK-40", Color: "White/Black", Size: "40", Price: 1099000, Stock: 18, ImageURL: "https://images.unsplash.com/photo-1595950653106-6c9ebd614d3a?w=800"})
	db.Create(&entity.ProductVariant{ProductID: skechersP2.ID, SKU: "SK-DL-WHTBLK-41", Color: "White/Black", Size: "41", Price: 1099000, Stock: 10, ImageURL: "https://images.unsplash.com/photo-1595950653106-6c9ebd614d3a?w=800"})

	// Skechers Product 3
	skechersP3 := entity.Product{
		StoreID:     skechersStore.ID,
		CategoryID:  catSandal.ID,
		Name:        "Skechers Arch Fit Slides",
		Description: "Sandal slide dengan dukungan lengkung kaki yang dirancang bersama podiatrist untuk kenyamanan seharian",
		ImageURL:    "https://images.unsplash.com/photo-1603487742131-4160ec999306?w=800",
	}
	db.Create(&skechersP3)

	db.Create(&entity.ProductVariant{ProductID: skechersP3.ID, SKU: "SK-AFS-BLK-38", Color: "Black", Size: "38", Price: 649000, Stock: 20, ImageURL: "https://images.unsplash.com/photo-1603487742131-4160ec999306?w=800"})
	db.Create(&entity.ProductVariant{ProductID: skechersP3.ID, SKU: "SK-AFS-BLK-39", Color: "Black", Size: "39", Price: 649000, Stock: 25, ImageURL: "https://images.unsplash.com/photo-1603487742131-4160ec999306?w=800"})
	db.Create(&entity.ProductVariant{ProductID: skechersP3.ID, SKU: "SK-AFS-BLK-40", Color: "Black", Size: "40", Price: 649000, Stock: 30, ImageURL: "https://images.unsplash.com/photo-1603487742131-4160ec999306?w=800"})
	db.Create(&entity.ProductVariant{ProductID: skechersP3.ID, SKU: "SK-AFS-NVY-38", Color: "Navy", Size: "38", Price: 649000, Stock: 15, ImageURL: "https://images.unsplash.com/photo-1603487742131-4160ec999306?w=800"})
	db.Create(&entity.ProductVariant{ProductID: skechersP3.ID, SKU: "SK-AFS-NVY-39", Color: "Navy", Size: "39", Price: 649000, Stock: 18, ImageURL: "https://images.unsplash.com/photo-1603487742131-4160ec999306?w=800"})

	// 5. Puma
	pumaUser := entity.User{
		Name:     "Puma",
		Email:    "official@puma.com",
		Phone:    "08001234004",
		Password: hashedPassword,
		Role:     "seller",
	}
	db.Create(&pumaUser)

	pumaStore := entity.Store{
		UserID: pumaUser.ID,
		Name:   "Puma Official Store",
	}
	db.Create(&pumaStore)

	// Puma Product 1
	pumaP1 := entity.Product{
		StoreID:     pumaStore.ID,
		CategoryID:  catKasual.ID,
		Name:        "Puma Suede Classic",
		Description: "Sepatu suede klasik yang telah menjadi ikon streetwear sejak 1968 dengan sol karet formstripe",
		ImageURL:    "https://images.unsplash.com/photo-1608667508764-33cf0726b13a?w=800",
	}
	db.Create(&pumaP1)

	db.Create(&entity.ProductVariant{ProductID: pumaP1.ID, SKU: "PM-SC-BLK-39", Color: "Black", Size: "39", Price: 1199000, Stock: 20, ImageURL: "https://images.unsplash.com/photo-1608667508764-33cf0726b13a?w=800"})
	db.Create(&entity.ProductVariant{ProductID: pumaP1.ID, SKU: "PM-SC-BLK-40", Color: "Black", Size: "40", Price: 1199000, Stock: 25, ImageURL: "https://images.unsplash.com/photo-1608667508764-33cf0726b13a?w=800"})
	db.Create(&entity.ProductVariant{ProductID: pumaP1.ID, SKU: "PM-SC-BLK-41", Color: "Black", Size: "41", Price: 1199000, Stock: 18, ImageURL: "https://images.unsplash.com/photo-1608667508764-33cf0726b13a?w=800"})
	db.Create(&entity.ProductVariant{ProductID: pumaP1.ID, SKU: "PM-SC-NVY-39", Color: "Navy", Size: "39", Price: 1199000, Stock: 15, ImageURL: "https://images.unsplash.com/photo-1608667508764-33cf0726b13a?w=800"})
	db.Create(&entity.ProductVariant{ProductID: pumaP1.ID, SKU: "PM-SC-NVY-40", Color: "Navy", Size: "40", Price: 1199000, Stock: 20, ImageURL: "https://images.unsplash.com/photo-1608667508764-33cf0726b13a?w=800"})

	// Puma Product 2
	pumaP2 := entity.Product{
		StoreID:     pumaStore.ID,
		CategoryID:  catLari.ID,
		Name:        "Puma Velocity Nitro 2",
		Description: "Sepatu lari dengan midsole NITRO foam yang ringan dan responsif untuk lari harian berkecepatan tinggi",
		ImageURL:    "https://images.unsplash.com/photo-1539109136881-3be0616acf4b?w=800",
	}
	db.Create(&pumaP2)

	db.Create(&entity.ProductVariant{ProductID: pumaP2.ID, SKU: "PM-VN2-ORG-40", Color: "Orange", Size: "40", Price: 1699000, Stock: 12, ImageURL: "https://images.unsplash.com/photo-1539109136881-3be0616acf4b?w=800"})
	db.Create(&entity.ProductVariant{ProductID: pumaP2.ID, SKU: "PM-VN2-ORG-41", Color: "Orange", Size: "41", Price: 1699000, Stock: 15, ImageURL: "https://images.unsplash.com/photo-1539109136881-3be0616acf4b?w=800"})
	db.Create(&entity.ProductVariant{ProductID: pumaP2.ID, SKU: "PM-VN2-ORG-42", Color: "Orange", Size: "42", Price: 1699000, Stock: 10, ImageURL: "https://images.unsplash.com/photo-1539109136881-3be0616acf4b?w=800"})
	db.Create(&entity.ProductVariant{ProductID: pumaP2.ID, SKU: "PM-VN2-BLK-40", Color: "Black", Size: "40", Price: 1699000, Stock: 18, ImageURL: "https://images.unsplash.com/photo-1542291026-7eec264c27ff?w=800"})
	db.Create(&entity.ProductVariant{ProductID: pumaP2.ID, SKU: "PM-VN2-BLK-41", Color: "Black", Size: "41", Price: 1699000, Stock: 20, ImageURL: "https://images.unsplash.com/photo-1542291026-7eec264c27ff?w=800"})

	// Puma Product 3
	pumaP3 := entity.Product{
		StoreID:     pumaStore.ID,
		CategoryID:  catPakaian.ID,
		Name:        "Puma Active Woven Shorts",
		Description: "Celana pendek olahraga dengan bahan woven ringan dan elastic waistband untuk latihan intensitas tinggi",
		ImageURL:    "https://images.unsplash.com/photo-1591195853828-11db59a44f43?w=800",
	}
	db.Create(&pumaP3)

	db.Create(&entity.ProductVariant{ProductID: pumaP3.ID, SKU: "PM-AWS-BLK-S", Color: "Black", Size: "S", Price: 399000, Stock: 30, ImageURL: "https://images.unsplash.com/photo-1591195853828-11db59a44f43?w=800"})
	db.Create(&entity.ProductVariant{ProductID: pumaP3.ID, SKU: "PM-AWS-BLK-M", Color: "Black", Size: "M", Price: 399000, Stock: 35, ImageURL: "https://images.unsplash.com/photo-1591195853828-11db59a44f43?w=800"})
	db.Create(&entity.ProductVariant{ProductID: pumaP3.ID, SKU: "PM-AWS-BLK-L", Color: "Black", Size: "L", Price: 399000, Stock: 25, ImageURL: "https://images.unsplash.com/photo-1591195853828-11db59a44f43?w=800"})
	db.Create(&entity.ProductVariant{ProductID: pumaP3.ID, SKU: "PM-AWS-RED-S", Color: "Red", Size: "S", Price: 399000, Stock: 20, ImageURL: "https://images.unsplash.com/photo-1591195853828-11db59a44f43?w=800"})
	db.Create(&entity.ProductVariant{ProductID: pumaP3.ID, SKU: "PM-AWS-RED-M", Color: "Red", Size: "M", Price: 399000, Stock: 22, ImageURL: "https://images.unsplash.com/photo-1591195853828-11db59a44f43?w=800"})

	log.Println("Seeder completed")
}