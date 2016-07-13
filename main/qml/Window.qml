import QtQuick 2.0
import QtQuick.Controls 1.0
import QtQuick.Controls.Styles 1.0
import QtQuick.Dialogs 1.2
import QtQuick.Layouts 1.0
import QtGraphicalEffects 1.0

ApplicationWindow {
    id: window
    width: 800
    height: 600
    title: "Lime"

    property var myWindow
    property string themeFolder: "../../packages/Soda/Soda Dark"

    function view() {
      return mainView.view();
    }

    function addTab(title, view) {
      return mainView.addTab(view);
    }

    function activateTab(tabIndex) {
      return mainView.activateTab(tabIndex);
    }

    function removeTab(tabIndex) {
      return mainView.removeTab(tabIndex);
    }

    function setTabTitle(tabIndex, title) {
      return mainView.setTabTitle(tabIndex, title);
    }

    menuBar: MenuBar {
        id: menu
        Menu {
            title: qsTr("File")
            MenuItem {
                text: qsTr("New File")
                onTriggered: frontend.runCommand("new_file");
            }
            MenuItem {
                text: qsTr("Open File...")
                onTriggered: frontend.runCommand("prompt_open_file");
            }
            MenuItem {
                text: qsTr("Save")
                onTriggered: frontend.runCommand("save");
            }
            MenuItem {
                text: qsTr("Save As...")
                // TODO(.) : qml doesn't have a ready dialog like FileDialog
                // onTriggered: saveAsDialog.open()
            }
            MenuItem {
                text: qsTr("Save All")
                onTriggered: frontend.runCommand("save_all")
            }
            MenuSeparator{}
            MenuItem {
                text: qsTr("New Window")
                onTriggered: frontend.runCommand("new_window");
            }
            MenuItem {
                text: qsTr("Close Window")
                onTriggered: frontend.runCommand("close_window");
            }
            MenuSeparator{}
            MenuItem {
                text: qsTr("Close File")
                onTriggered: frontend.runCommand("close");
            }
            MenuItem {
                text: qsTr("Close All Files")
                onTriggered: frontend.runCommand("close_all");
            }
            MenuSeparator{}
            MenuItem {
                text: qsTr("Quit")
            		// TODO: frontend.runCommand("quit");
            		onTriggered: Qt.quit();
            }
        }
        Menu {
            title: qsTr("Find")
            MenuItem {
                text: qsTr("Find Next")
                onTriggered: frontend.runCommand("find_next");
            }
        }
        Menu {
            title: qsTr("Edit")
            MenuItem {
                text: qsTr("Undo")
                onTriggered: frontend.runCommand("undo");
            }
            MenuItem {
                text: qsTr("Redo")
                onTriggered: frontend.runCommand("redo");
            }
            Menu {
                title: qsTr("Undo Selection")
                MenuItem {
                    text: qsTr("Soft Undo")
                    onTriggered: frontend.runCommand("soft_undo");
                }
                MenuItem {
                    text: qsTr("Soft Redo")
                    onTriggered: frontend.runCommand("soft_redo");
                }
            }
            MenuSeparator{}
            MenuItem {
                text: qsTr("Copy")
                onTriggered: frontend.runCommand("copy");
            }
            MenuItem {
                text: qsTr("Cut")
                onTriggered: frontend.runCommand("cut");
            }
            MenuItem {
                text: qsTr("Paste")
                onTriggered: frontend.runCommand("paste");
            }
        }
        Menu {
            title: qsTr("View")
            MenuItem {
                text: qsTr("Show/Hide Console")
                onTriggered: { consoleView.visible = !consoleView.visible }
            }
            MenuItem {
                text: qsTr("Show/Hide Minimap")
                onTriggered: {
                  var tab = tabs.getTab(tabs.currentIndex);

                  if (tab.item)
                    tab.item.minimapVisible = !tab.item.minimapVisible;
                }
            }
            MenuItem {
                text: qsTr("Show/Hide Statusbar")
                onTriggered: { statusBar.visible = !statusBar.visible }
            }
        }
    }

    property Tab currentTab: mainView.currentTab()
    property var statusBarMap: (currentTab == null || currentTab.item == null) ? null : currentTab.item.statusBar
    property var statusBarSorted: []
    onStatusBarMapChanged: {
      if (statusBarMap == null) {
        statusBarSorted = [["a", "git branch: master"], ["b", "INSERT MODE"], ["c", "Line xx, Column yy"]];
        return;
      }

      console.log("status bar map:", statusBarMap);
      var keys = Object.keys(statusBarMap);
      keys.sort();
      console.log("status bar keys:", keys);
      var sorted = [];
      for (var i = 0; i < keys.length; i++)
        sorted.push([keys[i], statusBarMap[keys[i]]]);

      statusBarSorted = sorted;
    }


    statusBar: StatusBar {
        id: statusBar
        style: StatusBarStyle {
            background: Image {
              source: themeFolder + "/status-bar-background.png"
            }
        }

        property color textColor: "#969696"

        RowLayout {
            anchors.fill: parent
            id: statusBarRowLayout
            spacing: 15

            RowLayout {
                anchors.fill: parent
                spacing: 3
                Repeater {
                  model: statusBarSorted
                  delegate:
                    Label {
                        text: modelData[1]
                        color: statusBar.textColor
                    }
                }
            }

            Label {
                id: statusBarIndent
                text: "Tab Size/Spaces: 4"
                color: statusBar.textColor
                Layout.alignment: Qt.AlignRight
            }

            Label {
                id: statusBarLanguage
                text: "Go"
                color: statusBar.textColor
                Layout.alignment: Qt.AlignRight
            }
        }
    }

    Component {
      id: tabTemplate

      View {}
    }

    Item {
        anchors.fill: parent
        Keys.onPressed: {
          var v = view(); if (v === undefined) return;
          v.ctrl = (event.key == Qt.Key_Control) ? true : false;
          event.accepted = frontend.handleInput(event.text, event.key, event.modifiers)
          // if (event.key == Qt.Key_Alt)
          event.accepted = true;
        }
        Keys.onReleased: {
          var v = view(); if (v === undefined) return;
          v.ctrl = (event.key == Qt.Key_Control) ? false : view().ctrl;
        }
        focus: true // Focus required for Keys.onPressed
        SplitView {
            anchors.fill: parent
            orientation: Qt.Vertical
              MainView {
                id: mainView
              }
              View {
                id: consoleView
                myView: frontend.console
                visible: false
                minimapVisible: false
                height: 100
              }
        }
    }
    MessageDialog {
        objectName: "messageDialog"
    }
    FileDialog {
        objectName: "fileDialog"
    }
}
