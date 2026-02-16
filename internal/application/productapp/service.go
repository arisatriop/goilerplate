package productapp

import (
	"context"
	"fmt"
	"goilerplate/internal/domain/category"
	"goilerplate/internal/domain/plantype"
	"goilerplate/internal/domain/plantyperule"
	"goilerplate/internal/domain/product"
	"goilerplate/internal/domain/productcategory"
	"goilerplate/internal/domain/productimage"
	"goilerplate/internal/domain/transaction"
	"goilerplate/internal/infrastructure/cache"
	"goilerplate/pkg/constants"
	"goilerplate/pkg/utils"
	"net/http"
	"time"

	"github.com/google/uuid"
)

var (
	boolTrue = true
)

type ApplicationService interface {
	GetProductDetails(ctx context.Context, id uuid.UUID, filter product.Filter) (*ProductDetails, error)
	GetProductListWithCategory(ctx context.Context, filter product.Filter) ([]ProductWithCategory, int64, error)
	GetCategoryListWithProducts(ctx context.Context, filter category.CategoryWithProductsFilter) ([]category.CategoryWithProducts, error)

	CreateProducts(ctx context.Context, products []Product) error
	CreateProductImages(ctx context.Context, productID uuid.UUID, images []productimage.ProductImage) error
	CreateProductCategories(ctx context.Context, productID uuid.UUID, categories []string) error
	UpdateProductCategories(ctx context.Context, productID uuid.UUID, categories []string) error

	UpdateToggleActive(ctx context.Context, productID uuid.UUID) (bool, error)

	DeleteProductByID(ctx context.Context, productID uuid.UUID) error
	DeleteProductImageByID(ctx context.Context, imageID uuid.UUID) error
	DeleteProductImagesByProductID(ctx context.Context, productID uuid.UUID) error
	DeleteProductCategoryByID(ctx context.Context, productID, categoryID uuid.UUID) error
	DeleteProductCategoryByProductID(ctx context.Context, productID uuid.UUID) error

	MarkImagesAsPrimary(ctx context.Context, imageID uuid.UUID) (bool, error)
}
type applicationService struct {
	txManager           transaction.Transaction
	cache               *cache.RedisService
	categoryRepo        category.Repository
	plantypeUC          plantype.Usecase
	plantypeRepo        plantype.Repository
	plantypeRuleUC      plantyperule.Usecase
	plantypeRuleRepo    plantyperule.Repository
	productRepo         product.Repository
	productImageRepo    productimage.Repository
	productCategoryRepo productcategory.Repository
}

func NewApplicationService(
	txManager transaction.Transaction,
	cacheService *cache.RedisService,
	categoryRepo category.Repository,
	plantypeUC plantype.Usecase,
	plantypeRepo plantype.Repository,
	plantypeRuleUC plantyperule.Usecase,
	plantypeRuleRepo plantyperule.Repository,
	productCategoryRepo productcategory.Repository,
	productImageRepo productimage.Repository,
	productRepo product.Repository,
) ApplicationService {
	return &applicationService{
		txManager:           txManager,
		cache:               cacheService,
		categoryRepo:        categoryRepo,
		plantypeUC:          plantypeUC,
		plantypeRepo:        plantypeRepo,
		plantypeRuleUC:      plantypeRuleUC,
		plantypeRuleRepo:    plantypeRuleRepo,
		productRepo:         productRepo,
		productImageRepo:    productImageRepo,
		productCategoryRepo: productCategoryRepo,
	}
}

func (s *applicationService) GetCategoryListWithProducts(ctx context.Context, filter category.CategoryWithProductsFilter) ([]category.CategoryWithProducts, error) {

	// Only cache when no filters are applied (default product list)
	shouldCache := filter.Keyword == "" && filter.CategoryID == nil
	cacheKey := fmt.Sprintf("store:%s:products", filter.StoreID.String())

	if shouldCache {
		var cacheData []category.CategoryWithProducts
		if err := s.cache.GetJSON(ctx, cacheKey, &cacheData); err != nil {
			return nil, fmt.Errorf("failed to get category list with products from cache: %w", err)
		}
		if cacheData != nil {
			return cacheData, nil
		}
	}

	data, err := s.categoryRepo.GetCategoryListWithProducts(ctx, &filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get category list with products: %w", err)
	}

	// Cache the result if it's the default query
	if shouldCache && len(data) > 0 {
		if err := s.cache.SetJSON(ctx, cacheKey, data, 15*time.Minute); err != nil {
			// Log error but don't fail the request
			fmt.Printf("failed to cache category list with products: %v\n", err)
		}
	}

	return data, nil
}

