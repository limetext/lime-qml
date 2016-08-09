package main

import (
	"fmt"

	"github.com/limetext/text"
)

type settingsWatcher struct {
	ptr             interface{} // pointer to root struct
	settings        *text.Settings
	watchedSettings map[string]watchedSetting
}

func newSettingsWatcher(ptr interface{}, settings *text.Settings) *settingsWatcher {
	w := &settingsWatcher{
		ptr:             ptr,
		settings:        settings,
		watchedSettings: make(map[string]watchedSetting),
	}
	settings.AddOnChange("settingsWatcher", w.onSettingChange)
	return w
}

func (w *settingsWatcher) onSettingChange(name string) {
	fmt.Printf("SettingChanged: %s %v\n", name, w.settings.Get(name))

	if s, ok := w.watchedSettings[name]; ok {
		s.setFromSettings(w.ptr, w.settings)
	}
}

type watchedSetting interface {
	setFromSettings(ptr interface{}, settings *text.Settings)
}

func (w *settingsWatcher) watchSettingInt(name string, ptr *int, defaultValue int, mapFn ...func(int) int) {
	s := &watchedSettingInt{
		name:         name,
		ptr:          ptr,
		defaultValue: defaultValue,
	}
	if len(mapFn) == 1 {
		s.mapFn = mapFn[0]
	}
	w.watchedSettings[name] = s
	s.setFromSettings(w.ptr, w.settings)
}

type watchedSettingInt struct {
	name         string
	ptr          *int
	defaultValue int
	mapFn        func(int) int
}

func (s *watchedSettingInt) setFromSettings(ptr interface{}, settings *text.Settings) {
	current := settings.Int(s.name, s.defaultValue)
	if s.mapFn != nil {
		current = s.mapFn(current)
	}
	*s.ptr = current
	fe.qmlChanged(ptr, s.ptr)
}

func (w *settingsWatcher) watchSettingString(name string, ptr *string, defaultValue string, mapFn ...func(string) string) {
	s := &watchedSettingString{
		name:         name,
		ptr:          ptr,
		defaultValue: defaultValue,
	}
	if len(mapFn) == 1 {
		s.mapFn = mapFn[0]
	}
	w.watchedSettings[name] = s
	s.setFromSettings(w.ptr, w.settings)
}

type watchedSettingString struct {
	name         string
	ptr          *string
	defaultValue string
	mapFn        func(string) string
}

func (s *watchedSettingString) setFromSettings(ptr interface{}, settings *text.Settings) {
	current := settings.String(s.name, s.defaultValue)
	if s.mapFn != nil {
		current = s.mapFn(current)
	}
	*s.ptr = current
	fe.qmlChanged(ptr, s.ptr)
}
