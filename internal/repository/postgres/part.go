package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/felipemaejima/backend-test/internal/domain"
)

type partModel struct {
	ID                uuid.UUID `gorm:"column:id;type:uuid;primaryKey"`
	Name              string    `gorm:"column:name;type:varchar(120);not null"`
	Category          string    `gorm:"column:category;type:varchar(60);not null;index"`
	CurrentStock      int       `gorm:"column:current_stock;not null"`
	MinimumStock      int       `gorm:"column:minimum_stock;not null"`
	AverageDailySales float64   `gorm:"column:average_daily_sales;type:numeric(12,4);not null"`
	LeadTimeDays      int       `gorm:"column:lead_time_days;not null"`
	UnitCost          float64   `gorm:"column:unit_cost;type:numeric(12,2);not null"`
	CriticalityLevel  int       `gorm:"column:criticality_level;not null"`
	CreatedAt         time.Time `gorm:"column:created_at;autoCreateTime:false;not null"`
	UpdatedAt         time.Time `gorm:"column:updated_at;autoUpdateTime:false;not null"`
}

func (partModel) TableName() string { return "parts" }

func toModel(part domain.Part) partModel {
	return partModel{
		ID:                part.ID,
		Name:              part.Name,
		Category:          part.Category,
		CurrentStock:      part.CurrentStock,
		MinimumStock:      part.MinimumStock,
		AverageDailySales: part.AverageDailySales,
		LeadTimeDays:      part.LeadTimeDays,
		UnitCost:          part.UnitCost,
		CriticalityLevel:  part.CriticalityLevel,
		CreatedAt:         part.CreatedAt,
		UpdatedAt:         part.UpdatedAt,
	}
}

func (m partModel) toDomain() domain.Part {
	return domain.Part{
		ID:                m.ID,
		Name:              m.Name,
		Category:          m.Category,
		CurrentStock:      m.CurrentStock,
		MinimumStock:      m.MinimumStock,
		AverageDailySales: m.AverageDailySales,
		LeadTimeDays:      m.LeadTimeDays,
		UnitCost:          m.UnitCost,
		CriticalityLevel:  m.CriticalityLevel,
		CreatedAt:         m.CreatedAt,
		UpdatedAt:         m.UpdatedAt,
	}
}

type PartRepository struct {
	db *gorm.DB
}

func NewPartRepository(db *gorm.DB) *PartRepository {
	return &PartRepository{db: db}
}

var _ domain.PartRepository = (*PartRepository)(nil)

func (r *PartRepository) Create(ctx context.Context, part *domain.Part) error {
	model := toModel(*part)
	return r.db.WithContext(ctx).Create(&model).Error
}

func (r *PartRepository) Update(ctx context.Context, part *domain.Part) error {
	model := toModel(*part)
	result := r.db.WithContext(ctx).Model(&partModel{}).Where("id = ?", model.ID).Updates(map[string]any{
		"name":                model.Name,
		"category":            model.Category,
		"current_stock":       model.CurrentStock,
		"minimum_stock":       model.MinimumStock,
		"average_daily_sales": model.AverageDailySales,
		"lead_time_days":      model.LeadTimeDays,
		"unit_cost":           model.UnitCost,
		"criticality_level":   model.CriticalityLevel,
		"updated_at":          model.UpdatedAt,
	})
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domain.ErrPartNotFound
	}
	return nil
}

func (r *PartRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&partModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrPartNotFound
	}
	return nil
}

func (r *PartRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Part, error) {
	var model partModel
	if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrPartNotFound
		}
		return nil, err
	}
	part := model.toDomain()
	return &part, nil
}

func (r *PartRepository) List(ctx context.Context, filter domain.PartFilter) ([]domain.Part, error) {
	query := r.db.WithContext(ctx).Model(&partModel{})
	if filter.Category != "" {
		query = query.Where("category = ?", filter.Category)
	}

	var models []partModel
	err := query.Order("name, id").Limit(filter.Limit).Offset(filter.Offset).Find(&models).Error
	if err != nil {
		return nil, err
	}

	parts := make([]domain.Part, 0, len(models))
	for _, model := range models {
		parts = append(parts, model.toDomain())
	}
	return parts, nil
}