// invalidateProductCache invalidates the cached product list for a store
func (s *applicationService) invalidateProductCache(ctx context.Context, storeID uuid.UUID) {
	cacheKey := fmt.Sprintf("store:%s:products", storeID.String())
	if err := s.cache.Delete(ctx, cacheKey); err != nil {
		// Log error but don't fail the request
		fmt.Printf("failed to invalidate product cache for store %s: %v\n", storeID.String(), err)
	}
}

func (s *applicationService) GetSubscriptionRules(ctx context.Context, storeID string) (*SubscriptionRule, error) {
	var subscriptionRules *SubscriptionRule
	if err := s.cache.GetJSON(ctx, fmt.Sprintf("subscription:tenant:%s:rules", storeID), &subscriptionRules); err != nil {
		return nil, fmt.Errorf("failed to get subcrition rules: %w", err)
	}
	if subscriptionRules != nil {
		return subscriptionRules, nil
	}

	storeUUID, err := uuid.Parse(storeID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse store ID: %w", err)
	}

	planTypes, err := s.plantypeRepo.GetStoreActiveSubsciption(ctx, storeUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get store active subscription: %w", err)
	}
	if len(planTypes) == 0 {
		return nil, utils.Error(http.StatusForbidden, "Invalid store ID")
	}

	rules, err := s.plantypeRuleRepo.GetPlanTypeRuleByPlanTypeID(ctx, s.plantypeUC.FilterHighestPlanType(planTypes).ID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get plan type rules: %w", err)
	}

	subscriptionRules = buildSubscriptionRules(rules)

	if err := s.cache.SetJSON(ctx, fmt.Sprintf("subscription:tenant:%s:rules", storeID), subscriptionRules, 1*24*time.Hour); err != nil {
		return nil, fmt.Errorf("failed to set subscription rules to cache: %w", err)
	}

	return subscriptionRules, nil
}

func (s *applicationService) CreateProductCategories(ctx context.Context, productID uuid.UUID, categories []string) error {
	storeID := uuid.MustParse(ctx.Value(constants.ContextKeyStoreID).(string))

	prod, err := s.productRepo.GetProductByID(ctx, productID, &product.Filter{StoreID: &storeID})
	if err != nil {
		return fmt.Errorf("failed to get product by ID: %v", err)
	}
	if prod == nil {
		return utils.Error(http.StatusNotFound, constants.MsgUnauthorizedAccess)
	}

	plantypes, err := s.plantypeRepo.GetStoreActiveSubsciption(ctx, storeID)
	if err != nil {
		return fmt.Errorf("failed to get store active subscription: %v", err)
	}

	rules, err := s.plantypeRuleRepo.GetPlanTypeRuleByPlanTypeID(ctx, s.plantypeUC.FilterHighestPlanType(plantypes).ID.String())
	if err != nil {
		return fmt.Errorf("failed to get plan type rules: %v", err)
	}

	maxProductPerCategory := s.plantypeRuleUC.GetMaxProductPerCategoryFromRules(rules)

	return s.txManager.Do(ctx, func(txCtx context.Context) error {
		txProductCategoryRepo := s.productCategoryRepo.WithTx(txCtx)
		txCategoryRepo := s.categoryRepo.WithTx(txCtx)

		for _, catID := range categories {
			categoryUUID, err := uuid.Parse(catID)
			if err != nil {
				return fmt.Errorf("invalid category ID: %v", err)
			}
			cat, err := txCategoryRepo.GetCategoryByID(ctx, categoryUUID, &category.Filter{StoreID: &storeID})
			if err != nil {
				return fmt.Errorf("failed to get category by ID: %v", err)
			}
			if cat == nil {
				return utils.Error(http.StatusNotFound, fmt.Sprintf("kategori %s tidak ditemukan", catID))
			}

			if maxProductPerCategory > 0 {
				products, err := txProductCategoryRepo.GetProductCategoriesByCategoryID(ctx, categoryUUID)
				if err != nil {
					return fmt.Errorf("failed to get product categories by category ID: %v", err)
				}
				if len(products) >= maxProductPerCategory {
					return utils.Error(http.StatusBadRequest, productcategory.MsgMaxProductPerCategoryReached)
				}
			}

			newCategories := &productcategory.ProductCategory{
				ProductID:  productID,
				CategoryID: categoryUUID,
				IsActive:   true,
			}
			if err := txProductCategoryRepo.CreateProductCategory(txCtx, newCategories); err != nil {
				return fmt.Errorf("failed to create product category: %v", err)
			}
		}

		s.invalidateProductCache(ctx, storeID)

		return nil
	})
}

