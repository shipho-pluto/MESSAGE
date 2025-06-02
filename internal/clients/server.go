package clients

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	client "message/internal/clients/service"
	"message/internal/config"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"
)

type ClientFabric struct {
	CRUD       *client.ClientCRUD
	HttpServer *http.Server
}

func ClientMustLoad(cnf *config.Config, logger *slog.Logger) *ClientFabric {
	crudClient, err := client.New(
		context.Background(),
		logger,
		cnf.Clients.CRUD.Addr,
		cnf.Clients.CRUD.Timeout,
		cnf.Clients.CRUD.RetriesCount,
	)
	if err != nil {
		logger.Error("failed to initialize CRUD client", err)
		os.Exit(1)
	}
	logger.Info("ClientCRUD initialized")

	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: setupRoutes(crudClient, logger),
	}

	return &ClientFabric{
		HttpServer: httpServer,
		CRUD:       crudClient,
	}
}

//go:embed all:front/*
var frontFS embed.FS

func setupRoutes(cli *client.ClientCRUD, logger *slog.Logger) *http.ServeMux {
	tokenHardCode := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhcHBJRCI6MSwiZW1haWwiOiJleGFtcGxlQGNvbSIsImV4cCI6MTc0Nzc1MjE2MywidXNlcklEIjo5fQ.Ma6gNqIsJcgT65CFt_SX1H8WO1lSn21LjyD0fkbLkdw"
	mux := http.NewServeMux()

	templates := template.Must(template.ParseFS(frontFS,
		"front/templates/index.html",
	))

	// Обработчик статических файлов (CSS, JS)
	staticFS, _ := fs.Sub(frontFS, "front/static")
	fileServer := http.FileServer(http.FS(staticFS))

	mux.Handle("/static/", http.StripPrefix("/static/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Устанавливаем правильные MIME-типы
		switch {
		case strings.HasSuffix(r.URL.Path, ".css"):
			w.Header().Set("Content-Type", "text/css")
		case strings.HasSuffix(r.URL.Path, ".js"):
			w.Header().Set("Content-Type", "application/javascript")
		case strings.HasSuffix(r.URL.Path, ".png"):
			w.Header().Set("Content-Type", "image/png")
		}
		fileServer.ServeHTTP(w, r)
	})))

	// Serve index.html for root path
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		if err := templates.ExecuteTemplate(w, "index.html", nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	// API endpoints for messages
	mux.HandleFunc("/api/messages", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			if err := r.ParseForm(); err != nil {
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}

			messageType := r.FormValue("type")
			content := r.FormValue("message-content")
			datetime := time.Now().String()[0:16]

			mid, err := cli.SentMessage(r.Context(), datetime, messageType, content, tokenHardCode)
			if err != nil {
				logger.Error("failed to send message", "error", err.Error())
				http.Error(w, "Failed to send message", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"status": "success", "message_id": %d, "datetime": %q}`, mid, datetime)

		case "GET":
			// Получаем все сообщения
			messages, err := cli.ShowAllMessages(r.Context(), tokenHardCode)
			if err != nil {
				logger.Error("failed to get messages",
					"error", err.Error())
				http.Error(w, "Failed to retrieve messages", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			response := map[string]interface{}{
				"status":   "success",
				"count":    len(messages),
				"messages": messages,
			}

			if err := json.NewEncoder(w).Encode(response); err != nil {
				logger.Error("failed to encode response",
					"error", err.Error())
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	return mux
}
