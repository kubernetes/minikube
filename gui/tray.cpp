#include "tray.h"

#include <QAction>
#include <QCoreApplication>
#include <QMenu>

Tray::Tray(QIcon icon)
{
    m_icon = icon;

    trayIconMenu = new QMenu();
    trayIcon = new QSystemTrayIcon(this);

    connect(trayIcon, &QSystemTrayIcon::activated, this, &Tray::iconActivated);

    minimizeAction = new QAction(tr("Mi&nimize"), this);
    connect(minimizeAction, &QAction::triggered, this, &Tray::hideWindow);

    restoreAction = new QAction(tr("&Restore"), this);
    connect(restoreAction, &QAction::triggered, this, &Tray::restoreWindow);

    quitAction = new QAction(tr("&Quit"), this);
    connect(quitAction, &QAction::triggered, qApp, &QCoreApplication::quit);

    startAction = new QAction(tr("Start"), this);
    connect(startAction, &QAction::triggered, this, &Tray::start);

    pauseAction = new QAction(tr("Pause"), this);
    connect(pauseAction, &QAction::triggered, this, &Tray::pauseOrUnpause);

    stopAction = new QAction(tr("Stop"), this);
    connect(stopAction, &QAction::triggered, this, &Tray::stop);

    statusAction = new QAction(tr("Status:"), this);
    statusAction->setEnabled(false);

    trayIconMenu->addAction(statusAction);
    trayIconMenu->addSeparator();
    trayIconMenu->addAction(startAction);
    trayIconMenu->addAction(pauseAction);
    trayIconMenu->addAction(stopAction);
    trayIconMenu->addSeparator();
    trayIconMenu->addAction(minimizeAction);
    trayIconMenu->addAction(restoreAction);
    trayIconMenu->addSeparator();
    trayIconMenu->addAction(quitAction);

    trayIcon->setContextMenu(trayIconMenu);
    trayIcon->setIcon(m_icon);
    trayIcon->show();
}

void Tray::iconActivated(QSystemTrayIcon::ActivationReason reason)
{
    switch (reason) {
    case QSystemTrayIcon::Trigger:
    case QSystemTrayIcon::DoubleClick:
        emit restoreWindow();
        break;
    default:;
    }
}

void Tray::updateStatus(Cluster cluster)
{
    QString status = cluster.status();
    if (status.isEmpty()) {
        status = "Stopped";
    }
    statusAction->setText("Status: " + status);
}

bool Tray::isVisible()
{
    return trayIcon->isVisible();
}

void Tray::setVisible(bool visible)
{
    minimizeAction->setEnabled(visible);
    restoreAction->setEnabled(!visible);
}

static QString getPauseLabel(bool isPaused)
{
    if (isPaused) {
        return "Unpause";
    }
    return "Pause";
}

static QString getStartLabel(bool isRunning)
{
    if (isRunning) {
        return "Restart";
    }
    return "Start";
}

void Tray::updateTrayActions(Cluster cluster)
{
    startAction->setEnabled(true);
    bool isRunning = cluster.status() == "Running";
    bool isPaused = cluster.status() == "Paused";
    pauseAction->setEnabled(isRunning || isPaused);
    stopAction->setEnabled(isRunning || isPaused);
    pauseAction->setText(getPauseLabel(isPaused));
    startAction->setText(getStartLabel(isRunning));
}

void Tray::disableActions()
{
    startAction->setEnabled(false);
    stopAction->setEnabled(false);
    pauseAction->setEnabled(false);
}
