package database

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"
	pakstore "github.com/LoveRetro/nextui-pak-store"
	"github.com/LoveRetro/nextui-pak-store/models"
	"github.com/LoveRetro/nextui-pak-store/utils"
	_ "modernc.org/sqlite"
)

var dbc *sql.DB
var queries *Queries

func Init() {
	logger := gabagool.GetLogger()
	ctx := context.Background()

	var err error
	dbPath := filepath.Join(utils.GetUserDataDir(), "pak-store.db")

	if os.Getenv("ENVIRONMENT") == "DEV" {
		dbPath = "pak-store.db"
	}

	logger.Debug("Database path resolved", "path", dbPath)

	dbDir := filepath.Dir(dbPath)
	if dbDir != "." && dbDir != "" {
		err := os.MkdirAll(dbDir, 0755)
		if err != nil {
			logger.Error("Unable to create database directory", "error", err, "dir", dbDir)
			os.Exit(1)
		}
	}

	dbc, err = sql.Open("sqlite", "file:"+dbPath)
	if err != nil {
		logger.Error("Unable to open database file", "error", err, "path", dbPath)
		os.Exit(1)
	}

	schemaExists, err := tableExists(dbc, "installed_paks")
	if !schemaExists {
		logger.Debug("Initializing database schema")
		if _, err := dbc.ExecContext(ctx, pakstore.DDL); err != nil {
			logger.Error("Unable to init schema", "error", err)
			os.Exit(1)
		}
	}

	columnMigration("installed_paks", "repo_url", "TEXT")
	columnMigration("installed_paks", "pak_id", "TEXT")

	queries = New(dbc)

	var pak models.Pak
	err = utils.ParseJSONFile("pak.json", &pak)
	if err != nil {
		log.Fatalf("Error parsing JSON file: %v", err)
	}

	existingPakStore, err := queries.GetInstalledByPakID(ctx, sql.NullString{String: models.PakStoreID, Valid: true})
	if errors.Is(err, sql.ErrNoRows) {
		queries.SyncPakStoreByName(ctx, SyncPakStoreByNameParams{
			DisplayName: pak.Name,
			Name:        pak.Name,
			Version:     pak.Version,
			RepoUrl:     sql.NullString{String: models.PakStoreRepo, Valid: true},
			PakID:       sql.NullString{String: models.PakStoreID, Valid: true},
			OldName:     "Pak Store",
		})

		_, err = queries.GetInstalledByPakID(ctx, sql.NullString{String: models.PakStoreID, Valid: true})
		if errors.Is(err, sql.ErrNoRows) {
			queries.Install(ctx, InstallParams{
				DisplayName:  pak.Name,
				Name:         pak.Name,
				PakID:        sql.NullString{String: models.PakStoreID, Valid: true},
				RepoUrl:      sql.NullString{String: models.PakStoreRepo, Valid: true},
				Version:      pak.Version,
				Type:         "TOOL",
				CanUninstall: 0,
			})
		}
	} else if err != nil {
		logger.Error("Unable to check for Pak Store record", "error", err)
		os.Exit(1)
	} else if existingPakStore.Version != pak.Version {
		queries.SyncPakStore(ctx, SyncPakStoreParams{
			DisplayName: pak.Name,
			Name:        pak.Name,
			Version:     pak.Version,
			RepoUrl:     sql.NullString{String: models.PakStoreRepo, Valid: true},
			PakID:       sql.NullString{String: models.PakStoreID, Valid: true},
		})
	}
}

func DBQ() *Queries {
	return queries
}

func CloseDB() {
	_ = dbc.Close()
}

func tableExists(db *sql.DB, tableName string) (bool, error) {
	query := `SELECT name FROM sqlite_master WHERE type='table' AND name=?`
	var name string
	err := db.QueryRow(query, tableName).Scan(&name)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

func columnExists(db *sql.DB, tableName, columnName string) (bool, error) {
	query := fmt.Sprintf("PRAGMA table_info(%s)", tableName)
	rows, err := db.Query(query)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull, pk int
		var defaultValue sql.NullString

		err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk)
		if err != nil {
			return false, err
		}

		if name == columnName {
			return true, nil
		}
	}

	return false, rows.Err()
}

func columnMigration(tableName, columnName, columnDefinition string) {
	logger := gabagool.GetLogger()
	ctx := context.Background()

	ce, err := columnExists(dbc, tableName, columnName)
	if err != nil {
		logger.Error("Unable to check column existence", "error", err)
		os.Exit(1)
	}

	if !ce {
		migrationSQL := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", tableName, columnName, columnDefinition)
		if _, err := dbc.ExecContext(ctx, migrationSQL); err != nil {
			logger.Error("Unable to run column migration", "error", err)
			os.Exit(1)
		}
		logger.Info("Successfully added column", "column", columnName)
	}
}
