import QtQuick 2.0
import QtQuick.Controls 1.0
import QtQuick.Controls.Styles 1.0
import QtQuick.Dialogs 1.0
import QtQuick.Layouts 1.0
import QtGraphicalEffects 1.0


Item {
  id: viewRoot

  property var myView
  property int fontSize: 12
  property string fontFace: "Helvetica"
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
      // if (!isMinimap) {
          console.log("LimeEditor: updateMyView: ", myView);
          linesModel.clear();
          // listView.myView = myView;
          if (myView)
            myView.fix(viewRoot);
      // }
  }

  function onSelectionModified() {
      if (myView == undefined) return;
      editorView.onSelectionModified()
      // minimap.onSelectionModified()
  }

  RowLayout {
    anchors.fill: parent
    LimeEditor {
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

        // Component.onCompleted: {
        //   console.log("initizing editor: ", this.x)
        // }
    }
    LimeMinimap {
        id: minimap
        Layout.maximumWidth: 200
        Layout.minimumWidth: 200
        Layout.preferredWidth: 200
        Layout.fillHeight: true
        width: 200


        visible: viewRoot.minimapVisible

        linesModel: viewRoot.linesModel
        myView: viewRoot.myView
        fontSize: viewRoot.fontSize
        fontFace: viewRoot.fontFace
        cursor: viewRoot.cursor
        ctrl: viewRoot.ctrl

        // Component.onCompleted: {
        //   console.log("initizing minimap: ", minimap.x)
        // }

        // property var oldView

        function scroll() {
            var p = percentage(myView);
            // children[1].contentY = p*(children[1].contentHeight-height);
            if (!ma.drag.active) {
                minimapArea.y =  p*(height-minimapArea.height)
            }
        }
        // onMyViewChanged: {
        //   // console.log("myViewChanged", myView);
        //     if (oldView && oldView.contentYChanged) {
        //         oldView.contentYChanged.disconnect(scroll);
        //     }
        //     if (myView && myView.contentYChanged) {
        //       myView.contentYChanged.connect(scroll);
        //     }
        //     oldView = myView;
        // }
        function percentage(view) {
            if (view === undefined) { return 10; }
            if (!view.visibleArea) return 10;
            return view.visibleArea.yPosition/(1-view.visibleArea.heightRatio);
        }
        Rectangle {
            id: minimapArea
            width: parent.width
            height: (myView && myView.visibleArea) ? myView.visibleArea.heightRatio*parent.children[1].contentHeight : parent.height
            color: "white"
            opacity: 0.1
            onYChanged: {
                if (ma.drag.active) {
                    myView.contentY = y*(myView.contentHeight-myView.height)/(parent.height-height);
                }
            }
            onHeightChanged: {
                parent.scroll();
            }
            MouseArea {
                id: ma
                drag.target: parent
                anchors.fill: parent
                drag.minimumX: 0
                drag.minimumY: 0
                drag.maximumY: parent.parent.height-height
                drag.maximumX: parent.parent.width-width
            }
        }
    }
  }
}
