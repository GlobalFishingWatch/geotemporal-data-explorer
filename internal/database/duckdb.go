package database

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/4wings/cli/internal"
	"github.com/4wings/cli/internal/utils"
	"github.com/4wings/cli/types"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/sync/semaphore"

	_ "github.com/marcboeker/go-duckdb"
)

type duckdb struct {
	db   *sqlx.DB
	lock *semaphore.Weighted
}

var heatmapTemplate *template.Template
var datasetsTemplate *template.Template
var createGroupedTableTemplate *template.Template
var createRawTableTemplate *template.Template
var selectRawTableTemplate *template.Template
var insert4wingsTableTemplate *template.Template
var insert4wingsDayTableTemplate *template.Template
var insert4wingsMonthTableTemplate *template.Template

func init() {
	funcMap := template.FuncMap{
		// The name "inc" is what the function will be called in the template text.
		"add": func(i int, val int) int {
			return i + val
		},
		"StringsJoin": strings.Join,
		"Iterate": func(count int) []int {
			var i int
			var Items []int
			for i = 0; i < count; i++ {
				Items = append(Items, i)
			}
			return Items
		},
	}
	tmpl := template.New("interactionQuery")
	tmpl.Funcs(funcMap)
	tmpl, err := tmpl.Parse(`select name, configuration from dataset_v1`)
	if err != nil {
		panic(err)
	}
	datasetsTemplate = tmpl

	tmpl = template.New("createGroupedTableTemplate")
	tmpl.Funcs(funcMap)
	tmpl, err = tmpl.Parse(`
	{{range $val := Iterate 13}}
CREATE TABLE IF NOT EXISTS "4wings_{{$.name}}_{{$.resolution}}_z{{$val}}"(
	VALUE FLOAT,
	TIMESTAMP TIMESTAMP,
	HTIME INTEGER,
	POS INTEGER,
	CELL INTEGER,
	{{range $k := $.fields}}
	{{index $k "name"}} {{index $k "type"}},
	{{end}}	
);
	{{end}}
	`)
	if err != nil {
		panic(err)
	}
	createGroupedTableTemplate = tmpl

	tmpl = template.New("createRawTableTemplate")
	tmpl.Funcs(funcMap)
	tmpl, err = tmpl.Parse(`
CREATE TABLE IF NOT EXISTS "4wings_{{$.name}}"(
	VALUE FLOAT,
	TIMESTAMP TIMESTAMP,
	HTIME INTEGER,
	POSITION BIGINT,
	POSITIONS INTEGER[],
	CELLS INTEGER[],
	{{range $k := $.fields}}
	{{index $k "name"}} {{index $k "type"}},
	{{end}}	
);
	`)
	if err != nil {
		panic(err)
	}
	createRawTableTemplate = tmpl

	tmpl = template.New("selectRawTableTemplate")
	tmpl.Funcs(funcMap)
	tmpl, err = tmpl.Parse(`
select 
	{{range $k, $v := .fields}}{{index $v}} as {{index $k}},{{end}}	
FROM "{{$.name}}";
	`)
	if err != nil {
		panic(err)
	}
	selectRawTableTemplate = tmpl

	tmpl = template.New("insert4wingsTableTemplate")
	tmpl.Funcs(funcMap)
	tmpl, err = tmpl.Parse(`
insert into  "4wings_{{.name}}" (HTIME, TIMESTAMP, POSITION, POSITIONS, CELLS, VALUE{{range $k := $.fields}},{{$k}}{{end}}) values 
	`)
	if err != nil {
		panic(err)
	}
	insert4wingsTableTemplate = tmpl

	tmpl = template.New("insert4wingsDayTableTemplate")
	tmpl.Funcs(funcMap)
	tmpl, err = tmpl.Parse(`
		insert into "4wings_{{.name}}_day_z{{.z}}" (TIMESTAMP,HTIME, POS, CELL{{range $k := $.fields}},{{$k}}{{end}},VALUE) 
		select date_trunc('day', "timestamp") as timestamp, floor(htime/24) as htime, pos_{{.z}} as pos, cell_{{.z}} as cell{{range $k := $.fields}},{{$k}}{{end}},{{.func}}(VALUE)
		from "4wings_{{.name}}" group by 1, 2, 3, 4{{range $i,$k := $.fields}},{{add $i 5}}{{end}}
	`)
	if err != nil {
		panic(err)
	}
	insert4wingsDayTableTemplate = tmpl

	tmpl = template.New("insert4wingsMonthTableTemplate")
	tmpl.Funcs(funcMap)
	tmpl, err = tmpl.Parse(`
		insert into "4wings_{{.name}}_month_z{{.z}}" (TIMESTAMP,HTIME, POS, CELL{{range $k := $.fields}},{{$k}}{{end}},VALUE) 
		select date_trunc('month', "timestamp") as timestamp, (extract('year' FROM timestamp)*12 + extract('month' FROM timestamp) - 1) as htime, pos_{{.z}} as pos, cell_{{.z}} as cell{{range $k := $.fields}},{{$k}}{{end}},{{.func}}(VALUE)
		from "4wings_{{.name}}" group by 1, 2, 3, 4{{range $i,$k := $.fields}},{{add $i 5}}{{end}}
	`)
	if err != nil {
		panic(err)
	}
	insert4wingsMonthTableTemplate = tmpl

	tmpl = template.New("heatmapTemplate")
	tmpl.Funcs(funcMap)
	tmpl, err = tmpl.Parse(`
select {{.cellColumn}}, {{if not .temporalAggr}}htime,{{end}} {{.dataColumn}} as count from "{{.tablename}}"
where {{.posFilter}}
{{if ne .filters ""}} and {{.filters}}{{end}}
group by 1 {{if not .temporalAggr}},2{{end}} 
`)
	if err != nil {
		panic(err)
	}
	heatmapTemplate = tmpl
}

