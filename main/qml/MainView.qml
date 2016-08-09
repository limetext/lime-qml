import QtQuick 2.0
import QtQuick.Controls 1.0
import QtQuick.Controls.Styles 1.0
import QtQuick.Dialogs 1.0
import QtQuick.Layouts 1.0
import QtGraphicalEffects 1.0

Item {
  id: mainView
  Layout.fillHeight: true
  Layout.fillWidth: true

  property var rows: [0, 0.7, 1]
  property var cols: [0.0, 0.25, 0.75, 1.0]
  property var cells: [[0, 0, 2, 3]] //[[0, 0, 1, 2], [1, 0, 2, 1], [0, 2, 1, 3], [1, 1, 2, 3]]

  property var tabsMap: ({})

  Component {
      id: tabTemplate
      Item {}
  }

  Component {
    id: viewTemplate
    View {
      id: tabView
      anchors.fill: parent
    }
  }

  property alias cellCount: cellHolder.count
  onCellCountChanged: {
    // TODO: Actually track current cell!
    currentCell = cellHolder.itemAt(0);
  }

  property Cell currentCell: cellHolder.itemAt(0)
  property Tab currentTab: currentCell && currentCell.currentTab
  property View currentView: currentCell && currentCell.currentView

  function getViewFromTab(tab) {
    return tab? tab.item.view : undefined;
  }

  function addTab(tabId, view) {
    var cell = currentCell;
    var tab = cell.addTab(Qt.binding(function() {
      console.log("view.title", view.title);
      return view.title? view.title : "untitled";
    }), tabTemplate);

    tabsMap[tabId] = tab;

    tab.active = true;

    var obj = viewTemplate.createObject(tab.item, {myView: view});
  }

  function activateTab(tabId) {
    var tab = tabsMap[tabId];
    var tabView = tab.parent.parent.parent;

    for (var i = 0; i < tabView.count; i++) {
      if (tabView.getTab(i) == tab) {
        tabView.currentIndex = i;
        return;
      }
    }
  }

  function removeTab(tabId) {
    var tab = tabsMap[tabId];
    var tabView = tab.parent.parent.parent;

    for (var i = 0; i < tabView.count; i++) {
      if (tabView.getTab(i) == tab) {
        tabView.removeTab(i);
        return;
      }
    }
  }

  function setTabTitle(tabId, title) {
    var tab = tabsMap[tabId];
    tab.title = title;
  }

  Repeater {
    id: cellHolder

    model: cells

    Cell {
      id: cellItem

      // property Cell cell: cellCell

      property real leftPercent: rows[modelData[0]]
      property real topPercent: cols[modelData[1]]
      property real rightPercent: rows[modelData[2]]
      property real bottomPercent: cols[modelData[3]]

      property real leftT: mainView.width * leftPercent
      property real topT: mainView.height * topPercent
      property real rightT: mainView.width * rightPercent
      property real bottomT: mainView.height * bottomPercent

      x: leftT + (leftPercent < 0.01 ? 0 : 2)
      width: rightT - x - (rightPercent > 0.99 ? 0 : 2)
      y: topT + (topPercent < 0.01 ? 0 : 2)
      height: bottomT - y - (bottomPercent > 0.99 ? 0 : 2)

      property Tab currentTab: count > 0 ? getTab(currentIndex) : null
      property View currentView: (currentTab && currentTab.item && currentTab.item.children.length) ?
        currentTab.item.children[0] : null

      property var currentMyView: currentView && currentView.myView
      onCurrentMyViewChanged: currentMyView && updateCurrent()


      function updateCurrent() {
          currentView.myView.setActive();
      }

    }
  }
}
