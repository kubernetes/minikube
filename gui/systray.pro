HEADERS       = window.h \
    instance.h
SOURCES       = main.cpp \
                instance.cpp \
                window.cpp
RESOURCES     = systray.qrc

QT += widgets
requires(qtConfig(combobox))
