#ifndef TRAY_H
#define TRAY_H

#include "cluster.h"

#include <QAction>
#include <QSystemTrayIcon>

class Tray : public QObject
{
    Q_OBJECT

public:
    explicit Tray(QIcon icon);
    bool isVisible();
    void setVisible(bool visible);
    void updateStatus(Cluster cluster);
    void updateTrayActions(Cluster cluster);
    void disableActions();

signals:
    void restoreWindow();
    void showWindow();
    void hideWindow();
    void start();
    void stop();
    void pauseOrUnpause();

private:
    void createTrayIcon();
    void createActions();
    void iconActivated(QSystemTrayIcon::ActivationReason reason);
    QAction *minimizeAction;
    QAction *restoreAction;
    QAction *quitAction;
    QAction *startAction;
    QAction *pauseAction;
    QAction *stopAction;
    QAction *statusAction;
    QSystemTrayIcon *trayIcon;
    QMenu *trayIconMenu;
    QIcon m_icon;
};

#endif // TRAY_H
