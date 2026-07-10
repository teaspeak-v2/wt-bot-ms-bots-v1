package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/teaspeak-v2/wt-bot-ms-bots-v1/internal/models"
	"github.com/teaspeak-v2/wt-bot-ms-bots-v1/internal/secure"
)

// BotRepository defines persistence for Bot records.
type BotRepository interface {
	Create(ctx context.Context, bot *models.Bot) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Bot, error)
	List(ctx context.Context, filter models.BotListFilter) ([]models.Bot, int64, error)
	Update(ctx context.Context, id uuid.UUID, req models.UpdateBotRequest) (*models.Bot, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type db interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type Repository struct {
	db     db
	encKey string
}

func New(db db, encKey string) *Repository {
	return &Repository{db: db, encKey: encKey}
}

const botColumns = `b.id, b.name, b.teamspeak_id, b.owner_id, b.nickname, b.greeting, b.help_message, b.enabled, b.status, b.api_key, b.created_at, b.updated_at`

func (r *Repository) Create(ctx context.Context, bot *models.Bot) error {
	encAPIKey, err := secure.Encrypt(bot.APIKey, r.encKey)
	if err != nil {
		return fmt.Errorf("encrypt api key: %w", err)
	}
	if bot.ID == uuid.Nil {
		bot.ID = uuid.New()
	}
	now := time.Now().UTC()
	bot.CreatedAt = now
	bot.UpdatedAt = now

	const q = `insert into bots (
		id, name, teamspeak_id, owner_id, nickname, greeting, help_message, enabled, status, api_key, created_at, updated_at
	) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`
	_, err = r.db.Exec(ctx, q,
		bot.ID, bot.Name, bot.TeamSpeakID, bot.OwnerID, bot.Nickname, bot.Greeting, bot.HelpMessage, bot.Enabled, bot.Status, encAPIKey, bot.CreatedAt, bot.UpdatedAt,
	)
	return err
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*models.Bot, error) {
	bot, err := r.getByID(ctx, id)
	if err != nil {
		return nil, err
	}
	bot.APIKey, err = secure.Decrypt(bot.APIKey, r.encKey)
	if err != nil {
		return nil, fmt.Errorf("decrypt api key: %w", err)
	}
	return bot, nil
}

func (r *Repository) getByID(ctx context.Context, id uuid.UUID) (*models.Bot, error) {
	q := fmt.Sprintf(`select %s from bots b where b.id=$1`, botColumns)
	row := r.db.QueryRow(ctx, q, id)
	return scanBot(row)
}

func (r *Repository) List(ctx context.Context, filter models.BotListFilter) ([]models.Bot, int64, error) {
	where, args := r.buildWhere(filter)
	sort := r.sanitizeSort(filter.SortBy, filter.SortOrder)
	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	listQ := fmt.Sprintf(`select %s from bots b %s order by %s limit $%d offset $%d`, botColumns, where, sort, len(args)+1, len(args)+2)
	allArgs := append(args, limit, offset)

	rows, err := r.db.Query(ctx, listQ, allArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []models.Bot
	for rows.Next() {
		bot, err := scanBot(rows)
		if err != nil {
			return nil, 0, err
		}
		bot.APIKey, err = secure.Decrypt(bot.APIKey, r.encKey)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, *bot)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	countQ := fmt.Sprintf(`select count(*) from bots %s`, where)
	var count int64
	if err := r.db.QueryRow(ctx, countQ, args...).Scan(&count); err != nil {
		return nil, 0, err
	}
	return list, count, nil
}

func (r *Repository) buildWhere(filter models.BotListFilter) (string, []any) {
	var conds []string
	var args []any
	nextIdx := func() int {
		return len(args) + 1
	}
	if filter.Search != "" {
		idx := nextIdx()
		conds = append(conds, fmt.Sprintf(`(b.name ilike $%d or b.nickname ilike $%d)`, idx, idx))
		args = append(args, "%"+filter.Search+"%")
	}
	if filter.TeamSpeakID != uuid.Nil {
		conds = append(conds, fmt.Sprintf("b.teamspeak_id = $%d", nextIdx()))
		args = append(args, filter.TeamSpeakID)
	}
	if filter.Enabled != nil {
		conds = append(conds, fmt.Sprintf("b.enabled = $%d", nextIdx()))
		args = append(args, *filter.Enabled)
	}
	if filter.OwnerID != uuid.Nil {
		conds = append(conds, fmt.Sprintf("b.owner_id = $%d", nextIdx()))
		args = append(args, filter.OwnerID)
	}
	if len(conds) == 0 {
		return "", args
	}
	return "where " + strings.Join(conds, " and "), args
}

var sortByWhitelist = map[string]string{
	"created_at": "created_at",
	"updated_at": "updated_at",
	"name":       "name",
	"status":     "status",
	"enabled":    "enabled",
}

func (r *Repository) sanitizeSort(sortBy, sortOrder string) string {
	col, ok := sortByWhitelist[strings.ToLower(sortBy)]
	if !ok {
		col = "created_at"
	}
	order := "desc"
	if strings.EqualFold(sortOrder, "asc") {
		order = "asc"
	}
	return col + " " + order
}

func (r *Repository) Update(ctx context.Context, id uuid.UUID, req models.UpdateBotRequest) (*models.Bot, error) {
	bot, err := r.getByID(ctx, id)
	if err != nil {
		return nil, err
	}
	name := coalesceStr(req.Name, bot.Name)
	teamspeakID := coalesceUUID(req.TeamSpeakID, bot.TeamSpeakID)
	nickname := coalesceStr(req.Nickname, bot.Nickname)
	greeting := coalesceStr(req.Greeting, bot.Greeting)
	helpMessage := coalesceStr(req.HelpMessage, bot.HelpMessage)
	enabled := coalesceBool(req.Enabled, bot.Enabled)
	apiKey, err := secure.Encrypt(bot.APIKey, r.encKey)
	if err != nil {
		return nil, err
	}

	const q = `update bots set name=$2, teamspeak_id=$3, nickname=$4, greeting=$5, help_message=$6, enabled=$7, api_key=$8, updated_at=now()
		where id=$1
		returning ` + botColumns
	row := r.db.QueryRow(ctx, q,
		id, name, teamspeakID, nickname, greeting, helpMessage, enabled, apiKey,
	)
	bot, err = scanBot(row)
	if err != nil {
		return nil, err
	}
	bot.APIKey, err = secure.Decrypt(bot.APIKey, r.encKey)
	if err != nil {
		return nil, err
	}
	return bot, nil
}

func (r *Repository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.db.Exec(ctx, `update bots set status=$2, updated_at=now() where id=$1`, id, status)
	return err
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `delete from bots where id=$1`, id)
	return err
}

func coalesceStr(v *string, def string) string {
	if v != nil {
		return *v
	}
	return def
}

func coalesceUUID(v *uuid.UUID, def uuid.UUID) uuid.UUID {
	if v != nil {
		return *v
	}
	return def
}

func coalesceBool(v *bool, def bool) bool {
	if v != nil {
		return *v
	}
	return def
}

type scanner interface {
	Scan(...any) error
}

func scanBot(row scanner) (*models.Bot, error) {
	var bot models.Bot
	if err := row.Scan(
		&bot.ID, &bot.Name, &bot.TeamSpeakID, &bot.OwnerID, &bot.Nickname, &bot.Greeting, &bot.HelpMessage, &bot.Enabled, &bot.Status, &bot.APIKey, &bot.CreatedAt, &bot.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &bot, nil
}

func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func IsNotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}