func (duckdb *duckdb) Close() {
	log.Info("Closing connections")
	duckdb.Close()
}

func loadDatabase(db *sqlx.DB) error {
	log.Debug("Initializing duckdb database")

	_, err := db.Exec("CREATE TABLE IF NOT EXISTS temp_file(name VARCHAR, status VARCHAR, message VARCHAR);")
	if err != nil {
		return err
	}
	return nil
}

func openDuckDB() (*duckdb, error) {
	log.Debug("Opening DuckDB ")
	db, err := sqlx.Connect("duckdb", fmt.Sprintf("%s?access_mode=READ_WRITE", viper.GetString("local-db")))
	if err != nil {
		return nil, err
	}
	err = loadDatabase(db)
	return &duckdb{
		db,
		semaphore.NewWeighted(1),
	}, err
}

func (duckdb *duckdb) GetAllDatasets() ([]types.Dataset, error) {
	log.Debug("Obtaining datasets from duckdb")
	var query bytes.Buffer
	err := datasetsTemplate.Execute(&query, map[string]interface{}{})
	if err != nil {
		return nil, err
	}

	datasets := make([]types.Dataset, 0)
	err = duckdb.db.Select(&datasets, query.String())
	if err != nil {
		return nil, err
	}

	return datasets, nil
}

func (duckdb *duckdb) existTable(tablename string) bool {
	rows := duckdb.db.QueryRow(fmt.Sprintf("select count(*) as exist from information_schema.tables where table_name = '%s';", tablename))
	var exist int
	err := rows.Scan(&exist)
	if err != nil || exist == 0 {
		return false
	}
	return true
}

func (duckdb *duckdb) IngestFile(path string, name string) error {
	log.Debugf("Ingesting CSV %s with tablename %s", path, name)
	log.Debugf("Checking if exist table %s", name)
	if duckdb.existTable(name) {
		log.Infof("Table with name %s already imported. Skip import", name)
		return nil
	}

	res, err := duckdb.db.Exec(fmt.Sprintf(`CREATE TABLE "%s" AS SELECT * FROM read_csv_auto('%s');`, name, path))
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	log.Infof("Imported csv correctly. Rows affected %d", rowsAffected)
	return nil
}

func (duckdb *duckdb) DropTable(tablename string) error {
	log.Debugf("Drop table %s", tablename)
	res, err := duckdb.db.Exec(fmt.Sprintf(`DROP TABLE IF EXISTS "%s" `, tablename))
	if err != nil {
		return err
	}
	_, err = res.RowsAffected()
	if err != nil {
		return err
	}
	log.Infof("Dropped table %d", tablename)
	return nil
}

func (duckdb *duckdb) GetSchema(tablename string) ([]types.ColumnSchema, error) {
	log.Debugf("Getting schema of table %s", tablename)
	rows, err := duckdb.db.Query(fmt.Sprintf("select column_name, data_type from information_schema.columns where table_name = '%s';", tablename))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	columns := make([]types.ColumnSchema, 0)
	for rows.Next() {
		var c types.ColumnSchema
		err := rows.Scan(&c.Name, &c.Type)
		if err != nil {
			log.Fatal(err)
		}
		columns = append(columns, c)
	}

	return columns, nil
}

