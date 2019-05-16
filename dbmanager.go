package lnksworks

import (
	"io"
	"strings"
	"sync"
	/*
		_ "github.com/denisenkom/go-mssqldb"
		_ "github.com/go-sql-driver/mysql"
		_ "github.com/jackc/pgx"

			_ "github.com/HouzuoGuo/tiedot/dberr"
			_ "github.com/syndtr/goleveldb/leveldb"
	*/)

//DbManager DbManager controller
type DbManager struct {
	dbaliases map[string]*DbConnection
}

//Open - open a db connection based and register aliased reference to it in the DbManager
//returns a DbConnnection controller of the connection
func (dbmngr *DbManager) Open(alias string, driver string, datasourcename string) (cn *DbConnection, err error) {
	if dbmngr.dbaliases == nil {
		dbmngr.dbaliases = make(map[string]*DbConnection)
	}
	if cn, _ = dbmngr.dbaliases[alias]; cn == nil {
		if cn, err = openConnection(dbmngr, driver, datasourcename); err == nil {
			dbmngr.dbaliases[alias] = cn
		}
	}
	return cn, err
}

//RegisterConnection a db connection refer to DbManager.Open
func (dbmngr *DbManager) RegisterConnection(alias string, driver string, datasourcename string) (err error) {
	_, err = dbmngr.Open(alias, driver, datasourcename)
	return err
}

var dbmngr *DbManager
var dbmngrlock = &sync.Mutex{}

//DatabaseManager global instance of DbManager
func DatabaseManager() *DbManager {
	if dbmngr == nil {
		dbmngrlock.Lock()
		if dbmngr == nil {
			dbmngr = &DbManager{}
		}
		dbmngrlock.Unlock()
	}
	return dbmngr
}

//Connection returns aliased DbConnection
func (dbmngr *DbManager) Connection(alias string) (cn *DbConnection) {
	if len(dbmngr.dbaliases) > 0 {
		cn, _ = dbmngr.dbaliases[alias]
	}
	return cn
}

//OutputResultSet - helper method that output res *DbResultSet to the following formats into a io.Writer
//contentext=.js => javascript
//contentext=.json => json
//contentext=.csv => .csv
func OutputResultSet(w io.Writer, name string, contentext string, res *DbResultSet, err error, setting ...string) {
	var out *IORW
	if outr, owok := w.(*IORW); owok {
		out = outr
	} else {
		out, _ = NewIORW(w)
	}
	if err == nil {
		if contentext == ".js" || contentext == ".json" {
			if contentext == ".js" {
				out.Print("var dataset_" + name + "=")
			}
			out.Print("[")
		} else if contentext == ".csv" {
		}
		for ci, col := range res.MetaData().Columns() {
			if contentext == ".js" || contentext == ".json" {
				if ci == 0 {
					out.Print("[")
				}
				if strings.Index(col, "\"") > -1 || strings.Index(col, "'") > -1 {
					if strings.Index(col, "\"") > -1 {
						col = strings.Replace(col, "\"", "\\\"", -1)
					}
					if strings.Index(col, "'") > -1 {
						col = strings.Replace(col, "'", "\\'", -1)
					}
				}
				out.Print("\"" + col + "\"")
				if ci == len(res.MetaData().Columns())-1 {
					out.Print("]")
				} else {
					out.Print(",")
				}
			} else if contentext == ".csv" {
				if strings.Index(col, "\"") > -1 || strings.Index(col, "'") > -1 {
					if strings.Index(col, "\"") > -1 {
						col = strings.Replace(col, "\"", "\"\"", -1)
					}
				}
				out.Print("\"" + col + "\"")
				if ci == len(res.MetaData().Columns())-1 {
					out.Print("\r\n")
				} else {
					out.Print(",")
				}
			}
		}
		if contentext == ".js" || contentext == ".json" {
			out.Print(",[")
		}
		foundResc := false
		for {
			if next, nexterr := res.Next(); next {
				for n, a := range res.Data() {
					if contentext == ".js" || contentext == ".json" {
						if n == 0 {
							if foundResc {
								out.Print(",[")
							} else {
								out.Print("[")
							}
						}
						if a == nil {
							if contentext == ".json" {
								out.Print("null")
							} else {
								out.Print("\"\"")
							}
						} else if a == "" {
							out.Print("\"\"")
						} else {
							if sa, oks := a.(string); oks {
								if res.MetaData().ColumnTypes()[n].Numeric() {
									out.Print(sa)
								} else {
									out.Print("\"", strings.Replace(strings.Replace(sa, "\"", "\\\"", -1), "'", "\\'", -1), "\"")
								}
							} else {
								out.Print(a)
							}
						}
						if n == len(res.MetaData().Columns())-1 {
							out.Print("]")
						} else {
							out.Print(",")
						}
					} else if contentext == ".csv" {
						if a != nil {
							if sa, oks := a.(string); oks {
								if sa != "" {
									if res.MetaData().ColumnTypes()[n].Numeric() {
										out.Print(sa)
									} else {
										out.Print("\"", strings.TrimSpace(strings.Replace(sa, "\"", "\"\"", -1)), "\"")
									}
								}
							}
						} else {
							out.Print(a)
						}

						if n == len(res.MetaData().Columns())-1 {
							out.Print("\r\n")
						} else {
							out.Print(",")
						}
					}
				}
				if !foundResc {
					foundResc = true
				}
			} else {
				if nexterr != nil {
					out.Print(nexterr.Error())
				}
				break
			}
		}
		if contentext == ".js" || contentext == ".json" {
			out.Print("]]")
			if contentext == ".js" {
				out.Println(";")
			}
		} else if contentext == "csv" {
		}

	} else {
		out.Print(err.Error())
	}
	if out != nil {
		if _, owok := w.(*IORW); owok {
			out.Close()
		}
	}
	out = nil
}

