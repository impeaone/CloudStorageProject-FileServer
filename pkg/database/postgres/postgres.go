package postgres

import (
	"CloudStorageProject-FileServer/pkg/tools"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	pool  *pgxpool.Pool
	ctxDB context.Context
	clsDB context.CancelFunc
}

func InitPostgres() (*Postgres, error) {
	pgUser := tools.GetEnv("PG_USER", "postgres")
	pgPassword := tools.GetEnv("PG_PASSWORD", "080455mN")
	pgHost := tools.GetEnv("PG_HOST", "localhost")
	pgPort := tools.GetEnvAsInt("PG_PORT", 5432)
	pgDatabase := tools.GetEnv("PG_DATABASE", "storage")

	connStr := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s", pgUser, pgPassword, pgHost, pgPort, pgDatabase)

	ctx := context.Background()
	pool, errPGX := pgxpool.New(ctx, connStr)
	if errPGX != nil {
		return nil, errPGX
	}
	err := createTables(pool)
	if err != nil {
		return nil, err
	}
	createExampleAPI(pool)

	ctxDB, cls := context.WithCancel(context.Background())
	return &Postgres{
		pool:  pool,
		ctxDB: ctxDB,
		clsDB: cls,
	}, nil

}
func createTables(pool *pgxpool.Pool) error {
	ctx := context.Background()
	_, err := pool.Query(ctx, `
		CREATE TABLE IF NOT EXISTS minio_keys (
    		id SERIAL PRIMARY KEY,
    		key_name VARCHAR(100) NOT NULL,
			cloud_access VARCHAR(5) DEFAULT '010',
    		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			last_login TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		return err
	}
	return nil
}
func createExampleAPI(pool *pgxpool.Pool) {
	ctx := context.Background()
	query := fmt.Sprintf("INSERT INTO minio_keys (key_name) VALUES ('%s') on conflict do nothing", "test")
	_ = pool.QueryRow(ctx, query)
}

func (p *Postgres) CheckApiExists(api string) bool {
	ctx := context.Background()
	var apiID string
	query := fmt.Sprintf("SELECT id FROM minio_keys WHERE key_name = '%s'", api)
	if p.pool.QueryRow(ctx, query).Scan(&apiID); apiID == "" {
		return false
	}
	return true
}

func (p *Postgres) UpdateLastLogin(api string) {

}

func (p *Postgres) Close() {
	p.clsDB()
}
