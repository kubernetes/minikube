/****************************************************************************
**
** Copyright 2022 The Kubernetes Authors All rights reserved.
**
** Copyright (C) 2021 Anders F Björklund
** Copyright (C) 2016 The Qt Company Ltd.
** Contact: https://www.qt.io/licensing/
**
** This file is part of the examples of the Qt Toolkit.
**
** $QT_BEGIN_LICENSE:BSD$
** Commercial License Usage
** Licensees holding valid commercial Qt licenses may use this file in
** accordance with the commercial license agreement provided with the
** Software or, alternatively, in accordance with the terms contained in
** a written agreement between you and The Qt Company. For licensing terms
** and conditions see https://www.qt.io/terms-conditions. For further
** information use the contact form at https://www.qt.io/contact-us.
**
** BSD License Usage
** Alternatively, you may use this file under the terms of the BSD license
** as follows:
**
** "Redistribution and use in source and binary forms, with or without
** modification, are permitted provided that the following conditions are
** met:
**   * Redistributions of source code must retain the above copyright
**     notice, this list of conditions and the following disclaimer.
**   * Redistributions in binary form must reproduce the above copyright
**     notice, this list of conditions and the following disclaimer in
**     the documentation and/or other materials provided with the
**     distribution.
**   * Neither the name of The Qt Company Ltd nor the names of its
**     contributors may be used to endorse or promote products derived
**     from this software without specific prior written permission.
**
**
** THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
** "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
** LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
** A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
** OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
** SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
** LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
** DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
** THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
** (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
** OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE."
**
** $QT_END_LICENSE$
**
****************************************************************************/

#ifndef WINDOW_H
#define WINDOW_H

#include <QSystemTrayIcon>
#include <QFormLayout>
#include <QStackedWidget>
#include <QProcessEnvironment>
#include <QVersionNumber>
#include <QtNetwork/QNetworkAccessManager>
#include <QtNetwork/QNetworkReply>

#ifndef QT_NO_SYSTEMTRAYICON

#include <QDialog>

QT_BEGIN_NAMESPACE
class QAction;
class QCheckBox;
class QComboBox;
class QGroupBox;
class QIcon;
class QLabel;
class QLineEdit;
class QMenu;
class QPushButton;
class QSpinBox;
class QTextEdit;
class QTableView;
class QProcess;
QT_END_NAMESPACE

#include "cluster.h"

class Window : public QDialog
{
    Q_OBJECT

public:
    Window();

    void setVisible(bool visible) override;

protected:
    void closeEvent(QCloseEvent *event) override;

private slots:
    void messageClicked();
    void updateButtons();
    void dashboardClose();

private:
    // Tray icon
    void createTrayIcon();
    void createActions();
    void updateStatus(Cluster cluster);
    void updateTrayActions(Cluster cluster);
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
    QIcon *trayIconIcon;

    // Basic view
    void createBasicView();
    void toBasicView();
    void updateBasicButtons(Cluster cluster);
    QPushButton *basicStartButton;
    QPushButton *basicStopButton;
    QPushButton *basicPauseButton;
    QPushButton *basicDeleteButton;
    QPushButton *basicRefreshButton;
    QPushButton *basicSSHButton;
    QPushButton *basicDashboardButton;

    // Advanced view
    void createAdvancedView();
    void toAdvancedView();
    void createClusterGroupBox();
    void updateAdvancedButtons(Cluster cluster);
    QPushButton *startButton;
    QPushButton *stopButton;
    QPushButton *pauseButton;
    QPushButton *deleteButton;
    QPushButton *refreshButton;
    QPushButton *createButton;
    QPushButton *sshButton;
    QPushButton *dashboardButton;
    QGroupBox *clusterGroupBox;

    // Cluster table
    QString selectedClusterName();
    void setSelectedClusterName(QString cluster);
    Cluster selectedCluster();
    void updateClusterList();
    void updateClustersTable();
    void showLoading();
    void hideLoading();
    ClusterModel *clusterModel;
    QTableView *clusterListView;
    ClusterList clusterList;
    QLabel *loading;

    // Create cluster
    void askCustom();
    void askName();
    QComboBox *driverComboBox;
    QComboBox *containerRuntimeComboBox;
    QComboBox *k8sVersionComboBox;

    // Commands
    void startMinikube(QStringList args);
    void startSelectedMinikube();
    void stopMinikube();
    void pauseMinikube();
    void unpauseMinikube();
    void pauseOrUnpauseMinikube();
    void deleteMinikube();
    bool sendMinikubeCommand(QStringList cmds);
    bool sendMinikubeCommand(QStringList cmds, QString &text);
    void initMachine();
    void sshConsole();
    void dashboardBrowser();
    Cluster createClusterObject(QJsonObject obj);
    QProcess *dashboardProcess;
    QProcessEnvironment env;
#if __APPLE__
    void hyperkitPermission();
    bool hyperkitPermissionFix(QStringList args, QString text);
    bool showHyperKitMessage();
#endif

    // Error messaging
    void outputFailedStart(QString text);
    QLabel *createLabel(QString title, QString text, QFormLayout *form, bool isLink);

    void checkForMinikube();
    void restoreWindow();
    QString getPauseLabel(bool isPaused);
    QString getStartLabel(bool isRunning);
    QProcessEnvironment setMacEnv();
    QStackedWidget *stackedWidget;
    bool isBasicView;
    void delay();
    int getCenter(int widgetSize, int parentSize);
    void checkForUpdates();
    QString getRequest(QString url);
    void notifyUpdate(QString latest, QString link);
};

#endif // QT_NO_SYSTEMTRAYICON

#endif