//DBExecuted controller
type DBExecuted struct {
	LastInsertId int64
	RowsAffected int64
	Err          error
}

//Execute execute query for alias connection
//return a DbExecute controller that represents the outcome of the executed request
func (dbmngr *DbManager) Execute(alias string, query string, args ...interface{}) (dbexecuted *DBExecuted) {
	if cn := dbmngr.Connection(alias); cn != nil {
		dbexecuted = &DBExecuted{}
		dbexecuted.LastInsertId, dbexecuted.RowsAffected, dbexecuted.Err = cn.Execute(query, args...)
	}
	return dbexecuted
}

//DBQuery DBQuery controller
type DBQuery struct {
	RSet         *DbResultSet
	readColsFunc ReadColumnsFunc
	readRwFunc   ReadRowFunc
	prcssFunc    ProcessingFunc
	Err          error
}

//ReadColumnsFunc definition
type ReadColumnsFunc = func(dbqry *DBQuery, columns []string, columntypes []*ColumnType)

//ReadRowFunc definition
type ReadRowFunc = func(dbqry *DBQuery, data []interface{}, firstRec bool, lastRec bool)

//ProcessingFunc definition
type ProcessingFunc = func(dbqry *DBQuery, stage QueryStage, a ...interface{})

//QueryStage stage
type QueryStage int

var qryStages = []string{"STARTED", "READING-COLUMNS", "COMPLETED-READING-COLUMNS", "READING-ROWS", "COMPLETED-READING-ROWS", "COMPLETED"}

func (qrystage QueryStage) String() (s string) {
	if qrystage > 0 && qrystage <= QueryStage(len(qryStages)) {
		s = qryStages[qrystage-1]
	} else {
		s = "UNKOWN"
	}
	return
}

//Process reading Columns then reading rows one by one till eof and finally indicate done processing
func (dbqry *DBQuery) Process() (err error) {
	var didProcess = false

	var columns = dbqry.MetaData().Columns()
	var coltypes = dbqry.MetaData().ColumnTypes()
	if dbqry.prcssFunc != nil {
		if !didProcess {
			didProcess = true
		}
		dbqry.prcssFunc(dbqry, 1)
	}
	if dbqry.prcssFunc != nil {
		if !didProcess {
			didProcess = true
		}
		dbqry.prcssFunc(dbqry, 2)
	}
	if dbqry.readColsFunc != nil {
		if !didProcess {
			didProcess = true
		}
		dbqry.readColsFunc(dbqry, columns, coltypes)
	}

	if dbqry.prcssFunc != nil {
		if !didProcess {
			didProcess = true
		}
		dbqry.prcssFunc(dbqry, 3)
	}

	if dbqry.prcssFunc != nil {
		if !didProcess {
			didProcess = true
		}
		dbqry.prcssFunc(dbqry, 4)
	}
	if dbqry.readRwFunc != nil {
		if !didProcess {
			didProcess = true
		}

		var hasRows = dbqry.Next()
		var rdata []interface{}
		var firstRow = true
		for hasRows {
			rdata = dbqry.Data()
			hasRows = dbqry.Next()
			dbqry.readRwFunc(dbqry, rdata, firstRow, !hasRows)
			if firstRow {
				firstRow = false
			}
		}
	}
	if dbqry.prcssFunc != nil {
		if !didProcess {
			didProcess = true
		}
		dbqry.prcssFunc(dbqry, 5, columns, coltypes)
	}

	if columns != nil {
		columns = nil
	}

	if coltypes != nil {
		coltypes = nil
	}

	if dbqry.prcssFunc != nil {
		if !didProcess {
			didProcess = true
		}
		dbqry.prcssFunc(dbqry, 6)
	}

	if didProcess {
		if dbqry.RSet != nil {
			err = dbqry.RSet.Close()
			dbqry.RSet = nil
		}
	}
	return
}

