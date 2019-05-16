package lnksworks

import (
	"context"
	"database/sql"
	"database/sql/driver"
)

type ActiveDBDriver struct {
	atvdbcns   map[string]*ActiveDBConn
	atvdbcntrs map[string]*ActiveDBConnector
}

type ActiveDBConn struct {
	atvdbdrvr *ActiveDBDriver
	name      string
}

//driver.Conn interface methods

func (atvdbcn *ActiveDBConn) Prepare(query string) (stmnt driver.Stmt, err error) {
	return
}

func (atvdbcn *ActiveDBConn) Close() (err error) {

	return
}
func (atvdbcn *ActiveDBConn) Begin() (tx driver.Tx, er error) {

	return
}

func (atvdbdrvr *ActiveDBDriver) Open(name string) (cn driver.Conn, err error) {
	if atvdbdrvr.atvdbcns == nil {
		atvdbdrvr.atvdbcns = map[string]*ActiveDBConn{}
		var atvdbcn = &ActiveDBConn{atvdbdrvr: atvdbdrvr, name: name}
		atvdbdrvr.atvdbcns[name] = atvdbcn
		cn = atvdbcn
	} else {
		if existingcntr, exists := atvdbdrvr.atvdbcns[name]; exists {
			cn = existingcntr
		} else {
			var atvdbcn = &ActiveDBConn{atvdbdrvr: atvdbdrvr, name: name}
			atvdbdrvr.atvdbcns[name] = atvdbcn
			cn = atvdbcn
		}
	}
	return
}

type ActiveDBConnector struct {
	atvdbdrvr *ActiveDBDriver
	name      string
}

func (atvdbcntr *ActiveDBConnector) Connect(cntx context.Context) (cn driver.Conn, err error) {
	cn, err = atvdbcntr.atvdbdrvr.Open(atvdbcntr.name)
	return
}

func (atvdbcntr *ActiveDBConnector) Driver() driver.Driver {
	return atvdbcntr.atvdbdrvr
}

func (atvdbdrvr *ActiveDBDriver) OpenConnector(name string) (cntr driver.Connector, err error) {
	var atvdbcntr = &ActiveDBConnector{atvdbdrvr: atvdbdrvr, name: name}

	if atvdbdrvr.atvdbcntrs == nil {
		atvdbdrvr.atvdbcntrs = map[string]*ActiveDBConnector{}
	}
	atvdbdrvr.atvdbcntrs[name] = atvdbcntr
	cntr = atvdbcntr
	return
}

func newActiveDBDriver(a ...interface{}) (atvdbdrvr *ActiveDBDriver) {
	atvdbdrvr = &ActiveDBDriver{}
	return
}

func init() {
	sql.Register("activedb", newActiveDBDriver())
}
