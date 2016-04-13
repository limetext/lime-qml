import QtQuick 2.0
import QtQuick.Controls 1.0
import QtQuick.Controls.Styles 1.0
import QtQuick.Dialogs 1.0
import QtQuick.Layouts 1.0
import QtGraphicalEffects 1.0

import "dialogs"

ApplicationWindow {
    id: window
    width: 800
    height: 600
    title: "Lime"

    property var myWindow
    property string themeFolder: "../../packages/Soda/Soda Dark"

    function view() {
      var tab = tabs.getTab(tabs.currentIndex);
       return tab === undefined ? undefined : tab.item;
    }

    function addTab(title, view) {
      var tab = tabs.addTab(title, tabTemplate);
      console.log("tab", tab, tab.item);

      var loadTab = function() {
        tab.item.myView = view;
      }

      if (tab.item != null) {
        loadTab();
      } else {
        tab.loaded.connect(loadTab);
      }
    }

    menuBar: MenuBar {
        id: menu
        Menu {
            title: qsTr("File")
            MenuItem {
                text: qsTr("New File")
                shortcut: "Ctrl+N"
                onTriggered: frontend.runCommand("new_file");
            }
            MenuItem {
                text: qsTr("Open File...")
                shortcut: "Ctrl+O"
                onTriggered: openDialog.open();
            }
            MenuItem {
                text: qsTr("Save")
                shortcut: "Ctrl+S"
                onTriggered: frontend.runCommand("save");
            }
            MenuItem {
                text: qsTr("Save As...")
                shortcut: "Shift+Ctrl+S"
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
                shortcut: "Shift+Ctrl+N"
                onTriggered: frontend.runCommand("new_window");
            }
            MenuItem {
                text: qsTr("Close Window")
                shortcut: "Shift+Ctrl+W"
                onTriggered: frontend.runCommand("close_window");
            }
            MenuSeparator{}
            MenuItem {
                text: qsTr("Close File")
                shortcut: "Ctrl+W"
                onTriggered: frontend.runCommand("close_view");
            }
            MenuItem {
                text: qsTr("Close All Files")
                onTriggered: frontend.runCommand("close_all");
            }
            MenuSeparator{}
            MenuItem {
                text: qsTr("Quit")
                shortcut: "Ctrl+Q"
                onTriggered: Qt.quit(); // frontend.runCommand("quit");
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
                shortcut: "Ctrl+Z"
                onTriggered: frontend.runCommand("undo");
            }
            MenuItem {
                text: qsTr("Redo")
                shortcut: "Ctrl+Y"
                onTriggered: frontend.runCommand("redo");
            }
            Menu {
                title: qsTr("Undo Selection")
                MenuItem {
                    text: qsTr("Soft Undo")
                    shortcut: "Ctrl+U"
                    onTriggered: frontend.runCommand("soft_undo");
                }
                MenuItem {
                    text: qsTr("Soft Redo")
                    shortcut: "Shift+Ctrl+U"
                    onTriggered: frontend.runCommand("soft_redo");
                }
            }
            MenuSeparator{}
            MenuItem {
                text: qsTr("Copy")
                shortcut: "Ctrl+C"
                onTriggered: frontend.runCommand("copy");
            }
            MenuItem {
                text: qsTr("Cut")
                shortcut: "Ctrl+X"
                onTriggered: frontend.runCommand("cut");
            }
            MenuItem {
                text: qsTr("Paste")
                shortcut: "Ctrl+V"
                onTriggered: frontend.runCommand("paste");
            }
        }
        Menu {
            title: qsTr("View")
            MenuItem {
                text: qsTr("Show/Hide Console")
                shortcut: "Ctrl+`"
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

                Label {
                    text: "git branch: master"
                    color: statusBar.textColor
                }

                Label {
                    text: "INSERT MODE"
                    color: statusBar.textColor
                }

                Label {
                    id: statusBarCaretPos
                    text: "Line xx, Column yy"
                    color: statusBar.textColor
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
              TabView {
                Layout.fillHeight: true
                Layout.fillWidth: true
                id: tabs
                objectName: "tabs"
                style: TabViewStyle {
                    frameOverlap: 0
                    tab: Item {
                        implicitWidth: 180
                        implicitHeight: 28
                        ToolTip {
                            backgroundColor: "#BECCCC66"
                            textColor: "black"
                            font.pointSize: 8
                            text: styleData.title
                            Component.onCompleted: {
                                this.parent = tabs;
                            }
                        }
                        BorderImage {
                            source: themeFolder + (styleData.selected ? "/tab-active.png" : "/tab-inactive.png")
                            border { left: 5; top: 5; right: 5; bottom: 5 }
                            width: 180
                            height: 25
                            Text {
                                id: tab_title
                                anchors.centerIn: parent
                                text: styleData.title.replace(/^.*[\\\/]/, '')
                                color: frontend.defaultFg()
                                anchors.verticalCenterOffset: 1
                            }
                        }
                    }
                    tabBar: Image {
                        fillMode: Image.TileHorizontally
                        source: themeFolder + "/tabset-background.png"
                    }
                    tabsMovable: true
                    frame: Rectangle { color: frontend.defaultBg() }
                    tabOverlap: 5
                }

              }
              View {
                id: consoleView
                myView: frontend.console
                minimapVisible: false
                height: 100
              }
        }
    }
    OpenDialog {
        id: openDialog
    }
}
