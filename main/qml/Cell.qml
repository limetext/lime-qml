import QtQuick 2.0
import QtQuick.Controls 1.0
import QtQuick.Controls.Styles 1.0
import QtQuick.Dialogs 1.0
import QtQuick.Layouts 1.0
import QtGraphicalEffects 1.0


TabView {
  Layout.fillHeight: true
  Layout.fillWidth: true
  id: tabs

  style: TabViewStyle {
      frameOverlap: 0
      tab: Item {
          implicitWidth: 180
          implicitHeight: 28

          property string titleText: (styleData.title != "") ? styleData.title : "untitled"

          ToolTip {
              backgroundColor: "#BECCCC66"
              textColor: "black"
              font.pointSize: 8
              text: titleText
              visibleParent: tabs
          }
          BorderImage {
              source: themeFolder + (styleData.selected ? "/tab-active.png" : "/tab-inactive.png")
              border { left: 5; top: 5; right: 5; bottom: 5 }
              width: 180
              height: 25
              Text {
                  id: tab_title
                  anchors.centerIn: parent
                  text: titleText.replace(/^.*[\\\/]/, '')
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
