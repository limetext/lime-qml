import QtQuick 2.0
import QtQuick.Controls 1.0
import QtQuick.Controls.Styles 1.0
import QtQuick.Dialogs 1.0
import QtQuick.Layouts 1.0
import QtGraphicalEffects 1.0


Item {
  id: viewRoot

  property var myView
  property int fontSize: 10
  property string fontFace: "Monospace"
  property var cursor: Qt.IBeamCursor
  property bool ctrl: false
  property bool minimapVisible: true


  property var linesModel: ListModel {}


  function addLine() {
      linesModel.append({});
  }

  function insertLine(idx) {
      linesModel.insert(idx, {});
  }

  Rectangle  {
      color: frontend.defaultBg()
      anchors.fill: parent
      z: -1
  }

  Component.onCompleted: {
    if (myView) {
      // console.log("myView onCompleted");
      // updateMyView();
    }
  }
  onMyViewChanged: {
    updateMyView();
  }

  function updateMyView() {
      console.log("Buffer: updateMyView: ", myView);
      linesModel.clear();
      if (myView)
        myView.fix(viewRoot);
  }

  function onSelectionModified() {
      if (myView == undefined) return;
      editorView.onSelectionModified();
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
