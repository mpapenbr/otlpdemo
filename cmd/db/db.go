package db

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/mpapenbr/otlpdemo/cmd/config"
	"github.com/mpapenbr/otlpdemo/log"
	"github.com/mpapenbr/otlpdemo/version"
)

//nolint:errcheck,lll // by design
func NewDBCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "db",
		Short: "tests database connectivity",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			doDBStuff()
			return nil
		},
	}
	cmd.Flags().StringVar(&config.DBConf.Host, "host", "localhost", "database host")
	cmd.Flags().IntVar(&config.DBConf.Port, "port", 5432, "database port")
	cmd.Flags().StringVar(&config.DBConf.Database, "database", "postgres", "database name")
	cmd.Flags().StringVar(&config.DBConf.StaticSecrets.User, "user", "", "database user")
	cmd.Flags().StringVar(&config.DBConf.StaticSecrets.Password, "password", "", "database password")
	cmd.Flags().StringVar(&config.DBConf.SecretsFile, "secrets-file", "", "path to secrets file")
	cmd.Flags().StringVar(&config.DBConf.SSLMode, "sslmode", "disable", "database SSL mode")
	cmd.Flags().StringVar(&config.DBConf.TLSCert, "tls-cert", "", "path to TLS certificate")
	cmd.Flags().StringVar(&config.DBConf.TLSKey, "tls-key", "", "path to TLS key")
	cmd.Flags().StringVar(&config.DBConf.TLSCA, "tls-ca", "", "path to TLS CA")
	cmd.Flags().DurationVar(&poolMaxLife, "pool-max-life", 5*time.Minute, "maximum lifetime of a connection in the pool")
	cmd.Flags().DurationVar(&tickerDuration, "ticker", 5*time.Second, "duration for ticker interval")
	cmd.Flags().DurationVar(&longRunningTickerDuration, "long-running-ticker", 30*time.Second, "duration for long running ticker interval")
	cmd.Flags().DurationVar(&longRunningDuration, "long-running", 30*time.Second, "duration for long running query")
	return &cmd
}

var tickerDuration,
	longRunningTickerDuration,
	longRunningDuration,
	poolMaxLife time.Duration

func doDBStuff() error {
	log.Debug("Connecting to database", log.Any("conf", config.DBConf))
	dbDemo, err := newDemoDB(config.DBConf)
	if err != nil {
		log.Error("could not create DB connection", log.ErrorField(err))
		return err
	}
	log.Debug("Connected to database", log.Any("db", dbDemo))
	defer dbDemo.pool.Close()
	dbDemo.run()
	return nil
}

type (
	demoDB struct {
		sync.RWMutex
		pool   *pgxpool.Pool
		appCfg config.DBConfig
	}
)

func newDemoDB(conf config.DBConfig) (*demoDB, error) {
	pool, err := createPool(conf)
	if err != nil {
		return nil, fmt.Errorf("could not create DB pool: %w", err)
	}
	return &demoDB{
		pool:   pool,
		appCfg: conf,
	}, nil
}

//nolint:funlen // lots of stuff to do here
func createPool(conf config.DBConfig) (*pgxpool.Pool, error) {
	param := make(map[string]string)
	param["host"] = conf.Host
	param["port"] = fmt.Sprintf("%d", conf.Port)
	param["dbname"] = conf.Database
	param["application_name"] = fmt.Sprintf("otlpdemo-%s", version.Version)
	param["sslmode"] = "disable"
	if conf.SSLMode != "" {
		param["sslmode"] = conf.SSLMode
	}

	if conf.StaticSecrets.User != "" || conf.StaticSecrets.Password != "" {
		param["user"] = conf.StaticSecrets.User
		param["password"] = conf.StaticSecrets.Password
	}
	if conf.SecretsFile != "" {
		// load secrets from file
		b, err := os.ReadFile(conf.SecretsFile)
		if err != nil {
			return nil, fmt.Errorf("could not read secrets file: %w", err)
		}
		var secrets config.DBSecrets
		if err := json.Unmarshal(b, &secrets); err != nil {
			return nil, fmt.Errorf("could not unmarshal secrets file: %w", err)
		}
		param["user"] = secrets.User
		param["password"] = secrets.Password
	}

	connStr := strings.Join(
		lo.MapToSlice(
			lo.OmitBy(param, func(k, v string) bool {
				return v == ""
			}),
			func(k, v string) string {
				return fmt.Sprintf("%s=%s", k, v)
			}), " ")

	// Parse the config from connection string
	poolCfg, poolErr := pgxpool.ParseConfig(connStr)
	if poolErr != nil {
		return nil, fmt.Errorf("failed to parse pool config: %w", poolErr)
	}
	//nolint:nestif // by design
	if conf.TLSCert != "" || conf.TLSKey != "" || conf.TLSCA != "" {
		caCertPool := x509.NewCertPool()
		if conf.TLSCA != "" {
			log.Debug("loading CA", log.String("ca", conf.TLSCA))
			caCert, caErr := os.ReadFile(conf.TLSCA)
			if caErr != nil {
				return nil, fmt.Errorf("could not read CA file: %w", caErr)
			}
			if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
				return nil, fmt.Errorf("failed to append server certificate")
			}
		}
		poolCfg.ConnConfig.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS13,
			//nolint:whitespace // editor/linter issue
			GetClientCertificate: func(cri *tls.CertificateRequestInfo) (
				*tls.Certificate, error,
			) {
				cert, certErr := tls.LoadX509KeyPair(conf.TLSCert, conf.TLSKey)
				if certErr != nil {
					return nil, certErr
				}
				log.Debug("Providing client certificate",
					log.String("cn", cert.Leaf.Subject.CommonName),
					log.String("serial", cert.Leaf.SerialNumber.String()),
					log.Time("notAfter", cert.Leaf.NotAfter),
					log.Time("notBefore", cert.Leaf.NotBefore),
				)
				return &cert, nil
			},

			RootCAs:    caCertPool,
			ServerName: conf.Host,
		}
	}
	poolCfg.MaxConns = 10
	poolCfg.MinConns = 2
	if poolMaxLife > 0 {
		poolCfg.MaxConnLifetime = poolMaxLife
	}
	poolCfg.BeforeConnect = func(ctx context.Context, cfg *pgx.ConnConfig) error {
		if param["user"] != cfg.User || param["password"] != cfg.Password {
			if param["user"] != cfg.User {
				log.Debug("Updating DB user",
					log.String("old", cfg.User),
					log.String("new", param["user"]),
				)
			}
			if param["password"] != cfg.Password {
				log.Debug("Updating DB password")
			}
			cfg.User = param["user"]
			cfg.Password = param["password"]
		}
		log.Debug("Establishing new database connection",
			log.String("user", cfg.User),
			log.String("pwd", cfg.Password),
		)
		return nil
	}

	// Create the connection pool
	pool, poolErr := pgxpool.NewWithConfig(context.Background(), poolCfg)
	if poolErr != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", poolErr)
	}

	// Test the connection
	if pingErr := pool.Ping(context.Background()); pingErr != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", pingErr)
	}

	return pool, nil
}

