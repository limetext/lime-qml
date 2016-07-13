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

  function addTab(view) {
    var cell = currentCell();
    var tab = cell.addTab(Qt.binding(function() {
      console.log("view.title", view.title);
      return view.title? view.title : "untitled";
    }), tabTemplate);
    console.log("addTab", tab, tab.item);

    var loadTab = function() {
      tab.item.myView = view;
    }

    if (tab.item != null) {
      loadTab();
    } else {
      tab.loaded.connect(loadTab);
    }
    tab.active = true;
  }

  function activateTab(tabIndex) {
    var cell = currentCell();
    cell.currentIndex = tabIndex;
  }

  function removeTab(tabIndex) {
    var cell = currentCell();
    cell.removeTab(tabIndex);
  }

  function setTabTitle(tabIndex, title) {
    var cell = currentCell();
    cell.getTab(tabIndex).title = title;
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

      Component.onCompleted: {

      }

      // onWidthChanged: {
      //   if (width < 100) return;
      //
      //   console.log("Cell %:", index, modelData,  leftPercent, rightPercent, topPercent, bottomPercent);
      //   console.log("Cell x:", index, modelData,  leftT, rightT, topT, bottomT);
      //   console.log("Cell s:", index, x, width, y, height);
      // }

      // Cell {
      //   id: cellCell
      //   anchors.fill: parent
      // }

      // Rectangle{
      //   color: "blue"
      //   anchors.fill: parent
      // }

    }
  }
}
