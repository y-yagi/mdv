package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/fsnotify/fsnotify"
	"github.com/y-yagi/dlogger"
	"github.com/yuin/goldmark"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type TemplateArgument struct {
	Body string
	Addr string
}

const (
	app = "mdv"
)

var (
	flags    *flag.FlagSet
	filename string
	addr     string
	watcher  *fsnotify.Watcher
	logger   = dlogger.New(os.Stdout)
)

func setFlags() {
	flags = flag.NewFlagSet(app, flag.ExitOnError)
	flags.StringVar(&addr, "addr", ":8888", "http service address")
}

func main() {
	setFlags()
	flags.Parse(os.Args[1:])

	if flags.NArg() != 1 {
		fmt.Printf("please specify filename\n")
		return
	}
	filename = flags.Args()[0]

	if err := startWatch(); err != nil {
		log.Println(err)
		return
	}
	defer watcher.Close()

	http.HandleFunc("/", handler)
	http.HandleFunc("/ws", wsHandler)
	log.Print("Listening on http://localhost:8888/")
	http.ListenAndServe(":8888", nil)
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

	for {
		select {
		case event, ok := <-watcher.Events:
			logger.Printf("watch %v\n", event)
			if !ok {
				return
			}

			err = wsjson.Write(ctx, c, "")
			if err != nil {
				logger.Printf("Write error %v\n", err)
				return
			}
		}
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	buf := new(bytes.Buffer)
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		errorResponse(err, w)
		return
	}

	if err = goldmark.Convert(body, buf); err != nil {
		errorResponse(err, w)
		return
	}

	t := TemplateArgument{Body: string(buf.Bytes()), Addr: r.Host}

	buf.Reset()
	tpl, err := template.New("html").Parse(html)
	if err != nil {
		errorResponse(err, w)
		return
	}

	tpl.Execute(buf, t)
	fmt.Fprintf(w, string(buf.Bytes()))
}

func errorResponse(err error, w http.ResponseWriter) {
	fmt.Fprintf(w, fmt.Sprintf("Error occurred: %v", err))
}

const html = `
<html>
  <head>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/github-markdown-css/4.0.0/github-markdown.min.css">
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
    <script type="text/javascript">
      (function() {
        var conn = new WebSocket("ws://{{.Addr}}/ws");
        conn.onmessage = function(evt) {
					location.reload();
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
