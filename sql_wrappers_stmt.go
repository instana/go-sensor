// (c) Copyright IBM Corp. 2022
package instana

import "database/sql/driver"

func wrapStmt(stmt driver.Stmt, query string, connDetails dbConnDetails, sensor *Sensor) driver.Stmt {
	Stmt, isStmt := stmt.(driver.Stmt)
	StmtExecContext, isStmtExecContext := stmt.(driver.StmtExecContext)
	StmtQueryContext, isStmtQueryContext := stmt.(driver.StmtQueryContext)
	NamedValueChecker, isNamedValueChecker := stmt.(driver.NamedValueChecker)
	ColumnConverter, isColumnConverter := stmt.(driver.ColumnConverter)
	if isStmt && isStmtExecContext && isStmtQueryContext && isNamedValueChecker && isColumnConverter {
		return &w_stmt_Stmt_StmtExecContext_StmtQueryContext_NamedValueChecker_ColumnConverter{
			Stmt: &wStmt{
				Stmt:        Stmt,
				connDetails: connDetails,
				query:       query,
				sensor:      sensor,
			},
			StmtExecContext: &wStmtExecContext{
				StmtExecContext: StmtExecContext,
				connDetails:     connDetails,
				query:           query,
				sensor:          sensor,
			},
			StmtQueryContext: &wStmtQueryContext{
				StmtQueryContext: StmtQueryContext,
				connDetails:      connDetails,
				query:            query,
				sensor:           sensor,
			},
			NamedValueChecker: NamedValueChecker,
			ColumnConverter:   ColumnConverter,
		}
	}
	if isStmt && isStmtExecContext && isStmtQueryContext && isNamedValueChecker {
		return &w_stmt_Stmt_StmtExecContext_StmtQueryContext_NamedValueChecker{
			Stmt: &wStmt{
				Stmt:        Stmt,
				connDetails: connDetails,
				query:       query,
				sensor:      sensor,
			},
			StmtExecContext: &wStmtExecContext{
				StmtExecContext: StmtExecContext,
				connDetails:     connDetails,
				query:           query,
				sensor:          sensor,
			},
			StmtQueryContext: &wStmtQueryContext{
				StmtQueryContext: StmtQueryContext,
				connDetails:      connDetails,
				query:            query,
				sensor:           sensor,
			},
			NamedValueChecker: NamedValueChecker,
		}
	}
	if isStmt && isStmtExecContext && isStmtQueryContext && isColumnConverter {
		return &w_stmt_Stmt_StmtExecContext_StmtQueryContext_ColumnConverter{
			Stmt: &wStmt{
				Stmt:        Stmt,
				connDetails: connDetails,
				query:       query,
				sensor:      sensor,
			},
			StmtExecContext: &wStmtExecContext{
				StmtExecContext: StmtExecContext,
				connDetails:     connDetails,
				query:           query,
				sensor:          sensor,
			},
			StmtQueryContext: &wStmtQueryContext{
				StmtQueryContext: StmtQueryContext,
				connDetails:      connDetails,
				query:            query,
				sensor:           sensor,
			},
			ColumnConverter: ColumnConverter,
		}
	}
	if isStmt && isStmtExecContext && isNamedValueChecker && isColumnConverter {
		return &w_stmt_Stmt_StmtExecContext_NamedValueChecker_ColumnConverter{
			Stmt: &wStmt{
				Stmt:        Stmt,
				connDetails: connDetails,
				query:       query,
				sensor:      sensor,
			},
			StmtExecContext: &wStmtExecContext{
				StmtExecContext: StmtExecContext,
				connDetails:     connDetails,
				query:           query,
				sensor:          sensor,
			},
			NamedValueChecker: NamedValueChecker,
			ColumnConverter:   ColumnConverter,
		}
	}
	if isStmt && isStmtQueryContext && isNamedValueChecker && isColumnConverter {
		return &w_stmt_Stmt_StmtQueryContext_NamedValueChecker_ColumnConverter{
			Stmt: &wStmt{
				Stmt:        Stmt,
				connDetails: connDetails,
				query:       query,
				sensor:      sensor,
			},
			StmtQueryContext: &wStmtQueryContext{
				StmtQueryContext: StmtQueryContext,
				connDetails:      connDetails,
				query:            query,
				sensor:           sensor,
			},
			NamedValueChecker: NamedValueChecker,
			ColumnConverter:   ColumnConverter,
		}
	}
	if isStmt && isStmtExecContext && isStmtQueryContext {
		return &w_stmt_Stmt_StmtExecContext_StmtQueryContext{
			Stmt: &wStmt{
				Stmt:        Stmt,
				connDetails: connDetails,
				query:       query,
				sensor:      sensor,
			},
			StmtExecContext: &wStmtExecContext{
				StmtExecContext: StmtExecContext,
				connDetails:     connDetails,
				query:           query,
				sensor:          sensor,
			},
			StmtQueryContext: &wStmtQueryContext{
				StmtQueryContext: StmtQueryContext,
				connDetails:      connDetails,
				query:            query,
				sensor:           sensor,
			},
		}
	}
	if isStmt && isStmtExecContext && isNamedValueChecker {
		return &w_stmt_Stmt_StmtExecContext_NamedValueChecker{
			Stmt: &wStmt{
				Stmt:        Stmt,
				connDetails: connDetails,
				query:       query,
				sensor:      sensor,
			},
			StmtExecContext: &wStmtExecContext{
				StmtExecContext: StmtExecContext,
				connDetails:     connDetails,
				query:           query,
				sensor:          sensor,
			},
			NamedValueChecker: NamedValueChecker,
		}
	}
	if isStmt && isStmtExecContext && isColumnConverter {
		return &w_stmt_Stmt_StmtExecContext_ColumnConverter{
			Stmt: &wStmt{
				Stmt:        Stmt,
				connDetails: connDetails,
				query:       query,
				sensor:      sensor,
			},
			StmtExecContext: &wStmtExecContext{
				StmtExecContext: StmtExecContext,
				connDetails:     connDetails,
				query:           query,
				sensor:          sensor,
			},
			ColumnConverter: ColumnConverter,
		}
	}
	if isStmt && isStmtQueryContext && isNamedValueChecker {
		return &w_stmt_Stmt_StmtQueryContext_NamedValueChecker{
			Stmt: &wStmt{
				Stmt:        Stmt,
				connDetails: connDetails,
				query:       query,
				sensor:      sensor,
			},
			StmtQueryContext: &wStmtQueryContext{
				StmtQueryContext: StmtQueryContext,
				connDetails:      connDetails,
				query:            query,
				sensor:           sensor,
			},
			NamedValueChecker: NamedValueChecker,
		}
	}
	if isStmt && isStmtQueryContext && isColumnConverter {
		return &w_stmt_Stmt_StmtQueryContext_ColumnConverter{
			Stmt: &wStmt{
				Stmt:        Stmt,
				connDetails: connDetails,
				query:       query,
				sensor:      sensor,
			},
			StmtQueryContext: &wStmtQueryContext{
				StmtQueryContext: StmtQueryContext,
				connDetails:      connDetails,
				query:            query,
				sensor:           sensor,
			},
			ColumnConverter: ColumnConverter,
		}
	}
	if isStmt && isNamedValueChecker && isColumnConverter {
		return &w_stmt_Stmt_NamedValueChecker_ColumnConverter{
			Stmt: &wStmt{
				Stmt:        Stmt,
				connDetails: connDetails,
				query:       query,
				sensor:      sensor,
			},
			NamedValueChecker: NamedValueChecker,
			ColumnConverter:   ColumnConverter,
		}
	}
	if isStmt && isStmtExecContext {
		return &w_stmt_Stmt_StmtExecContext{
			Stmt: &wStmt{
				Stmt:        Stmt,
				connDetails: connDetails,
				query:       query,
				sensor:      sensor,
			},
			StmtExecContext: &wStmtExecContext{
				StmtExecContext: StmtExecContext,
				connDetails:     connDetails,
				query:           query,
				sensor:          sensor,
			},
		}
	}
	if isStmt && isStmtQueryContext {
		return &w_stmt_Stmt_StmtQueryContext{
			Stmt: &wStmt{
				Stmt:        Stmt,
				connDetails: connDetails,
				query:       query,
				sensor:      sensor,
			},
			StmtQueryContext: &wStmtQueryContext{
				StmtQueryContext: StmtQueryContext,
				connDetails:      connDetails,
				query:            query,
				sensor:           sensor,
			},
		}
	}
	if isStmt && isNamedValueChecker {
		return &w_stmt_Stmt_NamedValueChecker{
			Stmt: &wStmt{
				Stmt:        Stmt,
				connDetails: connDetails,
				query:       query,
				sensor:      sensor,
			},
			NamedValueChecker: NamedValueChecker,
		}
	}
	if isStmt && isColumnConverter {
		return &w_stmt_Stmt_ColumnConverter{
			Stmt: &wStmt{
				Stmt:        Stmt,
				connDetails: connDetails,
				query:       query,
				sensor:      sensor,
			},
			ColumnConverter: ColumnConverter,
		}
	}
	if isStmt {
		return &w_stmt_Stmt{
			Stmt: &wStmt{
				Stmt:        Stmt,
				connDetails: connDetails,
				query:       query,
				sensor:      sensor,
			},
		}
	}
	return stmt
}

