package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/cli/browser"
	"github.com/fsnotify/fsnotify"
	"github.com/y-yagi/dlogger"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type TemplateArgument struct {
	Body  string
	Addr  string
	Style string
}

const (
	app = "mdv"
)

var (
	flags    *flag.FlagSet
	filename string
	addr     string
	dir      string
	css      string
	style    string
	watcher  *fsnotify.Watcher
	logger   = dlogger.New(os.Stdout)
)

func setFlags() {
	flags = flag.NewFlagSet(app, flag.ExitOnError)
	flags.StringVar(&addr, "addr", ":8888", "http service address")
	flags.StringVar(&dir, "dir", "", "directory that uses in the file server")
	flags.StringVar(&css, "css", "", "CSS file that uses in rendering")
}

func main() {
	setFlags()
	flags.Parse(os.Args[1:])

	if flags.NArg() != 1 {
		fmt.Println("please specify filename")
		return
	}
	filename = flags.Args()[0]

	if err := startWatch(); err != nil {
		log.Println(err)
		return
	}
	defer watcher.Close()

	var err error
	if style, err = buildStyle(); err != nil {
		log.Println(err)
		return

	}

	http.HandleFunc("/", handler)
	http.HandleFunc("/ws", wsHandler)
	url := "http://localhost" + addr
	log.Print("Listening on " + url)
	go browser.OpenURL(url)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Println(err)
		return
	}
}

func startWatch() error {
	var err error
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	return nil
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	defer c.Close(websocket.StatusInternalError, "the sky is falling")

	ctx := r.Context()
	ctx = c.CloseRead(ctx)

	if err := watcher.Add(filename); err != nil {
		log.Fatal(err)
	}

	var timer *time.Timer
	for {
		for event := range watcher.Events {
			logger.Printf("watch %v\n", event)
			// Vim will remove a file when a file was changed.  So we need to add a new file to the watcher.
			if event.Op&fsnotify.Remove == fsnotify.Remove {
				watcher.Add(filename)
			}

			if timer == nil {
				timer = time.NewTimer(100 * time.Millisecond)
				go func() {
					<-timer.C
					timer = nil

					buf := new(bytes.Buffer)
					body, err := os.ReadFile(filename)
					if err != nil {
						logger.Printf("Read error %v\n", err)
						return
					}

					if err = buildParser().Convert(body, buf); err != nil {
						logger.Printf("Convert error %v\n", err)
						return
					}

					err = wsjson.Write(ctx, c, buf.String())
					if err != nil {
						logger.Printf("Write error %v\n", err)
						return
					}
				}()
			}
		}
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	if len(dir) != 0 && r.URL.Path != "/" {
		fh := http.FileServer(http.Dir(dir))
		fh.ServeHTTP(w, r)
		return
	}

	buf := new(bytes.Buffer)
	body, err := os.ReadFile(filename)
	if err != nil {
		errorResponse(err, w)
		return
	}

	if err = buildParser().Convert(body, buf); err != nil {
		errorResponse(err, w)
		return
	}

	t := TemplateArgument{Body: buf.String(), Addr: r.Host, Style: style}

	buf.Reset()
	tpl, err := template.New("html").Parse(layout)
	if err != nil {
		errorResponse(err, w)
		return
	}

	tpl.Execute(buf, t)
	fmt.Fprint(w, buf.String())
}

func errorResponse(err error, w http.ResponseWriter) {
	fmt.Fprintf(w, "Error occurred: %v", err)
}

func buildParser() goldmark.Markdown {
	return goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)
}

func buildStyle() (string, error) {
	if len(css) == 0 {
		return defaultStyle, nil
	}

	if strings.HasPrefix(css, "http://") || strings.HasPrefix(css, "https://") {
		return `<link rel="stylesheet" href="` + css + `">`, nil
	}

	style, err := os.ReadFile(css)
	if err != nil {
		return "", err
	}

	return "<style>" + string(style) + "</style>", nil
}

const defaultStyle = `
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/github-markdown-css/5.5.1/github-markdown.min.css">
    <style>
      .markdown-body {
        box-sizing: border-box;
        min-width: 200px;
        max-width: 980px;
        margin: 0 auto;
        padding: 45px;
      }

      @media (max-width: 767px) {
        .markdown-body {
          padding: 15px;
        }
      }
    </style>
`

const layout = `
<html>
  <head>
    <meta name="viewport" content="width=device-width, initial-scale=1">
	{{.Style}}
    <script type="text/javascript">
      (function() {
        var conn = new WebSocket("ws://{{.Addr}}/ws");
        conn.onmessage = function(evt) {
					let element = document.getElementsByClassName('markdown-body')[0]
					element.innerHTML = JSON.parse(evt.data);
        }
      })();
    </script>
  </head>
  <body>
    <article class="markdown-body">
    {{.Body}}
    </article>
  </body>
</html>
`