func (duckdb *duckdb) GetTempFile(name string) (*types.TempFile, error) {
	log.Debugf("Obtaining temp file with name %s", name)
	rows := duckdb.db.QueryRowx(fmt.Sprintf("select * from temp_file where name = '%s';", name))

	var tempFile types.TempFile
	err := rows.StructScan(&tempFile)
	if err != nil {
		return nil, nil
	}
	return &tempFile, nil
}

func (duckdb *duckdb) CreateOrUpdateTempFile(tempFile types.TempFile) error {
	log.Debugf("Creating or updating temp file with name %s", tempFile.Name)
	exists, err := duckdb.GetTempFile(tempFile.Name)
	if err != nil {
		return err
	}

	if exists != nil {
		log.Debugf("exist temp file. Updating")
		_, err = duckdb.db.Exec("update temp_file set status = ?, message = ? where name = ?", tempFile.Status, tempFile.Message, tempFile.Name)
	} else {
		log.Debugf("NOT exist temp file. Creating")
		_, err = duckdb.db.Exec("insert into temp_file (name, status, message) values (?, ?, ?)", tempFile.Name, tempFile.Status, tempFile.Message)
	}
	return err
}

func (duckdb *duckdb) CreateGroupedTables(dataset types.Dataset, resolution string) error {
	log.Debugf("Creating tables for resolution %s and table name %s", resolution, dataset.Configuration.FileID)
	schema, err := duckdb.GetSchema(dataset.Configuration.FileID)
	if err != nil {
		log.Errorf("error obtaining schema %e", err)
		return err
	}
	fields := []map[string]interface{}{}
	for _, f := range dataset.Configuration.Fields.Filters {
		for _, s := range schema {
			if s.Name == f {
				fields = append(fields, map[string]interface{}{
					"name": f,
					"type": s.Type,
				})
			}
		}
	}
	var query bytes.Buffer
	err = createGroupedTableTemplate.Execute(&query, map[string]interface{}{
		"name":       dataset.Configuration.FileID,
		"resolution": resolution,
		"fields":     fields,
	})
	if err != nil {
		log.Errorf("error creating sql for tables %e", err)
		return err
	}

	_, err = duckdb.db.Exec(query.String())
	return err
}

func (duckdb *duckdb) CreateRawTable(dataset types.Dataset) error {
	log.Debugf("Creating raw table with table name %s", dataset.Configuration.FileID)
	schema, err := duckdb.GetSchema(dataset.Configuration.FileID)
	if err != nil {
		log.Errorf("error obtaining schema %e", err)
		return err
	}
	fields := []map[string]interface{}{}
	for _, f := range dataset.Configuration.Fields.Filters {
		for _, s := range schema {
			if s.Name == f {
				fields = append(fields, map[string]interface{}{
					"name": f,
					"type": s.Type,
				})
			}
		}
	}
	var query bytes.Buffer
	err = createRawTableTemplate.Execute(&query, map[string]interface{}{
		"name":   dataset.Configuration.FileID,
		"fields": fields,
	})
	if err != nil {
		log.Errorf("error creating sql for tables %e", err)
		return err
	}

	_, err = duckdb.db.Exec(query.String())
	return err
}

