package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"strings"
	"time"
)

type ProxyPass struct {
	Target string
}

var proxyConfig map[string]ProxyPass = make(map[string]ProxyPass)

func main() {

	config := LoadConfigFromEnv()

	proxy := &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			log.Printf("%v", r.In.Header)
			log.Printf("http in %s %v host %s", r.In.Method, r.In.URL, r.In.Host)

			prefix := getPrefix(r.In.URL.Path)
			log.Printf("checking config for prefix %s", prefix)
			if pp, ok := proxyConfig[prefix]; ok {
				log.Printf("found config for prefix %s", prefix)
				targetUrl, _ := url.Parse(pp.Target)

				if canReach(targetUrl) {
					log.Printf("routing to target %s", targetUrl)
					r.SetURL(targetUrl)
					return
				}
			}

			// Handle missing rules
			log.Printf("no route set for %s", prefix)
			holdingPageUrl, _ := url.Parse(fmt.Sprintf("http://localhost:%d/wait", config.Port))
			r.Out.URL = holdingPageUrl
			r.Out.URL.RawQuery = "id=" + url.QueryEscape(r.In.URL.Path)
			log.Printf("%v", r.Out)

		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/register", registerRoute)
	mux.HandleFunc("/wait", notReady)
	mux.HandleFunc("/routes", listRoutes)

	mux.Handle("/", proxy)

	log.Printf("starting reverse proxy on port %d", config.Port)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port),
		Handler: mux,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

// Gets the first segment of a path.
// e.g. `/1234/foo -> 1234`
func getPrefix(p string) string {
	parts := strings.Split(path.Clean(p), "/")
	if len(parts) < 2 {
		return "/"
	}
	return parts[1]
}

// Is the target reachable/resolvable?
func canReach(url *url.URL) bool {
	timeout := time.Second * 1
	target := fmt.Sprintf("%s:%s", url.Hostname(), url.Port())
	log.Printf("checking %s", target)
	_, err := net.DialTimeout("tcp", target, timeout)
	return err == nil
}

// Internal handler for showing some sort of holding page in the future
func notReady(w http.ResponseWriter, r *http.Request) {
	log.Printf("Not ready: %v", r.URL)
	w.WriteHeader(202)
	fmt.Fprintf(
		w,
		"DEBUG:\nThis will be a holding page that redirects the user back to /%s after x seconds",
		r.URL.Query().Get("id"),
	)
}

// Healthcheck handler
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
}

// Dynamically add a route
func registerRoute(w http.ResponseWriter, r *http.Request) {
	// get payload
	q := r.URL.Query()
	id := q.Get("id")
	target := q.Get("target")

	if id == "" || target == "" {
		w.WriteHeader(400)
		fmt.Fprintf(w, "invalid request")
		return
	}

	// create route
	target, err := url.QueryUnescape(target)
	if err != nil {
		w.WriteHeader(400)
		fmt.Fprintf(w, "failed to unescape %s", target)
		return
	}

	proxyConfig[id] = ProxyPass{target}
	log.Printf("added route %s to %s", id, target)

	w.WriteHeader(200)
}

func listRoutes(w http.ResponseWriter, r *http.Request) {

	fmt.Fprint(w, "DEUBG: ROUTES\n")
	for token, target := range proxyConfig {
		fmt.Fprintf(w, "%s -> %s\n", token, target)
	}
	w.WriteHeader(200)
}
