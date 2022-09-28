package mysql

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/shopastro/logs"

	"github.com/shopastro/go-common/common"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormopentracing "gorm.io/plugin/opentracing"
)

type (
	Model struct {
		CreatedAt int64 `json:"createdAt" gorm:"column:created_at"`
		UpdatedAt int64 `json:"updatedAt" gorm:"column:updated_at"`
		Valid     int32 `json:"valid" gorm:"column:valid"`
	}

	ConfigModel struct {
		CfgName                   string        `yaml:"cfgName"`
		Host                      string        `yaml:"host"`
		Port                      int64         `yaml:"port"`
		DbName                    string        `yaml:"dbname"`
		User                      string        `yaml:"user"`
		Password                  string        `yaml:"password"`
		Charset                   string        `yaml:"charset"`
		ParseTime                 bool          `yaml:"parseTime"`
		MaxIdle                   time.Duration `yaml:"maxIdle"`
		MaxLifetime               time.Duration `yaml:"maxLifetime"`
		MaxOpenConns              int           `yaml:"maxOpenConns"`
		MaxIdleConns              int           `yaml:"maxIdleConns"`
		Local                     bool          `yaml:"local"`
		Debug                     bool          `yaml:"debug"`
		InterpolateParams         bool          `yaml:"interpolateParams"`
		MultiStatements           bool          `yaml:"multiStatements"`
		DefaultStringSize         uint          `yaml:"defaultStringSize"`
		DontSupportRenameIndex    *bool         `yaml:"dontSupportRenameIndex"`
		DontSupportRenameColumn   *bool         `yaml:"dontSupportRenameColumn"`
		SkipInitializeWithVersion bool          `yaml:"skipInitializeWithVersion"`
	}
)

var (
	connMap sync.Map
	err     error
)

func NewMysql(cfg *ConfigModel) *ConfigModel {

	if cfg.DefaultStringSize <= 0 {
		cfg.DefaultStringSize = 512
	}

	if cfg.DontSupportRenameIndex == nil {
		cfg.DontSupportRenameIndex = common.NewTools().Bool(true)
	}

	if cfg.DontSupportRenameColumn == nil {
		cfg.DontSupportRenameColumn = common.NewTools().Bool(true)
	}

	return cfg
}

const (
	connectionDefault = "default"
)

func NewDBClient(ctx context.Context, key ...string) *gorm.DB {
	conn, ok := connMap.Load(getConnKey(key))
	if !ok {
		log.Println("The key corresponding to database was not found", key)
		return nil
	}

	if ctx == nil {
		ctx = context.Background()
	}

	return conn.(*gorm.DB)
}

func (m *ConfigModel) Connection(key ...string) *gorm.DB {
	dsn := fmt.Sprintf("%s:%s@(%s:%d)/%s?charset=%s&parseTime=%t",
		m.User,
		m.Password,
		m.Host,
		m.Port,
		m.DbName,
		m.Charset,
		m.ParseTime)

	if m.Local {
		dsn = fmt.Sprintf("%s&loc=%t", dsn, m.Local)
	}

	if m.InterpolateParams {
		dsn = fmt.Sprintf("%s&interpolateParams=%t", dsn, m.InterpolateParams)
	}

	if m.MultiStatements {
		dsn = fmt.Sprintf("%s&multiStatements=%t", dsn, m.MultiStatements)
	}

	var (
		gormConfig = new(gorm.Config)
	)
	if !m.Debug {
		gormConfig.Logger = NewLogger()
	}

	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       dsn,
		DefaultStringSize:         512,
		DontSupportRenameIndex:    common.NewTools().BoolValue(m.DontSupportRenameIndex),
		DontSupportRenameColumn:   common.NewTools().BoolValue(m.DontSupportRenameColumn),
		SkipInitializeWithVersion: m.SkipInitializeWithVersion,
	}), gormConfig)
	if err != nil {
		log.Fatal(err)
	}

	gormDB, err := db.DB()
	if err != nil {
		logs.Logger.Fatal("connection errors", zap.Error(err))
	}

	gormDB.SetConnMaxIdleTime(m.MaxIdle * time.Millisecond)
	gormDB.SetConnMaxLifetime(m.MaxLifetime * time.Hour)

	gormDB.SetMaxOpenConns(m.MaxOpenConns)
	gormDB.SetMaxIdleConns(m.MaxIdleConns)

	db.Set("gorm:table_options", "ENGINE=InnoDB")
	// if err := db.Use(prometheus.New(prometheus.Config{
	// 	DBName:          m.DbName,
	// 	RefreshInterval: 5,
	// 	MetricsCollector: []prometheus.MetricsCollector{
	// 		&prometheus.MySQL{
	// 			VariableNames: []string{"Threads_running"},
	// 		},
	// 	},
	// })); err != nil {
	// 	logs.Logger.Error("use prometheus plugin errors", zap.Error(err))
	// }

	if err := db.Use(gormopentracing.New()); err != nil {
		logs.Logger.Error("use gorm opentracing plugin errors", zap.Error(err))
	}

	if m.Debug {
		db = db.Debug()
	}

	connMap.LoadOrStore(getConnKey(key), db)
	return db
}

func (m *Model) BeforeSave(tx *gorm.DB) (err error) {
	m.UpdatedAt = common.NewTools().GetNowMillisecond()

	return nil
}

func (m *Model) BeforeCreate(tx *gorm.DB) (err error) {
	m.CreatedAt = common.NewTools().GetNowMillisecond()
	m.UpdatedAt = common.NewTools().GetNowMillisecond()

	return nil
}

func (m *Model) BeforeUpdate(tx *gorm.DB) (err error) {
	m.UpdatedAt = common.NewTools().GetNowMillisecond()

	return nil
}

func getConnKey(key []string) string {
	if len(key) == 1 {
		return key[0]
	}

	return connectionDefault
}