func (s *applicationService) UpdateProductCategories(ctx context.Context, productID uuid.UUID, categories []string) error {
	storeID := uuid.MustParse(ctx.Value(constants.ContextKeyStoreID).(string))

	prod, err := s.productRepo.GetProductByID(ctx, productID, &product.Filter{StoreID: &storeID})
	if err != nil {
		return fmt.Errorf("failed to get product by ID: %v", err)
	}
	if prod == nil {
		return utils.Error(http.StatusNotFound, constants.MsgUnauthorizedAccess)
	}

	plantypes, err := s.plantypeRepo.GetStoreActiveSubsciption(ctx, storeID)
	if err != nil {
		return fmt.Errorf("failed to get store active subscription: %v", err)
	}

	rules, err := s.plantypeRuleRepo.GetPlanTypeRuleByPlanTypeID(ctx, s.plantypeUC.FilterHighestPlanType(plantypes).ID.String())
	if err != nil {
		return fmt.Errorf("failed to get plan type rules: %v", err)
	}

	maxProductPerCategory := s.plantypeRuleUC.GetMaxProductPerCategoryFromRules(rules)

	return s.txManager.Do(ctx, func(txCtx context.Context) error {
		txProductCategoryRepo := s.productCategoryRepo.WithTx(txCtx)
		txCategoryRepo := s.categoryRepo.WithTx(txCtx)

		if err := txProductCategoryRepo.DeleteProductCategoryByProductID(ctx, productID); err != nil {
			return fmt.Errorf("failed to delete existing product categories: %v", err)
		}

		for _, catID := range categories {
			categoryUUID, err := uuid.Parse(catID)
			if err != nil {
				return fmt.Errorf("invalid category ID: %v", err)
			}
			cat, err := txCategoryRepo.GetCategoryByID(ctx, categoryUUID, &category.Filter{StoreID: &storeID})
			if err != nil {
				return fmt.Errorf("failed to get category by ID: %v", err)
			}
			if cat == nil {
				return utils.Error(http.StatusNotFound, fmt.Sprintf("kategori %s tidak ditemukan", catID))
			}

			if maxProductPerCategory > 0 {
				products, err := txProductCategoryRepo.GetProductCategoriesByCategoryID(ctx, categoryUUID)
				if err != nil {
					return fmt.Errorf("failed to get product categories by category ID: %v", err)
				}
				if len(products) >= maxProductPerCategory {
					return utils.Error(http.StatusBadRequest, productcategory.MsgMaxProductPerCategoryReached)
				}
			}

			newCategories := &productcategory.ProductCategory{
				ProductID:  productID,
				CategoryID: categoryUUID,
				IsActive:   true,
			}
			if err := txProductCategoryRepo.CreateProductCategory(txCtx, newCategories); err != nil {
				return fmt.Errorf("failed to create product category: %v", err)
			}
		}

		s.invalidateProductCache(ctx, storeID)

		return nil
	})

}

