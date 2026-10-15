// Package database provides multi-database connection management.
package database

import (
	"context"
	"fmt"
	"sync"
	"time"

	"cms-backend/internal/config"
	"cms-backend/internal/models"
	"cms-backend/pkg/logger"

	"github.com/glebarez/sqlite"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// DBConnection represents a live database connection
type DBConnection struct {
	ID              string
	Name            string
	Type            string
	Host            string
	Port            int
	Username        string
	Password        string
	Database        string
	Schema          string
	SSLMode         string
	SSLCert         string
	SSLKey          string
	SSLRootCert     string
	CustomURL       string
	IsActive        bool
	IsDefault       bool
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	DB              *gorm.DB
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Metadata        map[string]string
}

// ConnectionManager manages multiple database connections
type ConnectionManager struct {
	mu          sync.RWMutex
	connections map[string]*DBConnection
	defaultDB   string
	redisClient *redis.Client
	log         logger.Logger
}

var (
	managerInstance *ConnectionManager
	managerOnce     sync.Once
)

// GetConnectionManager returns the singleton connection manager
func GetConnectionManager() *ConnectionManager {
	managerOnce.Do(func() {
		managerInstance = &ConnectionManager{
			connections: make(map[string]*DBConnection),
			log:         logger.GetLogger(),
		}
	})
	return managerInstance
}

// Initialize sets up the default connection and runs auto-migrations
func (cm *ConnectionManager) Initialize(cfg *config.Config) error {
	conn := &DBConnection{
		ID:              "default",
		Name:            "Default",
		Type:            string(cfg.Database.Type),
		Host:            cfg.Database.Host,
		Port:            cfg.Database.Port,
		Username:        cfg.Database.Username,
		Password:        cfg.Database.Password,
		Database:        cfg.Database.Database,
		Schema:          cfg.Database.Schema,
		SSLMode:         cfg.Database.SSLMode,
		SSLCert:         cfg.Database.SSLCert,
		SSLKey:          cfg.Database.SSLKey,
		SSLRootCert:     cfg.Database.SSLRootCert,
		CustomURL:       cfg.Database.CustomURL,
		IsActive:        true,
		IsDefault:       true,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
		ConnMaxIdleTime: cfg.Database.ConnMaxIdleTime,
		Metadata:        make(map[string]string),
	}

	if err := cm.connect(conn); err != nil {
		return fmt.Errorf("failed to connect to default database: %w", err)
	}

	// Auto-migrate all models
	if err := conn.DB.AutoMigrate(models.AllModels()...); err != nil {
		return fmt.Errorf("failed to auto-migrate: %w", err)
	}

	cm.mu.Lock()
	cm.connections["default"] = conn
	cm.defaultDB = "default"
	cm.mu.Unlock()

	// Seed default data
	if err := cm.seedDefaults(conn.DB); err != nil {
		cm.log.Warn("Failed to seed default data", "error", err)
	}

	if cfg.Redis.Enabled {
		if err := cm.initRedis(cfg); err != nil {
			cm.log.Warn("Redis init failed", "error", err)
		}
	}

	cm.log.Info("Connection manager initialized", "db_type", cfg.Database.Type)
	return nil
}

// CreateConnection creates and stores a new DB connection (from API)
func (cm *ConnectionManager) CreateConnection(ctx context.Context, dbConn *models.DatabaseConnection, password string) (*DBConnection, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	id := fmt.Sprintf("conn_%d", dbConn.ID)
	if _, exists := cm.connections[id]; exists {
		return cm.connections[id], nil
	}

	conn := &DBConnection{
		ID:              id,
		Name:            dbConn.Name,
		Type:            dbConn.Type,
		Host:            dbConn.Host,
		Port:            dbConn.Port,
		Username:        dbConn.Username,
		Password:        password,
		Database:        dbConn.Database,
		Schema:          dbConn.Schema,
		SSLMode:         dbConn.SSLMode,
		IsActive:        true,
		IsDefault:       dbConn.IsDefault,
		MaxOpenConns:    max(dbConn.MaxOpenConns, 5),
		MaxIdleConns:    max(dbConn.MaxIdleConns, 2),
		ConnMaxLifetime: time.Duration(dbConn.ConnMaxLifetime) * time.Second,
		ConnMaxIdleTime: time.Duration(dbConn.ConnMaxIdleTime) * time.Second,
	}

	if err := cm.connect(conn); err != nil {
		return nil, err
	}
	cm.connections[id] = conn
	return conn, nil
}

// GetConnection retrieves a connection by model ID
func (cm *ConnectionManager) GetConnection(connID uint) (*DBConnection, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	id := fmt.Sprintf("conn_%d", connID)
	if c, ok := cm.connections[id]; ok {
		return c, nil
	}
	return nil, fmt.Errorf("connection %d not found", connID)
}

// GetDefaultConnection returns the default database connection
func (cm *ConnectionManager) GetDefaultConnection() (*DBConnection, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	if c, ok := cm.connections[cm.defaultDB]; ok {
		return c, nil
	}
	return nil, fmt.Errorf("no default connection available")
}

// GetConnectionForService returns the DB for a service (or default)
func (cm *ConnectionManager) GetConnectionForService(svc *models.Service) (*DBConnection, error) {
	if svc.DatabaseConnectionID != nil && *svc.DatabaseConnectionID > 0 {
		return cm.GetConnection(*svc.DatabaseConnectionID)
	}
	return cm.GetDefaultConnection()
}

// GetRedis returns the Redis client
func (cm *ConnectionManager) GetRedis() *redis.Client {
	return cm.redisClient
}

// Close closes all connections
func (cm *ConnectionManager) Close() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	for _, conn := range cm.connections {
		if conn.DB != nil {
			db, _ := conn.DB.DB()
			if db != nil {
				_ = db.Close()
			}
		}
	}
	if cm.redisClient != nil {
		_ = cm.redisClient.Close()
	}
}

