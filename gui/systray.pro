HEADERS       = window.h \
    cluster.h
SOURCES       = main.cpp \
                cluster.cpp \
                window.cpp
RESOURCES     = systray.qrc

QT += widgets
requires(qtConfig(combobox))
