package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// Setting reads a raw settings value.
func (d *DB) Setting(ctx context.Context, key string) (string, error) {
	var value string
	err := d.sql.QueryRowContext(ctx, `SELECT value FROM settings WHERE key = ?`, key).Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("read setting %q: %w", key, err)
	}
	return value, nil
}

// SetSetting writes a raw settings value.
func (d *DB) SetSetting(ctx context.Context, key, value string) error {
	_, err := d.sql.ExecContext(ctx, `
		INSERT INTO settings (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value`, key, value)
	if err != nil {
		return fmt.Errorf("write setting %q: %w", key, err)
	}
	return nil
}

// SetProviderConnection records the state of an external provider.
func (d *DB) SetProviderConnection(ctx context.Context, provider, state, metadata string) error {
	_, err := d.sql.ExecContext(ctx, `
		INSERT INTO provider_connections (provider, connection_state, metadata, updated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(provider) DO UPDATE SET
			connection_state = excluded.connection_state,
			metadata = excluded.metadata,
			updated_at = excluded.updated_at`, provider, state, metadata, nowMs())
	if err != nil {
		return fmt.Errorf("save provider connection: %w", err)
	}
	return nil
}

// ProviderConnection reads a provider's connection state.
func (d *DB) ProviderConnection(ctx context.Context, provider string) (state string, metadata string, err error) {
	err = d.sql.QueryRowContext(ctx,
		`SELECT connection_state, metadata FROM provider_connections WHERE provider = ?`, provider).
		Scan(&state, &metadata)
	if errors.Is(err, sql.ErrNoRows) {
		return "disconnected", "", nil
	}
	if err != nil {
		return "", "", fmt.Errorf("read provider connection: %w", err)
	}
	return state, metadata, nil
}
