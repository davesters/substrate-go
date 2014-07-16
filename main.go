package main

import (
	"github.com/conformal/gotk3/gtk"
	"github.com/idada/v8.go"
	"log"
	"reflect"
	"substrate/util"
)

func main() {
	app := &Substrate{}

	app.Init()
}

type Substrate struct {
	engine  *v8.Engine
	context *v8.Context
	global  *v8.ObjectTemplate
	window  *gtk.Window
	builder *gtk.Builder
}

type EventData struct {
	callback  *v8.Function
	name      string
	eventType string
}

func (s *Substrate) Init() {
	s.engine = v8.NewEngine()
	s.global = s.engine.NewObjectTemplate()

	s.global.Bind("exit", s.onAppQuit)
	s.global.Bind("alert", s.onAlert)
	s.global.Bind("bindEvent", s.bindEvent)
	s.global.Bind("loadMainWindow", s.loadMainWindow)

	source, err := util.ReadFile("main.js")
	if err != nil {
		log.Fatal("Unable to read main.js file", err)
	}

	script := s.engine.Compile(source, nil)

	gtk.Init(nil)
	s.builder, err = gtk.BuilderNew()

	s.context = s.engine.NewContext(s.global)
	s.context.Scope(func(cs v8.ContextScope) {
		cs.Run(script)

		onAppReady := cs.Global().GetProperty("on_app_ready")

		if onAppReady.IsFunction() {
			onAppReady.ToFunction().Call()
		}
	})

	gtk.Main()
}

func (s *Substrate) onAppQuit() {
	gtk.MainQuit()
}

func (s *Substrate) onAlert(msg string) {
	msgDialog := gtk.MessageDialogNew(s.window, gtk.DIALOG_MODAL, gtk.MESSAGE_INFO, gtk.BUTTONS_NONE, msg)
	msgDialog.Show()
}

func (s *Substrate) loadMainWindow(name string) {
	err := s.builder.AddFromFile(name + ".ui")
	if err != nil {
		log.Fatal("Unable to load main window ui.")
	}

	win, err := s.builder.GetObject(name)
	if err != nil {
		log.Fatal("Unable to find main window object.")
	}

	s.window = win.(*gtk.Window)
	s.window.ShowAll()

	s.window.Connect("destroy", func() {
		gtk.MainQuit()
	})
}

func (s *Substrate) bindEvent(name string, eventType string, callback *v8.Function) {
	obj, err := s.builder.GetObject(name)
	if err != nil {
		log.Fatal("Unable to find object " + name)
	}

	obj.(*gtk.Button).Connect(eventType, s.onEventCallback, EventData{callback: callback, name: name, eventType: eventType})
}

func (s *Substrate) onEventCallback(button *gtk.Button, eventData interface{}) {
	s.context.Scope(func(cs v8.ContextScope) {
		ed := eventData.(EventData)

		event := s.engine.GoValueToJsValue(reflect.ValueOf(map[string]interface{}{
			"name":          ed.name,
			"eventType":     ed.eventType,
			"currentTarget": button.GetBorderWidth(),
		}))

		if ed.callback.IsFunction() {
			ed.callback.ToFunction().Call(event)
		}
	})
}