type w_stmt_Stmt_StmtExecContext_StmtQueryContext_NamedValueChecker_ColumnConverter struct {
	driver.Stmt
	driver.StmtExecContext
	driver.StmtQueryContext
	driver.NamedValueChecker
	driver.ColumnConverter
}
type w_stmt_Stmt_StmtExecContext_StmtQueryContext_NamedValueChecker struct {
	driver.Stmt
	driver.StmtExecContext
	driver.StmtQueryContext
	driver.NamedValueChecker
}
type w_stmt_Stmt_StmtExecContext_StmtQueryContext_ColumnConverter struct {
	driver.Stmt
	driver.StmtExecContext
	driver.StmtQueryContext
	driver.ColumnConverter
}
type w_stmt_Stmt_StmtExecContext_NamedValueChecker_ColumnConverter struct {
	driver.Stmt
	driver.StmtExecContext
	driver.NamedValueChecker
	driver.ColumnConverter
}
type w_stmt_Stmt_StmtQueryContext_NamedValueChecker_ColumnConverter struct {
	driver.Stmt
	driver.StmtQueryContext
	driver.NamedValueChecker
	driver.ColumnConverter
}
type w_stmt_Stmt_StmtExecContext_StmtQueryContext struct {
	driver.Stmt
	driver.StmtExecContext
	driver.StmtQueryContext
}
type w_stmt_Stmt_StmtExecContext_NamedValueChecker struct {
	driver.Stmt
	driver.StmtExecContext
	driver.NamedValueChecker
}
type w_stmt_Stmt_StmtExecContext_ColumnConverter struct {
	driver.Stmt
	driver.StmtExecContext
	driver.ColumnConverter
}
type w_stmt_Stmt_StmtQueryContext_NamedValueChecker struct {
	driver.Stmt
	driver.StmtQueryContext
	driver.NamedValueChecker
}
type w_stmt_Stmt_StmtQueryContext_ColumnConverter struct {
	driver.Stmt
	driver.StmtQueryContext
	driver.ColumnConverter
}
type w_stmt_Stmt_NamedValueChecker_ColumnConverter struct {
	driver.Stmt
	driver.NamedValueChecker
	driver.ColumnConverter
}
type w_stmt_Stmt_StmtExecContext struct {
	driver.Stmt
	driver.StmtExecContext
}
type w_stmt_Stmt_StmtQueryContext struct {
	driver.Stmt
	driver.StmtQueryContext
}
type w_stmt_Stmt_NamedValueChecker struct {
	driver.Stmt
	driver.NamedValueChecker
}
type w_stmt_Stmt_ColumnConverter struct {
	driver.Stmt
	driver.ColumnConverter
}
type w_stmt_Stmt struct {
	driver.Stmt
}

