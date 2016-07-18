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

  property var cellObjects: []

  property var tabsMap: ({})


  function currentCell() {
    // TODO: Handle current cell?
    var cellItem = cellHolder.itemAt(0);
    if (cellItem == null) return null;
    return cellItem;
  }

  function currentTab() {
    var cell = currentCell();
    if (cell == null) return null;
    return cell.getTab(cell.currentIndex);
  }

  function view() {
    var tab = currentTab();
     return tab === undefined ? undefined : tab.item;
  }

  function addTab(tabId, view) {
    var cell = currentCell();
    var tab = cell.addTab(Qt.binding(function() {
      console.log("view.title", view.title);
      return view.title? view.title : "untitled";
    }), tabTemplate);
    console.log("addTab", tab, tab.item);

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

      onCurrentIndexChanged: {
        getTab(currentIndex).item.children[0].myView.setActive();
      }

    }
  }
}
