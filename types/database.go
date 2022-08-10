package types

type Database interface {
	// GetConnection(id string) *sql.DB
	HeatmapQuery(group []*Dataset, x, y, z, pos int64, intervalTable string, filters string, temporalAggr bool, rows *Rows) error
	// HeatmapQueryH3(group []*Dataset, x, y, z, pos int64, intervalTable string, filters string, temporalAggr bool, rows *Rows) error
	// InteractionQuery(dataset *Dataset, x, y, z, pos int64, filters string, cells []int, limit int, vesselGroupIds []string) ([]map[string]interface{}, error)
	// BindsQuery(dataset *Dataset, z int, filters string, temporalAggregation bool, interval string, numBinds int, vesselGroupIds []string) (map[int]float64, error)
	// StatsQuery(dataset []*Dataset, filters string, vesselGroupIds []string, fields []string) (map[string]interface{}, error)
	// StatsCellQuery(dataset []*Dataset, filters string, vesselGroupIds []string, fields []string, pos int64, cell, z, x, y int) (map[string]interface{}, error)
	// RunQuery(databaseInstance, query string) (*sql.Rows, error)
	// ReportQuery(reportType string, dataset *Dataset, geojson string, filter string, resolution int, resolutionDelta float64, dateFormat string, groupBy string, groupColumns, groupColumnsValid []string, vesselTables []Dataset) (*ReportRows, error)
	// ReportTifMinMaxQuery(dataset *Dataset, geojson string, filter string, resolution int, resolutionDelta float64) (*MinMax, error)
	// Legend(project string, filter string, table string) ([]map[string]interface{}, error)
	Close()
}

type LocalDatabase interface {
	Database
	GetAllDatasets() ([]Dataset, error)
	IngestFile(path string, name string) error
	GetSchema(tablename string) ([]ColumnSchema, error)
	DropTable(tablename string) error
	CreateOrUpdateTempFile(tempFile TempFile) error
	GetTempFile(name string) (*TempFile, error)
	CreateGroupedTables(dataset Dataset, interval string) error
	CreateRawTable(dataset Dataset) error
	IngestDataset(dataset Dataset) error
	GetDistinctValuesOfColumn(table string, field string) ([]interface{}, error)
}
