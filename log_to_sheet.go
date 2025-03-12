package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/option"
)

const (
	documentID = "1V4qshUJ24_J_VIFJpYKZ9ZT5bv8aDN2Q5WWUhLiJTns"
	logFile    = "app.txt"
)

type LogEntry struct {
	Timestamp string
	Level     string
	Message   string
}

func parseLogEntry(line string) LogEntry {
	parts := strings.SplitN(strings.Trim(line, "[]"), "]", 2)
	if len(parts) != 2 {
		return LogEntry{}
	}

	levelMsg := strings.SplitN(strings.TrimSpace(parts[1]), ":", 2)
	if len(levelMsg) != 2 {
		return LogEntry{}
	}

	return LogEntry{
		Timestamp: strings.TrimSpace(parts[0]),
		Level:     strings.TrimSpace(levelMsg[0]),
		Message:   strings.TrimSpace(levelMsg[1]),
	}
}

func initDocsService() (*docs.Service, error) {
	ctx := context.Background()

	credBytes, err := os.ReadFile("creds.json")
	if err != nil {
		return nil, fmt.Errorf("unable to read credentials file: %v", err)
	}

	config, err := google.JWTConfigFromJSON(credBytes, docs.DocumentsScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse credentials: %v", err)
	}

	client := config.Client(ctx)
	srv, err := docs.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to create docs service: %v", err)
	}

	return srv, nil
}

func appendToDoc(srv *docs.Service, entries []LogEntry) error {
	var text strings.Builder
	for _, entry := range entries {
		text.WriteString(fmt.Sprintf("[%s] %s: %s\n",
			entry.Timestamp,
			entry.Level,
			entry.Message))
	}

	req := &docs.BatchUpdateDocumentRequest{
		Requests: []*docs.Request{
			{
				InsertText: &docs.InsertTextRequest{
					Location: &docs.Location{
						Index: 1,
					},
					Text: text.String(),
				},
			},
		},
	}

	_, err := srv.Documents.BatchUpdate(documentID, req).Do()
	return err
}

func watchLogFile(srv *docs.Service) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("error creating watcher: %v", err)
	}
	defer watcher.Close()

	absPath, err := filepath.Abs(logFile)
	if err != nil {
		return fmt.Errorf("error getting absolute path: %v", err)
	}

	dir := filepath.Dir(absPath)
	err = watcher.Add(dir)
	if err != nil {
		return fmt.Errorf("error adding watch: %v", err)
	}

	var lastPos int64 = 0

	if entries, newPos, err := readNewEntries(absPath, lastPos); err == nil {
		if len(entries) > 0 {
			if err := appendToDoc(srv, entries); err != nil {
				log.Printf("Error appending initial entries: %v", err)
			}
		}
		lastPos = newPos
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return fmt.Errorf("watcher channel closed")
			}

			if event.Name == absPath && (event.Op&fsnotify.Write == fsnotify.Write) {
				time.Sleep(100 * time.Millisecond)

				entries, newPos, err := readNewEntries(absPath, lastPos)
				if err != nil {
					log.Printf("Error reading new entries: %v", err)
					continue
				}

				if len(entries) > 0 {
					if err := appendToDoc(srv, entries); err != nil {
						log.Printf("Error appending to doc: %v", err)
						continue
					}
					lastPos = newPos
				}
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return fmt.Errorf("watcher error channel closed")
			}
			log.Printf("Watcher error: %v", err)
		}
	}
}

func readNewEntries(filename string, lastPos int64) ([]LogEntry, int64, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, lastPos, err
	}
	defer file.Close()

	_, err = file.Seek(lastPos, 0)
	if err != nil {
		return nil, lastPos, err
	}

	var entries []LogEntry
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if entry := parseLogEntry(line); entry.Timestamp != "" {
			entries = append(entries, entry)
		}
	}

	newPos, err := file.Seek(0, 1)
	if err != nil {
		return entries, lastPos, err
	}

	return entries, newPos, scanner.Err()
}

func main() {
	srv, err := initDocsService()
	if err != nil {
		log.Fatalf("Unable to initialize docs service: %v", err)
	}

	log.Printf("Starting to watch log file: %s", logFile)
	if err := watchLogFile(srv); err != nil {
		log.Fatalf("Error watching log file: %v", err)
	}
}
