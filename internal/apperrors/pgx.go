package apperrors

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	pgErrUniqueViolation     = "23505"
	pgErrConstraintViolation = "23503"
)

// FromPgx traduce errores de pgx al dominio de la aplicación. Debe llamarse
// exactamente una vez por error, en el punto más cercano a la llamada a la DB
// (dentro del service, no en el handler). Llamarlo dos veces es inocuo porque
// un error que ya no es *pgconn.PgError llega directo al return err del final,
// pero indica un mal manejo del flujo de errores.
//
// Pasar constraintMap como nil es válido cuando la query no tiene constraints
// FK o unique relevantes (e.g. SELECTs, o INSERTs sin constraints de negocio).
func FromPgx(err error, constraintMap map[string]error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		if pgErr.Code == pgErrUniqueViolation {
			if mapped, ok := constraintMap[pgErr.ConstraintName]; ok {
				return mapped
			}
		}
		if pgErr.Code == pgErrConstraintViolation {
			if mapped, ok := constraintMap[pgErr.ConstraintName]; ok {
				return mapped
			}
		}
	}
	return err
}
