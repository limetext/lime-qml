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

    Component.onCompleted: {
      Qt.application.name = "LimeTextQML"
    }

    property var myWindow
    property string themeFolder: "../../packages/Soda/Soda Dark"

    function addTab(tabId, view) {
        return mainView.addTab(tabId, view);
    }

    function activateTab(tabId) {
        return mainView.activateTab(tabId);
    }

    function removeTab(tabId) {
        return mainView.removeTab(tabId);
    }

    function setTabTitle(tabId, title) {
        return mainView.setTabTitle(tabId, title);
    }

    function setFrontendStatus(text) {
        frontendStatus.text = text
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
                onTriggered: frontend.runCommand("prompt_save_as");
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

    property Tab currentTab: mainView.currentTab
    property View currentView: mainView.currentView

    Item {
        anchors.fill: parent
        Keys.onPressed: {
            var v = currentView; if (v === undefined) return;
            if (event.key == Qt.Key_Control) v.ctrl = true;
            event.accepted = frontend.handleInput(event.text, event.key, event.modifiers)
            event.accepted = true;
        }
        Keys.onReleased: {
            var v = currentView; if (v === undefined) return;
            if (event.key == Qt.Key_Control) v.ctrl = false;
        }
        focus: true // Focus required for Keys.onPressed
        SplitView {
            anchors.fill: parent
            orientation: Qt.Vertical
            SplitView {
                anchors.fill: parent
                orientation: Qt.Horizontal
                Sidebar {
                  id: sidebarView
                  width: 200
                  sidebarTree: myWindow && myWindow.sidebarTree
                }
                MainView {
                  id: mainView
                  Layout.fillWidth: true
                }
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


    statusBar: StatusBar {
        id: statusBar
        property color textColor: "#969696"

        style: StatusBarStyle {
            background: Image {
                source: themeFolder + "/status-bar-background.png"
            }
        }

        RowLayout {
            anchors.fill: parent
            spacing: 15

            RowLayout {
                anchors.fill: parent
                spacing: 3
                Repeater {
                    // model: statusBarSorted
                    delegate:
                    Label {
                        text: modelData[1]
                        color: statusBar.textColor
                    }
                }
                Label {
                    id: frontendStatus
                    color: statusBar.textColor
                }
            }

            Label {
                id: indentStatus
                color: statusBar.textColor
                Layout.alignment: Qt.AlignRight
                text: currentView && currentView.myView ? currentView.myView.tabSize : "0"
            }

            Label {
                id: syntaxStatus
                color: statusBar.textColor
                Layout.alignment: Qt.AlignRight
                text: currentView && currentView.myView ? currentView.myView.syntaxName : "0"
            }
        }
    }


    MessageDialog {
        objectName: "messageDialog"
        onAccepted: frontend.promptClosed("accepted")
        onApply: frontend.promptClosed("apply")
        onDiscard: frontend.promptClosed("discard")
        onHelp: frontend.promptClosed("help")
        onNo: frontend.promptClosed("no")
        onRejected: frontend.promptClosed("rejected")
        onReset: frontend.promptClosed("reset")
        onYes: frontend.promptClosed("yes")
    }

    FileDialog {
        objectName: "fileDialog"
        onAccepted: frontend.promptClosed("accepted")
        onRejected: frontend.promptClosed("rejected")
    }
}
