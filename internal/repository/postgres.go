package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/MosinFAM/vk-marketplace/internal/models"
	"github.com/google/uuid"
)

type PostgresRepo struct {
	db *sql.DB
}

func NewPostgresRepo(db *sql.DB) *PostgresRepo {
	return &PostgresRepo{db: db}
}

func (r *PostgresRepo) CreateAd(ad models.Ad) (models.Ad, error) {
	ad.ID = uuid.New().String()
	query := `
		INSERT INTO ads (id, title, description, image_url, price, user_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at`
	err := r.db.QueryRow(query, ad.ID, ad.Title, ad.Description, ad.ImageURL, ad.Price, ad.UserID).Scan(&ad.CreatedAt)
	return ad, err
}

func (r *PostgresRepo) ListAds(f AdsFilter) ([]models.Ad, error) {
	where := []string{"TRUE"}
	args := []interface{}{}
	idx := 1

	if f.MinPrice != nil {
		where = append(where, fmt.Sprintf("price >= $%d", idx))
		args = append(args, *f.MinPrice)
		idx++
	}
	if f.MaxPrice != nil {
		where = append(where, fmt.Sprintf("price <= $%d", idx))
		args = append(args, *f.MaxPrice)
		idx++
	}

	order := "created_at DESC"
	if f.SortBy != "" {
		direction := "ASC"
		if strings.ToLower(f.SortOrder) == "desc" {
			direction = "DESC"
		}
		if f.SortBy == "price" || f.SortBy == "created_at" {
			order = fmt.Sprintf("%s %s", f.SortBy, direction)
		}
	}

	args = append(args, f.Limit, f.Offset)
	query := fmt.Sprintf(`
		SELECT a.id, a.title, a.description, a.image_url, a.price, a.created_at, a.user_id, u.username
		FROM ads a
		JOIN users u ON a.user_id = u.id
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, strings.Join(where, " AND "), order, idx, idx+1)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.Ad
	for rows.Next() {
		var ad models.Ad
		if err := rows.Scan(&ad.ID, &ad.Title, &ad.Description, &ad.ImageURL, &ad.Price, &ad.CreatedAt, &ad.UserID, &ad.Username); err != nil {
			return nil, err
		}
		if f.UserID != nil && *f.UserID == ad.UserID {
			ad.IsOwner = true
		}
		result = append(result, ad)
	}
	return result, nil
}

func (r *PostgresRepo) GetAdByID(id string) (*models.Ad, error) {
	query := `SELECT id, title, description, image_url, price, created_at, user_id FROM ads WHERE id=$1`
	var ad models.Ad
	err := r.db.QueryRow(query, id).Scan(&ad.ID, &ad.Title, &ad.Description, &ad.ImageURL, &ad.Price, &ad.CreatedAt, &ad.UserID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &ad, err
}

func (r *PostgresRepo) RegisterUser(username, hash string) (string, error) {
	id := uuid.New().String()
	_, err := r.db.Exec("INSERT INTO users (id, username, password_hash) VALUES ($1, $2, $3)", id, username, hash)
	return id, err
}

func (r *PostgresRepo) GetUserByUsername(username string) (*models.User, error) {
	var u models.User
	err := r.db.QueryRow("SELECT id, username, password_hash FROM users WHERE username=$1", username).
		Scan(&u.ID, &u.Username, &u.PasswordHash)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &u, err
}
