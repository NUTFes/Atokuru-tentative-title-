package repository

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/domain"
)

// ErrDuplicateCommentTemplate は、同じ屋台に同じ定型文がすでにあることを表す。
var ErrDuplicateCommentTemplate = errors.New("同じ定型文がすでにあります")

// CommentTemplateRepository は comment_templates テーブルを読み書きする。
type CommentTemplateRepository struct {
	db *gorm.DB
}

func NewCommentTemplateRepository(db *gorm.DB) *CommentTemplateRepository {
	return &CommentTemplateRepository{db: db}
}

// ListByStall は屋台の定型文を、登録した順に取る。1店あたりの数には上限があるので、件数は絞らない。
func (r *CommentTemplateRepository) ListByStall(ctx context.Context, stallID uint) ([]domain.CommentTemplate, error) {
	var ts []domain.CommentTemplate
	err := r.db.WithContext(ctx).Where("stall_id = ?", stallID).Order("id").Find(&ts).Error
	return ts, err
}

// CountByStall は屋台の定型文の数を返す。
func (r *CommentTemplateRepository) CountByStall(ctx context.Context, stallID uint) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&domain.CommentTemplate{}).Where("stall_id = ?", stallID).Count(&n).Error
	return n, err
}

// FindByComment は、屋台の定型文のうち文がまったく同じものを取る。ないときは domain.ErrNotFound を返す。
// (stall_id, comment) のユニークインデックスで1行だけ読む。
func (r *CommentTemplateRepository) FindByComment(ctx context.Context, stallID uint, comment string) (*domain.CommentTemplate, error) {
	var t domain.CommentTemplate
	err := r.db.WithContext(ctx).Where("stall_id = ? AND comment = ?", stallID, comment).First(&t).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// Create は定型文を1件作る。同じ屋台に同じ文があるときは ErrDuplicateCommentTemplate を返す。
func (r *CommentTemplateRepository) Create(ctx context.Context, t *domain.CommentTemplate) error {
	err := r.db.WithContext(ctx).Omit(clause.Associations).Create(t).Error
	if err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed") {
		return ErrDuplicateCommentTemplate
	}
	return err
}

// Delete は定型文を1件消す。ないときは domain.ErrNotFound を返す。
func (r *CommentTemplateRepository) Delete(ctx context.Context, id uint) error {
	res := r.db.WithContext(ctx).Delete(&domain.CommentTemplate{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}
