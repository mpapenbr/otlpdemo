package config

import (
	"crypto/tls"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"

	"github.com/mpapenbr/otlpdemo/log"
)

type TLSReloader struct {
	certPath      string
	keyPath       string
	caPaths       []string
	clientCAPaths []string

	mu        sync.RWMutex
	tlsConfig *tls.Config
}

//nolint:whitespace //editor/linter issue
func NewTLSReloader() (
	*TLSReloader,
	error,
) {
	r := &TLSReloader{
		certPath:      TLSCert,
		keyPath:       TLSKey,
		caPaths:       TLSCAs,
		clientCAPaths: TLSClientCAs,
	}
	if err := r.reload(); err != nil {
		return nil, err
	}
	return r, nil
}

//nolint:funlen,nestif // by design
func (r *TLSReloader) reload() error {
	newCfg, err := buildTLSFromConfig()
	if err != nil {
		return err
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

	for _, clientCAPath := range r.clientCAPaths {
		watchPaths[filepath.Dir(clientCAPath)] = struct{}{}
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
