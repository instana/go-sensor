// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instana

import (
	"context"
	"database/sql/driver"
)

type wrappedSQLConnector struct {
	driver.Connector

	connDetails dbConnDetails
	sensor      *Sensor
}

// WrapSQLConnector wraps an existing sql.Connector and instruments the DB calls made using it
func WrapSQLConnector(sensor *Sensor, name string, connector driver.Connector) *wrappedSQLConnector {
	if c, ok := connector.(*wrappedSQLConnector); ok {
		return c
	}

	return &wrappedSQLConnector{
		Connector:   connector,
		connDetails: parseDBConnDetails(name),
		sensor:      sensor,
	}
}

func (c *wrappedSQLConnector) Connect(ctx context.Context) (driver.Conn, error) {
	conn, err := c.Connector.Connect(ctx)
	if err != nil {
		return conn, err
	}

	if connAlreadyWrapped(conn) {
		return conn, nil
	}

	w := wrapConn(c.connDetails, conn, c.sensor)

	return w, nil
}

func (c *wrappedSQLConnector) Driver() driver.Driver {
	if drv, ok := c.Connector.Driver().(*wrappedSQLDriver); ok {
		return drv
	}

	return &wrappedSQLDriver{
		Driver: c.Connector.Driver(),
		sensor: c.sensor,
	}
}

func (drv *wrappedSQLDriver) OpenConnector(name string) (driver.Connector, error) {
	var connector driver.Connector = dsnConnector{dsn: name, driver: drv}

	if d, ok := drv.Driver.(driver.DriverContext); ok {
		var err error

		connector, err = d.OpenConnector(name)
		if err != nil {
			return connector, err
		}
	}

	if connector, ok := connector.(*wrappedSQLConnector); ok {
		return connector, nil
	}

	return WrapSQLConnector(drv.sensor, name, connector), nil
}