func (db *demoDB) run() {
	if db.appCfg.SecretsFile != "" {
		log.Info("Using secrets from file", log.String("path", db.appCfg.SecretsFile))
		db.setupDynamicSecrets(db.appCfg.SecretsFile)
	}
	wg := sync.WaitGroup{}
	ticker := time.NewTicker(tickerDuration)
	defer ticker.Stop()
	longRunningTicker := time.NewTicker(longRunningTickerDuration)
	defer longRunningTicker.Stop()
	wg.Add(2)
	go func() {
		db.showDBUsers()
		for range ticker.C {
			db.showDBUsers()
		}
		wg.Done()
	}()
	go func() {
		db.simLongRunningQuery(longRunningDuration)
		for range longRunningTicker.C {
			db.simLongRunningQuery(longRunningDuration)
		}
		log.Info("Long running query ticker stopped")
		wg.Done()
	}()
	wg.Wait()
}

func (db *demoDB) setupDynamicSecrets(path string) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("json")
	if err := v.ReadInConfig(); err != nil {
		log.Error("could not read secrets file",
			log.String("path", path), log.ErrorField(err))
		return
	}
	v.OnConfigChange(func(in fsnotify.Event) {
		log.Info("secrets file changed, updating DB credentials",
			log.String("path", path),
			log.String("user", v.GetString("user")),
			log.String("password", v.GetString("password")),
		)
		// this is a workaround. The "right" way would be to keep reading the certificate
		// until it contains a matching common name to the user name
		// most of the time the delta between new db user and new certificate is <100ms
		time.Sleep(500 * time.Millisecond)
		newPool, err := createPool(db.appCfg)
		if err != nil {
			log.Error("failed to create new pool", log.ErrorField(err))
			return
		}
		db.Lock()
		oldPool := db.pool
		db.pool = newPool
		db.Unlock()
		if oldPool != nil {
			go func() {
				log.Info("closing old pool")
				oldPool.Close()
				log.Info("closed old pool")
			}()
		}
	})
	v.WatchConfig()
}

func (db *demoDB) simLongRunningQuery(d time.Duration) {
	log.Info("Starting long running query", log.Duration("duration", d))
	c, err := db.pool.Acquire(context.Background())
	if err != nil {
		log.Error("failed to acquire connection", log.ErrorField(err))
		return
	}
	defer c.Release()
	doInfo := func() {
		var now time.Time
		var username string
		err := c.QueryRow(
			context.Background(),
			"select now(),current_user").Scan(&now, &username)
		if err != nil {
			log.Error("query failed", log.ErrorField(err))
			return
		}
		log.Info("Long running query info",
			log.Time("now", now),
			log.String("username", username))
	}

	doInfo()
	time.Sleep(d)
	doInfo()
	log.Info("Long running query finished", log.Duration("duration", d))
}

func (db *demoDB) showDBUsers() {
	rows, err := db.pool.Query(context.Background(),
		"select usename,valuntil from pg_user")
	if err != nil {
		log.Error("query failed", log.ErrorField(err))
		return
	}
	defer rows.Close()

	curUsername, err := db.getCurrentUser()
	if err != nil {
		log.Error("failed to get current user", log.ErrorField(err))
		return
	}
	log.Info("Current User", log.String("username", curUsername))
	for rows.Next() {
		var username string
		var valuntil *time.Time
		if err := rows.Scan(&username, &valuntil); err != nil {
			log.Error("failed to scan row", log.ErrorField(err))
			return
		}

		log.Info("User",
			log.String("username", username),
			log.Timep("valuntil", valuntil))
	}
	if err := rows.Err(); err != nil {
		log.Error("rows iteration error", log.ErrorField(err))
	}
}

func (db *demoDB) getCurrentUser() (string, error) {
	rows, err := db.pool.Query(context.Background(),
		"select current_user")
	if err != nil {
		log.Error("query failed", log.ErrorField(err))
		return "", err
	}
	defer rows.Close()

	var username string
	for rows.Next() {
		if err := rows.Scan(&username); err != nil {
			log.Error("failed to scan row", log.ErrorField(err))
			return "", err
		}
	}
	if err := rows.Err(); err != nil {
		log.Error("rows iteration error", log.ErrorField(err))
	}
	return username, nil
}
