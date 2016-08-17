import QtQuick 2.0
import QtQuick.Controls 1.0
import QtQuick.Controls.Styles 1.0
import QtQuick.Dialogs 1.0
import QtQuick.Layouts 1.0
import QtGraphicalEffects 1.0


Item {
  id: viewRoot

  property var myView
  property int fontSize: myView.fontSize
  property string fontFace: "Monospace"
  property var cursor: Qt.IBeamCursor
  property bool ctrl: false
  property bool minimapVisible: true

  property var statusBar: {}


  property var linesModel: myView.formattedLines

  function setTitle(title) {
    parent.title = title;
  }

  Rectangle  {
      color: frontend.defaultBg()
      anchors.fill: parent
      z: -1
  }

  // Component.onCompleted: {
  //   if (myView) {
  //     console.log("myView onCompleted", myView);
  //     updateMyView();
  //   }
  // }
  onMyViewChanged: {
    console.log("myViewChanged", myView);
    updateMyView();
  }

  function updateMyView() {
      // linesModel.clear();
      if (myView) {
        console.log("updateMyView", myView);
        myView.fix(viewRoot);
      }
  }

  function onSelectionModified() {
      if (myView == undefined) return;
      editorView.onSelectionModified();
  }

  function onStatusChanged() {
    if (myView == undefined) return;
    var bv = myView.back();
    statusBar = bv.status();
  }

  RowLayout {
    anchors.fill: parent
    Buffer {
        id: editorView
        Layout.fillWidth: true
        Layout.fillHeight: true

        z: 100
        linesModel: viewRoot.linesModel
        myView: viewRoot.myView
        fontSize: viewRoot.fontSize
        fontFace: viewRoot.fontFace
        cursor: viewRoot.cursor
        ctrl: viewRoot.ctrl
    }
    Minimap {
        id: minimap
        Layout.maximumWidth: 200
        Layout.minimumWidth: 200
        Layout.preferredWidth: 200
        Layout.fillHeight: true
        width: 200


        visible: viewRoot.minimapVisible

        linesModel: viewRoot.linesModel
        myView: viewRoot.myView
        mainListView: editorView.listView
        fontSize: viewRoot.fontSize
        fontFace: viewRoot.fontFace
        cursor: viewRoot.cursor
        ctrl: viewRoot.ctrl
    }
  }
}
