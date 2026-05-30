package api_cultura_conecta

import "embed"

//go:embed all:migrations/*.sql
var MigrationFS embed.FS
