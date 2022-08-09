package database

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/4wings/cli/internal/utils"
	"github.com/4wings/cli/types"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	_ "github.com/marcboeker/go-duckdb"
)

type duckdb struct {
	db *sqlx.DB
}

var datasetsTemplate *template.Template
var createGroupedTableTemplate *template.Template
var createRawTableTemplate *template.Template
var selectRawTableTemplate *template.Template
var insert4wingsTableTemplate *template.Template
var insert4wingsDayTableTemplate *template.Template

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
		select date_trunc('day', "timestamp") as timestamp, floor(htime/24) as htime, positions[{{add .z 1}}] as pos, cells[{{add .z 1}}] as cell{{range $k := $.fields}},{{$k}}{{end}},{{.func}}(VALUE)
		from "4wings_{{.name}}" group by 1, 2, 3, 4{{range $i,$k := $.fields}},{{add $i 5}}{{end}}
	`)
	if err != nil {
		panic(err)
	}
	insert4wingsDayTableTemplate = tmpl
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

	res, err := duckdb.db.Exec(fmt.Sprintf(`CREATE TABLE "%s" AS SELECT * FROM '%s';`, name, path))
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
	ch := make(chan *types.Row, 10)
	var wg sync.WaitGroup
	wg.Add(1)
	go duckdb.insertRow(dataset, ch, &wg)
	for rows.Next() {
		err := rows.MapScan(rowMap)
		if err != nil {
			log.Errorf("error scanning source data %e", err)
			close(ch)
			return err
		}
		row, err := utils.SanitizeRow(rowMap, dataset.Configuration.Fields.Resolution)
		if err != nil {
			log.Errorf("error scanning source data %e", err)
			close(ch)
			return err
		}
		ch <- row
	}
	close(ch)
	wg.Wait()

	for i := 0; i <= 12; i++ {
		var query bytes.Buffer
		err = insert4wingsDayTableTemplate.Execute(&query, map[string]interface{}{
			"name":   dataset.Configuration.FileID,
			"z":      i,
			"func":   "SUM",
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
		fmt.Println(query.String())
	}

	return nil
}

func (duckdb *duckdb) insertRow(dataset types.Dataset, ch chan *types.Row, wg *sync.WaitGroup) {
	defer wg.Done()
	schema, err := duckdb.GetSchema(dataset.Configuration.FileID)
	if err != nil {
		log.Errorf("error obtaining schema %e", err)
		return
	}
	var query bytes.Buffer
	err = insert4wingsTableTemplate.Execute(&query, map[string]interface{}{
		"name":   dataset.Configuration.FileID,
		"fields": dataset.Configuration.Fields.Filters,
	})
	if err != nil {
		log.Errorf("error generating insert sql %e", err)
		return
	}
	for row := range ch {
		date := row.Timestamp.Format("2006-01-02T15:04:05.999Z")
		values := fmt.Sprintf(`%d, '%s', %d, %s, %s, %f`, row.HTime, date, row.Position, utils.ArrayToString(row.Positions), utils.ArrayToString(row.Cells), row.Value)

		for _, f := range dataset.Configuration.Fields.Filters {
			for _, s := range schema {
				if s.Name == f {
					if s.Type == "VARCHAR" {
						values = fmt.Sprintf(`%s, '%s'`, values, row.Others[f])
					} else if s.Type == "DATE" {
						values = fmt.Sprintf(`%s, '%s'`, values, row.Others[f].(time.Time).Format("2006-01-02T15:04:05.999Z"))
					} else if s.Type == "INTEGER" {
						values = fmt.Sprintf(`%s, %d`, values, row.Others[f])
					} else if s.Type == "REAL" {
						values = fmt.Sprintf(`%s, %f`, values, row.Others[f])
					}
				}
			}
		}
		finalQuery := fmt.Sprintf("%s (%s)", query.String(), values)
		_, err := duckdb.db.Exec(finalQuery)
		if err != nil {
			log.Errorf("error insert %e", err)
		}
	}

}
