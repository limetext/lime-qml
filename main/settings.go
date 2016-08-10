// Copyright 2013 The lime Authors.
// Use of this source code is governed by a 2-clause
// BSD-style license that can be found in the LICENSE file.

package main

import (
	"fmt"

	"github.com/limetext/text"
)

type settingsWatcher struct {
	structPtr       interface{} // pointer to root struct
	settings        *text.Settings
	watchedSettings map[string]watchedSetting
}

func newSettingsWatcher(structPtr interface{}, settings *text.Settings) *settingsWatcher {
	w := &settingsWatcher{
		structPtr:       structPtr,
		settings:        settings,
		watchedSettings: make(map[string]watchedSetting),
	}
	settings.AddOnChange("settingsWatcher", w.onSettingChange)
	return w
}

func (w *settingsWatcher) onSettingChange(name string) {
	fmt.Printf("SettingChanged: %s %v\n", name, w.settings.Get(name))

	if s, ok := w.watchedSettings[name]; ok {
		s.setFromSettings(w.structPtr, w.settings)
	}
}

type watchedSetting interface {
	setFromSettings(structPtr interface{}, settings *text.Settings)
}

func (w *settingsWatcher) watchInt(name string, fieldPtr *int, defaultValue int, mapFn ...func(int) int) {
	s := &watchedSettingInt{
		name:         name,
		fieldPtr:     fieldPtr,
		defaultValue: defaultValue,
	}
	if len(mapFn) == 1 {
		s.mapFn = mapFn[0]
	}
	w.watchedSettings[name] = s
	s.setFromSettings(w.structPtr, w.settings)
}

type watchedSettingInt struct {
	name         string
	fieldPtr     *int
	defaultValue int
	mapFn        func(int) int
}

func (s *watchedSettingInt) setFromSettings(structPtr interface{}, settings *text.Settings) {
	current := settings.Int(s.name, s.defaultValue)
	if s.mapFn != nil {
		current = s.mapFn(current)
	}
	*s.fieldPtr = current
	fe.qmlChanged(structPtr, s.fieldPtr)
}

func (w *settingsWatcher) watchString(name string, fieldPtr *string, defaultValue string, mapFn ...func(string) string) {
	s := &watchedSettingString{
		name:         name,
		fieldPtr:     fieldPtr,
		defaultValue: defaultValue,
	}
	if len(mapFn) == 1 {
		s.mapFn = mapFn[0]
	}
	w.watchedSettings[name] = s
	s.setFromSettings(w.structPtr, w.settings)
}

type watchedSettingString struct {
	name         string
	fieldPtr     *string
	defaultValue string
	mapFn        func(string) string
}

func (s *watchedSettingString) setFromSettings(structPtr interface{}, settings *text.Settings) {
	current := settings.String(s.name, s.defaultValue)
	if s.mapFn != nil {
		current = s.mapFn(current)
	}
	*s.fieldPtr = current
	fe.qmlChanged(structPtr, s.fieldPtr)
}
