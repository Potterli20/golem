package zdb

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Potterli20/golem/pkg/logger"
	"github.com/Potterli20/golem/pkg/zdb/zdbconfig"
	"github.com/Potterli20/golem/pkg/zdb/zdbconnector"
	"gorm.io/gorm/clause"

	"gorm.io/gorm"
)

const (
	retryDefault       = 3
	maxAttemptsDefault = 5
)

type ZDatabase interface {
	Find(out any, where ...any) ZDatabase
	First(dest any, where ...any) ZDatabase
	FirstOrCreate(dest any, where ...any) ZDatabase
	Scan(dest any) ZDatabase
	Rows() (*sql.Rows, error)
	ScanRows(rows *sql.Rows, result any) error
	Select(query any, args ...any) ZDatabase
	Where(query any, args ...any) ZDatabase
	Joins(query string, args ...any) ZDatabase
	UnionAll(subQuery1 ZDatabase, subQuery2 ZDatabase) ZDatabase
	UnionDistinct(subQuery1 ZDatabase, subQuery2 ZDatabase) ZDatabase
	Limit(limit int) ZDatabase
	Offset(offset int) ZDatabase
	Order(value any) ZDatabase
	Distinct(args ...any) ZDatabase
	Count(count *int64) ZDatabase
	Group(name string) ZDatabase
	Create(value any) ZDatabase
	Updates(value any) ZDatabase
	Update(column string, value any) ZDatabase
	Delete(value any, where ...any) ZDatabase
	Raw(sql string, values ...any) ZDatabase
	Exec(sql string, values ...any) ZDatabase
	Table(name string, args ...any) ZDatabase
	Transaction(fc func(tx ZDatabase) error, opts ...*sql.TxOptions) (err error)
	Clauses(conds ...clause.Expression) ZDatabase
	WithContext(ctx context.Context) ZDatabase
	Error() error
	Scopes(funcs ...func(ZDatabase) ZDatabase) ZDatabase
	RowsAffected() int64
	GetDbConnection() *gorm.DB
	GetDBStats() (sql.DBStats, error)
}

type zDatabase struct {
	db *gorm.DB
}

func NewInstance(dbType string, config *zdbconfig.Config) (ZDatabase, error) {
	if config.RetryInterval == 0 {
		config.RetryInterval = retryDefault
	}

	if config.MaxAttempts == 0 {
		config.MaxAttempts = maxAttemptsDefault
	}

	connector, ok := zdbconnector.Connectors[dbType]
	if !ok {
		return nil, fmt.Errorf("unsupported database type %s", dbType)
	}

	var dbConn *gorm.DB
	var err error

	for i := 0; i < config.MaxAttempts; i++ {
		dbConn, err = connector.Connect(config)
		if err == nil {
			verifyErr := connector.VerifyConnection(dbConn)
			if verifyErr == nil {
				// Setup OpenTelemetry instrumentation with configuration
				if err := setupOpenTelemetryInstrumentation(dbConn, &config.OpenTelemetry); err != nil {
					return nil, fmt.Errorf("failed to setup OpenTelemetry instrumentation: %w", err)
				}

				return &zDatabase{db: dbConn}, nil
			}

			err = verifyErr
		}

		logger.GetLoggerFromContext(context.Background()).Infof("Failed to establish database connection: %v. Attempt %d/%d. Retrying in %d seconds...", err, i+1, config.MaxAttempts, config.RetryInterval)
		time.Sleep(time.Duration(config.RetryInterval) * time.Second)
	}

	logger.GetLoggerFromContext(context.Background()).Infof("Unable to establish database connection after %d attempts.", config.MaxAttempts)
	return nil, err
}