func DbFormatColDelimSettings(coldelim ...string) (readsettings map[string]string) {
	readsettings["format-type"] = "csv"
	readsettings["text-par"] = "\""
	if coldelim != nil && len(coldelim) == 1 {
		readsettings["col-sep"] = coldelim[0]
	} else {
		readsettings["col-sep"] = ","
	}
	readsettings["row-sep"] = "\r\n"
	return
}

var dbReadFormats = map[string]func() map[string]string{}
var dbReadFormattingFuncs = map[string]func(map[string]string, *DbResultSet, io.Writer) error{}

var dbReadFormatsLock = &sync.RWMutex{}

func (dbqry *DBQuery) ReadAllCustom(w io.Writer, settings map[string]string, formatFunction func(map[string]string, *DbResultSet, io.Writer) error) {
	if formatFunction != nil && settings != nil && len(settings) > 0 && w != nil {
		formatFunction(settings, dbqry.RSet, w)
	}
}

func (dbqry *DBQuery) ReadAll(w io.Writer, format string) {
	if w == nil || format == "" {
		return
	}
	dbReadFormatsLock.RLock()
	var settings func() map[string]string
	if settings = dbReadFormats[format]; settings != nil {
		var formatFunction = dbReadFormattingFuncs[format]
		if formatFunction != nil {
			defer dbReadFormatsLock.RUnlock()
			formatFunction(settings(), dbqry.RSet, w)
			formatFunction = nil
		} else {
			dbReadFormatsLock.RUnlock()
		}
		settings = nil
	} else {
		dbReadFormatsLock.RUnlock()
	}
}

func RegisterDbReadFormat(formatname string, settings map[string]string, formatFunction func(map[string]string, *DbResultSet, io.Writer) error) {
	if formatname != "" && settings != nil && len(settings) > 0 && formatFunction != nil {
		dbReadFormatsLock.RLock()
		defer dbReadFormatsLock.RUnlock()
	}
}

//PrintResult [refer to OutputResultSet] - helper method that output res *DbResultSet to the following formats into a io.Writer
//contentext=.js => javascript
//contentext=.json => json
//contentext=.csv => .csv
func (dbqry *DBQuery) PrintResult(out *IORW, name string, contentext string, setting ...string) {
	OutputResultSet(out, name, contentext, dbqry.RSet, dbqry.Err, setting...)
}

//MetaData return a DbResultSetMetaData object of the resultset that is wrapped by this DBQuery controller
func (dbqry *DBQuery) MetaData() *DbResultSetMetaData {
	if dbqry.RSet == nil {
		return nil
	}
	return dbqry.RSet.MetaData()
}

//Data returns an array if data of the current row from the underlying resultset
func (dbqry *DBQuery) Data() []interface{} {
	if dbqry.RSet == nil {
		return nil
	}
	return dbqry.RSet.Data()
}

//Next execute the Next record method of the underlying resultset
func (dbqry *DBQuery) Next() bool {
	if dbqry.RSet == nil {
		return false
	}
	next, err := dbqry.RSet.Next()
	if err != nil {
		dbqry.Err = err
		dbqry.RSet = nil
	}
	return next
}

//Query query aliased connection and returns a DBQuery controller for the underlying resultset
func (dbmngr *DbManager) Query(alias string, query string, args ...interface{}) (dbquery *DBQuery) {
	if cn := dbmngr.Connection(alias); cn != nil {
		var rdColsFunc ReadColumnsFunc
		var rdRowFunc ReadRowFunc
		var prcessFunc ProcessingFunc
		var foundOk = false
		if len(args) > 0 {
			var n = 0
			for n < len(args) {
				if rdColsFunc == nil {
					if rdColsFunc, foundOk = args[n].(ReadColumnsFunc); foundOk {
						if len(args) > 1 {
							args = append(args[:n], args[n+1:]...)
						} else {
							args = nil
						}
					} else {
						n++
					}
				} else if rdRowFunc == nil {
					if rdRowFunc, foundOk = args[n].(ReadRowFunc); foundOk {
						if len(args) > 1 {
							args = append(args[:n], args[n+1:]...)
						} else {
							args = nil
						}
					} else {
						n++
					}
				} else if prcessFunc == nil {
					if prcessFunc, foundOk = args[n].(ProcessingFunc); foundOk {
						if len(args) > 1 {
							args = append(args[:n], args[n+1:]...)
						} else {
							args = nil
						}
					} else {
						n++
					}
				} else {
					n++
				}
			}
		}
		dbquery = &DBQuery{readColsFunc: rdColsFunc, readRwFunc: rdRowFunc, prcssFunc: prcessFunc}

		dbquery.RSet, dbquery.Err = cn.Query(query, args...)
	}
	return dbquery
}
