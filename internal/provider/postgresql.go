package provider

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/tuunit/external-database-operator/api/v1"
	"github.com/tuunit/external-database-operator/api/v1alpha1"
)

type PostgreSQL struct {
	v1.DatabaseHostSpec
}

func NewPostgresClient(spec v1.DatabaseHostSpec) *PostgreSQL {
	return &PostgreSQL{spec}
}

func (p *PostgreSQL) CheckConnection() error {
	connectionString := fmt.Sprintf("host=%s port=%d user=%s password=%s database=postgres sslmode=disable", p.Host, p.Port, p.Superuser, p.Password)

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return fmt.Errorf("Failed to connect to '%s@%s': %w", p.Superuser, p.Host, err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("Failed to ping '%s@%s': %w", p.Superuser, p.Host, err)
	}

	return nil
}

func (p *PostgreSQL) CreateDB(spec *v1alpha1.DatabaseSpec) error {
	connectionString := fmt.Sprintf("host=%s port=%d user=%s password=%s database=postgres sslmode=disable", p.Host, p.Port, p.Superuser, p.Password)

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return fmt.Errorf("Failed to connect to '%s@%s': %w", p.Superuser, p.Host, err)
	}
	defer db.Close()

	var datname string
	err = db.QueryRow(`SELECT datname FROM pg_database WHERE datname = $1`, spec.Name).Scan(&datname)

	owner := p.Superuser
	charset := "UTF8"
	collation := "en_US.UTF-8"

	if spec.Owner != "" {
		owner = spec.Owner
	}

	if spec.Charset != "" {
		charset = spec.Charset
	}

	if spec.Collation != "" {
		collation = spec.Collation
	}

	if err == sql.ErrNoRows {
		_, err = db.Exec(`CREATE DATABASE "` + spec.Name + `"
												  WITH OWNER "` + owner + `" 
													ENCODING '` + charset + `'
													LC_COLLATE '` + collation + `'
													LC_CTYPE '` + collation + `'`)
	}

	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("Failed to create database '%s': %w", spec.Name, err)
	}

	return nil
}

func (p *PostgreSQL) CreateUser(spec *v1alpha1.DatabaseUserSpec) error {
	return nil
}
