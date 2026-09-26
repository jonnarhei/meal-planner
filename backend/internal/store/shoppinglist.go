package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/jonnarhei/meal-planner/backend/internal/store/models"
)

type ShoppinglistStore struct {
	db *sql.DB
}

// one ($1, $2, ...) group per item, plus the arguments to bind to them
func buildInsertValues(items []models.ShoppinglistItem) (string, []any) {
	valueStrings := make([]string, len(items))
	valueArgs := make([]any, 0, len(items)*6)

	for i, item := range items {
		n := i * 6
		valueStrings[i] = fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d)", n+1, n+2, n+3, n+4, n+5, n+6)
		valueArgs = append(valueArgs, item.UserID, item.IngredientID, item.Name, item.Amount, item.Unit, item.Source)
	}

	return strings.Join(valueStrings, ","), valueArgs
}

func (s *ShoppinglistStore) AddItems(ctx context.Context, userID int64, items []models.ShoppinglistItem) error {
	if len(items) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	values, valueArgs := buildInsertValues(items)

	query := fmt.Sprintf(`
	INSERT INTO shopping_list_items (user_id, ingredient_id, name, amount, unit, source)
	VALUES %s
	ON CONFLICT (user_id, agg_key, unit)
	DO UPDATE SET amount = shopping_list_items.amount + EXCLUDED.amount, name = LEAST(shopping_list_items.name, EXCLUDED.name)
	`, values)

	_, err = tx.ExecContext(ctx, query, valueArgs...)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *ShoppinglistStore) GetAll(ctx context.Context, userID int64) ([]models.ShoppinglistItem, error) {
	query := `
	SELECT id, user_id, ingredient_id, name, amount, unit, checked, source, created_at
	FROM shopping_list_items
	WHERE user_id = $1
	`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.ShoppinglistItem

	for rows.Next() {
		var item models.ShoppinglistItem
		err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.IngredientID,
			&item.Name,
			&item.Amount,
			&item.Unit,
			&item.Checked,
			&item.Source,
			&item.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (s *ShoppinglistStore) ToggleChecked(ctx context.Context, itemID int64, userID int64) error {
	query := `
	UPDATE shopping_list_items
	SET checked = NOT checked
	WHERE id = $1 AND user_id = $2
	`

	_, err := s.db.ExecContext(ctx, query, itemID, userID)

	return err
}

func (s *ShoppinglistStore) DeleteItem(ctx context.Context, itemID int64, userID int64) error {
	query := `
	DELETE FROM shopping_list_items
	WHERE id = $1 AND user_id = $2
	`

	_, err := s.db.ExecContext(ctx, query, itemID, userID)
	return err
}

func (s *ShoppinglistStore) DeleteChecked(ctx context.Context, userID int64) error {
	query := `
	DELETE FROM shopping_list_items
	WHERE user_id = $1 AND checked = true
	`

	_, err := s.db.ExecContext(ctx, query, userID)
	return err
}

func (s *ShoppinglistStore) DeleteBySource(ctx context.Context, userID int64, source string) error {
	query := `
	DELETE FROM shopping_list_items
	WHERE user_id = $1 AND source = $2
	`

	_, err := s.db.ExecContext(ctx, query, userID, source)
	return err
}
