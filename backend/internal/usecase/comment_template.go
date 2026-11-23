package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/domain"
	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/repository"
)

// Translator は日本語の文を、指定した言語に訳す。実装は infra/translate にあり、main.go で渡す。
// 戻り値のキーは言語コード。訳せなかった言語は入れなくてよい。
type Translator interface {
	Translate(ctx context.Context, text string, langs []string) (domain.Translations, error)
}

// translateLanguages は定型文を訳す先の言語（FE の locales/ にある、日本語以外の言語と合わせる）。
var translateLanguages = []string{"en"}

// maxTemplatesPerStall は1店あたりの定型文の上限（仕様書と合わせる）。発券の画面にボタンとして並べられる数にする。
const maxTemplatesPerStall = 30

// translateTimeout は、定型文1件を訳すのを待つ上限の時間。過ぎたら訳なしで登録する。
const translateTimeout = 15 * time.Second

// CommentTemplateUsecase は備考の定型文についての処理をまとめる。
type CommentTemplateUsecase struct {
	stalls    *repository.StallRepository
	templates *repository.CommentTemplateRepository
	// translator は定型文を訳す。nil なら訳さない（翻訳の API キーを設定していないとき）。
	translator Translator
}

// NewCommentTemplateUsecase を作る。translator が nil なら訳さずに登録する。
func NewCommentTemplateUsecase(stalls *repository.StallRepository, templates *repository.CommentTemplateRepository, translator Translator) *CommentTemplateUsecase {
	return &CommentTemplateUsecase{stalls: stalls, templates: templates, translator: translator}
}

// List は屋台の定型文を、登録した順に返す。
func (u *CommentTemplateUsecase) List(ctx context.Context, stallID uint) ([]domain.CommentTemplate, error) {
	if _, err := findStall(ctx, u.stalls, stallID); err != nil {
		return nil, err
	}
	return u.templates.ListByStall(ctx, stallID)
}

// Create は定型文を登録する。登録する前に一度だけ訳し、訳も一緒に保存する。
// 訳せなかったときも、定型文の登録は止めずに訳なしで保存する（店の作業を止めないため）。
func (u *CommentTemplateUsecase) Create(ctx context.Context, stallID uint, comment string) (*domain.CommentTemplate, error) {
	comment = strings.TrimSpace(comment)
	if comment == "" {
		return nil, domain.InvalidInput("定型文を入力してください")
	}
	if utf8.RuneCountInString(comment) > maxCommentLength {
		return nil, domain.InvalidInput(fmt.Sprintf("定型文は %d 文字以内にしてください", maxCommentLength))
	}
	if _, err := findStall(ctx, u.stalls, stallID); err != nil {
		return nil, err
	}

	// 訳すのには時間とお金がかかるので、登録できないものは先に弾く。
	n, err := u.templates.CountByStall(ctx, stallID)
	if err != nil {
		return nil, err
	}
	if n >= maxTemplatesPerStall {
		return nil, domain.Conflict(fmt.Sprintf("定型文は1店あたり %d 件までです", maxTemplatesPerStall))
	}
	if _, err := u.templates.FindByComment(ctx, stallID, comment); err == nil {
		return nil, domain.Conflict("同じ定型文がすでにあります")
	} else if !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}

	t := &domain.CommentTemplate{StallID: stallID, Comment: comment, Translations: u.translate(ctx, comment)}
	if err := u.templates.Create(ctx, t); err != nil {
		if errors.Is(err, repository.ErrDuplicateCommentTemplate) {
			return nil, domain.Conflict("同じ定型文がすでにあります")
		}
		return nil, err
	}
	return t, nil
}

// translate は定型文を訳す。訳せなかったときはログに出して空を返す。
func (u *CommentTemplateUsecase) translate(ctx context.Context, comment string) domain.Translations {
	if u.translator == nil {
		return domain.Translations{}
	}
	ctx, cancel := context.WithTimeout(ctx, translateTimeout)
	defer cancel()
	tr, err := u.translator.Translate(ctx, comment, translateLanguages)
	if err != nil {
		log.Printf("定型文を訳せませんでした（訳なしで登録します）: %v", err)
		return domain.Translations{}
	}

	// 頼んだ言語の、空でなく長すぎない訳だけを残す。
	out := domain.Translations{}
	for _, lang := range translateLanguages {
		s := strings.TrimSpace(tr[lang])
		if s != "" && utf8.RuneCountInString(s) <= maxTranslationLength {
			out[lang] = s
		}
	}
	return out
}

// maxTranslationLength は訳1つの上限の文字数。日本語より長くなるので、備考の上限の2倍にする。
const maxTranslationLength = maxCommentLength * 2

// Delete は定型文を消す。発券済みの整理券の備考と訳はそのまま残る。
func (u *CommentTemplateUsecase) Delete(ctx context.Context, id uint) error {
	err := u.templates.Delete(ctx, id)
	if errors.Is(err, domain.ErrNotFound) {
		return domain.NotFound("定型文が見つかりません")
	}
	return err
}
