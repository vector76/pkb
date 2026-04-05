package server

import (
	"context"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"time"

	"github.com/vector76/pkb/internal/kb"
	"github.com/vector76/pkb/internal/render"
	"github.com/vector76/pkb/web"
)

const staleThreshold = 60 * time.Second

// templates holds one parsed template set per page so that each page's
// {{define "content"}} block correctly overrides the base template's
// {{block "content"}} without conflicting with other pages.
type templates struct {
	wiki     *template.Template
	conv     *template.Template
	convList *template.Template
	search   *template.Template
	errPage  *template.Template
}

// Server is the PKB HTTP server.
type Server struct {
	kb       *kb.KB
	renderer *render.Renderer
	hub      *Hub
	mux      *http.ServeMux
	tmpl     *templates
	addr     string
}

// New creates a Server. addr is the listen address, e.g. "127.0.0.1:4242".
func New(kbase *kb.KB, addr string) (*Server, error) {
	tmpl, err := parseTemplates()
	if err != nil {
		return nil, err
	}

	s := &Server{
		kb:       kbase,
		renderer: render.New("/wiki/"),
		hub:      NewHub(),
		mux:      http.NewServeMux(),
		tmpl:     tmpl,
		addr:     addr,
	}
	s.registerRoutes()
	return s, nil
}

func parseTemplates() (*templates, error) {
	sub, err := fs.Sub(web.Assets, "templates")
	if err != nil {
		return nil, err
	}
	parse := func(names ...string) (*template.Template, error) {
		return template.New("").ParseFS(sub, names...)
	}
	var t templates
	if t.wiki, err = parse("base.html", "wiki.html"); err != nil {
		return nil, err
	}
	if t.conv, err = parse("base.html", "conversation.html"); err != nil {
		return nil, err
	}
	if t.convList, err = parse("base.html", "conversation_list.html"); err != nil {
		return nil, err
	}
	if t.search, err = parse("base.html", "search.html"); err != nil {
		return nil, err
	}
	if t.errPage, err = parse("base.html", "error.html"); err != nil {
		return nil, err
	}
	return &t, nil
}

func (s *Server) registerRoutes() {
	// Static assets.
	staticFS, err := fs.Sub(web.Assets, "static")
	if err != nil {
		panic("pkb: embedded static assets missing: " + err.Error())
	}
	s.mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	// Wiki.
	s.mux.HandleFunc("GET /", s.handleRoot)
	s.mux.HandleFunc("GET /wiki/", s.handleWikiIndex)
	s.mux.HandleFunc("GET /wiki/{page}", s.handleWikiPage)

	// Conversations.
	s.mux.HandleFunc("GET /conversations/", s.handleConversationList)
	s.mux.HandleFunc("GET /conversations/{id}", s.handleConversation)
	s.mux.HandleFunc("GET /ephemeral/{id}", s.handleConversation)
	s.mux.HandleFunc("POST /conversations/new", s.handleNewConversation)
	s.mux.HandleFunc("POST /conversations/{id}/reply", s.handleReply)
	s.mux.HandleFunc("POST /conversations/{id}/draft", s.handleSaveDraft)
	s.mux.HandleFunc("POST /conversations/{id}/ingest", s.handleIngest)
	s.mux.HandleFunc("POST /conversations/{id}/promote", s.handlePromote)

	// Queue triggers.
	s.mux.HandleFunc("POST /queue/lint", s.handleLint)
	s.mux.HandleFunc("POST /queue/commit", s.handleCommit)

	// Attachments.
	s.mux.HandleFunc("POST /attachments/upload", s.handleUpload)
	s.mux.HandleFunc("GET /attachments/{name}", s.handleAttachment)

	// Search and log.
	s.mux.HandleFunc("GET /search", s.handleSearch)
	s.mux.HandleFunc("GET /log", s.handleLog)

	// SSE.
	s.mux.HandleFunc("GET /events", s.handleSSE)
}

// Hub returns the SSE hub so the watcher can publish to it.
func (s *Server) Hub() *Hub { return s.hub }

// Start runs the HTTP server until ctx is cancelled.
func (s *Server) Start(ctx context.Context) error {
	srv := &http.Server{
		Addr:    s.addr,
		Handler: s.mux,
	}

	go func() {
		<-ctx.Done()
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(shutCtx)
	}()

	log.Printf("pkb listening on http://%s", s.addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// baseData holds fields common to every page template.
type baseData struct {
	Title         string
	PagePath      string
	RaymondActive bool
}

func (s *Server) newBaseData(title, pagePath string) baseData {
	return baseData{
		Title:         title,
		PagePath:      pagePath,
		RaymondActive: kb.RaymondActive(s.kb, staleThreshold),
	}
}

func (s *Server) renderError(w http.ResponseWriter, status int, msg string) {
	type errData struct {
		baseData
		StatusCode int
		StatusText string
		Message    string
	}
	d := errData{
		baseData:   s.newBaseData("Error", ""),
		StatusCode: status,
		StatusText: http.StatusText(status),
		Message:    msg,
	}
	w.WriteHeader(status)
	if err := s.tmpl.errPage.ExecuteTemplate(w, "base", d); err != nil {
		// Headers already sent; can't change the response. Just log.
		log.Printf("renderError: template execution failed: %v", err)
	}
}