// connect establishes an actual DB connection
func (cm *ConnectionManager) connect(conn *DBConnection) error {
	dialector, err := cm.createDialector(conn)
	if err != nil {
		return err
	}

	gormCfg := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{SingularTable: false},
		Logger: gormlogger.New(
			&gormWriterAdapter{cm.log},
			gormlogger.Config{
				SlowThreshold:             200 * time.Millisecond,
				LogLevel:                  gormlogger.Warn,
				IgnoreRecordNotFoundError: true,
			},
		),
	}

	db, err := gorm.Open(dialector, gormCfg)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	maxOpen := conn.MaxOpenConns
	if maxOpen <= 0 {
		maxOpen = 25
	}
	maxIdle := conn.MaxIdleConns
	if maxIdle <= 0 {
		maxIdle = 10
	}
	lifetime := conn.ConnMaxLifetime
	if lifetime <= 0 {
		lifetime = 5 * time.Minute
	}

	sqlDB.SetMaxOpenConns(maxOpen)
	sqlDB.SetMaxIdleConns(maxIdle)
	sqlDB.SetConnMaxLifetime(lifetime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	conn.DB = db
	return nil
}

func (cm *ConnectionManager) createDialector(conn *DBConnection) (gorm.Dialector, error) {
	dsn := GetDSN(conn)
	switch conn.Type {
	case string(config.PostgreSQL), "postgres":
		return postgres.Open(dsn), nil
	case string(config.MySQL):
		return mysql.Open(dsn), nil
	case string(config.SQLite), "sqlite3":
		return sqlite.Open(conn.Database), nil
	case string(config.SQLServer):
		return sqlserver.Open(dsn), nil
	default:
		return nil, fmt.Errorf("unsupported database type: %s", conn.Type)
	}
}

// GetDSN builds the connection string for a given connection
func GetDSN(conn *DBConnection) string {
	if conn.CustomURL != "" {
		return conn.CustomURL
	}
	switch conn.Type {
	case string(config.PostgreSQL), "postgres":
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			conn.Host, conn.Port, conn.Username, conn.Password, conn.Database, conn.SSLMode)
		if conn.Schema != "" && conn.Schema != "public" {
			dsn += " search_path=" + conn.Schema
		}
		return dsn
	case string(config.MySQL):
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			conn.Username, conn.Password, conn.Host, conn.Port, conn.Database)
	case string(config.SQLite), "sqlite3":
		return conn.Database
	case string(config.SQLServer):
		return fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s",
			conn.Username, conn.Password, conn.Host, conn.Port, conn.Database)
	}
	return ""
}

func (cm *ConnectionManager) initRedis(cfg *config.Config) error {
	cm.redisClient = redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
		MaxRetries:   cfg.Redis.MaxRetries,
		DialTimeout:  cfg.Redis.DialTimeout,
		ReadTimeout:  cfg.Redis.ReadTimeout,
		WriteTimeout: cfg.Redis.WriteTimeout,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := cm.redisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}
	cm.log.Info("Redis connected")
	return nil
}

// seedDefaults creates the default superadmin and roles
func (cm *ConnectionManager) seedDefaults(db *gorm.DB) error {
	var count int64
	db.Model(&models.Role{}).Count(&count)
	if count > 0 {
		return nil
	}

	roles := []models.Role{
		{Name: "super_admin", Description: "Full system access", IsActive: true},
		{Name: "admin", Description: "Admin access", IsActive: true},
		{Name: "editor", Description: "Can create and edit content", IsActive: true},
		{Name: "viewer", Description: "Read-only access", IsActive: true},
	}
	return db.CreateInBatches(roles, 10).Error
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// TestConnection tests a connection without persisting it
func TestConnection(connType, host string, port int, user, pass, dbName, sslMode, customURL string) error {
	conn := &DBConnection{
		Type:      connType,
		Host:      host,
		Port:      port,
		Username:  user,
		Password:  pass,
		Database:  dbName,
		SSLMode:   sslMode,
		CustomURL: customURL,
	}
	cm := GetConnectionManager()
	if err := cm.connect(conn); err != nil {
		return err
	}
	if conn.DB != nil {
		db, _ := conn.DB.DB()
		if db != nil {
			_ = db.Close()
		}
	}
	return nil
}

// gormWriterAdapter adapts our Logger to satisfy gorm's logger.Writer interface
type gormWriterAdapter struct {
	log logger.Logger
}

func (g *gormWriterAdapter) Printf(format string, args ...any) {
	g.log.Info(fmt.Sprintf(format, args...))
}
