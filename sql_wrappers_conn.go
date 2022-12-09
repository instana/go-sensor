// (c) Copyright IBM Corp. 2022
package instana

import "database/sql/driver"

func wrapConn(connDetails dbConnDetails, conn driver.Conn, sensor *Sensor) driver.Conn {
	Conn, isConn := conn.(driver.Conn)
	Execer, isExecer := conn.(driver.Execer)
	ExecerContext, isExecerContext := conn.(driver.ExecerContext)
	Queryer, isQueryer := conn.(driver.Queryer)
	QueryerContext, isQueryerContext := conn.(driver.QueryerContext)
	ConnPrepareContext, isConnPrepareContext := conn.(driver.ConnPrepareContext)
	NamedValueChecker, isNamedValueChecker := conn.(driver.NamedValueChecker)
	ColumnConverter, isColumnConverter := conn.(driver.ColumnConverter)
	if isConn && isExecer && isExecerContext && isQueryer && isQueryerContext && isConnPrepareContext && isNamedValueChecker && isColumnConverter {
		return &w_conn_Conn_Execer_ExecerContext_Queryer_QueryerContext_ConnPrepareContext_NamedValueChecker_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
			NamedValueChecker: NamedValueChecker,
			ColumnConverter:   ColumnConverter,
		}
	}
	if isConn && isExecer && isExecerContext && isQueryer && isQueryerContext && isConnPrepareContext && isNamedValueChecker {
		return &w_conn_Conn_Execer_ExecerContext_Queryer_QueryerContext_ConnPrepareContext_NamedValueChecker{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
			NamedValueChecker: NamedValueChecker,
		}
	}
	if isConn && isExecer && isExecerContext && isQueryer && isQueryerContext && isConnPrepareContext && isColumnConverter {
		return &w_conn_Conn_Execer_ExecerContext_Queryer_QueryerContext_ConnPrepareContext_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
			ColumnConverter: ColumnConverter,
		}
	}
	if isConn && isExecer && isExecerContext && isQueryer && isQueryerContext && isNamedValueChecker && isColumnConverter {
		return &w_conn_Conn_Execer_ExecerContext_Queryer_QueryerContext_NamedValueChecker_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
			NamedValueChecker: NamedValueChecker,
			ColumnConverter:   ColumnConverter,
		}
	}
	if isConn && isExecer && isExecerContext && isQueryer && isConnPrepareContext && isNamedValueChecker && isColumnConverter {
		return &w_conn_Conn_Execer_ExecerContext_Queryer_ConnPrepareContext_NamedValueChecker_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
			NamedValueChecker: NamedValueChecker,
			ColumnConverter:   ColumnConverter,
		}
	}
	if isConn && isExecer && isExecerContext && isQueryerContext && isConnPrepareContext && isNamedValueChecker && isColumnConverter {
		return &w_conn_Conn_Execer_ExecerContext_QueryerContext_ConnPrepareContext_NamedValueChecker_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
			NamedValueChecker: NamedValueChecker,
			ColumnConverter:   ColumnConverter,
		}
	}
	if isConn && isExecer && isQueryer && isQueryerContext && isConnPrepareContext && isNamedValueChecker && isColumnConverter {
		return &w_conn_Conn_Execer_Queryer_QueryerContext_ConnPrepareContext_NamedValueChecker_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
			NamedValueChecker: NamedValueChecker,
			ColumnConverter:   ColumnConverter,
		}
	}
	if isConn && isExecerContext && isQueryer && isQueryerContext && isConnPrepareContext && isNamedValueChecker && isColumnConverter {
		return &w_conn_Conn_ExecerContext_Queryer_QueryerContext_ConnPrepareContext_NamedValueChecker_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
			NamedValueChecker: NamedValueChecker,
			ColumnConverter:   ColumnConverter,
		}
	}
	if isConn && isExecer && isExecerContext && isQueryer && isQueryerContext && isConnPrepareContext {
		return &w_conn_Conn_Execer_ExecerContext_Queryer_QueryerContext_ConnPrepareContext{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
		}
	}
	if isConn && isExecer && isExecerContext && isQueryer && isQueryerContext && isNamedValueChecker {
		return &w_conn_Conn_Execer_ExecerContext_Queryer_QueryerContext_NamedValueChecker{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
			NamedValueChecker: NamedValueChecker,
		}
	}
	if isConn && isExecer && isExecerContext && isQueryer && isQueryerContext && isColumnConverter {
		return &w_conn_Conn_Execer_ExecerContext_Queryer_QueryerContext_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
			ColumnConverter: ColumnConverter,
		}
	}
	if isConn && isExecer && isExecerContext && isQueryer && isConnPrepareContext && isNamedValueChecker {
		return &w_conn_Conn_Execer_ExecerContext_Queryer_ConnPrepareContext_NamedValueChecker{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
			NamedValueChecker: NamedValueChecker,
		}
	}
	if isConn && isExecer && isExecerContext && isQueryer && isConnPrepareContext && isColumnConverter {
		return &w_conn_Conn_Execer_ExecerContext_Queryer_ConnPrepareContext_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
			ColumnConverter: ColumnConverter,
		}
	}
	if isConn && isExecer && isExecerContext && isQueryer && isNamedValueChecker && isColumnConverter {
		return &w_conn_Conn_Execer_ExecerContext_Queryer_NamedValueChecker_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			NamedValueChecker: NamedValueChecker,
			ColumnConverter:   ColumnConverter,
		}
	}
	if isConn && isExecer && isExecerContext && isQueryerContext && isConnPrepareContext && isNamedValueChecker {
		return &w_conn_Conn_Execer_ExecerContext_QueryerContext_ConnPrepareContext_NamedValueChecker{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
			NamedValueChecker: NamedValueChecker,
		}
	}
	if isConn && isExecer && isExecerContext && isQueryerContext && isConnPrepareContext && isColumnConverter {
		return &w_conn_Conn_Execer_ExecerContext_QueryerContext_ConnPrepareContext_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
			ColumnConverter: ColumnConverter,
		}
	}
	if isConn && isExecer && isExecerContext && isQueryerContext && isNamedValueChecker && isColumnConverter {
		return &w_conn_Conn_Execer_ExecerContext_QueryerContext_NamedValueChecker_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
			NamedValueChecker: NamedValueChecker,
			ColumnConverter:   ColumnConverter,
		}
	}
	if isConn && isExecer && isExecerContext && isConnPrepareContext && isNamedValueChecker && isColumnConverter {
		return &w_conn_Conn_Execer_ExecerContext_ConnPrepareContext_NamedValueChecker_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
			NamedValueChecker: NamedValueChecker,
			ColumnConverter:   ColumnConverter,
		}
	}
	if isConn && isExecer && isQueryer && isQueryerContext && isConnPrepareContext && isNamedValueChecker {
		return &w_conn_Conn_Execer_Queryer_QueryerContext_ConnPrepareContext_NamedValueChecker{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
			NamedValueChecker: NamedValueChecker,
		}
	}
	if isConn && isExecer && isQueryer && isQueryerContext && isConnPrepareContext && isColumnConverter {
		return &w_conn_Conn_Execer_Queryer_QueryerContext_ConnPrepareContext_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
			ColumnConverter: ColumnConverter,
		}
	}
	if isConn && isExecer && isQueryer && isQueryerContext && isNamedValueChecker && isColumnConverter {
		return &w_conn_Conn_Execer_Queryer_QueryerContext_NamedValueChecker_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
			NamedValueChecker: NamedValueChecker,
			ColumnConverter:   ColumnConverter,
		}
	}
	if isConn && isExecer && isQueryer && isConnPrepareContext && isNamedValueChecker && isColumnConverter {
		return &w_conn_Conn_Execer_Queryer_ConnPrepareContext_NamedValueChecker_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
			NamedValueChecker: NamedValueChecker,
			ColumnConverter:   ColumnConverter,
		}
	}
	if isConn && isExecer && isQueryerContext && isConnPrepareContext && isNamedValueChecker && isColumnConverter {
		return &w_conn_Conn_Execer_QueryerContext_ConnPrepareContext_NamedValueChecker_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
			NamedValueChecker: NamedValueChecker,
			ColumnConverter:   ColumnConverter,
		}
	}
	if isConn && isExecerContext && isQueryer && isQueryerContext && isConnPrepareContext && isNamedValueChecker {
		return &w_conn_Conn_ExecerContext_Queryer_QueryerContext_ConnPrepareContext_NamedValueChecker{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
			NamedValueChecker: NamedValueChecker,
		}
	}
	if isConn && isExecerContext && isQueryer && isQueryerContext && isConnPrepareContext && isColumnConverter {
		return &w_conn_Conn_ExecerContext_Queryer_QueryerContext_ConnPrepareContext_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
			ColumnConverter: ColumnConverter,
		}
	}
	if isConn && isExecerContext && isQueryer && isQueryerContext && isNamedValueChecker && isColumnConverter {
		return &w_conn_Conn_ExecerContext_Queryer_QueryerContext_NamedValueChecker_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
			NamedValueChecker: NamedValueChecker,
			ColumnConverter:   ColumnConverter,
		}
	}
	if isConn && isExecerContext && isQueryer && isConnPrepareContext && isNamedValueChecker && isColumnConverter {
		return &w_conn_Conn_ExecerContext_Queryer_ConnPrepareContext_NamedValueChecker_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
			NamedValueChecker: NamedValueChecker,
			ColumnConverter:   ColumnConverter,
		}
	}
	if isConn && isExecerContext && isQueryerContext && isConnPrepareContext && isNamedValueChecker && isColumnConverter {
		return &w_conn_Conn_ExecerContext_QueryerContext_ConnPrepareContext_NamedValueChecker_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
			NamedValueChecker: NamedValueChecker,
			ColumnConverter:   ColumnConverter,
		}
	}
	if isConn && isQueryer && isQueryerContext && isConnPrepareContext && isNamedValueChecker && isColumnConverter {
		return &w_conn_Conn_Queryer_QueryerContext_ConnPrepareContext_NamedValueChecker_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
			NamedValueChecker: NamedValueChecker,
			ColumnConverter:   ColumnConverter,
		}
	}
	if isConn && isExecer && isExecerContext && isQueryer && isQueryerContext {
		return &w_conn_Conn_Execer_ExecerContext_Queryer_QueryerContext{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
		}
	}
	if isConn && isExecer && isExecerContext && isQueryer && isConnPrepareContext {
		return &w_conn_Conn_Execer_ExecerContext_Queryer_ConnPrepareContext{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
		}
	}
	if isConn && isExecer && isExecerContext && isQueryer && isNamedValueChecker {
		return &w_conn_Conn_Execer_ExecerContext_Queryer_NamedValueChecker{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			NamedValueChecker: NamedValueChecker,
		}
	}
	if isConn && isExecer && isExecerContext && isQueryer && isColumnConverter {
		return &w_conn_Conn_Execer_ExecerContext_Queryer_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ColumnConverter: ColumnConverter,
		}
	}
	if isConn && isExecer && isExecerContext && isQueryerContext && isConnPrepareContext {
		return &w_conn_Conn_Execer_ExecerContext_QueryerContext_ConnPrepareContext{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
		}
	}
	if isConn && isExecer && isExecerContext && isQueryerContext && isNamedValueChecker {
		return &w_conn_Conn_Execer_ExecerContext_QueryerContext_NamedValueChecker{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
			NamedValueChecker: NamedValueChecker,
		}
	}
	if isConn && isExecer && isExecerContext && isQueryerContext && isColumnConverter {
		return &w_conn_Conn_Execer_ExecerContext_QueryerContext_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
			ColumnConverter: ColumnConverter,
		}
	}
	if isConn && isExecer && isExecerContext && isConnPrepareContext && isNamedValueChecker {
		return &w_conn_Conn_Execer_ExecerContext_ConnPrepareContext_NamedValueChecker{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
			NamedValueChecker: NamedValueChecker,
		}
	}
	if isConn && isExecer && isExecerContext && isConnPrepareContext && isColumnConverter {
		return &w_conn_Conn_Execer_ExecerContext_ConnPrepareContext_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
			ColumnConverter: ColumnConverter,
		}
	}
	if isConn && isExecer && isExecerContext && isNamedValueChecker && isColumnConverter {
		return &w_conn_Conn_Execer_ExecerContext_NamedValueChecker_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			NamedValueChecker: NamedValueChecker,
			ColumnConverter:   ColumnConverter,
		}
	}
	if isConn && isExecer && isQueryer && isQueryerContext && isConnPrepareContext {
		return &w_conn_Conn_Execer_Queryer_QueryerContext_ConnPrepareContext{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
		}
	}
	if isConn && isExecer && isQueryer && isQueryerContext && isNamedValueChecker {
		return &w_conn_Conn_Execer_Queryer_QueryerContext_NamedValueChecker{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
			NamedValueChecker: NamedValueChecker,
		}
	}
	if isConn && isExecer && isQueryer && isQueryerContext && isColumnConverter {
		return &w_conn_Conn_Execer_Queryer_QueryerContext_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
			ColumnConverter: ColumnConverter,
		}
	}
	if isConn && isExecer && isQueryer && isConnPrepareContext && isNamedValueChecker {
		return &w_conn_Conn_Execer_Queryer_ConnPrepareContext_NamedValueChecker{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
			NamedValueChecker: NamedValueChecker,
		}
	}
	if isConn && isExecer && isQueryer && isConnPrepareContext && isColumnConverter {
		return &w_conn_Conn_Execer_Queryer_ConnPrepareContext_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
			ColumnConverter: ColumnConverter,
		}
	}
	if isConn && isExecer && isQueryer && isNamedValueChecker && isColumnConverter {
		return &w_conn_Conn_Execer_Queryer_NamedValueChecker_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			NamedValueChecker: NamedValueChecker,
			ColumnConverter:   ColumnConverter,
		}
	}
	if isConn && isExecer && isQueryerContext && isConnPrepareContext && isNamedValueChecker {
		return &w_conn_Conn_Execer_QueryerContext_ConnPrepareContext_NamedValueChecker{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
			NamedValueChecker: NamedValueChecker,
		}
	}
	if isConn && isExecer && isQueryerContext && isConnPrepareContext && isColumnConverter {
		return &w_conn_Conn_Execer_QueryerContext_ConnPrepareContext_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
			ColumnConverter: ColumnConverter,
		}
	}
	if isConn && isExecer && isQueryerContext && isNamedValueChecker && isColumnConverter {
		return &w_conn_Conn_Execer_QueryerContext_NamedValueChecker_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
			NamedValueChecker: NamedValueChecker,
			ColumnConverter:   ColumnConverter,
		}
	}
	if isConn && isExecer && isConnPrepareContext && isNamedValueChecker && isColumnConverter {
		return &w_conn_Conn_Execer_ConnPrepareContext_NamedValueChecker_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
			NamedValueChecker: NamedValueChecker,
			ColumnConverter:   ColumnConverter,
		}
	}
	if isConn && isExecerContext && isQueryer && isQueryerContext && isConnPrepareContext {
		return &w_conn_Conn_ExecerContext_Queryer_QueryerContext_ConnPrepareContext{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
		}
	}
	if isConn && isExecerContext && isQueryer && isQueryerContext && isNamedValueChecker {
		return &w_conn_Conn_ExecerContext_Queryer_QueryerContext_NamedValueChecker{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
			NamedValueChecker: NamedValueChecker,
		}
	}
	if isConn && isExecerContext && isQueryer && isQueryerContext && isColumnConverter {
		return &w_conn_Conn_ExecerContext_Queryer_QueryerContext_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
			ColumnConverter: ColumnConverter,
		}
	}
	if isConn && isExecerContext && isQueryer && isConnPrepareContext && isNamedValueChecker {
		return &w_conn_Conn_ExecerContext_Queryer_ConnPrepareContext_NamedValueChecker{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
			NamedValueChecker: NamedValueChecker,
		}
	}
	if isConn && isExecerContext && isQueryer && isConnPrepareContext && isColumnConverter {
		return &w_conn_Conn_ExecerContext_Queryer_ConnPrepareContext_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
			ColumnConverter: ColumnConverter,
		}
	}
	if isConn && isExecerContext && isQueryer && isNamedValueChecker && isColumnConverter {
		return &w_conn_Conn_ExecerContext_Queryer_NamedValueChecker_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			NamedValueChecker: NamedValueChecker,
			ColumnConverter:   ColumnConverter,
		}
	}
	if isConn && isExecerContext && isQueryerContext && isConnPrepareContext && isNamedValueChecker {
		return &w_conn_Conn_ExecerContext_QueryerContext_ConnPrepareContext_NamedValueChecker{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
			NamedValueChecker: NamedValueChecker,
		}
	}
	if isConn && isExecerContext && isQueryerContext && isConnPrepareContext && isColumnConverter {
		return &w_conn_Conn_ExecerContext_QueryerContext_ConnPrepareContext_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
			ColumnConverter: ColumnConverter,
		}
	}
	if isConn && isExecerContext && isQueryerContext && isNamedValueChecker && isColumnConverter {
		return &w_conn_Conn_ExecerContext_QueryerContext_NamedValueChecker_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
			NamedValueChecker: NamedValueChecker,
			ColumnConverter:   ColumnConverter,
		}
	}
	if isConn && isExecerContext && isConnPrepareContext && isNamedValueChecker && isColumnConverter {
		return &w_conn_Conn_ExecerContext_ConnPrepareContext_NamedValueChecker_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
			NamedValueChecker: NamedValueChecker,
			ColumnConverter:   ColumnConverter,
		}
	}
	if isConn && isQueryer && isQueryerContext && isConnPrepareContext && isNamedValueChecker {
		return &w_conn_Conn_Queryer_QueryerContext_ConnPrepareContext_NamedValueChecker{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
			NamedValueChecker: NamedValueChecker,
		}
	}
	if isConn && isQueryer && isQueryerContext && isConnPrepareContext && isColumnConverter {
		return &w_conn_Conn_Queryer_QueryerContext_ConnPrepareContext_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
			ColumnConverter: ColumnConverter,
		}
	}
	if isConn && isQueryer && isQueryerContext && isNamedValueChecker && isColumnConverter {
		return &w_conn_Conn_Queryer_QueryerContext_NamedValueChecker_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
			NamedValueChecker: NamedValueChecker,
			ColumnConverter:   ColumnConverter,
		}
	}
	if isConn && isQueryer && isConnPrepareContext && isNamedValueChecker && isColumnConverter {
		return &w_conn_Conn_Queryer_ConnPrepareContext_NamedValueChecker_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
			NamedValueChecker: NamedValueChecker,
			ColumnConverter:   ColumnConverter,
		}
	}
	if isConn && isQueryerContext && isConnPrepareContext && isNamedValueChecker && isColumnConverter {
		return &w_conn_Conn_QueryerContext_ConnPrepareContext_NamedValueChecker_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
			NamedValueChecker: NamedValueChecker,
			ColumnConverter:   ColumnConverter,
		}
	}
	if isConn && isExecer && isExecerContext && isQueryer {
		return &w_conn_Conn_Execer_ExecerContext_Queryer{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
		}
	}
	if isConn && isExecer && isExecerContext && isQueryerContext {
		return &w_conn_Conn_Execer_ExecerContext_QueryerContext{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
		}
	}
	if isConn && isExecer && isExecerContext && isConnPrepareContext {
		return &w_conn_Conn_Execer_ExecerContext_ConnPrepareContext{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
		}
	}
	if isConn && isExecer && isExecerContext && isNamedValueChecker {
		return &w_conn_Conn_Execer_ExecerContext_NamedValueChecker{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			NamedValueChecker: NamedValueChecker,
		}
	}
	if isConn && isExecer && isExecerContext && isColumnConverter {
		return &w_conn_Conn_Execer_ExecerContext_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			ColumnConverter: ColumnConverter,
		}
	}
	if isConn && isExecer && isQueryer && isQueryerContext {
		return &w_conn_Conn_Execer_Queryer_QueryerContext{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
		}
	}
	if isConn && isExecer && isQueryer && isConnPrepareContext {
		return &w_conn_Conn_Execer_Queryer_ConnPrepareContext{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
		}
	}
	if isConn && isExecer && isQueryer && isNamedValueChecker {
		return &w_conn_Conn_Execer_Queryer_NamedValueChecker{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			NamedValueChecker: NamedValueChecker,
		}
	}
	if isConn && isExecer && isQueryer && isColumnConverter {
		return &w_conn_Conn_Execer_Queryer_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ColumnConverter: ColumnConverter,
		}
	}
	if isConn && isExecer && isQueryerContext && isConnPrepareContext {
		return &w_conn_Conn_Execer_QueryerContext_ConnPrepareContext{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
		}
	}
	if isConn && isExecer && isQueryerContext && isNamedValueChecker {
		return &w_conn_Conn_Execer_QueryerContext_NamedValueChecker{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
			NamedValueChecker: NamedValueChecker,
		}
	}
	if isConn && isExecer && isQueryerContext && isColumnConverter {
		return &w_conn_Conn_Execer_QueryerContext_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
			ColumnConverter: ColumnConverter,
		}
	}
	if isConn && isExecer && isConnPrepareContext && isNamedValueChecker {
		return &w_conn_Conn_Execer_ConnPrepareContext_NamedValueChecker{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
			NamedValueChecker: NamedValueChecker,
		}
	}
	if isConn && isExecer && isConnPrepareContext && isColumnConverter {
		return &w_conn_Conn_Execer_ConnPrepareContext_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
			ColumnConverter: ColumnConverter,
		}
	}
	if isConn && isExecer && isNamedValueChecker && isColumnConverter {
		return &w_conn_Conn_Execer_NamedValueChecker_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			NamedValueChecker: NamedValueChecker,
			ColumnConverter:   ColumnConverter,
		}
	}
	if isConn && isExecerContext && isQueryer && isQueryerContext {
		return &w_conn_Conn_ExecerContext_Queryer_QueryerContext{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
		}
	}
	if isConn && isExecerContext && isQueryer && isConnPrepareContext {
		return &w_conn_Conn_ExecerContext_Queryer_ConnPrepareContext{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
		}
	}
	if isConn && isExecerContext && isQueryer && isNamedValueChecker {
		return &w_conn_Conn_ExecerContext_Queryer_NamedValueChecker{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			NamedValueChecker: NamedValueChecker,
		}
	}
	if isConn && isExecerContext && isQueryer && isColumnConverter {
		return &w_conn_Conn_ExecerContext_Queryer_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ColumnConverter: ColumnConverter,
		}
	}
	if isConn && isExecerContext && isQueryerContext && isConnPrepareContext {
		return &w_conn_Conn_ExecerContext_QueryerContext_ConnPrepareContext{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
		}
	}
	if isConn && isExecerContext && isQueryerContext && isNamedValueChecker {
		return &w_conn_Conn_ExecerContext_QueryerContext_NamedValueChecker{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
			NamedValueChecker: NamedValueChecker,
		}
	}
	if isConn && isExecerContext && isQueryerContext && isColumnConverter {
		return &w_conn_Conn_ExecerContext_QueryerContext_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
			ColumnConverter: ColumnConverter,
		}
	}
	if isConn && isExecerContext && isConnPrepareContext && isNamedValueChecker {
		return &w_conn_Conn_ExecerContext_ConnPrepareContext_NamedValueChecker{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
			NamedValueChecker: NamedValueChecker,
		}
	}
	if isConn && isExecerContext && isConnPrepareContext && isColumnConverter {
		return &w_conn_Conn_ExecerContext_ConnPrepareContext_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
			ColumnConverter: ColumnConverter,
		}
	}
	if isConn && isExecerContext && isNamedValueChecker && isColumnConverter {
		return &w_conn_Conn_ExecerContext_NamedValueChecker_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			NamedValueChecker: NamedValueChecker,
			ColumnConverter:   ColumnConverter,
		}
	}
	if isConn && isQueryer && isQueryerContext && isConnPrepareContext {
		return &w_conn_Conn_Queryer_QueryerContext_ConnPrepareContext{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
		}
	}
	if isConn && isQueryer && isQueryerContext && isNamedValueChecker {
		return &w_conn_Conn_Queryer_QueryerContext_NamedValueChecker{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
			NamedValueChecker: NamedValueChecker,
		}
	}
	if isConn && isQueryer && isQueryerContext && isColumnConverter {
		return &w_conn_Conn_Queryer_QueryerContext_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
			ColumnConverter: ColumnConverter,
		}
	}
	if isConn && isQueryer && isConnPrepareContext && isNamedValueChecker {
		return &w_conn_Conn_Queryer_ConnPrepareContext_NamedValueChecker{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
			NamedValueChecker: NamedValueChecker,
		}
	}
	if isConn && isQueryer && isConnPrepareContext && isColumnConverter {
		return &w_conn_Conn_Queryer_ConnPrepareContext_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
			ColumnConverter: ColumnConverter,
		}
	}
	if isConn && isQueryer && isNamedValueChecker && isColumnConverter {
		return &w_conn_Conn_Queryer_NamedValueChecker_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			NamedValueChecker: NamedValueChecker,
			ColumnConverter:   ColumnConverter,
		}
	}
	if isConn && isQueryerContext && isConnPrepareContext && isNamedValueChecker {
		return &w_conn_Conn_QueryerContext_ConnPrepareContext_NamedValueChecker{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
			NamedValueChecker: NamedValueChecker,
		}
	}
	if isConn && isQueryerContext && isConnPrepareContext && isColumnConverter {
		return &w_conn_Conn_QueryerContext_ConnPrepareContext_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
			ColumnConverter: ColumnConverter,
		}
	}
	if isConn && isQueryerContext && isNamedValueChecker && isColumnConverter {
		return &w_conn_Conn_QueryerContext_NamedValueChecker_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
			NamedValueChecker: NamedValueChecker,
			ColumnConverter:   ColumnConverter,
		}
	}
	if isConn && isConnPrepareContext && isNamedValueChecker && isColumnConverter {
		return &w_conn_Conn_ConnPrepareContext_NamedValueChecker_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
			NamedValueChecker: NamedValueChecker,
			ColumnConverter:   ColumnConverter,
		}
	}
	if isConn && isExecer && isExecerContext {
		return &w_conn_Conn_Execer_ExecerContext{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
		}
	}
	if isConn && isExecer && isQueryer {
		return &w_conn_Conn_Execer_Queryer{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
		}
	}
	if isConn && isExecer && isQueryerContext {
		return &w_conn_Conn_Execer_QueryerContext{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
		}
	}
	if isConn && isExecer && isConnPrepareContext {
		return &w_conn_Conn_Execer_ConnPrepareContext{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
		}
	}
	if isConn && isExecer && isNamedValueChecker {
		return &w_conn_Conn_Execer_NamedValueChecker{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			NamedValueChecker: NamedValueChecker,
		}
	}
	if isConn && isExecer && isColumnConverter {
		return &w_conn_Conn_Execer_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ColumnConverter: ColumnConverter,
		}
	}
	if isConn && isExecerContext && isQueryer {
		return &w_conn_Conn_ExecerContext_Queryer{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
		}
	}
	if isConn && isExecerContext && isQueryerContext {
		return &w_conn_Conn_ExecerContext_QueryerContext{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
		}
	}
	if isConn && isExecerContext && isConnPrepareContext {
		return &w_conn_Conn_ExecerContext_ConnPrepareContext{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
		}
	}
	if isConn && isExecerContext && isNamedValueChecker {
		return &w_conn_Conn_ExecerContext_NamedValueChecker{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			NamedValueChecker: NamedValueChecker,
		}
	}
	if isConn && isExecerContext && isColumnConverter {
		return &w_conn_Conn_ExecerContext_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
			ColumnConverter: ColumnConverter,
		}
	}
	if isConn && isQueryer && isQueryerContext {
		return &w_conn_Conn_Queryer_QueryerContext{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
		}
	}
	if isConn && isQueryer && isConnPrepareContext {
		return &w_conn_Conn_Queryer_ConnPrepareContext{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
		}
	}
	if isConn && isQueryer && isNamedValueChecker {
		return &w_conn_Conn_Queryer_NamedValueChecker{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			NamedValueChecker: NamedValueChecker,
		}
	}
	if isConn && isQueryer && isColumnConverter {
		return &w_conn_Conn_Queryer_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ColumnConverter: ColumnConverter,
		}
	}
	if isConn && isQueryerContext && isConnPrepareContext {
		return &w_conn_Conn_QueryerContext_ConnPrepareContext{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
		}
	}
	if isConn && isQueryerContext && isNamedValueChecker {
		return &w_conn_Conn_QueryerContext_NamedValueChecker{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
			NamedValueChecker: NamedValueChecker,
		}
	}
	if isConn && isQueryerContext && isColumnConverter {
		return &w_conn_Conn_QueryerContext_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
			ColumnConverter: ColumnConverter,
		}
	}
	if isConn && isConnPrepareContext && isNamedValueChecker {
		return &w_conn_Conn_ConnPrepareContext_NamedValueChecker{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
			NamedValueChecker: NamedValueChecker,
		}
	}
	if isConn && isConnPrepareContext && isColumnConverter {
		return &w_conn_Conn_ConnPrepareContext_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
			ColumnConverter: ColumnConverter,
		}
	}
	if isConn && isNamedValueChecker && isColumnConverter {
		return &w_conn_Conn_NamedValueChecker_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			NamedValueChecker: NamedValueChecker,
			ColumnConverter:   ColumnConverter,
		}
	}
	if isConn && isExecer {
		return &w_conn_Conn_Execer{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Execer: &wExecer{
				Execer:      Execer,
				connDetails: connDetails,
				sensor:      sensor,
			},
		}
	}
	if isConn && isExecerContext {
		return &w_conn_Conn_ExecerContext{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ExecerContext: &wExecerContext{
				ExecerContext: ExecerContext,
				connDetails:   connDetails,
				sensor:        sensor,
			},
		}
	}
	if isConn && isQueryer {
		return &w_conn_Conn_Queryer{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			Queryer: &wQueryer{
				Queryer:     Queryer,
				connDetails: connDetails,
				sensor:      sensor,
			},
		}
	}
	if isConn && isQueryerContext {
		return &w_conn_Conn_QueryerContext{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			QueryerContext: &wQueryerContext{
				QueryerContext: QueryerContext,
				connDetails:    connDetails,
				sensor:         sensor,
			},
		}
	}
	if isConn && isConnPrepareContext {
		return &w_conn_Conn_ConnPrepareContext{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ConnPrepareContext: &wConnPrepareContext{
				ConnPrepareContext: ConnPrepareContext,
				connDetails:        connDetails,
				sensor:             sensor,
			},
		}
	}
	if isConn && isNamedValueChecker {
		return &w_conn_Conn_NamedValueChecker{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			NamedValueChecker: NamedValueChecker,
		}
	}
	if isConn && isColumnConverter {
		return &w_conn_Conn_ColumnConverter{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
			ColumnConverter: ColumnConverter,
		}
	}
	if isConn {
		return &w_conn_Conn{
			Conn: &wConn{
				Conn:        Conn,
				connDetails: connDetails,
				sensor:      sensor,
			},
		}
	}
	return conn
}

