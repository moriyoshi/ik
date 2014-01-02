package main

import (
	"log"
	"time"
	"strconv"
	"net"
	"net/http"
	"html/template"
	"unsafe"
	"bytes"
	"reflect"
	"sync/atomic"
	"github.com/moriyoshi/ik"
	"github.com/moriyoshi/ik/markup"
)

var mainTemplate = `<!DOCTYPE html>
<html>
<head>
<title>Ik Scoreboard</title>
<style type="text/css">
body {
  margin: 0 0;
  background-color: #eee;
  color: #222;
}

header {
  background-color: #888;
  padding: 4px 8px;
}

main {
  padding: 8px 8px;
}

h1 {
  margin: 0 0;
}

.table {
  margin: 0 0;
  padding: 0 0;
  border: 1px solid #ccc;
  border-collapse: collapse;
}

.table.fullwidth {
  width: 100%;
}

.table > thead > tr > td,
.table > thead > tr > th {
  background-color: #888;
  color: #eee;
}

.table > thead > tr > th,
.table > tbody > tr > td,
.table > tbody > tr > th {
  border: 1px solid #ccc;
  padding: 4px 4px;
}

.exitStatus.running {
  background-color: #eec;
}

.exitStatus.error {
  background-color: #ecc;
  color: #f00;
}
</style>
</head>
<body>
<header>
<h1>Ik Scoreboard</h1>
</header>
<main>
<h2>Plugins</h2>
<h3>Input Plugins</h3>
<ul>
{{range .InputPlugins}}
<li>{{.Name}}</li>
{{end}}
</ul>
<h3>Output Plugins</h3>
<ul>
{{range .OutputPlugins}}
<li>{{.Name}}</li>
{{end}}
</ul>
<h3>Scoreboard Plugins</h3>
<ul>
{{range .ScoreboardPlugins}}
<li>{{.Name}}</li>
{{end}}
</ul>
<h2>Spawnees</h2>
<table class="table">
  <thead>
    <tr>
      <th>#</th>
      <th>Name</th>
      <th>Status</th>
    </tr>
  </thead>
  <tbody>
    {{range .SpawneeStatuses}}
    <tr>
      <th>{{.Id|printf "%d"}}</th>
      <th>{{spawneeName .Spawnee}}</th>
      <td class="exitStatus {{renderExitStatusStyle .ExitStatus}}">{{renderExitStatusLabel .ExitStatus}}</td>
    </tr>
    {{end}}
  </tbody>
</table>
<h2>Topics</h2>
{{range .Plugins}}
{{$topics := getTopics .}}
<h3>{{.Name}}</h3>
{{if len $topics}}
<table class="table">
  <thead>
    <tr>
      <th>Name</th>
      <th>Value</th>
      <th>Description</th>
    </tr>
  </thead>
  <tbody>
    {{range $topics}}
    <tr>
      <th>{{.DisplayName}}</th>
      <td>{{.Fetcher.Markup|renderMarkup}}</td>
      <td>{{.Description}}</td>
    </tr>
    {{end}}
  </tbody>
</table>
{{else}}
No topics available
{{end}}
{{end}}
</main>
</body>
</html>`

type HTMLHTTPScoreboard struct {
	template *template.Template
	factory  *HTMLHTTPScoreboardFactory
	logger   *log.Logger
	engine   ik.Engine
	registry ik.PluginRegistry
	listener net.Listener
	server   http.Server
	requests int64
}

type HTMLHTTPScoreboardFactory struct {
}

type viewModel struct {
	InputPlugins []ik.InputFactory
	OutputPlugins []ik.OutputFactory
	ScoreboardPlugins []ik.ScoreboardFactory
	Plugins []ik.Plugin
	SpawneeStatuses []ik.SpawneeStatus
}

type requestCountFetcher struct {
	scoreboard *HTMLHTTPScoreboard
}

func (fetcher *requestCountFetcher) Markup() (ik.Markup, error) {
	text, err := fetcher.PlainText()
	if err != nil {
		return ik.Markup {}, err
	}
	return ik.Markup { []ik.MarkupChunk { { Attrs: 0, Text: text } } }, nil
}

func (fetcher *requestCountFetcher) PlainText() (string, error) {
	return strconv.FormatInt(fetcher.scoreboard.requests, 10), nil
}

func spawneeName(spawnee ik.Spawnee) string {
	switch spawnee_ := spawnee.(type) {
	case ik.Input:
		return spawnee_.Factory().Name()
	case ik.Output:
		return spawnee_.Factory().Name()
	case ik.Scoreboard:
		return spawnee_.Factory().Name()
	default:
		return reflect.TypeOf(spawnee_).Name()
	}
}

func renderExitStatusStyle(err error) string {
	switch err_ := err.(type) {
	case *ik.ContinueType:
		_ = err_
		return "running"
	default:
		return "error"
	}
}

func renderExitStatusLabel(err error) string {
	switch err_ := err.(type) {
	case *ik.ContinueType:
		_ = err_
		return `Running`
	default:
		return err.Error()
	}
}

