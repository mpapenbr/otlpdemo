package config

import (
	"crypto/tls"
	"path/filepath"
	"slices"
	"sync"
	"time"

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

	filesToWatch := []string{}
	watchPaths := map[string]struct{}{
		filepath.Dir(r.certPath): {},
		filepath.Dir(r.keyPath):  {},
	}
	filesToWatch = append(filesToWatch, r.certPath)
	filesToWatch = append(filesToWatch, r.keyPath)

	for _, caPath := range r.caPaths {
		watchPaths[filepath.Dir(caPath)] = struct{}{}
		filesToWatch = append(filesToWatch, caPath)
	}

	for _, clientCAPath := range r.clientCAPaths {
		watchPaths[filepath.Dir(clientCAPath)] = struct{}{}
		filesToWatch = append(filesToWatch, clientCAPath)
	}

	for dir := range watchPaths {
		if err := watcher.Add(dir); err != nil {
			log.Fatal("fsnotify watch error ",
				log.String("dir", dir),
				log.ErrorField(err))
		}
		// check if link ..data exists in dir (e.g., when using cert-manager in k8s)
		dataLink := filepath.Join(dir, "..data")
		if x, err := filepath.Glob(dataLink); err == nil && len(x) > 0 {
			log.Debug("Added ..data link to watch",
				log.String("dir", dataLink))
			filesToWatch = append(filesToWatch, dataLink)
		}
	}
	// resolve symlinks
	for _, file := range filesToWatch {
		resolved, err := filepath.EvalSymlinks(file)
		if err != nil {
			log.Error("error resolving symlink",
				log.String("file", file),
				log.ErrorField(err))
			continue
		}
		filesToWatch = append(filesToWatch, resolved)
	}

	log.Debug("Watching filenames for changes",
		log.Any("files", filesToWatch))

	isCertFile := func(name string) bool {
		ret := slices.Contains(filesToWatch, name)
		log.Debug("Checking if file is cert file",
			log.String("name", name),
			log.Bool("result", ret))
		return ret
	}
	var lastReload time.Time
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			log.Debug("event", log.Any("name", event))
			if isCertFile(event.Name) && time.Since(lastReload) > time.Second {
				lastReload = time.Now()
				time.Sleep(time.Second) // debounce
				log.Debug("Change in dir detected. reloading certs",
					log.String("name", event.Name))
				if err := r.reload(); err != nil {
					log.Error("error reloading", log.ErrorField(err))
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