type w_conn_Conn_Execer_ExecerContext_Queryer_QueryerContext_ConnPrepareContext_NamedValueChecker_ColumnConverter struct {
	driver.Conn
	driver.Execer
	driver.ExecerContext
	driver.Queryer
	driver.QueryerContext
	driver.ConnPrepareContext
	driver.NamedValueChecker
	driver.ColumnConverter
}
type w_conn_Conn_Execer_ExecerContext_Queryer_QueryerContext_ConnPrepareContext_NamedValueChecker struct {
	driver.Conn
	driver.Execer
	driver.ExecerContext
	driver.Queryer
	driver.QueryerContext
	driver.ConnPrepareContext
	driver.NamedValueChecker
}
type w_conn_Conn_Execer_ExecerContext_Queryer_QueryerContext_ConnPrepareContext_ColumnConverter struct {
	driver.Conn
	driver.Execer
	driver.ExecerContext
	driver.Queryer
	driver.QueryerContext
	driver.ConnPrepareContext
	driver.ColumnConverter
}
type w_conn_Conn_Execer_ExecerContext_Queryer_QueryerContext_NamedValueChecker_ColumnConverter struct {
	driver.Conn
	driver.Execer
	driver.ExecerContext
	driver.Queryer
	driver.QueryerContext
	driver.NamedValueChecker
	driver.ColumnConverter
}
type w_conn_Conn_Execer_ExecerContext_Queryer_ConnPrepareContext_NamedValueChecker_ColumnConverter struct {
	driver.Conn
	driver.Execer
	driver.ExecerContext
	driver.Queryer
	driver.ConnPrepareContext
	driver.NamedValueChecker
	driver.ColumnConverter
}
type w_conn_Conn_Execer_ExecerContext_QueryerContext_ConnPrepareContext_NamedValueChecker_ColumnConverter struct {
	driver.Conn
	driver.Execer
	driver.ExecerContext
	driver.QueryerContext
	driver.ConnPrepareContext
	driver.NamedValueChecker
	driver.ColumnConverter
}
type w_conn_Conn_Execer_Queryer_QueryerContext_ConnPrepareContext_NamedValueChecker_ColumnConverter struct {
	driver.Conn
	driver.Execer
	driver.Queryer
	driver.QueryerContext
	driver.ConnPrepareContext
	driver.NamedValueChecker
	driver.ColumnConverter
}
type w_conn_Conn_ExecerContext_Queryer_QueryerContext_ConnPrepareContext_NamedValueChecker_ColumnConverter struct {
	driver.Conn
	driver.ExecerContext
	driver.Queryer
	driver.QueryerContext
	driver.ConnPrepareContext
	driver.NamedValueChecker
	driver.ColumnConverter
}
type w_conn_Conn_Execer_ExecerContext_Queryer_QueryerContext_ConnPrepareContext struct {
	driver.Conn
	driver.Execer
	driver.ExecerContext
	driver.Queryer
	driver.QueryerContext
	driver.ConnPrepareContext
}
type w_conn_Conn_Execer_ExecerContext_Queryer_QueryerContext_NamedValueChecker struct {
	driver.Conn
	driver.Execer
	driver.ExecerContext
	driver.Queryer
	driver.QueryerContext
	driver.NamedValueChecker
}
type w_conn_Conn_Execer_ExecerContext_Queryer_QueryerContext_ColumnConverter struct {
	driver.Conn
	driver.Execer
	driver.ExecerContext
	driver.Queryer
	driver.QueryerContext
	driver.ColumnConverter
}
type w_conn_Conn_Execer_ExecerContext_Queryer_ConnPrepareContext_NamedValueChecker struct {
	driver.Conn
	driver.Execer
	driver.ExecerContext
	driver.Queryer
	driver.ConnPrepareContext
	driver.NamedValueChecker
}
type w_conn_Conn_Execer_ExecerContext_Queryer_ConnPrepareContext_ColumnConverter struct {
	driver.Conn
	driver.Execer
	driver.ExecerContext
	driver.Queryer
	driver.ConnPrepareContext
	driver.ColumnConverter
}
type w_conn_Conn_Execer_ExecerContext_Queryer_NamedValueChecker_ColumnConverter struct {
	driver.Conn
	driver.Execer
	driver.ExecerContext
	driver.Queryer
	driver.NamedValueChecker
	driver.ColumnConverter
}
type w_conn_Conn_Execer_ExecerContext_QueryerContext_ConnPrepareContext_NamedValueChecker struct {
	driver.Conn
	driver.Execer
	driver.ExecerContext
	driver.QueryerContext
	driver.ConnPrepareContext
	driver.NamedValueChecker
}
type w_conn_Conn_Execer_ExecerContext_QueryerContext_ConnPrepareContext_ColumnConverter struct {
	driver.Conn
	driver.Execer
	driver.ExecerContext
	driver.QueryerContext
	driver.ConnPrepareContext
	driver.ColumnConverter
}
type w_conn_Conn_Execer_ExecerContext_QueryerContext_NamedValueChecker_ColumnConverter struct {
	driver.Conn
	driver.Execer
	driver.ExecerContext
	driver.QueryerContext
	driver.NamedValueChecker
	driver.ColumnConverter
}
type w_conn_Conn_Execer_ExecerContext_ConnPrepareContext_NamedValueChecker_ColumnConverter struct {
	driver.Conn
	driver.Execer
	driver.ExecerContext
	driver.ConnPrepareContext
	driver.NamedValueChecker
	driver.ColumnConverter
}
type w_conn_Conn_Execer_Queryer_QueryerContext_ConnPrepareContext_NamedValueChecker struct {
	driver.Conn
	driver.Execer
	driver.Queryer
	driver.QueryerContext
	driver.ConnPrepareContext
	driver.NamedValueChecker
}
type w_conn_Conn_Execer_Queryer_QueryerContext_ConnPrepareContext_ColumnConverter struct {
	driver.Conn
	driver.Execer
	driver.Queryer
	driver.QueryerContext
	driver.ConnPrepareContext
	driver.ColumnConverter
}
type w_conn_Conn_Execer_Queryer_QueryerContext_NamedValueChecker_ColumnConverter struct {
	driver.Conn
	driver.Execer
	driver.Queryer
	driver.QueryerContext
	driver.NamedValueChecker
	driver.ColumnConverter
}
type w_conn_Conn_Execer_Queryer_ConnPrepareContext_NamedValueChecker_ColumnConverter struct {
	driver.Conn
	driver.Execer
	driver.Queryer
	driver.ConnPrepareContext
	driver.NamedValueChecker
	driver.ColumnConverter
}
type w_conn_Conn_Execer_QueryerContext_ConnPrepareContext_NamedValueChecker_ColumnConverter struct {
	driver.Conn
	driver.Execer
	driver.QueryerContext
	driver.ConnPrepareContext
	driver.NamedValueChecker
	driver.ColumnConverter
}
type w_conn_Conn_ExecerContext_Queryer_QueryerContext_ConnPrepareContext_NamedValueChecker struct {
	driver.Conn
	driver.ExecerContext
	driver.Queryer
	driver.QueryerContext
	driver.ConnPrepareContext
	driver.NamedValueChecker
}
type w_conn_Conn_ExecerContext_Queryer_QueryerContext_ConnPrepareContext_ColumnConverter struct {
	driver.Conn
	driver.ExecerContext
	driver.Queryer
	driver.QueryerContext
	driver.ConnPrepareContext
	driver.ColumnConverter
}
type w_conn_Conn_ExecerContext_Queryer_QueryerContext_NamedValueChecker_ColumnConverter struct {
	driver.Conn
	driver.ExecerContext
	driver.Queryer
	driver.QueryerContext
	driver.NamedValueChecker
	driver.ColumnConverter
}
type w_conn_Conn_ExecerContext_Queryer_ConnPrepareContext_NamedValueChecker_ColumnConverter struct {
	driver.Conn
	driver.ExecerContext
	driver.Queryer
	driver.ConnPrepareContext
	driver.NamedValueChecker
	driver.ColumnConverter
}
type w_conn_Conn_ExecerContext_QueryerContext_ConnPrepareContext_NamedValueChecker_ColumnConverter struct {
	driver.Conn
	driver.ExecerContext
	driver.QueryerContext
	driver.ConnPrepareContext
	driver.NamedValueChecker
	driver.ColumnConverter
}
type w_conn_Conn_Queryer_QueryerContext_ConnPrepareContext_NamedValueChecker_ColumnConverter struct {
	driver.Conn
	driver.Queryer
	driver.QueryerContext
	driver.ConnPrepareContext
	driver.NamedValueChecker
	driver.ColumnConverter
}
type w_conn_Conn_Execer_ExecerContext_Queryer_QueryerContext struct {
	driver.Conn
	driver.Execer
	driver.ExecerContext
	driver.Queryer
	driver.QueryerContext
}
type w_conn_Conn_Execer_ExecerContext_Queryer_ConnPrepareContext struct {
	driver.Conn
	driver.Execer
	driver.ExecerContext
	driver.Queryer
	driver.ConnPrepareContext
}
type w_conn_Conn_Execer_ExecerContext_Queryer_NamedValueChecker struct {
	driver.Conn
	driver.Execer
	driver.ExecerContext
	driver.Queryer
	driver.NamedValueChecker
}
type w_conn_Conn_Execer_ExecerContext_Queryer_ColumnConverter struct {
	driver.Conn
	driver.Execer
	driver.ExecerContext
	driver.Queryer
	driver.ColumnConverter
}
type w_conn_Conn_Execer_ExecerContext_QueryerContext_ConnPrepareContext struct {
	driver.Conn
	driver.Execer
	driver.ExecerContext
	driver.QueryerContext
	driver.ConnPrepareContext
}
type w_conn_Conn_Execer_ExecerContext_QueryerContext_NamedValueChecker struct {
	driver.Conn
	driver.Execer
	driver.ExecerContext
	driver.QueryerContext
	driver.NamedValueChecker
}
type w_conn_Conn_Execer_ExecerContext_QueryerContext_ColumnConverter struct {
	driver.Conn
	driver.Execer
	driver.ExecerContext
	driver.QueryerContext
	driver.ColumnConverter
}
type w_conn_Conn_Execer_ExecerContext_ConnPrepareContext_NamedValueChecker struct {
	driver.Conn
	driver.Execer
	driver.ExecerContext
	driver.ConnPrepareContext
	driver.NamedValueChecker
}
type w_conn_Conn_Execer_ExecerContext_ConnPrepareContext_ColumnConverter struct {
	driver.Conn
	driver.Execer
	driver.ExecerContext
	driver.ConnPrepareContext
	driver.ColumnConverter
}
type w_conn_Conn_Execer_ExecerContext_NamedValueChecker_ColumnConverter struct {
	driver.Conn
	driver.Execer
	driver.ExecerContext
	driver.NamedValueChecker
	driver.ColumnConverter
}
type w_conn_Conn_Execer_Queryer_QueryerContext_ConnPrepareContext struct {
	driver.Conn
	driver.Execer
	driver.Queryer
	driver.QueryerContext
	driver.ConnPrepareContext
}
type w_conn_Conn_Execer_Queryer_QueryerContext_NamedValueChecker struct {
	driver.Conn
	driver.Execer
	driver.Queryer
	driver.QueryerContext
	driver.NamedValueChecker
}
type w_conn_Conn_Execer_Queryer_QueryerContext_ColumnConverter struct {
	driver.Conn
	driver.Execer
	driver.Queryer
	driver.QueryerContext
	driver.ColumnConverter
}
type w_conn_Conn_Execer_Queryer_ConnPrepareContext_NamedValueChecker struct {
	driver.Conn
	driver.Execer
	driver.Queryer
	driver.ConnPrepareContext
	driver.NamedValueChecker
}
type w_conn_Conn_Execer_Queryer_ConnPrepareContext_ColumnConverter struct {
	driver.Conn
	driver.Execer
	driver.Queryer
	driver.ConnPrepareContext
	driver.ColumnConverter
}
type w_conn_Conn_Execer_Queryer_NamedValueChecker_ColumnConverter struct {
	driver.Conn
	driver.Execer
	driver.Queryer
	driver.NamedValueChecker
	driver.ColumnConverter
}
type w_conn_Conn_Execer_QueryerContext_ConnPrepareContext_NamedValueChecker struct {
	driver.Conn
	driver.Execer
	driver.QueryerContext
	driver.ConnPrepareContext
	driver.NamedValueChecker
}
type w_conn_Conn_Execer_QueryerContext_ConnPrepareContext_ColumnConverter struct {
	driver.Conn
	driver.Execer
	driver.QueryerContext
	driver.ConnPrepareContext
	driver.ColumnConverter
}
type w_conn_Conn_Execer_QueryerContext_NamedValueChecker_ColumnConverter struct {
	driver.Conn
	driver.Execer
	driver.QueryerContext
	driver.NamedValueChecker
	driver.ColumnConverter
}
type w_conn_Conn_Execer_ConnPrepareContext_NamedValueChecker_ColumnConverter struct {
	driver.Conn
	driver.Execer
	driver.ConnPrepareContext
	driver.NamedValueChecker
	driver.ColumnConverter
}
type w_conn_Conn_ExecerContext_Queryer_QueryerContext_ConnPrepareContext struct {
	driver.Conn
	driver.ExecerContext
	driver.Queryer
	driver.QueryerContext
	driver.ConnPrepareContext
}
type w_conn_Conn_ExecerContext_Queryer_QueryerContext_NamedValueChecker struct {
	driver.Conn
	driver.ExecerContext
	driver.Queryer
	driver.QueryerContext
	driver.NamedValueChecker
}
type w_conn_Conn_ExecerContext_Queryer_QueryerContext_ColumnConverter struct {
	driver.Conn
	driver.ExecerContext
	driver.Queryer
	driver.QueryerContext
	driver.ColumnConverter
}
type w_conn_Conn_ExecerContext_Queryer_ConnPrepareContext_NamedValueChecker struct {
	driver.Conn
	driver.ExecerContext
	driver.Queryer
	driver.ConnPrepareContext
	driver.NamedValueChecker
}
type w_conn_Conn_ExecerContext_Queryer_ConnPrepareContext_ColumnConverter struct {
	driver.Conn
	driver.ExecerContext
	driver.Queryer
	driver.ConnPrepareContext
	driver.ColumnConverter
}
type w_conn_Conn_ExecerContext_Queryer_NamedValueChecker_ColumnConverter struct {
	driver.Conn
	driver.ExecerContext
	driver.Queryer
	driver.NamedValueChecker
	driver.ColumnConverter
}
type w_conn_Conn_ExecerContext_QueryerContext_ConnPrepareContext_NamedValueChecker struct {
	driver.Conn
	driver.ExecerContext
	driver.QueryerContext
	driver.ConnPrepareContext
	driver.NamedValueChecker
}
type w_conn_Conn_ExecerContext_QueryerContext_ConnPrepareContext_ColumnConverter struct {
	driver.Conn
	driver.ExecerContext
	driver.QueryerContext
	driver.ConnPrepareContext
	driver.ColumnConverter
}
type w_conn_Conn_ExecerContext_QueryerContext_NamedValueChecker_ColumnConverter struct {
	driver.Conn
	driver.ExecerContext
	driver.QueryerContext
	driver.NamedValueChecker
	driver.ColumnConverter
}
type w_conn_Conn_ExecerContext_ConnPrepareContext_NamedValueChecker_ColumnConverter struct {
	driver.Conn
	driver.ExecerContext
	driver.ConnPrepareContext
	driver.NamedValueChecker
	driver.ColumnConverter
}
type w_conn_Conn_Queryer_QueryerContext_ConnPrepareContext_NamedValueChecker struct {
	driver.Conn
	driver.Queryer
	driver.QueryerContext
	driver.ConnPrepareContext
	driver.NamedValueChecker
}
type w_conn_Conn_Queryer_QueryerContext_ConnPrepareContext_ColumnConverter struct {
	driver.Conn
	driver.Queryer
	driver.QueryerContext
	driver.ConnPrepareContext
	driver.ColumnConverter
}
type w_conn_Conn_Queryer_QueryerContext_NamedValueChecker_ColumnConverter struct {
	driver.Conn
	driver.Queryer
	driver.QueryerContext
	driver.NamedValueChecker
	driver.ColumnConverter
}
type w_conn_Conn_Queryer_ConnPrepareContext_NamedValueChecker_ColumnConverter struct {
	driver.Conn
	driver.Queryer
	driver.ConnPrepareContext
	driver.NamedValueChecker
	driver.ColumnConverter
}
type w_conn_Conn_QueryerContext_ConnPrepareContext_NamedValueChecker_ColumnConverter struct {
	driver.Conn
	driver.QueryerContext
	driver.ConnPrepareContext
	driver.NamedValueChecker
	driver.ColumnConverter
}
type w_conn_Conn_Execer_ExecerContext_Queryer struct {
	driver.Conn
	driver.Execer
	driver.ExecerContext
	driver.Queryer
}
type w_conn_Conn_Execer_ExecerContext_QueryerContext struct {
	driver.Conn
	driver.Execer
	driver.ExecerContext
	driver.QueryerContext
}
type w_conn_Conn_Execer_ExecerContext_ConnPrepareContext struct {
	driver.Conn
	driver.Execer
	driver.ExecerContext
	driver.ConnPrepareContext
}
type w_conn_Conn_Execer_ExecerContext_NamedValueChecker struct {
	driver.Conn
	driver.Execer
	driver.ExecerContext
	driver.NamedValueChecker
}
type w_conn_Conn_Execer_ExecerContext_ColumnConverter struct {
	driver.Conn
	driver.Execer
	driver.ExecerContext
	driver.ColumnConverter
}
type w_conn_Conn_Execer_Queryer_QueryerContext struct {
	driver.Conn
	driver.Execer
	driver.Queryer
	driver.QueryerContext
}
type w_conn_Conn_Execer_Queryer_ConnPrepareContext struct {
	driver.Conn
	driver.Execer
	driver.Queryer
	driver.ConnPrepareContext
}
type w_conn_Conn_Execer_Queryer_NamedValueChecker struct {
	driver.Conn
	driver.Execer
	driver.Queryer
	driver.NamedValueChecker
}
type w_conn_Conn_Execer_Queryer_ColumnConverter struct {
	driver.Conn
	driver.Execer
	driver.Queryer
	driver.ColumnConverter
}
type w_conn_Conn_Execer_QueryerContext_ConnPrepareContext struct {
	driver.Conn
	driver.Execer
	driver.QueryerContext
	driver.ConnPrepareContext
}
type w_conn_Conn_Execer_QueryerContext_NamedValueChecker struct {
	driver.Conn
	driver.Execer
	driver.QueryerContext
	driver.NamedValueChecker
}
type w_conn_Conn_Execer_QueryerContext_ColumnConverter struct {
	driver.Conn
	driver.Execer
	driver.QueryerContext
	driver.ColumnConverter
}
type w_conn_Conn_Execer_ConnPrepareContext_NamedValueChecker struct {
	driver.Conn
	driver.Execer
	driver.ConnPrepareContext
	driver.NamedValueChecker
}
type w_conn_Conn_Execer_ConnPrepareContext_ColumnConverter struct {
	driver.Conn
	driver.Execer
	driver.ConnPrepareContext
	driver.ColumnConverter
}
type w_conn_Conn_Execer_NamedValueChecker_ColumnConverter struct {
	driver.Conn
	driver.Execer
	driver.NamedValueChecker
	driver.ColumnConverter
}
type w_conn_Conn_ExecerContext_Queryer_QueryerContext struct {
	driver.Conn
	driver.ExecerContext
	driver.Queryer
	driver.QueryerContext
}
type w_conn_Conn_ExecerContext_Queryer_ConnPrepareContext struct {
	driver.Conn
	driver.ExecerContext
	driver.Queryer
	driver.ConnPrepareContext
}
type w_conn_Conn_ExecerContext_Queryer_NamedValueChecker struct {
	driver.Conn
	driver.ExecerContext
	driver.Queryer
	driver.NamedValueChecker
}
type w_conn_Conn_ExecerContext_Queryer_ColumnConverter struct {
	driver.Conn
	driver.ExecerContext
	driver.Queryer
	driver.ColumnConverter
}
type w_conn_Conn_ExecerContext_QueryerContext_ConnPrepareContext struct {
	driver.Conn
	driver.ExecerContext
	driver.QueryerContext
	driver.ConnPrepareContext
}
type w_conn_Conn_ExecerContext_QueryerContext_NamedValueChecker struct {
	driver.Conn
	driver.ExecerContext
	driver.QueryerContext
	driver.NamedValueChecker
}
type w_conn_Conn_ExecerContext_QueryerContext_ColumnConverter struct {
	driver.Conn
	driver.ExecerContext
	driver.QueryerContext
	driver.ColumnConverter
}
type w_conn_Conn_ExecerContext_ConnPrepareContext_NamedValueChecker struct {
	driver.Conn
	driver.ExecerContext
	driver.ConnPrepareContext
	driver.NamedValueChecker
}
type w_conn_Conn_ExecerContext_ConnPrepareContext_ColumnConverter struct {
	driver.Conn
	driver.ExecerContext
	driver.ConnPrepareContext
	driver.ColumnConverter
}
type w_conn_Conn_ExecerContext_NamedValueChecker_ColumnConverter struct {
	driver.Conn
	driver.ExecerContext
	driver.NamedValueChecker
	driver.ColumnConverter
}
type w_conn_Conn_Queryer_QueryerContext_ConnPrepareContext struct {
	driver.Conn
	driver.Queryer
	driver.QueryerContext
	driver.ConnPrepareContext
}
type w_conn_Conn_Queryer_QueryerContext_NamedValueChecker struct {
	driver.Conn
	driver.Queryer
	driver.QueryerContext
	driver.NamedValueChecker
}
type w_conn_Conn_Queryer_QueryerContext_ColumnConverter struct {
	driver.Conn
	driver.Queryer
	driver.QueryerContext
	driver.ColumnConverter
}
type w_conn_Conn_Queryer_ConnPrepareContext_NamedValueChecker struct {
	driver.Conn
	driver.Queryer
	driver.ConnPrepareContext
	driver.NamedValueChecker
}
type w_conn_Conn_Queryer_ConnPrepareContext_ColumnConverter struct {
	driver.Conn
	driver.Queryer
	driver.ConnPrepareContext
	driver.ColumnConverter
}
type w_conn_Conn_Queryer_NamedValueChecker_ColumnConverter struct {
	driver.Conn
	driver.Queryer
	driver.NamedValueChecker
	driver.ColumnConverter
}
type w_conn_Conn_QueryerContext_ConnPrepareContext_NamedValueChecker struct {
	driver.Conn
	driver.QueryerContext
	driver.ConnPrepareContext
	driver.NamedValueChecker
}
type w_conn_Conn_QueryerContext_ConnPrepareContext_ColumnConverter struct {
	driver.Conn
	driver.QueryerContext
	driver.ConnPrepareContext
	driver.ColumnConverter
}
type w_conn_Conn_QueryerContext_NamedValueChecker_ColumnConverter struct {
	driver.Conn
	driver.QueryerContext
	driver.NamedValueChecker
	driver.ColumnConverter
}
type w_conn_Conn_ConnPrepareContext_NamedValueChecker_ColumnConverter struct {
	driver.Conn
	driver.ConnPrepareContext
	driver.NamedValueChecker
	driver.ColumnConverter
}
type w_conn_Conn_Execer_ExecerContext struct {
	driver.Conn
	driver.Execer
	driver.ExecerContext
}
type w_conn_Conn_Execer_Queryer struct {
	driver.Conn
	driver.Execer
	driver.Queryer
}
type w_conn_Conn_Execer_QueryerContext struct {
	driver.Conn
	driver.Execer
	driver.QueryerContext
}
type w_conn_Conn_Execer_ConnPrepareContext struct {
	driver.Conn
	driver.Execer
	driver.ConnPrepareContext
}
type w_conn_Conn_Execer_NamedValueChecker struct {
	driver.Conn
	driver.Execer
	driver.NamedValueChecker
}
type w_conn_Conn_Execer_ColumnConverter struct {
	driver.Conn
	driver.Execer
	driver.ColumnConverter
}
type w_conn_Conn_ExecerContext_Queryer struct {
	driver.Conn
	driver.ExecerContext
	driver.Queryer
}
type w_conn_Conn_ExecerContext_QueryerContext struct {
	driver.Conn
	driver.ExecerContext
	driver.QueryerContext
}
type w_conn_Conn_ExecerContext_ConnPrepareContext struct {
	driver.Conn
	driver.ExecerContext
	driver.ConnPrepareContext
}
type w_conn_Conn_ExecerContext_NamedValueChecker struct {
	driver.Conn
	driver.ExecerContext
	driver.NamedValueChecker
}
type w_conn_Conn_ExecerContext_ColumnConverter struct {
	driver.Conn
	driver.ExecerContext
	driver.ColumnConverter
}
type w_conn_Conn_Queryer_QueryerContext struct {
	driver.Conn
	driver.Queryer
	driver.QueryerContext
}
type w_conn_Conn_Queryer_ConnPrepareContext struct {
	driver.Conn
	driver.Queryer
	driver.ConnPrepareContext
}
type w_conn_Conn_Queryer_NamedValueChecker struct {
	driver.Conn
	driver.Queryer
	driver.NamedValueChecker
}
type w_conn_Conn_Queryer_ColumnConverter struct {
	driver.Conn
	driver.Queryer
	driver.ColumnConverter
}
type w_conn_Conn_QueryerContext_ConnPrepareContext struct {
	driver.Conn
	driver.QueryerContext
	driver.ConnPrepareContext
}
type w_conn_Conn_QueryerContext_NamedValueChecker struct {
	driver.Conn
	driver.QueryerContext
	driver.NamedValueChecker
}
type w_conn_Conn_QueryerContext_ColumnConverter struct {
	driver.Conn
	driver.QueryerContext
	driver.ColumnConverter
}
type w_conn_Conn_ConnPrepareContext_NamedValueChecker struct {
	driver.Conn
	driver.ConnPrepareContext
	driver.NamedValueChecker
}
type w_conn_Conn_ConnPrepareContext_ColumnConverter struct {
	driver.Conn
	driver.ConnPrepareContext
	driver.ColumnConverter
}
type w_conn_Conn_NamedValueChecker_ColumnConverter struct {
	driver.Conn
	driver.NamedValueChecker
	driver.ColumnConverter
}
type w_conn_Conn_Execer struct {
	driver.Conn
	driver.Execer
}
type w_conn_Conn_ExecerContext struct {
	driver.Conn
	driver.ExecerContext
}
type w_conn_Conn_Queryer struct {
	driver.Conn
	driver.Queryer
}
type w_conn_Conn_QueryerContext struct {
	driver.Conn
	driver.QueryerContext
}
type w_conn_Conn_ConnPrepareContext struct {
	driver.Conn
	driver.ConnPrepareContext
}
type w_conn_Conn_NamedValueChecker struct {
	driver.Conn
	driver.NamedValueChecker
}
type w_conn_Conn_ColumnConverter struct {
	driver.Conn
	driver.ColumnConverter
}
type w_conn_Conn struct {
	driver.Conn
}

