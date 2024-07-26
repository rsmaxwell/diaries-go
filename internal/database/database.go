/* see:
 *    https://github.com/eclipse/paho.golang/blob/v0.21.0/autopaho/examples/basics/basics.go
 *    https://github.com/eclipse/paho.golang/blob/master/autopaho/examples/rpc/main.go
 */

package database

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/rsmaxwell/diaries/internal/config"

	_ "github.com/lib/pq"
)

func Connect(dBConfig *config.DBConfig) (*sql.DB, error) {

	driverName := dBConfig.DriverName()
	databaseName := dBConfig.Database

	connectionString := dBConfig.ConnectionString(databaseName)

	db, err := sql.Open(driverName, connectionString)
	if err != nil {
		slog.Error(err.Error())
		slog.Error(fmt.Sprintf("driverName: %s", driverName))
		slog.Error(fmt.Sprintf("connectionString: %s", connectionString))
		return nil, fmt.Errorf("could not connect to database")
	}

	return db, err
}
