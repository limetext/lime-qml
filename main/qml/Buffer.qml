import QtQuick 2.5
import QtQuick.Layouts 1.0

Item {
    id: editorRoot

    property var linesModel
    property var myView
    property var listView: listView
    // property bool isMinimap: false
    property int fontSize: 10
    property string fontFace: "Monospace"
    property var cursor: Qt.IBeamCursor
    property bool ctrl: false

    function getCurrentSelection() {
        if (!myView || !myView.back()) {
          console.log("returning null selection", myView, myView? myView.back() : false);
          return null;
        }
        return myView.back().sel();
    }

    FontMetrics {
      id: fontMetrics
      font.family: editorRoot.fontFace
      font.pointSize: editorRoot.fontSize
    }

    property var editorFont: editorRoot.fontSize + "pt \"" + editorRoot.fontFace + "\"" + ", monospace"
    property var spaceWidth: 25
    property var tabWidth: spaceWidth * 4
    property var lineHeight: fontMetrics.lineSpacing

    property var lineNumbersWidth: fontMetrics.advanceWidth(''+(myView? myView.lines() : "0"))
    onEditorFontChanged: {
      editorRoot.spaceWidth = fontMetrics.advanceWidth(' ');
    }

    Component {
      id: gutterComponent
      Row {
        // id: gutter
        // width: childrenRect.width + 25
        height: lineHeight + 2
        // property string lineNumberText: "0"
        clip: true


        Text {
          id: lineNumber
          width: lineNumbersWidth
          height: parent.height

          text: lineNumberText
          color: "white"
        }

        Item {
          // spacing
          width: 10
          height: 10
        }

        Rectangle {
          width: 1
          height: parent.height

          color: "white"
        }

        Item {
          // spacing
          width: 10
          height: 10
        }
      }
    }

    Loader {
      id: gutterTest
      visible: false
      sourceComponent: gutterComponent
      property string lineNumberText: "0"
    }


    property var gutterWidth: gutterTest.childrenRect.width

    function measureChunk(chunk) {
      var ctext = chunk.text;
      var ctlen = ctext.length;
      var j = 0;

      var skipWidth = 0;
      var tabWidth = editorRoot.tabWidth;

      // TODO: Are tabs always at the beginning of the chunk?
      while (j < ctlen && ctext[j] === '\t') {
        j++;
        skipWidth += tabWidth;
      }

      ctext = ctext.slice(j);
      var cwidth = fontMetrics.advanceWidth(ctext);

      chunk.skipWidth = skipWidth;
      chunk.width = cwidth;
      chunk.measured = true;
    }


    ListView {
        id: listView
        model: linesModel
        anchors.fill: parent
        boundsBehavior: Flickable.StopAtBounds
        // cacheBuffer: contentHeight < 0? 0 : contentHeight
        interactive: false
        clip: true
        z: 100

        property bool showBars: false
        property var cursor: editorRoot.cursor

        delegate:
          Item {
            width: parent.width
            height: lineHeight

            Loader {
              id: gutter
              sourceComponent: gutterComponent
              width: gutterWidth

              property string lineNumberText: index+1
            }

            Canvas {
              id: canvas
              property var line: myView && index > -1 ? myView.line(index) : null
              property var lineText: !line ? null : line.text

              onLineTextChanged: {
                requestPaint();
              }

              // TODO: why doesn't this work?
              // property var spaceWidth: fontMetrics.advanceWidth(' ')
              // property var tabWidth: spaceWidth * 4

              anchors {
                left: gutter.right
                right: parent.right
              }
              height: lineHeight + 2


              onPaint: {
                var l = line;
                if (l == null) return;

                var startTime = new Date().getTime();

                var drawTabs = false;

                var tabWidth = editorRoot.tabWidth;

                var ctx = canvas.getContext("2d");
                ctx.reset();

                ctx.font = editorFont;

                var defaultColor = "#ffffff";
                var currentColor = defaultColor;
                ctx.fillStyle = currentColor;

                ctx.clearRect(0, 0, canvas.width, canvas.height);

                var len = l.chunksLen();
                var x = 0;
                var y = fontMetrics.ascent;
                var yval = y - fontMetrics.strikeOutPosition;
                for (var i = 0; i < len; i++) {
                  var c = l.chunk(i);

                  if (c.foreground === "") {
                    if (currentColor !== defaultColor)
                      ctx.fillStyle = currentColor = defaultColor;
                  } else {
                    if (currentColor !== c.foreground)
                      ctx.fillStyle =  '#' + (currentColor = c.foreground);
                  }

                  if (!c.measured) {
                    measureChunk(c)
                  }

                  // ctx.fillRect(x, 0, 1, canvas.height); // Chunk left border

                  var ctext = c.text;
                  var ctlen = ctext.length;
                  var j = 0;

                  // TODO: Are tabs always at the beginning of the chunk?
                  while (j < ctlen && ctext[j] === '\t') {
                    j++;
                    if (drawTabs) ctx.fillRect(x+2, yval, tabWidth-4, 1);
                    x += tabWidth;
                  }

                  ctext = ctext.slice(j);

                  ctx.fillText(ctext, x, y);
                  x += c.width;
                }

                l.width = x;

                var endTime = new Date().getTime();

                var duration = endTime - startTime;
                // console.log("Paint took", duration);
                reportPaint(duration);
              }
            }
          }

        states: [
            State {
                name: "ShowBars"
                when: listView.movingVertically || listView.movingHorizontally || listView.flickingVertically || listView.flickingHorizontally
                PropertyChanges {
                    target: listView
                    showBars: true
                }
            },
            State {
                name: "HideBars"
                when: !listView.movingVertically && !listView.movingHorizontally && !listView.flickingVertically && !listView.flickingHorizontally
                PropertyChanges {
                    target: listView
                    showBars: false
                }
            }
        ]

        MouseArea {
            property var point: new Object()

            x: 0
            y: 0
            cursorShape: parent.cursor
            propagateComposedEvents: true
            height: parent.height
            width: parent.width-verticalScrollBar.width


            function colFromMouseX(lineIndex, mouseX) {

                const printDebug = false;

                var line = myView.line(lineIndex);
                var lineText = line.rawText;

                var fullWidth = line.width;

                mouseX -= gutterWidth;
                if (mouseX <= 0) {
                  return 0;
                }


                // if the click was farther right than the last character of
                // the line then return the last character's column
                if(mouseX > fullWidth) {
                  return lineText.length;
                }

                // calculate a column from a given mouse x coordinate and the line text.
                var col;

                // Trying to find closest column to clicked position
                var len = line.chunksLen();
                if (len == 0) return 0;
                var ci = 0;
                var partialWidth = 0;
                var partialCol = 0;
                var chunk;
                while (ci < len) {
                  chunk = line.chunk(ci);
                  var afterX = partialWidth + chunk.skipWidth + chunk.width;

                  if (mouseX < afterX)
                    break;

                  partialWidth = afterX;
                  partialCol += chunk.text.length;
                  ci += 1;
                }

                var chunkX = mouseX - partialWidth;
                if (chunkX < chunk.skipWidth) {
                  col = partialCol + Math.round(chunkX / tabWidth);
                }
                else {
                  partialWidth += chunk.skipWidth;
                  chunkX -= chunk.skipWidth;

                  var ctext = chunk.text;

                  var j = 0;
                  while (j < ctext.length && ctext[j] == '\t') j++;

                  ctext = ctext.slice(j);

                  col = partialCol + j + Math.round(ctext.length * (chunkX / chunk.width));

                }

                if (col > lineText.length) col = lineText.length;

                if (printDebug) console.log("cols: ", oldcol, col)

                return col;
            }


            onPositionChanged: {
                var item  = listView.itemAt(0, mouse.y+listView.contentY),
                    index = listView.indexAt(0, mouse.y+listView.contentY),
                    selection = getCurrentSelection();

                if (item != null && selection != null) {
                    var col = colFromMouseX(index, mouse.x);
                    point.r = myView.back().textPoint(index, col);
                    if (point.p != null && point.p != point.r) {
                        // Remove the last region and replace it with new one
                        var r = selection.get(selection.len()-1);
                        selection.substract(r);
                        selection.add(myView.region(point.p, point.r));
                        onSelectionModified();
                    }
                }
                point.r = null;
            }

            onPressed: {
                // TODO:
                // Changing caret position doesn't work on empty lines

                var item  = listView.itemAt(0, mouse.y+listView.contentY),
                    index = listView.indexAt(0, mouse.y+listView.contentY),
                    selection = getCurrentSelection();

                if (item != null) {
                    var col = colFromMouseX(index, mouse.x);
                    point.p = myView.back().textPoint(index, col)

                    if (!ctrl) {
                        selection.clear();
                    }

                    selection.add(myView.region(point.p, point.p));

                    onSelectionModified();
                }
            }

            onDoubleClicked: {

                var item  = listView.itemAt(0, mouse.y+listView.contentY),
                    index = listView.indexAt(0, mouse.y+listView.contentY);

                if (item != null) {
                    var col = colFromMouseX(index, mouse.x);
                    point.p = myView.back().textPoint(index, col)

                    if (!ctrl) {
                        getCurrentSelection().clear();
                    }

                    getCurrentSelection().add(myView.back().expandByClass(myView.region(point.p, point.p), 1|2|4|8))
                    onSelectionModified();
                }
            }

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

        Rectangle {
            id: verticalScrollBar

            width: 10
            radius: width
            height: listView.visibleArea.heightRatio * listView.height
            anchors.right: listView.right
            opacity: (listView.showBars || ma.containsMouse || ma.drag.active) ? 0.5 : 0.05

            onYChanged: {
                if (ma.drag.active) {
                    listView.contentY = y*(listView.contentHeight-listView.height)/(listView.height-height);
                }
            }

            states: [
                State {
                    when: !ma.drag.active
                    PropertyChanges {
                        target: verticalScrollBar
                        y: listView.visibleArea.yPosition*listView.height
                    }
                }
            ]

            Behavior on opacity { PropertyAnimation {} }
        }

        MouseArea {
            id: ma
            enabled: true
            width: verticalScrollBar.width
            height: listView.height
            anchors.right: parent.right
            hoverEnabled: true
            drag.target: verticalScrollBar
            drag.minimumY: 0
            drag.maximumY: listView.height-verticalScrollBar.height
        }
    }

    Flickable  {
      anchors.fill: parent
      contentY: listView.contentY
      interactive: false

      Repeater {
        id: highlightedLines
        property var currentSelection: null
        model: currentSelection ? currentSelection.len() : 0

        onCurrentSelectionChanged: {
          model = currentSelection ? currentSelection.len() : 0;
          if (!currentSelection) console.log("Bad currentSelection!", currentSelection);
        }

        delegate: Component {
          id: selComp
          Canvas {
            id: selCanvas

            Connections {
              target: highlightedLines
              onCurrentSelectionChanged: updateSelection()
            }
            Component.onCompleted: {
              updateSelection();
            }


            function updateSelection() {
              var csel = highlightedLines.currentSelection;
              if (csel) {
                selection = csel.get(index);
              }
              else {
                console.log("cc bad cs ", csel);
              }
            }

            property var selection: null
            property var safeSelection: null
            property bool isBlinking: false

            height: lineHeight
            width: listView.width
            y: 0
            z: listView.z-1

            property var lastSelection: null

            onSelectionChanged: {
              if (!selection) return;

              if (lastSelection && lastSelection.a == selection.a && lastSelection.b == selection.b) {
                console.log("Selection not changed");
                return;
              }

              var back = myView.back();

              safeSelection = toSafeSelection(selection);

              var first = back.rowCol(safeSelection.a);
              var last = back.rowCol(safeSelection.b);

              var firstLine = first[0];
              var lastLine = last[0];

              y = firstLine * lineHeight;
              height = (lastLine - firstLine + 1) * lineHeight + 1;

              selCanvas.requestPaint();
            }

            function getYPosition(rowCol) {
                if(rowCol) {
                    return rowcol[0] * lineHeight;
                }
                return 0;
            }

            onPaint: {
                if (!selection) {
                  console.log("Skipping paint");
                  return;
                }
                var back = myView.back();

                var ctx = selCanvas.getContext("2d");

                ctx.clearRect(0, 0, selCanvas.width, selCanvas.height);

                var outlineColor = "#ffffff";
                var fillColor = "#888888";
                ctx.fillStyle = outlineColor;

                var first = back.rowCol(safeSelection.a);
                var last = back.rowCol(safeSelection.b);

                var firstLine = first[0];
                var lastLine = last[0];

                var lh = lineHeight;

                var y = 0;

                var lastxA = -1;
                var lastxB = -1;

                if (first[0] == last[0] && first[1] == last[1]) {
                  var xA = getCursorOffset(first, back);


                  var caretStyle = myView.setting("caret_style"),
                      inverseCaretState = myView.setting("inverse_caret_state");

                  if (!isBlinking) {
                    selCanvas.opacity = Qt.binding(function() { return cursorOpacity; });
                    isBlinking = true;
                  }

                  if (caretStyle == "underscore") {
                    if (inverseCaretState) {
                      ctx.fillRect(xA, lh-1, editorRoot.spaceWidth, 1);
                    } else {
                      ctx.fillRect(xA, 0, 1, lh);
                    }
                  }

                  return;
                } else {
                  if (isBlinking) {
                    selCanvas.opacity = 1;
                    isBlinking = false;
                  }
                }

                for(var i = firstLine; i <= lastLine; i++) {
                  var lr = back.line(back.textPoint(i, 0));
                  var a = (i == firstLine)? safeSelection.a : lr.a;
                  var b  = (i == lastLine)? safeSelection.b : lr.b;

                  var rowcolA = back.rowCol(a);
                  var rowcolB = back.rowCol(b);

                  var xA = getCursorOffset(rowcolA, back);
                  var xB = getCursorOffset(rowcolB, back);

                  ctx.fillStyle = fillColor;
                  ctx.fillRect(xA+1, y, xB-xA-1, lh+1);

                  ctx.fillStyle = outlineColor;
                  ctx.fillRect(xA, y+1, 1, lh-1);
                  ctx.fillRect(xB, y+1, 1, lh-1);

                  if (i == firstLine) {
                    ctx.fillRect(xA+1, y, xB - xA-1, 1);
                  } else {
                    ctx.fillRect(Math.min(xA, lastxA)+1, y, Math.abs(lastxA - xA)-1, 1);
                    ctx.fillRect(Math.min(xB, lastxB)+1, y, Math.abs(lastxB - xB)-1, 1);
                  }

                  y += lh;

                  if (i == lastLine) {
                    ctx.fillRect(xA+1, y, xB - xA-1, 1);
                  }

                  lastxA = xA;
                  lastxB = xB;
                }

              }
            }
        }
    }
  }

    function toSafeSelection(selection) {
      return (selection.b > selection.a) ?
                  { a: selection.a, b: selection.b, reversed: false }:
                  { a: selection.b, b: selection.a, reversed: true };
    }

    function onSelectionModified() {
      highlightedLines.currentSelection = getCurrentSelection();
      resetBlink();
    }

    // getCursorOffset returns the x coordinate for the cursor.
    function getCursorOffset(rowcol, buf) {
        var line = myView.line(rowcol[0]);

        var len = line.chunksLen();
        var partialWidth = gutterWidth;
        if (len == 0) return partialWidth;
        var partialCol = 0;
        var ci = 0;
        var chunk;
        while (ci < len) {
          chunk = line.chunk(ci);
          if (chunk.measured == false) {
            measureChunk(chunk);
          }
          var totalCol = partialCol + chunk.text.length;
          if (totalCol > rowcol[1])
            break;

          partialCol = totalCol;
          partialWidth += chunk.skipWidth + chunk.width;
          ci ++;
        }

        var chunkCol = rowcol[1] - partialCol;

        var ctext = chunk.text;

        var j = 0;
        var tlen = ctext.length;
        while (j < tlen) {
          if (j == chunkCol) {
            return partialWidth;
          }
          if (ctext[j] != '\t') break;
          partialWidth += tabWidth;
          j++;
        }

        // j is now the start of non-tab characters in the chunk
        var textToCursor = ctext.slice(j, chunkCol);

        var cursorOffset = fontMetrics.advanceWidth(textToCursor);

        return partialWidth + cursorOffset;

    }

    property int numPaints: 0
    property real totalDuration: 0

    function reportPaint(duration) {
      numPaints += 1;
      totalDuration += duration;

      // paintTimer.running = true;
    }

    // Timer {
    //   id: paintTimer
    //     interval: 100
    //     repeat: false
    //     running: false
    //     onTriggered: {
    //         if (numPaints == 0) return;
    //
    //         console.log("Paints", numPaints, "in the last", interval, ", total", totalDuration, "average", totalDuration / numPaints);
    //
    //         numPaints = 0;
    //         totalDuration = 0;
    //     }
    // }


    function resetBlink() {
      startTime = Date.now()
    }
    property real cursorOpacity: 1
    property real startTime: Date.now()
    Timer {
        interval: 100
        repeat: true
        running: true
        onTriggered: {
            cursorOpacity = 0.5 + 0.5 * Math.cos((Date.now() - startTime)*0.008);
        }
    }
}
