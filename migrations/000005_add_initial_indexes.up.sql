-- Índice inverso en groups_focus_types para el filtro OR por categorías.
-- La PK (group_id, focus_type_id) sirve para "categorías de un grupo",
-- pero el filtro va al revés: dado un focus_type_id buscar sus group_ids.
CREATE INDEX idx_gft_focus_type ON groups_focus_types (focus_type_id, group_id);

-- Búsqueda por nombre con ILIKE requiere pg_trgm para usar un índice GIN.
-- Sin esta extensión, ILIKE '%texto%' hace seq-scan siempre.
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX idx_groups_name_trgm       ON groups USING GIN (name gin_trgm_ops);
CREATE INDEX idx_groups_depth_level_trgm ON groups USING GIN (depth_level gin_trgm_ops);

