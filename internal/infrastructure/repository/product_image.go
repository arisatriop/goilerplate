package repository

import (
	"context"
	"goilerplate/internal/domain/productimage"
	"goilerplate/internal/infrastructure/model"
	"goilerplate/internal/infrastructure/transaction"
	"goilerplate/pkg/utils"
	"net/http"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type productImageRepo struct {
	db *gorm.DB
}

func NewProductImage(db *gorm.DB) productimage.Repository {
	return &productImageRepo{
		db: db,
	}
}

func (r *productImageRepo) WithTx(ctx context.Context) productimage.Repository {
	tx := transaction.GetTxFromContext(ctx)
	if tx != nil {
		return NewProductImage(tx)
	}
	return r
}

func (r *productImageRepo) ResetPrimaryImage(ctx context.Context, productID uuid.UUID) error {
	if err := r.db.WithContext(ctx).
		Model(&model.ProductImage{}).
		Where("product_id = ?", productID).
		Update("is_primary", false).Error; err != nil {
		return err
	}
	return nil
}

func (r *productImageRepo) UpdateProductImage(ctx context.Context, pi *productimage.ProductImage) error {
	var img model.ProductImage
	if err := r.db.WithContext(ctx).Where("id = ?", pi.ID).First(&img).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.Error(http.StatusNotFound, productimage.MsgImageNotFound)
		}
		return err
	}

	img.FileType = pi.FileType
	img.FileStorage = pi.FileStorage
	img.FileName = pi.FileName
	img.FilePath = pi.FilePath
	img.FileURL = pi.FileURL
	img.IsPrimary = pi.IsPrimary
	img.IsActive = pi.IsActive

	if err := r.db.WithContext(ctx).Save(&img).Error; err != nil {
		return err
	}

	return nil
}

func (r *productImageRepo) GetProductImageByID(ctx context.Context, id uuid.UUID) (*productimage.ProductImage, error) {
	var m model.ProductImage
	if err := r.db.WithContext(ctx).
		Select("id", "product_id", "file_type", "file_storage", "file_name", "file_path", "file_url", "is_primary", "is_active").
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	entity := r.toDomainEntity(&m)
	return &entity, nil
}

func (r *productImageRepo) DeleteProductImagesByProductID(ctx context.Context, productID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Where("product_id = ?", productID).Delete(&model.ProductImage{}).Error; err != nil {
		return err
	}
	return nil
}

func (r *productImageRepo) DeleteProductImageByID(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.ProductImage{}).Error; err != nil {
		return err
	}
	return nil
}

func (r *productImageRepo) GetProductImagesByProductID(ctx context.Context, productID uuid.UUID) ([]productimage.ProductImage, error) {
	var productImages []model.ProductImage
	if err := r.db.WithContext(ctx).
		Where("product_id = ?", productID).
		Where("deleted_at IS NULL").
		Find(&productImages).Error; err != nil {
		return nil, err
	}

	return r.toDomainEntities(productImages), nil
}

func (r *productImageRepo) GetProductImageByStoreID(ctx context.Context, storeID uuid.UUID) ([]productimage.ProductImage, error) {
	var productImages []model.ProductImage
	if err := r.db.WithContext(ctx).
		Joins("JOIN products ON products.id = product_images.product_id").
		Where("products.store_id = ?", storeID).
		Where("products.deleted_at IS NULL").
		Where("product_images.deleted_at IS NULL").
		Find(&productImages).Error; err != nil {
		return nil, err
	}

	return r.toDomainEntities(productImages), nil
}

func (r *productImageRepo) CreateProductImage(ctx context.Context, pi *productimage.ProductImage) error {
	model := &model.ProductImage{
		ProductID:   pi.ProductID,
		FileType:    pi.FileType,
		FileStorage: pi.FileStorage,
		FileName:    pi.FileName,
		FilePath:    pi.FilePath,
		FileURL:     pi.FileURL,
		IsPrimary:   pi.IsPrimary,
		IsActive:    pi.IsActive,
	}

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}

	return nil
}

func (r *productImageRepo) toDomainEntities(models []model.ProductImage) []productimage.ProductImage {
	entities := make([]productimage.ProductImage, len(models))
	for i, m := range models {
		entities[i] = r.toDomainEntity(&m)
	}
	return entities
}

func (r *productImageRepo) toDomainEntity(m *model.ProductImage) productimage.ProductImage {
	return productimage.ProductImage{
		ID:          m.ID,
		ProductID:   m.ProductID,
		FileType:    m.FileType,
		FileStorage: m.FileStorage,
		FileName:    m.FileName,
		FilePath:    m.FilePath,
		FileURL:     m.FileURL,
		IsPrimary:   m.IsPrimary,
		IsActive:    m.IsActive,
	}
}
