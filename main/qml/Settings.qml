import QtQuick 2.0

QtObject {
  id: settingsRoot

  property var settingsObject

  // property var _settingsHolder: null

  onSettingsObjectChanged: {
    if (settingsObject)
      loadSettings();
  }

  property var _settingsMap: ({})
  property var _settingsValueMap: ({})
  property int _intSettingCounter: 19542

  function settingInt(name, defaultVal) {
    if (_settingsMap[name]) {
      return _settingsMap[name].id;
    }

    var counter = _intSettingCounter;
    _intSettingCounter += 1;

    _settingsValueMap["int" + counter] = name;

    _settingsMap[name] = {id: counter, type: "int", defaultVal: defaultVal};

    // without this, the _settingsMap change doesn't happen soon enough for the next call to settingInt
    _settingsMap = _settingsMap;

    // console.log("setting!", name, counter, Object.keys(_settingsMap));
    return counter;
  }

  property bool _firstLoadRan: false

  // Component.onCompleted: loadSettings()

  function firstLoad() {
    if (_firstLoadRan) return;
    _firstLoadRan = true;

    var keys = Object.keys(settingsRoot);

    for (var i = 0; i < keys.length; i++) {
      var property = keys[i];

      var val = settingsRoot[property];
      if (typeof val === "number") {
        var settingVal = _settingsValueMap["int" + val];
        if (settingVal) {
          var setting = _settingsMap[settingVal];

          setting.prop = property;
          // console.log("Found setting: ", property, typeof val);
        }
      }
    }
  }

  function loadSettings() {
    firstLoad();

    for (var settingKey in _settingsMap) {
      var setting = _settingsMap[settingKey];
      var property = setting.prop;
      if (!property) continue;

      var hasSetting = settingsObject.hasSetting(property);
      var value = hasSetting? settingsObject.getSetting(property) : setting.defaultVal;
      console.log("Load ", settingKey, property, value);
      settingsRoot[property] = value;
    }
  }

  function settingChanged(name, value) {
    if (settingsRoot.hasOwnProperty(name)) {
      console.log("Setting", name, "changed to:", value);
      settingsRoot[name] = value;
    } else {
      console.log("Unhandled setting: ", name, "changed to:", value);
    }
  }
}
