package sqlite

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/acme/certpilot/internal/shared/database"
	"github.com/acme/certpilot/internal/shared/eventlog/domain"
)

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

func (r *Repository) Create(ctx context.Context, exec database.Executor, event domain.Event) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	metadata, _ := json.Marshal(event.Metadata)
	_, err := exec.ExecContext(ctx, `INSERT INTO event_logs (id, actor, action, entity_type, entity_id, metadata_json, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		event.ID, event.Actor, event.Action, event.EntityType, event.EntityID, string(metadata), event.CreatedAt.Format(time.RFC3339Nano))
	if err != nil {
		return fmt.Errorf("insert event log: %w", err)
	}
	return nil
}

func (r *Repository) List(ctx context.Context, exec database.Executor, options domain.ListOptions) ([]domain.Event, int, error) {
	where, args := buildWhere(options.Filters)
	var total int
	if err := exec.QueryRowContext(ctx, "SELECT COUNT(*) FROM event_logs"+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count events: %w", err)
	}
	sort := sanitizeSort(options.Sort)
	order := strings.ToUpper(options.Order)
	if order != "DESC" {
		order = "ASC"
	}
	rows, err := exec.QueryContext(ctx, "SELECT id, actor, action, entity_type, entity_id, metadata_json, created_at FROM event_logs"+where+" ORDER BY "+sort+" "+order+" LIMIT ? OFFSET ?", append(args, options.PageSize, (options.Page-1)*options.PageSize)...)
	if err != nil {
		return nil, 0, fmt.Errorf("list events: %w", err)
	}
	defer rows.Close()
	items := make([]domain.Event, 0, options.PageSize)
	for rows.Next() {
		var item domain.Event
		var metadataRaw string
		var createdAt string
		if err := rows.Scan(&item.ID, &item.Actor, &item.Action, &item.EntityType, &item.EntityID, &metadataRaw, &createdAt); err != nil {
			return nil, 0, fmt.Errorf("scan event: %w", err)
		}
		_ = json.Unmarshal([]byte(metadataRaw), &item.Metadata)
		item.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
		items = append(items, item)
	}
	return items, total, rows.Err()
}
