package config

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"os"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"

	"github.com/mpapenbr/otlpdemo/log"
)

type TLSReloader struct {
	certPath string
	keyPath  string
	caPaths  []string

	mu        sync.RWMutex
	tlsConfig *tls.Config
}

//nolint:whitespace //editor/linter issue
func NewTLSReloader(certPath, keyPath string, caPaths []string) (
	*TLSReloader,
	error,
) {
	r := &TLSReloader{
		certPath: certPath,
		keyPath:  keyPath,
		caPaths:  caPaths,
	}
	if err := r.reload(); err != nil {
		return nil, err
	}
	return r, nil
}

//nolint:funlen,nestif // by design
func (r *TLSReloader) reload() error {
	newCfg := &tls.Config{
		MinVersion: tls.VersionTLS13,
	}
	if r.certPath != "" && r.keyPath != "" {
		log.Debug("cert and key provided")
		cert, err := tls.LoadX509KeyPair(r.certPath, r.keyPath)
		if err != nil {
			return err
		}
		newCfg.Certificates = []tls.Certificate{cert}
	}

	if len(r.caPaths) > 0 {
		caPool := x509.NewCertPool()
		for _, path := range r.caPaths {
			if path != "" {
				log.Debug("ca provided", log.String("path", path))
				caBytes, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				if !caPool.AppendCertsFromPEM(caBytes) {
					return errors.New("invalid client CA in: " + path)
				}
			}
		}
		newCfg.ClientCAs = caPool
	}

	if TLSClientAuth != "" {
		log.Debug("clientAuth provided")
		var clientAuth tls.ClientAuthType
		clientAuth, err := ParseClientAuth(TLSClientAuth)
		if err != nil {
			return err
		}
		newCfg.ClientAuth = clientAuth
	}
	if TLSSkipVerify {
		log.Debug("skipVerify enabled")
		newCfg.InsecureSkipVerify = true
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tlsConfig = newCfg
	log.Debug("Reloaded TLS config (server cert + multiple client CAs)")
	return nil
}

func (r *TLSReloader) GetConfigForClient(_ *tls.ClientHelloInfo) (*tls.Config, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	log.Debug("GetConfigForClient callback invoked")
	return r.tlsConfig, nil
}

//nolint:funlen // by design
func (r *TLSReloader) watch() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("watcher error:", log.ErrorField(err))
	}
	defer watcher.Close()

	watchPaths := map[string]struct{}{
		filepath.Dir(r.certPath): {},
		filepath.Dir(r.keyPath):  {},
	}
	for _, caPath := range r.caPaths {
		watchPaths[filepath.Dir(caPath)] = struct{}{}
	}

	for dir := range watchPaths {
		if err := watcher.Add(dir); err != nil {
			log.Fatal("fsnotify watch error ",
				log.String("dir", dir),
				log.ErrorField(err))
		}
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			filesToWatch := append([]string{r.certPath, r.keyPath}, r.caPaths...)
			for _, watchFile := range filesToWatch {
				if filepath.Base(event.Name) == filepath.Base(watchFile) &&
					(event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename)) != 0 {

					log.Debug("Detected change in", log.String("name", event.Name))
					if err := r.reload(); err != nil {
						log.Error("error reloading", log.ErrorField(err))
					}
					break
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Error("watcher error", log.ErrorField(err))
		}
	}
}
