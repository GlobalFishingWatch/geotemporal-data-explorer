package types

import (
	"database/sql"
	"strconv"
	"strings"
	"time"

	bq "google.golang.org/api/bigquery/v2"
)

type Dataset struct {
	ID            string        `json:"id,omitempty"  binding:"-"`
	Name          string        `json:"name,omitempty" binding:"required"`
	Description   string        `json:"description,omitempty" binding:"-"`
	Source        string        `json:"source,omitempty" binding:"required"`
	Type          string        `json:"type,omitempty" binding:"required"`
	StartDate     time.Time     `json:"startDate,omitempty" binding:"-" time_format:"2006-01-02"`
	EndDate       time.Time     `json:"endDate,omitempty" binding:"-" time_format:"2006-01-02"`
	Unit          string        `json:"unit,omitempty" binding:"-" `
	Status        string        `json:"status,omitempty" binding:"-"`
	Configuration Configuration `json:"configuration,omitempty" binding:"required,dive,required"`
}

type Configuration struct {
	Intervals            []string `json:"intervals,omitempty" binding:"-"`
	Images               []string `json:"images,omitempty" binding:"-"`
	Band                 string   `json:"band,omitempty" binding:"-"`
	Min                  float64  `json:"min" binding:"-"`
	Max                  float64  `json:"max" binding:"-"`
	Scale                float64  `json:"scale" binding:"-"`
	Offset               float64  `json:"offset" binding:"-"`
	PolygonID            string   `json:"polygonId,omitempty" binding:"-"`
	FileID               string   `json:"fileId,omitempty"  binding:"required"`
	MaxZoom              int      `json:"maxZoom,omitempty"  binding:"-"`
	Table                string   `json:"table,omitempty"`
	Source               string   `json:"source,omitempty"  binding:"-"`
	Fields               Fields   `json:"fields,omitempty"`
	AggregationOperation string   `json:"aggregationOperation,omitempty" binding:"-"`
	ValueMultiplier      float64  `json:"valueMultiplier,omitempty" binding:"-"`
}

type Fields struct {
	ID         string   `json:"id"`
	Lat        string   `json:"lat"`
	Lon        string   `json:"lon"`
	Timestamp  string   `json:"timestamp"`
	Value      string   `json:"value"`
	Resolution string   `json:"resolution"`
	Filters    []string `json:"filters"`
}

type Pagination[T any] struct {
	Total      int                    `json:"total"`
	Limit      *int                   `json:"limit"`
	Offset     *int                   `json:"offset"`
	NextOffset *int                   `json:"nextOffset"`
	Metadata   map[string]interface{} `json:"metadata"`
	Entries    []T                    `json:"entries"`
}

func NewPagination[T any](entries []T) *Pagination[T] {
	return &Pagination[T]{
		Total:    len(entries),
		Entries:  entries,
		Metadata: map[string]interface{}{},
	}
}

type ColumnSchema struct {
	Name string `json:"name,omitempty"`
	Type string `json:"type,omitempty"`
}

const (
	Created   = "CREATED"
	Importing = "IMPORTING"
	Error     = "ERROR"
	Completed = "COMPLETED"
)

type TempFile struct {
	Name    string `json:"name,omitempty"`
	Status  string `json:"status,omitempty"`
	Message string `json:"message,omitempty"`
}

type RowObj struct {
	Cell    int     `bigquery:"cell"`
	Htime   int     `bigquery:"htime"`
	Value   float64 `bigquery:"value"`
	H3Index string
}

type Row struct {
	Lat       float64
	Lon       float64
	Position  int64
	Timestamp time.Time
	Cells     []int
	Positions []int64
	HTime     int64
	Value     float64
	Others    map[string]interface{}
}

type Rows struct {
	iterators         []*sql.Rows
	bqiterators       [][]RowObj
	current           int
	bqcurrent         int
	RowObj            RowObj
	IndexCellAndValue int
	Data              []string
}

func NewRows() *Rows {
	return &Rows{
		iterators:         make([]*sql.Rows, 0),
		bqiterators:       make([][]RowObj, 0),
		current:           0,
		bqcurrent:         0,
		IndexCellAndValue: -1,
	}
}
func (r *Rows) AddPG(row *sql.Rows) {
	r.iterators = append(r.iterators, row)
}
func (r *Rows) AddBQ(tableRows []*bq.TableRow) {
	data := make([]RowObj, 0)
	for _, tr := range tableRows {
		if len(tr.F) == 1 {
			res, ok := tr.F[0].V.(string)
			if !ok {
				break
			}
			parts := strings.Split(res, ",")
			for i := 0; i < len(parts); i = i + 2 {
				cell, _ := strconv.Atoi(parts[i])
				value, _ := strconv.ParseFloat(parts[i+1], 64)
				data = append(data, RowObj{
					Cell:    cell,
					Value:   value,
					H3Index: parts[i],
				})
			}
		} else {
			res, ok := tr.F[0].V.(string)
			if !ok {
				break
			}
			htime, _ := strconv.Atoi(res)
			parts := strings.Split(tr.F[1].V.(string), ",")
			for i := 0; i < len(parts); i = i + 2 {
				cell, _ := strconv.Atoi(parts[i])
				value, _ := strconv.ParseFloat(parts[i+1], 64)
				data = append(data, RowObj{
					Cell:    cell,
					Value:   value,
					Htime:   htime,
					H3Index: parts[i],
				})
			}
		}

	}

	r.bqiterators = append(r.bqiterators, data)
}

func (r *Rows) Close() {
	for _, r := range r.iterators {
		if r != nil {
			r.Close()
		}
	}
}

func (r *Rows) Next() bool {

	for {
		if r.current >= len(r.iterators) {
			break
		}
		if r.iterators[r.current].Next() {
			return true
		}
		r.iterators[r.current].Close()
		r.current++
	}

	for {
		if r.bqcurrent >= len(r.bqiterators) {
			return false
		}
		r.IndexCellAndValue++

		if r.IndexCellAndValue >= len(r.bqiterators[r.bqcurrent]) {
			r.bqcurrent++
			r.IndexCellAndValue = 0
			if r.bqcurrent >= len(r.bqiterators) {
				return false
			}
		} else {
			return true
		}

	}

}

func (r *Rows) Scan(totalValuesToScan int) (RowObj, error) {
	if len(r.iterators) == 0 || r.current >= len(r.iterators) {
		return r.bqiterators[r.bqcurrent][r.IndexCellAndValue], nil
	} else {
		var obj RowObj
		var err error
		if totalValuesToScan == 2 {
			err = r.iterators[r.current].Scan(&obj.Cell, &obj.Value)
		} else {
			err = r.iterators[r.current].Scan(&obj.Cell, &obj.Htime, &obj.Value)
		}
		return obj, err
	}
}
