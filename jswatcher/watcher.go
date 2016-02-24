package jswatcher

import (
	"fmt"
	"github.com/Jumpscale/agentcontroller8/configs"
	"github.com/Jumpscale/agentcontroller8/core"
	"github.com/Jumpscale/agentcontroller8/scheduling"
	"github.com/Jumpscale/pygo"
	"github.com/rjeczalik/notify"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	jumpScriptExtension       = ".py"
	jumpScriptAutoScheduleKey = "auto.scheduler.jumpscript.%s.%s"
	jumpScriptCronSpec        = "@every %ds"
)

type JSWatcher interface {
	Start()
}

type jsWatcher struct {
	enabled   bool
	folder    string
	scheduler *scheduling.Scheduler
	module    pygo.Pygo
}

func NewJSWatcher(config *configs.Extension, scheduler *scheduling.Scheduler) (JSWatcher, error) {

	//load module according to extension.
	var folder string
	var module pygo.Pygo
	var err error
	var ok bool

	if config.Enabled {
		folder, ok = config.Settings["jumpscripts_path"]
		if !ok {
			return nil, fmt.Errorf("Jumpscript ext: jumpscripts_path is not set")
		}

		opts := &pygo.PyOpts{
			PythonBinary: config.GetPythonBinary(),
			PythonPath:   config.PythonPath,
			Env: []string{
				fmt.Sprintf("HOME=%s", os.Getenv("HOME")),
			},
		}

		module, err = pygo.NewPy(config.Module, opts)
		if err != nil {
			return nil, err
		}
	}

	watcher := &jsWatcher{
		enabled:   config.Enabled,
		folder:    folder,
		scheduler: scheduler,
		module:    module,
	}

	return watcher, nil
}

func (watcher *jsWatcher) getJumpscriptID(file string) (string, string) {
	filename := path.Base(file)
	domain := path.Base(path.Dir(file))

	ext := path.Ext(filename)
	name := filename[0 : len(filename)-len(ext)]

	return domain, name
}

func (watcher *jsWatcher) unschedule(domain string, name string) error {
	jobid := fmt.Sprintf(jumpScriptAutoScheduleKey, domain, name)
	_, err := watcher.scheduler.RemoveByID(jobid)
	return err
}

func (watcher *jsWatcher) schedule(attributes jsAttributes, domain string, name string) error {
	if !attributes.Enable() {
		//script is not enabled
		return nil
	}
	period := attributes.Period()

	if period <= 0 {
		return nil
	}

	roles := attributes.Roles()
	if len(roles) == 0 {
		roles = []string{"*"}
	}

	//build command.
	content := core.CommandContent{
		Cmd:    "jumpscript",
		Roles:  roles,
		Fanout: true,
		Data:   "{}", //this must be a serialized json dict for the jumpscript to work
		Tags:   attributes.Category(),
	}

	content.Args.Domain = domain
	content.Args.Name = name
	content.Args.Queue = attributes.Queue()
	content.Args.MaxTime = attributes.Timeout()

	command := core.CommandFromContent(&content)
	job := scheduling.Job{
		ID:   fmt.Sprintf(jumpScriptAutoScheduleKey, domain, name),
		Cron: fmt.Sprintf(jumpScriptCronSpec, period),
		Cmd:  command.Raw,
	}

	return watcher.scheduler.AddJob(&job)
}

func (watcher *jsWatcher) process(path string, domain string, name string) error {
	result, err := watcher.module.Call("get_info", path)

	if err != nil {
		return err
	}

	info, ok := result.(map[string]interface{})
	if !ok {
		return fmt.Errorf("Expecing jumpscript attributes to be a dict")
	}

	attributes := jsAttributes(info)

	err = watcher.schedule(attributes, domain, name)
	if err != nil {
		return fmt.Errorf("Failed to scheduled jumpscript %s/%s: %s", domain, name, err)
	}

	return nil
}

func (watcher *jsWatcher) watch() {
	// Make the channel buffered to ensure no event is dropped. Notify will drop
	// an event if the receiver is not able to keep up the sending pace.
	events := make(chan notify.EventInfo, 4)

	// Set up a watchpoint listening on events within current working directory.
	// Dispatch each create and remove events separately to c.
	folder := path.Join(watcher.folder, "...")
	if err := notify.Watch(folder, events, notify.Write, notify.Create, notify.Remove); err != nil {
		log.Fatal(err)
	}

	defer notify.Stop(events)

	//read fs events
	for event := range events {
		path := event.Path()

		if !strings.HasSuffix(path, jumpScriptExtension) {
			//file name too short to be a config file (shorter than the extension)
			continue
		}

		event := event.Event()
		domain, name := watcher.getJumpscriptID(path)

		if err := watcher.unschedule(domain, name); err != nil {
			log.Println("Failed to unschedule jumpscript", domain, name)
			continue
		}

		if event == notify.Create || event == notify.Write {
			err := watcher.process(path, domain, name)
			if err != nil {
				log.Println(err)
			}
		}
	}
}

func (watcher *jsWatcher) apply() {
	//process all jumpscripts in the given location.
	filepath.Walk(watcher.folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Println(err)
			return nil
		}

		if info.Mode().IsRegular() {
			if strings.HasSuffix(path, jumpScriptExtension) {
				domain, name := watcher.getJumpscriptID(path)

				err := watcher.process(path, domain, name)
				if err != nil {
					log.Printf("Processing script '%s' error: %s\n", path, err)
				}
			}
		}

		return nil
	})
}

func (watcher *jsWatcher) start() {
	if !watcher.enabled {
		return
	}

	if _, err := os.Stat(watcher.folder); os.IsNotExist(err) {
		os.MkdirAll(watcher.folder, 755)
	}

	watcher.apply()
	watcher.watch()

}

func (watcher *jsWatcher) Start() {
	go watcher.start()
}
