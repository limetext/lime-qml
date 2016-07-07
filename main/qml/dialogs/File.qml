import QtQuick 2.0
import QtQuick.Dialogs 1.0

FileDialog {
    title: title
    selectMultiple: true
    folder: {
        if (tabs.count == 0) return shortcuts.home;

        // show active view folder
        var fn = myWindow.view(tabs.currentIndex).title.text;
        var sp = (Qt.platform.os == "windows") ? "\\" : "/";
        return fn.substring(0, fn.lastIndexOf(sp)+1);
    }
}
