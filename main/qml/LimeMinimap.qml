import QtQuick 2.5
import QtQuick.Layouts 1.0

Item {
    id: viewItem

    property var linesModel
    property var myView
    property var mainListView
    property int fontSize: 12
    property string fontFace: "Menlo"
    property var cursor: Qt.IBeamCursor
    property bool ctrl: false

    function getCurrentSelection() {
        if (!myView || !myView.back()) return null;
        return myView.back().sel();
    }

    // Rectangle  {
    //     color: "#55555555" // frontend.defaultBg()
    //     anchors.fill: parent
    // }

    // onFontSizeChanged: {
    //     dummy.font.pointSize = fontSize;
    // }


    FontMetrics {
      id: fontMetrics
      font.family: viewItem.fontFace
      font.pointSize: viewItem.fontSize
    }

    ListView {
        id: listView
        model: linesModel
        anchors.fill: parent
        boundsBehavior: Flickable.StopAtBounds
        // cacheBuffer: contentHeight < 0? 0 : contentHeight
        interactive: false
        clip: true
        z: 4

        property bool showBars: false
        property var cursor: parent.cursor

        delegate:
            Canvas {
              id: canvas
              property var line: !myView ? null : myView.line(index)
              property var lineText: !line ? null : line.text
              onLineTextChanged: {
                // console.log("Changed");
                requestPaint();
              }
              width: parent.width
              height: 3 //fontMetrics.lineSpacing
              onPaint: {
                if (line == null) return;
                var l = line;
                // console.log("Paint: ", l.rawText, l.chunks());

                var ctx = canvas.getContext("2d");

                var spaceWidth = fontMetrics.advanceWidth(' ');
                var tabWidth = spaceWidth * 4;

                // if (viewItem.fontFace != "Helvetica") {
                //   ctx.font = viewItem.fontSize + "pt \"" + viewItem.fontFace + "\"";
                // }
                var defaultColor = "#ffffff";
                var currentColor = defaultColor;
                ctx.fillStyle = currentColor;

                ctx.clearRect(0, 0, canvas.width, canvas.height);

                // ctx.fillText(l.rawText, 0, fontMetrics.height);

                var chunkDebug = false;
                // if (l.rawText[1] == '"')
                //   chunkDebug = true;

                var len = l.chunksLen();
                var x = 0;
                // var y = fontMetrics.ascent;
                var chunks = "";
                var lineHeight = 2;
                // var widthMul = lineHeight / fontMetrics.lineSpacing;
                // console.log("widthMul: ", widthMul, spaceWidth, fontMetrics.lineSpacing);


                for (var ci = 0; ci < len; ci++) {
                  var c = l.chunk(ci);
                  // if (chunkDebug) chunks += "(" + c.text + ")";

                  if (c.foreground === "") {
                    if (currentColor !== defaultColor)
                      ctx.fillStyle = currentColor = defaultColor;
                  } else {
                    if (currentColor !== c.foreground)
                      ctx.fillStyle =  '#' + (currentColor = c.foreground);
                  }
                  var ctext = c.text;
                  var ctlen = ctext.length;
                  // var cwidth = fontMetrics.advanceWidth(ctext);

                  var i = 0;

                  while (i < ctlen) {
                    var c = ctext[i];
                    if (c == ' ') {
                      x += 1; i++; continue;
                    }
                    if (c == '\t') {
                      x += 4; i++; continue;
                    }

                    var j = i+1;
                    while (j < ctlen) {
                      var cj = ctext[j];
                      if (cj == ' ' || cj == '\t')
                        break;
                      j++;
                    }

                    var ptext = ctext.slice(i, j);
                    // var pwidth = fontMetrics.advanceWidth(ptext);
                    // ctx.fillRect((x * widthMul) | 0, 0, (pwidth * widthMul) |0, lineHeight);
                    ctx.fillRect(x, 0, j - i, lineHeight);


                    if (chunkDebug) console.log("(" + ptext + ")");

                    x += j - i;
                    i = j;
                  }

                  // if (chunkDebug) chunks += cwidth + " ";
                  // ctx.fillText(ctext, x, y);
                  // ctx.fillRect((x * widthMul) | 0, 0, (cwidth * widthMul) |0, lineHeight);
                  // if (chunkDebug) console.log("(", c.text, ")", (x * widthMul) | 0, 0, (cwidth * widthMul) |0, lineHeight);
                  // console.log(x, 0, x + cwidth, height);
                  // x += cwidth;
                }

                // if (chunkDebug) console.log("Paint: ", chunks);
              }
            }


    }


    property var oldView

    function scroll() {
        var p = percentage(mainListView);
        // console.log("scroll:", p);
        // children[1].contentY = p*(children[1].contentHeight-height);
        if (!ma.drag.active) {
            minimapArea.y =  p*(Math.min(height, listView.contentHeight)-minimapArea.height)
        }
    }
    onMainListViewChanged: {
      // console.log("mainListViewChanged", mainListView);
      if (oldView && oldView.contentYChanged) {
          oldView.contentYChanged.disconnect(scroll);
      }
      if (mainListView && mainListView.contentYChanged) {
        mainListView.contentYChanged.connect(scroll);
        // console.log("connected");
      }
      oldView = mainListView;
      scroll();
    }
    function percentage(view) {
      if (view === undefined) { return 10; }
      if (!view.visibleArea) return 10;
      // console.log("visibleArea: ", view.visibleArea.yPosition, view.visibleArea.heightRatio);
      return view.visibleArea.yPosition/(1-view.visibleArea.heightRatio);
    }


    Rectangle {
        id: minimapArea
        width: parent.width
        height: (mainListView && mainListView.visibleArea) ? mainListView.visibleArea.heightRatio*listView.contentHeight : parent.height
        color: "white"
        opacity: 0.1
        onYChanged: {
            if (ma.drag.active) {
                mainListView.contentY = y*(mainListView.contentHeight-mainListView.height)/ma.drag.maximumY;
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
            drag.maximumY: Math.min(parent.parent.height, listView.contentHeight)-height
            drag.maximumX: parent.parent.width-width
        }
    }

}
