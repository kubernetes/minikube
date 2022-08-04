HEADERS       = window.h \
                advancedview.h \
                basicview.h \
                cluster.h \
                commandrunner.h \
                errormessage.h \
                hyperkit.h \
                logger.h \
                operator.h \
                progresswindow.h \
                tray.h \
                updater.h
SOURCES       = main.cpp \
                advancedview.cpp \
                basicview.cpp \
                cluster.cpp \
                commandrunner.cpp \
                errormessage.cpp \
                hyperkit.cpp \
                logger.cpp \
                operator.cpp \
                progresswindow.cpp \
                tray.cpp \
                updater.cpp \
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