func connAlreadyWrapped(conn driver.Conn) bool {
	switch conn.(type) {
	case *w_conn_Conn_Execer_ExecerContext_Queryer_QueryerContext_ConnPrepareContext_NamedValueChecker_ColumnConverter, *w_conn_Conn_Execer_ExecerContext_Queryer_QueryerContext_ConnPrepareContext_NamedValueChecker, *w_conn_Conn_Execer_ExecerContext_Queryer_QueryerContext_ConnPrepareContext_ColumnConverter, *w_conn_Conn_Execer_ExecerContext_Queryer_QueryerContext_NamedValueChecker_ColumnConverter, *w_conn_Conn_Execer_ExecerContext_Queryer_ConnPrepareContext_NamedValueChecker_ColumnConverter, *w_conn_Conn_Execer_ExecerContext_QueryerContext_ConnPrepareContext_NamedValueChecker_ColumnConverter, *w_conn_Conn_Execer_Queryer_QueryerContext_ConnPrepareContext_NamedValueChecker_ColumnConverter, *w_conn_Conn_ExecerContext_Queryer_QueryerContext_ConnPrepareContext_NamedValueChecker_ColumnConverter, *w_conn_Conn_Execer_ExecerContext_Queryer_QueryerContext_ConnPrepareContext, *w_conn_Conn_Execer_ExecerContext_Queryer_QueryerContext_NamedValueChecker, *w_conn_Conn_Execer_ExecerContext_Queryer_QueryerContext_ColumnConverter, *w_conn_Conn_Execer_ExecerContext_Queryer_ConnPrepareContext_NamedValueChecker, *w_conn_Conn_Execer_ExecerContext_Queryer_ConnPrepareContext_ColumnConverter, *w_conn_Conn_Execer_ExecerContext_Queryer_NamedValueChecker_ColumnConverter, *w_conn_Conn_Execer_ExecerContext_QueryerContext_ConnPrepareContext_NamedValueChecker, *w_conn_Conn_Execer_ExecerContext_QueryerContext_ConnPrepareContext_ColumnConverter, *w_conn_Conn_Execer_ExecerContext_QueryerContext_NamedValueChecker_ColumnConverter, *w_conn_Conn_Execer_ExecerContext_ConnPrepareContext_NamedValueChecker_ColumnConverter, *w_conn_Conn_Execer_Queryer_QueryerContext_ConnPrepareContext_NamedValueChecker, *w_conn_Conn_Execer_Queryer_QueryerContext_ConnPrepareContext_ColumnConverter, *w_conn_Conn_Execer_Queryer_QueryerContext_NamedValueChecker_ColumnConverter, *w_conn_Conn_Execer_Queryer_ConnPrepareContext_NamedValueChecker_ColumnConverter, *w_conn_Conn_Execer_QueryerContext_ConnPrepareContext_NamedValueChecker_ColumnConverter, *w_conn_Conn_ExecerContext_Queryer_QueryerContext_ConnPrepareContext_NamedValueChecker, *w_conn_Conn_ExecerContext_Queryer_QueryerContext_ConnPrepareContext_ColumnConverter, *w_conn_Conn_ExecerContext_Queryer_QueryerContext_NamedValueChecker_ColumnConverter, *w_conn_Conn_ExecerContext_Queryer_ConnPrepareContext_NamedValueChecker_ColumnConverter, *w_conn_Conn_ExecerContext_QueryerContext_ConnPrepareContext_NamedValueChecker_ColumnConverter, *w_conn_Conn_Queryer_QueryerContext_ConnPrepareContext_NamedValueChecker_ColumnConverter, *w_conn_Conn_Execer_ExecerContext_Queryer_QueryerContext, *w_conn_Conn_Execer_ExecerContext_Queryer_ConnPrepareContext, *w_conn_Conn_Execer_ExecerContext_Queryer_NamedValueChecker, *w_conn_Conn_Execer_ExecerContext_Queryer_ColumnConverter, *w_conn_Conn_Execer_ExecerContext_QueryerContext_ConnPrepareContext, *w_conn_Conn_Execer_ExecerContext_QueryerContext_NamedValueChecker, *w_conn_Conn_Execer_ExecerContext_QueryerContext_ColumnConverter, *w_conn_Conn_Execer_ExecerContext_ConnPrepareContext_NamedValueChecker, *w_conn_Conn_Execer_ExecerContext_ConnPrepareContext_ColumnConverter, *w_conn_Conn_Execer_ExecerContext_NamedValueChecker_ColumnConverter, *w_conn_Conn_Execer_Queryer_QueryerContext_ConnPrepareContext, *w_conn_Conn_Execer_Queryer_QueryerContext_NamedValueChecker, *w_conn_Conn_Execer_Queryer_QueryerContext_ColumnConverter, *w_conn_Conn_Execer_Queryer_ConnPrepareContext_NamedValueChecker, *w_conn_Conn_Execer_Queryer_ConnPrepareContext_ColumnConverter, *w_conn_Conn_Execer_Queryer_NamedValueChecker_ColumnConverter, *w_conn_Conn_Execer_QueryerContext_ConnPrepareContext_NamedValueChecker, *w_conn_Conn_Execer_QueryerContext_ConnPrepareContext_ColumnConverter, *w_conn_Conn_Execer_QueryerContext_NamedValueChecker_ColumnConverter, *w_conn_Conn_Execer_ConnPrepareContext_NamedValueChecker_ColumnConverter, *w_conn_Conn_ExecerContext_Queryer_QueryerContext_ConnPrepareContext, *w_conn_Conn_ExecerContext_Queryer_QueryerContext_NamedValueChecker, *w_conn_Conn_ExecerContext_Queryer_QueryerContext_ColumnConverter, *w_conn_Conn_ExecerContext_Queryer_ConnPrepareContext_NamedValueChecker, *w_conn_Conn_ExecerContext_Queryer_ConnPrepareContext_ColumnConverter, *w_conn_Conn_ExecerContext_Queryer_NamedValueChecker_ColumnConverter, *w_conn_Conn_ExecerContext_QueryerContext_ConnPrepareContext_NamedValueChecker, *w_conn_Conn_ExecerContext_QueryerContext_ConnPrepareContext_ColumnConverter, *w_conn_Conn_ExecerContext_QueryerContext_NamedValueChecker_ColumnConverter, *w_conn_Conn_ExecerContext_ConnPrepareContext_NamedValueChecker_ColumnConverter, *w_conn_Conn_Queryer_QueryerContext_ConnPrepareContext_NamedValueChecker, *w_conn_Conn_Queryer_QueryerContext_ConnPrepareContext_ColumnConverter, *w_conn_Conn_Queryer_QueryerContext_NamedValueChecker_ColumnConverter, *w_conn_Conn_Queryer_ConnPrepareContext_NamedValueChecker_ColumnConverter, *w_conn_Conn_QueryerContext_ConnPrepareContext_NamedValueChecker_ColumnConverter, *w_conn_Conn_Execer_ExecerContext_Queryer, *w_conn_Conn_Execer_ExecerContext_QueryerContext, *w_conn_Conn_Execer_ExecerContext_ConnPrepareContext, *w_conn_Conn_Execer_ExecerContext_NamedValueChecker, *w_conn_Conn_Execer_ExecerContext_ColumnConverter, *w_conn_Conn_Execer_Queryer_QueryerContext, *w_conn_Conn_Execer_Queryer_ConnPrepareContext, *w_conn_Conn_Execer_Queryer_NamedValueChecker, *w_conn_Conn_Execer_Queryer_ColumnConverter, *w_conn_Conn_Execer_QueryerContext_ConnPrepareContext, *w_conn_Conn_Execer_QueryerContext_NamedValueChecker, *w_conn_Conn_Execer_QueryerContext_ColumnConverter, *w_conn_Conn_Execer_ConnPrepareContext_NamedValueChecker, *w_conn_Conn_Execer_ConnPrepareContext_ColumnConverter, *w_conn_Conn_Execer_NamedValueChecker_ColumnConverter, *w_conn_Conn_ExecerContext_Queryer_QueryerContext, *w_conn_Conn_ExecerContext_Queryer_ConnPrepareContext, *w_conn_Conn_ExecerContext_Queryer_NamedValueChecker, *w_conn_Conn_ExecerContext_Queryer_ColumnConverter, *w_conn_Conn_ExecerContext_QueryerContext_ConnPrepareContext, *w_conn_Conn_ExecerContext_QueryerContext_NamedValueChecker, *w_conn_Conn_ExecerContext_QueryerContext_ColumnConverter, *w_conn_Conn_ExecerContext_ConnPrepareContext_NamedValueChecker, *w_conn_Conn_ExecerContext_ConnPrepareContext_ColumnConverter, *w_conn_Conn_ExecerContext_NamedValueChecker_ColumnConverter, *w_conn_Conn_Queryer_QueryerContext_ConnPrepareContext, *w_conn_Conn_Queryer_QueryerContext_NamedValueChecker, *w_conn_Conn_Queryer_QueryerContext_ColumnConverter, *w_conn_Conn_Queryer_ConnPrepareContext_NamedValueChecker, *w_conn_Conn_Queryer_ConnPrepareContext_ColumnConverter, *w_conn_Conn_Queryer_NamedValueChecker_ColumnConverter, *w_conn_Conn_QueryerContext_ConnPrepareContext_NamedValueChecker, *w_conn_Conn_QueryerContext_ConnPrepareContext_ColumnConverter, *w_conn_Conn_QueryerContext_NamedValueChecker_ColumnConverter, *w_conn_Conn_ConnPrepareContext_NamedValueChecker_ColumnConverter, *w_conn_Conn_Execer_ExecerContext, *w_conn_Conn_Execer_Queryer, *w_conn_Conn_Execer_QueryerContext, *w_conn_Conn_Execer_ConnPrepareContext, *w_conn_Conn_Execer_NamedValueChecker, *w_conn_Conn_Execer_ColumnConverter, *w_conn_Conn_ExecerContext_Queryer, *w_conn_Conn_ExecerContext_QueryerContext, *w_conn_Conn_ExecerContext_ConnPrepareContext, *w_conn_Conn_ExecerContext_NamedValueChecker, *w_conn_Conn_ExecerContext_ColumnConverter, *w_conn_Conn_Queryer_QueryerContext, *w_conn_Conn_Queryer_ConnPrepareContext, *w_conn_Conn_Queryer_NamedValueChecker, *w_conn_Conn_Queryer_ColumnConverter, *w_conn_Conn_QueryerContext_ConnPrepareContext, *w_conn_Conn_QueryerContext_NamedValueChecker, *w_conn_Conn_QueryerContext_ColumnConverter, *w_conn_Conn_ConnPrepareContext_NamedValueChecker, *w_conn_Conn_ConnPrepareContext_ColumnConverter, *w_conn_Conn_NamedValueChecker_ColumnConverter, *w_conn_Conn_Execer, *w_conn_Conn_ExecerContext, *w_conn_Conn_Queryer, *w_conn_Conn_QueryerContext, *w_conn_Conn_ConnPrepareContext, *w_conn_Conn_NamedValueChecker, *w_conn_Conn_ColumnConverter, *w_conn_Conn:
		return true
	}
	return false
}
