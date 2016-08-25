import QtQuick 2.0
import QtQuick.Controls 1.0
import QtQuick.Controls.Styles 1.0
import QtQuick.Dialogs 1.0
import QtQuick.Layouts 1.0
import QtGraphicalEffects 1.0

Rectangle {
  id: sidebarRoot
  color: frontend.defaultBg()

  property var sidebarTree

  ListView {
    id: listView

    anchors.fill: parent
    model: sidebarTree && sidebarTree.model
    focus: true
    interactive: false
    // keyNavigationEnabled: true
    highlightFollowsCurrentItem: false

    highlight: Rectangle {
      anchors.left: parent.left
      anchors.right: parent.right
      y: listView.currentItem.y
      height: listView.currentItem.height
      color: "lightsteelblue"
    }

    add: Transition {
      id: addTrans
      SequentialAnimation {
        PropertyAction { property: "opacity"; value: 0 }
        PauseAnimation {
            duration: 25 //(dispTrans.ViewTransition.index -
                    // dispTrans.ViewTransition.targetIndexes[0]) * 10
        }
        ParallelAnimation {
          NumberAnimation { property: "opacity"; from: 0; to: 1.0; duration: 100 }
          NumberAnimation { properties: "x"; from: addTrans.ViewTransition.destination.x-100; duration: 100 }
        }
      }
    }

    remove: Transition {
        NumberAnimation { property: "opacity"; from: 1.0; to: 0; duration: 100 }
        NumberAnimation { properties: "x"; to: addTrans.ViewTransition.destination.x+100; duration: 100 }
    }

    // displaced: Transition {
    //     NumberAnimation { properties: "x,y"; duration: 400; easing.type: Easing.OutBounce }
    // }

    addDisplaced: Transition {
        id: addDispTrans
        SequentialAnimation {
            PauseAnimation {
                duration: 0 //(dispTrans.ViewTransition.index -
                        // dispTrans.ViewTransition.targetIndexes[0]) * 10
            }
            NumberAnimation { properties: "x,y"; duration: 50 }
        }
    }

    removeDisplaced: Transition {
        id: rmDispTrans
        SequentialAnimation {
            PauseAnimation {
                duration: 25 //(dispTrans.ViewTransition.index -
                        // dispTrans.ViewTransition.targetIndexes[0]) * 10
            }
            NumberAnimation { properties: "x,y"; duration: 50 }
        }
    }

    delegate: MouseArea {
      width: parent.width
      height: childrenRect.height
      onClicked: {
        display.item.click()
        listView.currentIndex = index;
      }

      RowLayout {
        spacing: 0

        // Component.onCompleted: {
        //   console.log("completed: ", index, display, display.item.name())
        // }

        Item {
          // indent
          height: 2
          width: display.indent * 20
        }
        Item {
          id: expandoHolder
          Layout.fillHeight: true
          Layout.topMargin: 2
          Layout.bottomMargin: 2
          Layout.preferredWidth: expandoImg.width

          Image {
            id: expandoImg
            height: parent.height
            width: 100
            source: themeFolder + (display.expanded? "/fold-open@2x.png" : "/fold-closed@2x.png")
            visible: status != Image.Ready || display.item.hasChildren()
            fillMode: Image.PreserveAspectFit

            // without this, width has a binding loop
            onPaintedWidthChanged: {
              if (paintedWidth > 0)
                width = paintedWidth;
            }

            MouseArea {
              anchors.fill: parent
              onClicked: {
                if (display.expanded) {
                  sidebarTree.collapseNode(index)
                } else {
                  sidebarTree.expandNode(index)
                }
              }
            }
          }
        }
        Item {
          id: iconHolder
          Layout.fillHeight: true
          Layout.topMargin: 2
          Layout.bottomMargin: 2
          Layout.preferredWidth: iconImg.width

          Image {
            id: iconImg
            height: parent.height
            width: 100
            source: themeFolder + (display.expanded? "/folder-open@2x.png" : "/folder-closed@2x.png")
            visible: status != Image.Ready || display.item.hasChildren()
            fillMode: Image.PreserveAspectFit

            // without this, width has a binding loop
            onPaintedWidthChanged: {
              if (paintedWidth > 0)
                width = paintedWidth;
            }
          }
        }
        // Rectangle {
        //   id: icon
        //   color: "green"
        //   width: 15
        //   height: 15
        // }
        Text {
          text: display.item.text()
          color: "white"
        }
      }
    }
    MouseArea {
      anchors.fill: parent
      property var point: new Object()

      x: 0
      y: 0
      // cursorShape: parent.cursor
      propagateComposedEvents: true

      onWheel: {
          var delta = wheel.pixelDelta,
              scaleFactor = 30;

          if (delta.x == 0 && delta.y == 0) {
              delta = wheel.angleDelta;
              scaleFactor = 15;
          }

          scaleFactor /= 3;

          listView.flick(delta.x*scaleFactor, delta.y*scaleFactor);
          wheel.accepted = true;
      }
    }
  }
}
