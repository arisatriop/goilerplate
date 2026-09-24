package repository

import (
	"context"
	"errors"
	"fmt"

	"goilerplate/internal/domain/bar"
	"goilerplate/internal/infrastructure/model"
	"goilerplate/internal/infrastructure/transaction"
	"goilerplate/pkg/constants"
	"goilerplate/pkg/utils"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

// uniqueBarCodeConstraint keeps a code unique across every bar, soft-deleted ones included:
// code is the business key, so it is never reassigned.
const uniqueBarCodeConstraint = "uq_bars_code"

// translateWriteErr turns a violation of the code constraint into the domain's conflict error.
// The index is what guarantees uniqueness — a check before the insert cannot, because two
// concurrent requests can both pass it — so this is where a duplicate becomes a 409 rather
// than a 500.
func translateWriteErr(op string, err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == uniqueBarCodeConstraint {
		return bar.ErrCodeAlreadyExists
	}
	return fmt.Errorf("%s: %w", op, err)
}

type barRepo struct {
	db *gorm.DB
}

func NewBar(db *gorm.DB) bar.Repository {
	return &barRepo{
		db: db,
	}
}

func (r *barRepo) WithTx(ctx context.Context) bar.Repository {
	tx := transaction.GetTxFromContext(ctx)
	if tx != nil {
		return &barRepo{db: tx}
	}
	return r
}

func (r *barRepo) CreateBar(ctx context.Context, entity *bar.Bar) (*bar.Bar, error) {
	now := utils.Now()
	user := ctx.Value(constants.ContextKeyUserID).(string)
	model := &model.Bar{
		ID:        utils.GenerateUUID(),
		Code:      entity.Code,
		Bar:       entity.Bar,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
		CreatedBy: user,
		UpdatedBy: user,
	}

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return nil, translateWriteErr("inserting bar", err)
	}

	return r.modelToEntity(model), nil
}

func (r *barRepo) UpdateBar(ctx context.Context, entity *bar.Bar) error {
	model, err := r.getBarByID(ctx, entity.ID)
	if err != nil {
		return err
	}

	// Code is not written: it is the business key and never changes once assigned.
	model.Bar = entity.Bar
	model.UpdatedAt = utils.Now()
	model.UpdatedBy = ctx.Value(constants.ContextKeyUserID).(string)

	if err = r.db.WithContext(ctx).Save(model).Error; err != nil {
		return fmt.Errorf("updating bar: %w", err)
	}

	return nil
}

func (r *barRepo) DeleteBar(ctx context.Context, entity *bar.Bar) error {
	model, err := r.getBarByID(ctx, entity.ID)
	if err != nil {
		return err
	}

	now := utils.Now()
	user := ctx.Value(constants.ContextKeyUserID).(string)
	model.DeletedAt = &now
	model.DeletedBy = &user

	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		return fmt.Errorf("deleting bar: %w", err)
	}

	return nil
}

func (r *barRepo) GetBarByID(ctx context.Context, id string) (*bar.Bar, error) {

	model, err := r.getBarByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return r.modelToEntity(model), nil
}

func (r *barRepo) GetBarList(ctx context.Context, filter *bar.Filter) ([]*bar.Bar, error) {
	var models []model.Bar

	query := r.db.WithContext(ctx).
		Select("id", "code", "bar").
		Where("deleted_at IS NULL").
		// Offset pagination needs a total order, or rows repeat or vanish between pages. IDs are
		// UUIDv7, so this is newest first, and the primary key index serves it.
		Order("id DESC")

	r.applyBarFilters(query, filter, true) // true = apply pagination

	err := query.Find(&models).Error
	if err != nil {
		return nil, fmt.Errorf("listing bars: %w", err)
	}

	entities := make([]*bar.Bar, len(models))
	for i, model := range models {
		entities[i] = r.modelToEntity(&model)
	}

	return entities, nil
}

func (r *barRepo) CountBar(ctx context.Context, filter *bar.Filter) (int64, error) {
	var count int64

	query := r.db.WithContext(ctx).
		Model(&model.Bar{}).
		Where("deleted_at IS NULL")

	r.applyBarFilters(query, filter, false) // false = don't apply pagination

	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("counting bars: %w", err)
	}

	return count, nil
}

func (r *barRepo) BulkCreate(ctx context.Context, entities []*bar.Bar) error {
	if len(entities) == 0 {
		return nil
	}

	now := utils.Now()
	user := ctx.Value(constants.ContextKeyUserID).(string) // Use proper context key

	models := make([]model.Bar, len(entities))
	for i, entity := range entities {
		models[i] = model.Bar{
			ID:        utils.GenerateUUID(),
			Code:      entity.Code,
			Bar:       entity.Bar,
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
			CreatedBy: user,
			UpdatedBy: user,
		}
	}

	// One INSERT statement, so a duplicate anywhere in the batch rejects all of it.
	if err := r.db.WithContext(ctx).Create(&models).Error; err != nil {
		return translateWriteErr("inserting bars", err)
	}

	return nil
}

func (r *barRepo) getBarByID(ctx context.Context, id string) (*model.Bar, error) {

	var data model.Bar

	err := r.db.WithContext(ctx).
		Where("id = ? and deleted_at IS NULL", id).
		First(&data).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, bar.ErrNotFound
		}
		return nil, fmt.Errorf("loading bar: %w", err)
	}

	return &data, nil
}

func (r *barRepo) applyBarFilters(query *gorm.DB, filter *bar.Filter, applyPagination bool) {
	if filter == nil {
		return
	}

	if filter.Keyword != "" {
		keyword := "%" + filter.Keyword + "%"
		query.Where("code ILIKE ? OR bar ILIKE ?", keyword, keyword)
	}

	if filter.Code != "" {
		query.Where("code = ?", filter.Code)
	}

	if applyPagination && filter.Pagination != nil {
		query.Offset(filter.Pagination.GetOffset()).Limit(filter.Pagination.GetLimit())
	}
}

func (r *barRepo) modelToEntity(model *model.Bar) *bar.Bar {
	return &bar.Bar{
		ID:   model.ID,
		Code: model.Code,
		Bar:  model.Bar,
	}
}
