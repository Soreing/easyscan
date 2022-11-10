package easyscan

import "database/sql"

type Scannable interface {
	Scan(dest ...any) error
	ColumnTypes() ([]*sql.ColumnType, error)
}

type ScanOne interface {
	ScanRow(Scannable) error
}

type ScanMany interface {
	ScanAppendRow(Scannable) error
}
