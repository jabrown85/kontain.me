// Copyright 2018 Google LLC All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/ko/pkg/build"
	"github.com/imjasonh/kontain.me/pkg/run"
	"github.com/imjasonh/kontain.me/pkg/serve"
)

func main() {
	http.Handle("/v2/", &server{
		info:  log.New(os.Stdout, "I ", log.Ldate|log.Ltime|log.Lshortfile),
		error: log.New(os.Stderr, "E ", log.Ldate|log.Ltime|log.Lshortfile),
	})

	log.Println("Starting...")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}
	log.Printf("Listening on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

type server struct {
	info, error *log.Logger
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.info.Println("handler:", r.Method, r.URL)
	path := strings.TrimPrefix(r.URL.String(), "/v2/")

	switch {
	case path == "":
		// API Version check.
		w.Header().Set("Docker-Distribution-API-Version", "registry/2.0")
		return
	case strings.Contains(path, "/blobs/"),
		strings.Contains(path, "/manifests/sha256:"):
		// Extract requested blob digest and redirect to serve it from GCS.
		// If it doesn't exist, this will return 404.
		parts := strings.Split(r.URL.Path, "/")
		digest := parts[len(parts)-1]
		serve.Blob(w, r, digest)
	case strings.Contains(path, "/manifests/"):
		s.serveKoManifest(w, r)
	default:
		serve.Error(w, serve.ErrNotFound)
	}
}

func getDefaultBaseImage(string) (v1.Image, error) {
	// TODO: memoize
	ref, err := name.ParseReference("gcr.io/distroless/base", name.WeakValidation)
	if err != nil {
		return nil, err
	}
	return remote.Image(ref)
}

// konta.in/ko/github.com/knative/build/cmd/controller -> ko build and serve
func (s *server) serveKoManifest(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v2/ko/")
	parts := strings.Split(path, "/")
	ip := strings.Join(parts[:len(parts)-2], "/")

	tag := parts[len(parts)-1]
	s.info.Printf("requested image tag :%s", tag)

	// go get the package.
	s.info.Printf("go get %s...", ip)
	if err := run.Do(s.info.Writer(), fmt.Sprintf("go get %s", ip)); err != nil {
		s.error.Printf("ERROR (go get): %s", err)
		serve.Error(w, err)
		return
	}
	// TODO: Check image tag for version, resolve branches -> commits and redirect to img:<commit>
	// TODO: For requests for commit SHAs, check if it's already built and serve that instead.
	// TODO: Look for $GOPATH/$importPath/.ko.yaml and up, to base image config.

	// ko build the package.
	g, err := build.NewGo(
		build.WithBaseImages(getDefaultBaseImage),
		build.WithCreationTime(v1.Time{time.Unix(0, 0)}),
	)
	if err != nil {
		s.error.Printf("ERROR (build.NewGo): %s", err)
		serve.Error(w, serve.ErrInvalid)
		return
	}
	if !g.IsSupportedReference(ip) {
		s.error.Printf("ERROR (IsSupportedReference): %s", err)
		serve.Error(w, serve.ErrInvalid)
		return
	}
	s.info.Printf("ko build %s...", ip)
	img, err := g.Build(ip)
	if err != nil {
		s.error.Printf("ERROR (ko build): %s", err)
		serve.Error(w, serve.ErrInvalid)
		return
	}
	serve.Manifest(w, r, img)
}