func (duckdb *duckdb) IngestDataset(dataset types.Dataset) error {
	log.Debugf("Ingesting data of table name %s", dataset.Configuration.FileID)

	df := dataset.Configuration.Fields
	fields := map[string]interface{}{
		"lat":       fmt.Sprintf(`"%s"`, df.Lat),
		"lon":       fmt.Sprintf(`"%s"`, df.Lon),
		"timestamp": fmt.Sprintf(`"%s"`, df.Timestamp),
	}
	if df.Value == "" {
		fields["value"] = "1.0"
	} else {
		fields["value"] = fmt.Sprintf(`"%s"`, df.Value)
	}
	for _, f := range dataset.Configuration.Fields.Filters {
		fields[f] = f
	}
	var query bytes.Buffer
	err := selectRawTableTemplate.Execute(&query, map[string]interface{}{
		"name":   dataset.Configuration.FileID,
		"fields": fields,
	})
	if err != nil {
		log.Errorf("error creating select sql for tables %e", err)
		return err
	}
	rows, err := duckdb.db.Queryx(query.String())
	if err != nil {
		log.Errorf("error obtaining source data %e", err)
		return err
	}
	rowMap := make(map[string]interface{})
	ch := make(chan types.Row, 100000)
	var wg sync.WaitGroup
	wg.Add(1)
	i := 0
	go duckdb.insertRow(dataset, ch, &wg)
	for rows.Next() {
		err := rows.MapScan(rowMap)
		if err != nil {
			log.Errorf("error scanning source data %e", err)
			close(ch)
			return err
		}
		row, err := utils.SanitizeRow(rowMap, dataset.Configuration.Fields.Resolution, dataset.Configuration.Fields.Filters)
		if err != nil {
			log.Errorf("error scanning source data %e", err)
			close(ch)
			return err
		}
		ch <- *row
		i++

	}
	close(ch)
	wg.Wait()

	err = duckdb.IngestFile(fmt.Sprintf("./%s/4wings_%s.csv", internal.DATA_FOLDER, dataset.Configuration.FileID), dataset.Configuration.Table)
	if err != nil {
		log.Errorf("error loading csv raw data %e", err)
		return err
	}
	// os.Remove(fmt.Sprintf("./%s/4wings_%s.csv", internal.DATA_FOLDER, dataset.Configuration.FileID))
	log.Debugf("Ingested data in raw table 4wings_%s correctly", dataset.Configuration.FileID)
	if dataset.Configuration.Fields.Resolution == "hour" {
		log.Debugf("Generating day tables for table name %s", dataset.Configuration.FileID)
		for i := 0; i <= 12; i++ {
			log.Debugf("Generating day table for zoom %d and table name %s", i, dataset.Configuration.FileID)
			var query bytes.Buffer
			err = insert4wingsDayTableTemplate.Execute(&query, map[string]interface{}{
				"name":   dataset.Configuration.FileID,
				"z":      i,
				"func":   dataset.Configuration.AggregationOperation,
				"fields": dataset.Configuration.Fields.Filters,
			})
			if err != nil {
				log.Errorf("error obtaining schema %e", err)
				return err
			}
			_, err := duckdb.db.Exec(query.String())
			if err != nil {
				log.Errorf("error insert %e", err)
				return err
			}
		}
		log.Debugf("Generated day tables for table name %s correctly", dataset.Configuration.FileID)
	}
	if dataset.Configuration.Fields.Resolution == "day" || dataset.Configuration.Fields.Resolution == "hour" {
		log.Debugf("Generating month tables for table name %s", dataset.Configuration.FileID)
		for i := 0; i <= 12; i++ {
			log.Debugf("Generating month table for zoom %d and table name %s", i, dataset.Configuration.FileID)
			var query bytes.Buffer
			err = insert4wingsMonthTableTemplate.Execute(&query, map[string]interface{}{
				"name":   dataset.Configuration.FileID,
				"z":      i,
				"func":   dataset.Configuration.AggregationOperation,
				"fields": dataset.Configuration.Fields.Filters,
			})
			if err != nil {
				log.Errorf("error obtaining schema %e", err)
				return err
			}
			_, err := duckdb.db.Exec(query.String())
			if err != nil {
				log.Errorf("error insert %e", err)
				return err
			}
		}
		log.Debugf("Generated month tables for table name %s correctly", dataset.Configuration.FileID)
	}

	return nil
}