func renderMarkup(markup_ ik.Markup) template.HTML {
	buf := &bytes.Buffer {}
	renderer := &markup.HTMLRenderer { Out: buf }
	renderer.Render(&markup_)
	return template.HTML(buf.String())
}

func (scoreboard *HTMLHTTPScoreboard) Run() error {
	scoreboard.server.Serve(scoreboard.listener)
	return nil
}

func (scoreboard *HTMLHTTPScoreboard) Shutdown() error {
	return scoreboard.listener.Close()
}

func (scoreboard *HTMLHTTPScoreboard) Factory() ik.ScoreboardFactory {
	return scoreboard.factory
}

func (factory *HTMLHTTPScoreboardFactory) Name() string {
	return "html_http"
}

func (scoreboard *HTMLHTTPScoreboard) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	atomic.AddInt64(&scoreboard.requests, 1)
	resp.Header().Set("Content-Type", "text/html; charset=utf-8")
	resp.WriteHeader(200)
	spawneeStatuses, err := scoreboard.engine.SpawneeStatuses()
	if err != nil {
		spawneeStatuses = nil
	}
	plugins := scoreboard.registry.Plugins()
	inputPlugins := make([]ik.InputFactory, 0)
	outputPlugins := make([]ik.OutputFactory, 0)
	scoreboardPlugins := make([]ik.ScoreboardFactory, 0)
	for _, plugin := range plugins {
		switch plugin_ := plugin.(type) {
		case ik.InputFactory:
			inputPlugins = append(inputPlugins, plugin_)
		case ik.OutputFactory:
			outputPlugins = append(outputPlugins, plugin_)
		case ik.ScoreboardFactory:
			scoreboardPlugins = append(scoreboardPlugins, plugin_)
		}
	}
	scoreboard.template.Execute(resp, viewModel {
		InputPlugins: inputPlugins,
		OutputPlugins: outputPlugins,
		ScoreboardPlugins: scoreboardPlugins,
		Plugins: plugins,
		SpawneeStatuses: spawneeStatuses,
	})
}

func newHTMLHTTPScoreboard(factory *HTMLHTTPScoreboardFactory, logger *log.Logger, engine ik.Engine, registry ik.PluginRegistry, bind string, readTimeout time.Duration, writeTimeout time.Duration) (*HTMLHTTPScoreboard, error) {
	template_, err := template.New("main").Funcs(template.FuncMap {
			"spawneeName": spawneeName,
			"renderExitStatusStyle": renderExitStatusStyle,
			"renderExitStatusLabel": renderExitStatusLabel,
			"renderMarkup": renderMarkup,
			"getTopics": func(plugin ik.Plugin) []ik.ScorekeeperTopic {
				return engine.Scorekeeper().GetTopics(plugin)
			},
		}).Parse(mainTemplate)
	if err != nil {
		logger.Print(err.Error())
		return nil, err
	}
	server := http.Server {
		Addr:           bind,
		Handler:        nil,
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		MaxHeaderBytes: 0,
		TLSConfig:      nil,
		TLSNextProto:   nil,
	}
	listener, err := net.Listen("tcp", bind)
	if err != nil {
		logger.Print(err.Error())
		return nil, err
	}
	retval := &HTMLHTTPScoreboard {
		template: template_,
		factory:  factory,
		logger:   logger,
		engine:   engine,
		registry: registry,
		server:   server,
		listener: listener,
		requests: 0,
	}
	engine.Scorekeeper().AddTopic(ik.ScorekeeperTopic {
		Plugin: factory,
		Name: "requests",
		DisplayName: "Requests",
		Description: "Number of requests accepted",
		Fetcher: &requestCountFetcher {scoreboard: retval},
	})
	retval.server.Handler = retval
	return retval, nil
}

var durationSize = unsafe.Sizeof(time.Duration(0))

func (factory *HTMLHTTPScoreboardFactory) New(engine ik.Engine, registry ik.PluginRegistry, config *ik.ConfigElement) (ik.Scoreboard, error) {
	listen, ok := config.Attrs["listen"]
	if !ok {
		listen = ""
	}
	netPort, ok := config.Attrs["port"]
	if !ok {
		netPort = "24226"
	}
	bind := listen + ":" + netPort
	readTimeout := time.Duration(0)
	{
		valueStr, ok := config.Attrs["read_timeout"]
		if ok {
			value, err := strconv.ParseInt(valueStr, 10, int(durationSize * 8))
			if err != nil {
				return nil, err
			}
			readTimeout = time.Duration(value)
		}
	}
	writeTimeout := time.Duration(0)
	{
		writeTimeoutStr, ok := config.Attrs["write_timeout"]
		if ok {
			value, err := strconv.ParseInt(writeTimeoutStr, 10, int(durationSize * 8))
			if err != nil {
				return nil, err
			}
			writeTimeout = time.Duration(value)
		}
	}
	return newHTMLHTTPScoreboard(factory, engine.Logger(), engine, registry, bind, readTimeout, writeTimeout)
}