func stmtAlreadyWrapped(stmt driver.Stmt) bool {
	switch stmt.(type) {
	case *w_stmt_Stmt_StmtExecContext_StmtQueryContext_NamedValueChecker_ColumnConverter, *w_stmt_Stmt_StmtExecContext_StmtQueryContext_NamedValueChecker, *w_stmt_Stmt_StmtExecContext_StmtQueryContext_ColumnConverter, *w_stmt_Stmt_StmtExecContext_NamedValueChecker_ColumnConverter, *w_stmt_Stmt_StmtQueryContext_NamedValueChecker_ColumnConverter, *w_stmt_Stmt_StmtExecContext_StmtQueryContext, *w_stmt_Stmt_StmtExecContext_NamedValueChecker, *w_stmt_Stmt_StmtExecContext_ColumnConverter, *w_stmt_Stmt_StmtQueryContext_NamedValueChecker, *w_stmt_Stmt_StmtQueryContext_ColumnConverter, *w_stmt_Stmt_NamedValueChecker_ColumnConverter, *w_stmt_Stmt_StmtExecContext, *w_stmt_Stmt_StmtQueryContext, *w_stmt_Stmt_NamedValueChecker, *w_stmt_Stmt_ColumnConverter, *w_stmt_Stmt:
		return true
	}
	return false
}