func (duckdb *duckdb) insertRow(dataset types.Dataset, ch chan types.Row, wg *sync.WaitGroup) {
	defer wg.Done()
	schema, err := duckdb.GetSchema(dataset.Configuration.FileID)
	if err != nil {
		log.Errorf("error obtaining schema %e", err)
		return
	}
	csvFile, err := os.Create(fmt.Sprintf("%s/4wings_%s.csv", internal.DATA_FOLDER, dataset.Configuration.FileID))
	if err != nil {
		log.Errorf("Error generating csv file %e", err)
		return
	}
	csvwriter := csv.NewWriter(csvFile)

	header := []string{"lat", "lon", "htime", "timestamp", "position", "value"}
	for _, f := range dataset.Configuration.Fields.Filters {
		header = append(header, f)
	}
	for i := 0; i <= 12; i++ {
		header = append(header, fmt.Sprintf("pos_%d", i))
	}
	for i := 0; i <= 12; i++ {
		header = append(header, fmt.Sprintf("cell_%d", i))
	}
	csvwriter.Write(header)
	i := 0
	for row := range ch {
		date := row.Timestamp.Format("2006-01-02T15:04:05.999Z")
		rowCSV := []string{fmt.Sprintf("%f", row.Lat), fmt.Sprintf("%f", row.Lon), fmt.Sprintf("%d", row.HTime), fmt.Sprintf(`%s`, date), fmt.Sprintf("%d", row.Position), fmt.Sprintf("%f", row.Value)}

		for _, f := range dataset.Configuration.Fields.Filters {
			for _, s := range schema {
				if s.Name == f {
					if s.Type == "VARCHAR" {
						rowCSV = append(rowCSV, fmt.Sprintf(`%s`, row.Others[f]))
					} else if s.Type == "DATE" {
						rowCSV = append(rowCSV, fmt.Sprintf(`%s`, row.Others[f].(time.Time).Format("2006-01-02T15:04:05.999Z")))
					} else if s.Type == "INTEGER" {
						rowCSV = append(rowCSV, fmt.Sprintf(`%d`, row.Others[f]))
					} else if s.Type == "REAL" {
						rowCSV = append(rowCSV, fmt.Sprintf(`%f`, row.Others[f]))
					}

				}
			}
		}
		for i := 0; i <= 12; i++ {
			rowCSV = append(rowCSV, fmt.Sprintf("%d", row.Positions[i]))
		}
		for i := 0; i <= 12; i++ {
			rowCSV = append(rowCSV, fmt.Sprintf("%d", row.Cells[i]))
		}
		err = csvwriter.Write(rowCSV)
		if err != nil {
			log.Error("error writing row %e", err)
		}
		i++

	}
	log.Debugf("wrote %d rows in CSV", i)
	csvwriter.Flush()
	csvFile.Close()

}

func (duckdb *duckdb) HeatmapQueryOfDataset(d *types.Dataset, x, y, z, pos int64, intervalTable string, filters string, temporalAggr bool) (*sql.Rows, error) {
	duckdb.lock.Acquire(context.TODO(), 1)
	defer duckdb.lock.Release(1)
	var query bytes.Buffer
	dataColumn := "sum(value)"
	if d.Configuration.AggregationOperation != "" {
		dataColumn = fmt.Sprintf("%s(value)", d.Configuration.AggregationOperation)
	}
	resolution := intervalTable
	if intervalTable == "" {
		resolution = "hour"
	}

	tablename := ""
	cellColumn := "cell"
	posFilter := fmt.Sprintf("pos = %d", pos)

	if resolution == "hour" {
		tablename = d.Configuration.Table
		cellColumn = fmt.Sprintf("cell_%d", z)
		min, max := utils.GetMinMaxPositionByTile(x, y, z, 12)
		posFilter = fmt.Sprintf("position between %s and %s", min, max)
	} else {
		tablename = fmt.Sprintf("%s_%s_z%d", d.Configuration.Table, resolution, z)

	}
	options := map[string]interface{}{
		"dataset":       d.Configuration.Table,
		"dataColumn":    dataColumn,
		"cellColumn":    cellColumn,
		"zoom":          z,
		"intervalTable": resolution,
		"filters":       filters,
		"temporalAggr":  temporalAggr,
		"tablename":     tablename,
		"posFilter":     posFilter,
	}

	err := heatmapTemplate.Execute(&query, options)
	if err != nil {
		return nil, fmt.Errorf("Error generating query %e", err)
	}
	finalQuery := query.String()
	log.Debugf("Executing query %s", finalQuery)
	return duckdb.db.Query(finalQuery)

}

func (duckdb *duckdb) HeatmapQuery(group []*types.Dataset, x, y, z, pos int64, intervalTable string, filters string, temporalAggr bool, rows *types.Rows) error {
	var wg sync.WaitGroup
	var err error
	for _, d := range group {
		wg.Add(1)
		go func(d *types.Dataset, wg *sync.WaitGroup) {
			sqlRows, errInter := duckdb.HeatmapQueryOfDataset(d, x, y, z, pos, intervalTable, filters, temporalAggr)
			if errInter != nil {
				err = errInter
				log.Error("Error executing query ", errInter)
			} else {
				rows.AddPG(sqlRows)
			}
			wg.Done()
		}(d, &wg)
	}
	wg.Wait()

	return err
}

func (duckdb *duckdb) GetDistinctValuesOfColumn(table string, field string) ([]interface{}, error) {
	log.Debugf("obtaining distinct values of column %s in table %s", field, table)

	fields := make([]interface{}, 0)
	query := fmt.Sprintf(`select distinct %s from "%s"`, field, table)
	err := duckdb.db.Select(&fields, query)
	return fields, err
}
