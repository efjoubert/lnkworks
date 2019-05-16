package lnksworks

import (
	"database/sql"
	"fmt"
)

//DbStatement container representing the underlying DbConnection and allocated sql.Tx transaction
type DbStatement struct {
	cn *DbConnection
	tx *sql.Tx
}

//NewDbStatement invoke a new DbStatement from the DbConnection
func NewDbStatement(cn *DbConnection) (stmnt *DbStatement, err error) {
	if err = cn.db.Ping(); err == nil {
		stmnt = &DbStatement{cn: cn}
	}
	return stmnt, err
}

//Begin Invoke a transaction, sql.Tx, from the for ths DbStatement
func (stmnt *DbStatement) Begin() (err error) {
	if tx, txerr := stmnt.cn.db.Begin(); txerr == nil {
		stmnt.tx = tx
	} else {
		err = txerr
	}
	return err
}

//Execute execute a none DbResultSet query
//Usually used for statement like update and insert, or executing db procedures that dont return a record
//In the case of Insert or Update if the underlying db driver
//return the lastInsertID and rowsAffected if supported
func (stmnt *DbStatement) Execute(query string, args ...interface{}) (lastInsertID int64, rowsAffected int64, err error) {
	if stmnt.tx == nil {
		err = stmnt.Begin()
	}
	if err == nil {
		if r, rerr := stmnt.tx.Exec(query, args...); rerr == nil {
			lastInsertID, err = r.LastInsertId()
			rowsAffected, err = r.RowsAffected()
			r = nil
			err = stmnt.tx.Commit()
		} else {
			err = rerr
		}
	}
	if err != nil {
		err = stmnt.tx.Rollback()
	}
	return lastInsertID, rowsAffected, err
}

//Query and return a DbResultSet
func (stmnt *DbStatement) Query(query string, args ...interface{}) (rset *DbResultSet, err error) {
	if stmnt.tx == nil {
		err = stmnt.Begin()
	}
	if rs, rserr := stmnt.tx.Query(query, args...); rserr == nil {
		if cols, colserr := rs.Columns(); colserr == nil {
			for n, col := range cols {
				if col == "" {
					cols[n] = "COLUMN" + fmt.Sprint(n+1)
				}
			}
			if coltypes, coltypeserr := rs.ColumnTypes(); coltypeserr == nil {
				rset = &DbResultSet{rset: rs, stmnt: stmnt, rsmetadata: &DbResultSetMetaData{cols: cols[:], colTypes: columnTypes(coltypes[:])}, dosomething: make(chan bool, 1)}
			} else {
				err = coltypeserr
			}
		} else {
			err = colserr
		}
	} else {
		stmnt.tx.Rollback()
		err = rserr
	}
	return rset, err
}

//Close the allocated transaction, sql.Tx associated to this DbStatement
//It will by default perform a commit before releasing the transaction reference
func (stmnt *DbStatement) Close() (err error) {
	if stmnt.tx != nil {
		err = stmnt.tx.Commit()
		stmnt.tx = nil
	}
	if stmnt.cn != nil {
		stmnt.cn = nil
	}
	return
}

func columnTypes(sqlcoltypes []*sql.ColumnType) (coltypes []*ColumnType) {
	coltypes = make([]*ColumnType, len(sqlcoltypes))
	for n, ctype := range sqlcoltypes {
		coltype := &ColumnType{}
		coltype.databaseType = ctype.DatabaseTypeName()
		coltype.length, coltype.hasLength = ctype.Length()
		coltype.name = ctype.Name()
		coltype.databaseType = ctype.DatabaseTypeName()
		coltype.nullable, coltype.hasNullable = ctype.Nullable()
		coltype.precision, coltype.scale, coltype.hasPrecisionScale = ctype.DecimalSize()
		coltype.scanType = ctype.ScanType()
		coltypes[n] = coltype
	}
	return coltypes
}
