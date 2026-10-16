package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"path"

	"go-codeexplorer/migrations"

	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"

	_ "github.com/jackc/pgx/v5/stdlib" // регистрирует database/sql драйвер "pgx"
)

// Migrate применяет все непримененные миграции из migrations.FS.
//
// Версия схемы учитывается в таблице goose_db_version. Состояния "dirty" у goose нет:
// если миграция падает внутри транзакции, она откатывается целиком вместе с записью версии.
func Migrate(dsn string, log *slog.Logger) (err error) {
	ctx := context.Background()

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	// Пул соединений принадлежит этой функции, закрываем его сами.
	// Важно: goose.Provider.Close() закрывает тот же самый *sql.DB, поэтому мы его
	// намеренно не вызываем — иначе закроем пул дважды (или, если бы db был общим
	// для приложения, убили бы чужой пул).
	defer func() {
		if cerr := db.Close(); err == nil {
			err = cerr
		}
	}()

	// goose по умолчанию не блокирует БД — при нескольких репликах приложения
	// миграции поедут параллельно. Воспроизводим прежнее поведение migrate
	// (advisory lock с таймаутом 2 минуты): опрос каждые 2с, до 60 попыток.
	sessionLocker, err := lock.NewPostgresSessionLocker(
		lock.WithLockTimeout(2, 60),   // 2s × 60 = 2m на захват
		lock.WithUnlockTimeout(2, 30), // 2s × 30 = 1m на освобождение
	)
	if err != nil {
		return fmt.Errorf("create session locker: %w", err)
	}

	provider, err := goose.NewProvider(
		goose.DialectPostgres,
		db,
		migrations.FS,
		goose.WithSlog(log),
		goose.WithSessionLocker(sessionLocker),
		// Подробные логи включаем, только если их кто-то читает.
		goose.WithVerbose(log.Enabled(ctx, slog.LevelDebug)),
	)
	switch {
	case errors.Is(err, goose.ErrNoMigrations):
		// Билд без //go:embed migrations/*.sql — это ошибка конфигурации сборки,
		// молча стартовать с пустой схемой нельзя.
		return fmt.Errorf("no migrations found in embedded FS: %w", err)
	case err != nil:
		return fmt.Errorf("init migrator: %w", err)
	}

	results, err := provider.Up(ctx)
	if err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}

	if len(results) == 0 {
		log.Info("schema is up to date")
		return nil
	}

	for _, r := range results {
		log.Info("migration applied",
			"version", r.Source.Version,
			"file", path.Base(r.Source.Path),
			"duration", r.Duration,
		)
	}

	if v, verr := provider.GetDBVersion(ctx); verr == nil {
		log.Info("migrations applied", "count", len(results), "version", v)
	}
	return nil
}
