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
	RSet *DbResultSet
	Err  error
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
		dbquery = &DBQuery{}
		dbquery.RSet, dbquery.Err = cn.Query(query, args...)
	}
	return dbquery
}
