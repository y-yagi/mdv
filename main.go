package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/yuin/goldmark"
)

type TemplateArgument struct {
	Body string
}

var filename string

func main() {
	if len(os.Args) == 1 {
		fmt.Printf("please specify filename\n")
		return
	}
	filename = os.Args[1]

	http.HandleFunc("/", handler)
	log.Print("Listening on http://localhost:8888/")
	http.ListenAndServe(":8888", nil)
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

	t := TemplateArgument{Body: string(buf.Bytes())}

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
  </head>
  <body>
    <article class="markdown-body">
		{{.Body}}
    </article>
  </body>
</html>
`
