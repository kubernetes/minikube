HEADERS       = window.h \
                cluster.h
SOURCES       = main.cpp \
                cluster.cpp \
                window.cpp
RESOURCES     = systray.qrc
ICON          = images/minikube.icns

QT += widgets network
requires(qtConfig(combobox))

DISTFILES += \
    LICENSE

# Enabling qtermwidget requires GPL-v2 license
#CONFIG += gpl_licensed

gpl_licensed {
  win32: DEFINES += QT_NO_TERMWIDGET

  unix: CONFIG += link_pkgconfig
  unix: PKGCONFIG += qtermwidget5
} else {
  DEFINES += QT_NO_TERMWIDGET
}