func (s *applicationService) DeleteProductByID(ctx context.Context, productID uuid.UUID) error {
	storeIDStr := ctx.Value(constants.ContextKeyStoreID).(string)
	storeID, _ := uuid.Parse(storeIDStr)

	prod, err := s.productRepo.GetProductByID(ctx, productID, &product.Filter{StoreID: &storeID})
	if err != nil {
		return fmt.Errorf("failed to get product by ID: %v", err)
	}
	if prod == nil {
		return utils.Error(http.StatusNotFound, product.ErrProductNotFound)
	}

	err = s.txManager.Do(ctx, func(txCtx context.Context) error {
		txProductCategoryRepo := s.productCategoryRepo.WithTx(txCtx)
		txProductImageRepo := s.productImageRepo.WithTx(txCtx)
		txProductRepo := s.productRepo.WithTx(txCtx)

		if err := txProductCategoryRepo.DeleteProductCategoryByProductID(ctx, productID); err != nil {
			return fmt.Errorf("failed to get product categories by product ID: %v", err)
		}

		if err := txProductImageRepo.DeleteProductImagesByProductID(ctx, productID); err != nil {
			return fmt.Errorf("failed to get product images by product ID: %v", err)
		}

		if err := txProductRepo.DeleteProduct(ctx, productID); err != nil {
			return fmt.Errorf("failed to delete product: %v", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	// Invalidate cache after successful deletion
	s.invalidateProductCache(ctx, storeID)

	return nil
}

func (s *applicationService) DeleteProductCategoryByProductID(ctx context.Context, productID uuid.UUID) error {
	storeID := uuid.MustParse(ctx.Value(constants.ContextKeyStoreID).(string))

	prod, err := s.productRepo.GetProductByID(ctx, productID, &product.Filter{StoreID: &storeID})
	if err != nil {
		return fmt.Errorf("failed to get product by ID: %v", err)
	}
	if prod == nil {
		return utils.Error(http.StatusNotFound, constants.MsgUnauthorizedAccess)
	}

	if err := s.productCategoryRepo.DeleteProductCategoryByProductID(ctx, productID); err != nil {
		return fmt.Errorf("failed to delete product categories by product ID: %v", err)
	}

	return nil
}

func (s *applicationService) DeleteProductCategoryByID(ctx context.Context, productID, categoryID uuid.UUID) error {
	storeID := uuid.MustParse(ctx.Value(constants.ContextKeyStoreID).(string))

	prodCat, err := s.productCategoryRepo.GetProductCategoryByID(ctx, productID, categoryID)
	if err != nil {
		return fmt.Errorf("failed to get product category by ID: %v", err)
	}
	if prodCat == nil {
		return utils.Error(http.StatusNotFound, productcategory.MsgProductCategoryNotFound)
	}

	prod, err := s.productRepo.GetProductByID(ctx, productID, &product.Filter{StoreID: &storeID})
	if err != nil {
		return fmt.Errorf("failed to get product by ID: %v", err)
	}
	if prod == nil {
		return utils.Error(http.StatusNotFound, constants.MsgUnauthorizedAccess)
	}

	if err := s.productCategoryRepo.DeleteProductCategoryByID(ctx, productID, categoryID); err != nil {
		return fmt.Errorf("failed to delete product category by ID: %v", err)
	}

	return nil
}

func (s *applicationService) MarkImagesAsPrimary(ctx context.Context, imageID uuid.UUID) (bool, error) {
	storeID := uuid.MustParse(ctx.Value(constants.ContextKeyStoreID).(string))

	image, err := s.productImageRepo.GetProductImageByID(ctx, imageID)
	if err != nil {
		return false, fmt.Errorf("failed to get product image by ID: %v", err)
	}
	if image == nil {
		return false, utils.Error(http.StatusNotFound, productimage.MsgImageNotFound)
	}

	prod, err := s.productRepo.GetProductByID(ctx, image.ProductID, &product.Filter{StoreID: &storeID})
	if err != nil {
		return false, fmt.Errorf("failed to get product by ID: %v", err)
	}
	if prod == nil {
		return false, utils.Error(http.StatusForbidden, constants.MsgUnauthorizedAccess)
	}

	err = s.txManager.Do(ctx, func(txCtx context.Context) error {
		productImageRepoTx := s.productImageRepo.WithTx(txCtx)
		productRepoTx := s.productRepo.WithTx(txCtx)

		// Reset all images for this product to not be primary
		if err := productImageRepoTx.ResetPrimaryImage(txCtx, image.ProductID); err != nil {
			return fmt.Errorf("failed to reset primary images: %v", err)
		}

		// Mark the selected image as primary
		image.IsPrimary = true
		if err := productImageRepoTx.UpdateProductImage(txCtx, image); err != nil {
			return fmt.Errorf("failed to mark image as primary: %v", err)
		}

		prod.Images = &image.FileURL
		if err := productRepoTx.UpdateProduct(txCtx, prod); err != nil {
			return fmt.Errorf("failed to update product primary image: %v", err)
		}

		return nil
	})

	if err != nil {
		return false, err
	}

	return true, nil
}

func (s *applicationService) CreateProductImages(ctx context.Context, productID uuid.UUID, images []productimage.ProductImage) error {
	storeID := uuid.MustParse(ctx.Value(constants.ContextKeyStoreID).(string))

	prod, err := s.productRepo.GetProductByID(ctx, productID, &product.Filter{StoreID: &storeID})
	if err != nil {
		return fmt.Errorf("failed to get product by id: %v", err)
	}
	if prod == nil {
		return utils.Error(http.StatusNotFound, product.ErrProductNotFound)
	}

	plantypes, err := s.plantypeRepo.GetStoreActiveSubsciption(ctx, storeID)
	if err != nil {
		return fmt.Errorf("failed to get store active subscription: %v", err)
	}

	rules, err := s.plantypeRuleRepo.GetPlanTypeRuleByPlanTypeID(ctx, s.plantypeUC.FilterHighestPlanType(plantypes).ID.String())
	if err != nil {
		return fmt.Errorf("failed to get plan type rules: %v", err)
	}

	existingImagesPerProduct, err := s.productImageRepo.GetProductImagesByProductID(ctx, productID)
	if err != nil {
		return fmt.Errorf("failed to get product images by product ID: %v", err)
	}

	maxAllowedImagesPerProduct := s.plantypeRuleUC.GetMaxImagePerProductFromRules(rules)

	if maxAllowedImagesPerProduct > 0 && len(existingImagesPerProduct)+len(images) > maxAllowedImagesPerProduct {
		return utils.Error(http.StatusBadRequest, productimage.MsgMaxImagesPerProductReached)
	}

	existingTotalImages, err := s.productImageRepo.GetProductImageByStoreID(ctx, storeID)
	if err != nil {
		return fmt.Errorf("failed to get product images by store ID: %v", err)
	}

	maxAllowedImages := s.plantypeRuleUC.GetMaxImagesFromRules(rules)
	if maxAllowedImages > 0 && len(existingTotalImages)+len(images) > maxAllowedImages {
		return utils.Error(http.StatusBadRequest, productimage.MsgMaxImagesReached)
	}

	return s.txManager.Do(ctx, func(txCtx context.Context) error {
		productImageRepoWithTx := s.productImageRepo.WithTx(txCtx)
		for _, img := range images {
			if err := productImageRepoWithTx.CreateProductImage(txCtx, &productimage.ProductImage{
				ProductID:   productID,
				FileType:    img.FileType,
				FileStorage: img.FileStorage,
				FileName:    img.FileName,
				FilePath:    img.FilePath,
				FileURL:     img.FileURL,
				IsPrimary:   false,
				IsActive:    true,
			}); err != nil {
				return fmt.Errorf("failed to create product image: %v", err)
			}
		}
		return nil
	})
}

func (s *applicationService) DeleteProductImageByID(ctx context.Context, imageID uuid.UUID) error {
	storeID := uuid.MustParse(ctx.Value(constants.ContextKeyStoreID).(string))

	image, err := s.productImageRepo.GetProductImageByID(ctx, imageID)
	if err != nil {
		return fmt.Errorf("failed to get product image by ID: %v", err)
	}
	if image == nil {
		return utils.Error(http.StatusNotFound, productimage.MsgImageNotFound)
	}

	prod, err := s.productRepo.GetProductByID(ctx, image.ProductID, &product.Filter{StoreID: &storeID})
	if err != nil {
		return fmt.Errorf("failed to get product by ID: %v", err)
	}
	if prod == nil {
		return utils.Error(http.StatusNotFound, product.ErrProductNotFound)
	}

	if err := s.productImageRepo.DeleteProductImageByID(ctx, imageID); err != nil {
		return fmt.Errorf("failed to delete product image by ID: %v", err)
	}

	return nil
}

func (s *applicationService) DeleteProductImagesByProductID(ctx context.Context, productID uuid.UUID) error {
	storeIDStr := ctx.Value(constants.ContextKeyStoreID).(string)
	storeID, _ := uuid.Parse(storeIDStr)

	prod, err := s.productRepo.GetProductByID(ctx, productID, &product.Filter{StoreID: &storeID})
	if err != nil {
		return fmt.Errorf("failed to get product by ID: %v", err)
	}
	if prod == nil {
		return utils.Error(http.StatusNotFound, product.ErrProductNotFound)
	}

	return s.txManager.Do(ctx, func(txCtx context.Context) error {
		productImageRepoWithTx := s.productImageRepo.WithTx(txCtx)
		productRepoWithTx := s.productRepo.WithTx(txCtx)

		if err := productImageRepoWithTx.DeleteProductImagesByProductID(ctx, productID); err != nil {
			return fmt.Errorf("failed to delete product images by product ID: %v", err)
		}

		prod.Images = nil
		if err := productRepoWithTx.UpdateProduct(txCtx, prod); err != nil {
			return fmt.Errorf("failed to update product image: %v", err)
		}

		// TODO delete exising image from file storage (S3)

		return nil
	})
}

func (s *applicationService) UpdateToggleActive(ctx context.Context, productID uuid.UUID) (bool, error) {
	storeIDStr := ctx.Value(constants.ContextKeyStoreID).(string)
	storeID, _ := uuid.Parse(storeIDStr)

	prod, err := s.productRepo.GetProductByID(ctx, productID, &product.Filter{StoreID: &storeID})
	if err != nil {
		return false, fmt.Errorf("failed to get product by ID: %v", err)
	}
	if prod == nil {
		return false, utils.Error(http.StatusNotFound, product.ErrProductNotFound)
	}

	prod.IsActive = !prod.IsActive
	if !prod.IsActive {
		if err := s.productRepo.UpdateToggleActive(ctx, prod.ID, prod.IsActive, ctx.Value(constants.ContextKeyUserID).(string)); err != nil {
			return false, fmt.Errorf("failed to update product toggle active: %v", err)
		}
		s.invalidateProductCache(ctx, storeID)
		return prod.IsActive, nil
	}

	planTypes, err := s.plantypeRepo.GetStoreActiveSubsciption(ctx, storeID)
	if err != nil {
		return false, fmt.Errorf("failed to get store active subscription: %v", err)
	}

	rules, err := s.plantypeRuleRepo.GetPlanTypeRuleByPlanTypeID(ctx, s.plantypeUC.FilterHighestPlanType(planTypes).ID.String())
	if err != nil {
		return false, fmt.Errorf("failed to get plan type rules: %v", err)
	}

	existingActiveProducts, err := s.productRepo.GetProductList(ctx, &product.Filter{StoreID: &storeID, IsActive: &boolTrue})
	if err != nil {
		return false, fmt.Errorf("failed to get existing products: %v", err)
	}

	maxProducts := s.plantypeRuleUC.GetMaxProductFromRules(rules)
	if maxProducts > 0 && len(existingActiveProducts) >= maxProducts {
		return false, utils.Error(http.StatusBadRequest, product.MsgMaxActiveProductsReached)
	}

	if err := s.productRepo.UpdateToggleActive(ctx, prod.ID, prod.IsActive, ctx.Value(constants.ContextKeyUserID).(string)); err != nil {
		return false, fmt.Errorf("failed to update product toggle active: %v", err)
	}

	s.invalidateProductCache(ctx, storeID)

	return prod.IsActive, nil
}

func (s *applicationService) GetProductDetails(ctx context.Context, id uuid.UUID, filter product.Filter) (*ProductDetails, error) {
	prod, err := s.productRepo.GetProductByID(ctx, id, &filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get product by ID: %v", err)
	}
	if prod == nil {
		return nil, utils.Error(http.StatusNotFound, product.ErrProductNotFound)
	}

	categories, err := s.categoryRepo.GetCategoryByProductID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories by product ID: %v", err)
	}

	productImages, err := s.productImageRepo.GetProductImagesByProductID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get product images by product ID: %v", err)
	}

	return &ProductDetails{
		ID:          prod.ID,
		Name:        prod.Name,
		Desc:        prod.Description,
		Image:       prod.Images,
		IsAvailable: prod.IsAvailable,
		IsActive:    prod.IsActive,
		Price:       prod.Price.String(),
		Categories:  categories,
		Images:      productImages,
	}, nil
}

func (s *applicationService) GetProductListWithCategory(ctx context.Context, filter product.Filter) ([]ProductWithCategory, int64, error) {
	products, err := s.productRepo.GetProductListWithCategories(ctx, &filter)
	if err != nil {
		return nil, 0, err
	}

	result := make([]ProductWithCategory, len(products))
	for i, prod := range products {

		// Map CategoryData to category.Category
		categories := make([]category.Category, len(prod.Categories))
		for j, cat := range prod.Categories {
			categories[j] = category.Category{
				ID:   cat.ID,
				Name: cat.Name,
			}
		}

		result[i] = ProductWithCategory{
			ID:          prod.ID,
			Name:        prod.Name,
			Desc:        prod.Description,
			Price:       prod.Price.String(),
			Images:      prod.Images,
			IsAvailable: prod.IsActive, // Using IsActive as IsAvailable
			IsActive:    prod.IsActive,
			Categories:  categories,
		}
	}

	total, err := s.productRepo.CountProductListWithCategories(ctx, &filter)
	if err != nil {
		return nil, 0, err
	}

	return result, total, nil
}

func (s *applicationService) CreateProducts(ctx context.Context, products []Product) error {

	storeIDStr := ctx.Value(constants.ContextKeyStoreID).(string)
	storeID, _ := uuid.Parse(storeIDStr)

	if err := s.validateInputCategories(ctx, storeID, products); err != nil {
		return err
	}

	planTypes, err := s.plantypeRepo.GetStoreActiveSubsciption(ctx, storeID)
	if err != nil {
		return fmt.Errorf("failed to get store active subscription: %v", err)
	}

	rules, err := s.plantypeRuleRepo.GetPlanTypeRuleByPlanTypeID(ctx, s.plantypeUC.FilterHighestPlanType(planTypes).ID.String())
	if err != nil {
		return fmt.Errorf("failed to get plan type rules: %v", err)
	}

	existingActiveProducts, err := s.productRepo.GetProductList(ctx, &product.Filter{StoreID: &storeID, IsActive: &boolTrue})
	if err != nil {
		return fmt.Errorf("failed to get existing products: %v", err)
	}

	productImages, err := s.productImageRepo.GetProductImageByStoreID(ctx, storeID)
	if err != nil {
		return fmt.Errorf("failed to get product images: %v", err)
	}

	activeProductsTotal := len(existingActiveProducts)
	maxAllowedProducts := s.plantypeRuleUC.GetMaxProductFromRules(rules)
	maxAllowedProductPerCategory := s.plantypeRuleUC.GetMaxProductPerCategoryFromRules(rules)
	imageTotal := len(productImages)
	maxAllowedImages := s.plantypeRuleUC.GetMaxImagesFromRules(rules)
	maxAllowedImagesPerProduct := s.plantypeRuleUC.GetMaxImagePerProductFromRules(rules)

	return s.txManager.Do(ctx, func(txCtx context.Context) error {
		txProductRepo := s.productRepo.WithTx(txCtx)
		txProductCategoryRepo := s.productCategoryRepo.WithTx(txCtx)
		txProductImageRepo := s.productImageRepo.WithTx(txCtx)

		for _, prod := range products {
			// process product
			isActiveProduct := activeProductsTotal == 0 || activeProductsTotal < maxAllowedProducts
			if isActiveProduct {
				activeProductsTotal++ // Increment count for next iteration
			}
			productID := uuid.New()
			newProduct := &product.Product{
				ID:          productID,
				Name:        prod.Name,
				Description: prod.Desc,
				Price:       prod.Price,
				StoreID:     storeID,
				IsActive:    isActiveProduct,
				Images:      getPrimaryImages(prod.Images),
				IsAvailable: true,
			}
			if err := txProductRepo.CreateProduct(txCtx, newProduct); err != nil {
				return fmt.Errorf("failed to create product: %v", err)
			}

			// process product categories
			for _, catID := range prod.Categories {
				productCategories, err := s.productCategoryRepo.GetProductCategoriesByCategoryID(ctx, uuid.MustParse(catID))
				if err != nil {
					return fmt.Errorf("failed to get product categories by category ID: %v", err)
				}
				if maxAllowedProductPerCategory > 0 && len(productCategories) >= maxAllowedProductPerCategory {
					return utils.Error(http.StatusBadRequest, productcategory.MsgMaxProductPerCategoryReached)
				}

				newCategories := &productcategory.ProductCategory{
					ProductID:  productID,
					CategoryID: uuid.MustParse(catID),
					IsActive:   true,
				}
				if err := txProductCategoryRepo.CreateProductCategory(txCtx, newCategories); err != nil {
					return fmt.Errorf("failed to create product category: %v", err)
				}

			}

			// process product images
			if maxAllowedImagesPerProduct > 0 && len(prod.Images) > maxAllowedImagesPerProduct {
				return utils.Error(http.StatusBadRequest, productimage.MsgMaxImagesPerProductReached)
			}
			for _, img := range prod.Images {
				if maxAllowedImages > 0 && imageTotal >= maxAllowedImages {
					return utils.Error(http.StatusBadRequest, productimage.MsgMaxImagesReached)
				}
				newImage := &productimage.ProductImage{
					ProductID:   productID,
					FileType:    img.FileType,
					FileStorage: img.FileStorage,
					FileName:    img.FileName,
					FilePath:    img.FilePath,
					FileURL:     img.FileURL,
					IsActive:    true,
				}
				if err := txProductImageRepo.CreateProductImage(txCtx, newImage); err != nil {
					return fmt.Errorf("failed to create product image: %v", err)
				}
				imageTotal++ // Increment count for next iteration

				if img.IsPrimary {
					newProduct.Images = &img.FileURL
				}
			}
		}

		s.invalidateProductCache(ctx, storeID)

		return nil
	})
}

func (s *applicationService) validateInputCategories(ctx context.Context, storeID uuid.UUID, products []Product) error {
	for _, prod := range products {
		for _, catID := range prod.Categories {
			catID, err := uuid.Parse(catID)
			if err != nil {
				return utils.Error(http.StatusBadRequest, "invalid category")
			}
			cat, err := s.categoryRepo.GetCategoryByID(ctx, catID, &category.Filter{StoreID: &storeID})
			if err != nil {
				return fmt.Errorf("failed to get category by ID: %v", err)
			}
			if cat == nil {
				return utils.Error(http.StatusBadRequest, "invalid category")
			}
		}
	}
	return nil
}

func getPrimaryImages(images []productimage.ProductImage) *string {
	for _, img := range images {
		if img.IsPrimary {
			return &img.FileURL
		}
	}
	return nil
}

func buildSubscriptionRules(rules []plantyperule.PlanTypeRule) *SubscriptionRule {
	subscriptionRules := &SubscriptionRule{}
	for _, rule := range rules {
		switch rule.Rule {
		case constants.LimitCategory:
			subscriptionRules.LimitCategory = rule.RuleValue
		case constants.LimitProduct:
			subscriptionRules.LimitProduct = rule.RuleValue
		case constants.LimitProductPerCategory:
			subscriptionRules.LimitProductPerCategory = rule.RuleValue
		case constants.LimitImage:
			subscriptionRules.LimitImage = rule.RuleValue
		case constants.LimitImagePerProduct:
			subscriptionRules.LimitImagePerProduct = rule.RuleValue
		case constants.LimitBanner:
			subscriptionRules.LimitBanner = rule.RuleValue
		case constants.DirectChatByWa:
			subscriptionRules.DirectChatByWa = rule.RuleValue
		case constants.OrderService:
			subscriptionRules.OrderService = rule.RuleValue
		case constants.PaymentService:
			subscriptionRules.PaymentService = rule.RuleValue
		case constants.TableService:
			subscriptionRules.TableService = rule.RuleValue
		}
	}
	return subscriptionRules
}
