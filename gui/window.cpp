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

#include "window.h"

#ifndef QT_NO_SYSTEMTRAYICON

#include <QAction>
#include <QCheckBox>
#include <QComboBox>
#include <QCoreApplication>
#include <QCloseEvent>
#include <QGroupBox>
#include <QLabel>
#include <QLineEdit>
#include <QMenu>
#include <QPushButton>
#include <QSpinBox>
#include <QTextEdit>
#include <QVBoxLayout>
#include <QMessageBox>
#include <QProcess>
#include <QDebug>
#include <QJsonDocument>
#include <QJsonObject>
#include <QJsonArray>
#include <QTableView>
#include <QHeaderView>
#include <QFormLayout>
#include <QDialogButtonBox>
#include <QStandardPaths>
#include <QDir>
#include <QFontDialog>
#include <QStackedWidget>
#include <QProcessEnvironment>
#include <QNetworkAccessManager>

#ifndef QT_NO_TERMWIDGET
#include <QApplication>
#include <QMainWindow>
#include "qtermwidget.h"
#endif

const QVersionNumber version = QVersionNumber::fromString("0.0.1");

Window::Window()
{
    trayIconIcon = new QIcon(":/images/minikube.png");
    checkForMinikube();
    isBasicView = true;

    stackedWidget = new QStackedWidget;
    QVBoxLayout *layout = new QVBoxLayout;
    dashboardProcess = 0;

    createClusterGroupBox();

    createActions();
    createTrayIcon();

    createBasicView();
    createAdvancedView();
    trayIcon->show();
    updateButtons();
    layout->addWidget(stackedWidget);
    setLayout(layout);
    resize(200, 275);

    setWindowTitle(tr("minikube"));
    setWindowIcon(*trayIconIcon);
    setupUpdateChecking();
}

void Window::setupUpdateChecking()
{
    checkForUpdates();
    connect(this, &Window::updateCheck, this, &Window::checkForUpdates);
}

QProcessEnvironment Window::setMacEnv()
{
    QProcessEnvironment env = QProcessEnvironment::systemEnvironment();
    QString path = env.value("PATH");
    env.insert("PATH", path + ":/usr/local/bin");
    return env;
}

void Window::createBasicView()
{
    QWidget *basicView = new QWidget();

    basicStartButton = new QPushButton(tr("Start"));
    basicStopButton = new QPushButton(tr("Stop"));
    basicPauseButton = new QPushButton(tr("Pause"));
    basicDeleteButton = new QPushButton(tr("Delete"));
    basicRefreshButton = new QPushButton(tr("Refresh"));
    basicSSHButton = new QPushButton(tr("SSH"));
    basicDashboardButton = new QPushButton(tr("Dashboard"));
    QPushButton *advancedViewButton = new QPushButton(tr("Advanced View"));

    QVBoxLayout *buttonLayout = new QVBoxLayout;
    basicView->setLayout(buttonLayout);
    buttonLayout->addWidget(basicStartButton);
    buttonLayout->addWidget(basicStopButton);
    buttonLayout->addWidget(basicPauseButton);
    buttonLayout->addWidget(basicDeleteButton);
    buttonLayout->addWidget(basicRefreshButton);
    buttonLayout->addWidget(basicSSHButton);
    buttonLayout->addWidget(basicDashboardButton);
    buttonLayout->addWidget(advancedViewButton);
    basicView->setSizePolicy(QSizePolicy::Ignored, QSizePolicy::Ignored);
    stackedWidget->addWidget(basicView);

    connect(basicSSHButton, &QAbstractButton::clicked, this, &Window::sshConsole);
    connect(basicDashboardButton, &QAbstractButton::clicked, this, &Window::dashboardBrowser);
    connect(basicStartButton, &QAbstractButton::clicked, this, &Window::startSelectedMinikube);
    connect(basicStopButton, &QAbstractButton::clicked, this, &Window::stopMinikube);
    connect(basicPauseButton, &QAbstractButton::clicked, this, &Window::pauseOrUnpauseMinikube);
    connect(basicDeleteButton, &QAbstractButton::clicked, this, &Window::deleteMinikube);
    connect(basicRefreshButton, &QAbstractButton::clicked, this, &Window::updateClustersTable);
    connect(advancedViewButton, &QAbstractButton::clicked, this, &Window::toAdvancedView);
}

void Window::toAdvancedView()
{
    isBasicView = false;
    stackedWidget->setCurrentIndex(1);
    resize(670, 400);
    updateButtons();
}

void Window::toBasicView()
{
    isBasicView = true;
    stackedWidget->setCurrentIndex(0);
    resize(200, 275);
    updateButtons();
}

void Window::createAdvancedView()
{
    connect(sshButton, &QAbstractButton::clicked, this, &Window::sshConsole);
    connect(dashboardButton, &QAbstractButton::clicked, this, &Window::dashboardBrowser);
    connect(startButton, &QAbstractButton::clicked, this, &Window::startSelectedMinikube);
    connect(stopButton, &QAbstractButton::clicked, this, &Window::stopMinikube);
    connect(pauseButton, &QAbstractButton::clicked, this, &Window::pauseOrUnpauseMinikube);
    connect(deleteButton, &QAbstractButton::clicked, this, &Window::deleteMinikube);
    connect(refreshButton, &QAbstractButton::clicked, this, &Window::updateClustersTable);
    connect(createButton, &QAbstractButton::clicked, this, &Window::initMachine);

    advancedView->setSizePolicy(QSizePolicy::Ignored, QSizePolicy::Ignored);
    stackedWidget->addWidget(advancedView);
}

void Window::setVisible(bool visible)
{
    minimizeAction->setEnabled(visible);
    restoreAction->setEnabled(!visible);
    QDialog::setVisible(visible);
}

void Window::closeEvent(QCloseEvent *event)
{
#if __APPLE__
    if (!event->spontaneous() || !isVisible()) {
        return;
    }
#endif
    if (trayIcon->isVisible()) {
        QMessageBox::information(this, tr("Systray"),
                                 tr("The program will keep running in the "
                                    "system tray. To terminate the program, "
                                    "choose <b>Quit</b> in the context menu "
                                    "of the system tray entry."));
        hide();
        event->ignore();
    }
}

void Window::messageClicked()
{
    QMessageBox::information(0, tr("Systray"),
                             tr("Sorry, I already gave what help I could.\n"
                                "Maybe you should try asking a human?"));
}

void Window::createActions()
{
    minimizeAction = new QAction(tr("Mi&nimize"), this);
    connect(minimizeAction, &QAction::triggered, this, &QWidget::hide);

    restoreAction = new QAction(tr("&Restore"), this);
    connect(restoreAction, &QAction::triggered, this, &Window::restoreWindow);
    connect(restoreAction, &QAction::triggered, this, &Window::updateCheck);

    quitAction = new QAction(tr("&Quit"), this);
    connect(quitAction, &QAction::triggered, qApp, &QCoreApplication::quit);

    startAction = new QAction(tr("Start"), this);
    connect(startAction, &QAction::triggered, this, &Window::startSelectedMinikube);

    pauseAction = new QAction(tr("Pause"), this);
    connect(pauseAction, &QAction::triggered, this, &Window::pauseOrUnpauseMinikube);

    stopAction = new QAction(tr("Stop"), this);
    connect(stopAction, &QAction::triggered, this, &Window::stopMinikube);

    statusAction = new QAction(tr("Status:"), this);
    statusAction->setEnabled(false);
}

void Window::updateStatus(Cluster cluster)
{
    QString status = cluster.status();
    if (status.isEmpty()) {
        status = "Stopped";
    }
    statusAction->setText("Status: " + status);
}

void Window::iconActivated(QSystemTrayIcon::ActivationReason reason)
{
    switch (reason) {
    case QSystemTrayIcon::Trigger:
    case QSystemTrayIcon::DoubleClick:
        Window::restoreWindow();
        break;
    default:;
    }
}

void Window::restoreWindow()
{
    bool wasVisible = isVisible();
    QWidget::showNormal();
    activateWindow();
    if (wasVisible) {
        return;
    }
    // without this delay window doesn't render until updateClusters() completes
    delay();
    updateClustersTable();
}

void Window::delay()
{
    QCoreApplication::processEvents(QEventLoop::AllEvents, 100);
}

static QString minikubePath()
{
    QString program = QStandardPaths::findExecutable("minikube");
    if (program.isEmpty()) {
        QStringList paths = { "/usr/local/bin" };
        program = QStandardPaths::findExecutable("minikube", paths);
    }
    return program;
}

void Window::createTrayIcon()
{
    trayIconMenu = new QMenu(this);
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

    trayIcon = new QSystemTrayIcon(this);
    trayIcon->setContextMenu(trayIconMenu);
    trayIcon->setIcon(*trayIconIcon);

    connect(trayIcon, &QSystemTrayIcon::activated, this, &Window::iconActivated);
}

void Window::startMinikube(QStringList moreArgs)
{
    QString text;
    QStringList args = { "start", "-o", "json" };
    args << moreArgs;
    bool success = sendMinikubeStart(args, text);
#if __APPLE__
    hyperkitPermissionFix(args, text);
#endif
    updateClustersTable();
    if (success) {
        return;
    }
    outputFailedStart(text);
}

void Window::startSelectedMinikube()
{
    QStringList args = { "-p", selectedClusterName() };
    return startMinikube(args);
}

void Window::stopMinikube()
{
    QStringList args = { "stop", "-p", selectedClusterName() };
    sendMinikubeCommand(args);
    updateClustersTable();
}

void Window::pauseMinikube()
{
    QStringList args = { "pause", "-p", selectedClusterName() };
    sendMinikubeCommand(args);
    updateClustersTable();
}

void Window::unpauseMinikube()
{
    QStringList args = { "unpause", "-p", selectedClusterName() };
    sendMinikubeCommand(args);
    updateClustersTable();
}

void Window::deleteMinikube()
{
    QStringList args = { "delete", "-p", selectedClusterName() };
    sendMinikubeCommand(args);
    updateClustersTable();
}

void Window::updateClustersTable()
{
    showLoading();
    QString cluster = selectedClusterName();
    updateClusterList();
    clusterModel->setClusters(clusterList);
    setSelectedClusterName(cluster);
    updateButtons();
    loading->setHidden(true);
    clusterListView->setEnabled(true);
    hideLoading();
}

void Window::showLoading()
{
    clusterListView->setEnabled(false);
    loading->setHidden(false);
    loading->raise();
    int width = getCenter(loading->width(), clusterListView->width());
    int height = getCenter(loading->height(), clusterListView->height());
    loading->move(width, height);
    delay();
}

void Window::hideLoading()
{
    loading->setHidden(true);
    clusterListView->setEnabled(true);
}

int Window::getCenter(int widgetSize, int parentSize)
{
    return parentSize / 2 - widgetSize / 2;
}

void Window::updateClusterList()
{
    ClusterList clusters;
    QStringList args = { "profile", "list", "-o", "json" };
    QString text;
    sendMinikubeCommand(args, text);
    QStringList lines;
#if QT_VERSION >= QT_VERSION_CHECK(5, 14, 0)
    lines = text.split("\n", Qt::SkipEmptyParts);
#else
    lines = text.split("\n", QString::SkipEmptyParts);
#endif
    for (int i = 0; i < lines.size(); i++) {
        QString line = lines.at(i);
        QJsonParseError error;
        QJsonDocument json = QJsonDocument::fromJson(line.toUtf8(), &error);
        if (json.isNull()) {
            qDebug() << error.errorString();
            continue;
        }
        if (!json.isObject()) {
            continue;
        }
        QJsonObject par = json.object();
        QJsonArray valid = par["valid"].toArray();
        QJsonArray invalid = par["invalid"].toArray();
        for (int i = 0; i < valid.size(); i++) {
            QJsonObject obj = valid[i].toObject();
            Cluster cluster = createClusterObject(obj);
            clusters << cluster;
        }
        for (int i = 0; i < invalid.size(); i++) {
            QJsonObject obj = invalid[i].toObject();
            Cluster cluster = createClusterObject(obj);
            cluster.setStatus("Invalid");
            clusters << cluster;
        }
    }
    clusterList = clusters;
}

Cluster Window::createClusterObject(QJsonObject obj)
{
    QString name;
    if (obj.contains("Name")) {
        name = obj["Name"].toString();
    }
    Cluster cluster(name);
    if (obj.contains("Status")) {
        QString status = obj["Status"].toString();
        cluster.setStatus(status);
    }
    if (!obj.contains("Config")) {
        return cluster;
    }
    QJsonObject config = obj["Config"].toObject();
    if (config.contains("CPUs")) {
        int cpus = config["CPUs"].toInt();
        cluster.setCpus(cpus);
    }
    if (config.contains("Memory")) {
        int memory = config["Memory"].toInt();
        cluster.setMemory(memory);
    }
    if (config.contains("Driver")) {
        QString driver = config["Driver"].toString();
        cluster.setDriver(driver);
    }
    if (!config.contains("KubernetesConfig")) {
        return cluster;
    }
    QJsonObject k8sConfig = config["KubernetesConfig"].toObject();
    if (k8sConfig.contains("ContainerRuntime")) {
        QString containerRuntime = k8sConfig["ContainerRuntime"].toString();
        cluster.setContainerRuntime(containerRuntime);
    }
    if (k8sConfig.contains("KubernetesVersion")) {
        QString k8sVersion = k8sConfig["KubernetesVersion"].toString();
        cluster.setK8sVersion(k8sVersion);
    }
    return cluster;
}

QString Window::selectedClusterName()
{
    if (isBasicView) {
        return "minikube";
    }
    QModelIndex index = clusterListView->currentIndex();
    QVariant variant = index.siblingAtColumn(0).data(Qt::DisplayRole);
    if (variant.isNull()) {
        return QString();
    }
    return variant.toString();
}

void Window::setSelectedClusterName(QString cluster)
{
    QAbstractItemModel *model = clusterListView->model();
    QModelIndex start = model->index(0, 0);
    QModelIndexList index = model->match(start, Qt::DisplayRole, cluster);
    if (index.size() == 0) {
        return;
    }
    clusterListView->setCurrentIndex(index[0]);
}

void Window::createClusterGroupBox()
{
    advancedView = new QWidget();

    updateClusterList();
    ClusterList clusters = clusterList;
    clusterModel = new ClusterModel(clusters);

    clusterListView = new QTableView();
    clusterListView->setModel(clusterModel);
    clusterListView->setSelectionMode(QAbstractItemView::SingleSelection);
    clusterListView->setSelectionBehavior(QAbstractItemView::SelectRows);
    clusterListView->horizontalHeader()->setSectionResizeMode(0, QHeaderView::Stretch);
    clusterListView->horizontalHeader()->setSectionResizeMode(1, QHeaderView::ResizeToContents);
    clusterListView->horizontalHeader()->setSectionResizeMode(2, QHeaderView::ResizeToContents);
    clusterListView->horizontalHeader()->setSectionResizeMode(3, QHeaderView::ResizeToContents);
    clusterListView->horizontalHeader()->setSectionResizeMode(4, QHeaderView::ResizeToContents);
    clusterListView->horizontalHeader()->setSectionResizeMode(5, QHeaderView::ResizeToContents);
    clusterListView->horizontalHeader()->setSectionResizeMode(6, QHeaderView::ResizeToContents);
    setSelectedClusterName("default");

    connect(clusterListView, SIGNAL(clicked(QModelIndex)), this, SLOT(updateButtons()));

    startButton = new QPushButton(tr("Start"));
    stopButton = new QPushButton(tr("Stop"));
    pauseButton = new QPushButton(tr("Pause"));
    deleteButton = new QPushButton(tr("Delete"));
    refreshButton = new QPushButton(tr("Refresh"));
    createButton = new QPushButton(tr("Create"));
    sshButton = new QPushButton(tr("SSH"));
    dashboardButton = new QPushButton(tr("Dashboard"));
    QPushButton *basicViewButton = new QPushButton(tr("Basic View"));
    connect(basicViewButton, &QAbstractButton::clicked, this, &Window::toBasicView);

    QHBoxLayout *topButtonLayout = new QHBoxLayout;
    topButtonLayout->addWidget(createButton);
    topButtonLayout->addWidget(refreshButton);
    topButtonLayout->addWidget(basicViewButton);
    topButtonLayout->addSpacing(340);

    QHBoxLayout *bottomButtonLayout = new QHBoxLayout;
    bottomButtonLayout->addWidget(startButton);
    bottomButtonLayout->addWidget(stopButton);
    bottomButtonLayout->addWidget(pauseButton);
    bottomButtonLayout->addWidget(deleteButton);
    bottomButtonLayout->addWidget(sshButton);
    bottomButtonLayout->addWidget(dashboardButton);

    QVBoxLayout *clusterLayout = new QVBoxLayout;
    clusterLayout->addLayout(topButtonLayout);
    clusterLayout->addWidget(clusterListView);
    clusterLayout->addLayout(bottomButtonLayout);
    advancedView->setLayout(clusterLayout);

    QFont *loadingFont = new QFont();
    loadingFont->setPointSize(30);
    loading = new QLabel("Loading...");
    loading->setFont(*loadingFont);
    loading->setParent(clusterListView);
    loading->setHidden(true);
}

void Window::updateButtons()
{
    Cluster cluster = selectedCluster();
    if (isBasicView) {
        updateBasicButtons(cluster);
    } else {
        updateAdvancedButtons(cluster);
    }
    updateTrayActions(cluster);
    updateStatus(cluster);
}

void Window::updateTrayActions(Cluster cluster)
{
    bool isRunning = cluster.status() == "Running";
    bool isPaused = cluster.status() == "Paused";
    pauseAction->setEnabled(isRunning || isPaused);
    stopAction->setEnabled(isRunning || isPaused);
    pauseAction->setText(getPauseLabel(isPaused));
    startAction->setText(getStartLabel(isRunning));
}

Cluster Window::selectedCluster()
{
    QString clusterName = selectedClusterName();
    if (clusterName.isEmpty()) {
        return Cluster();
    }
    ClusterList clusters = clusterList;
    ClusterHash clusterHash;
    for (int i = 0; i < clusters.size(); i++) {
        Cluster cluster = clusters.at(i);
        clusterHash[cluster.name()] = cluster;
    }
    return clusterHash[clusterName];
}

void Window::updateBasicButtons(Cluster cluster)
{
    bool exists = !cluster.isEmpty();
    bool isRunning = cluster.status() == "Running";
    bool isPaused = cluster.status() == "Paused";
    basicStopButton->setEnabled(isRunning || isPaused);
    basicPauseButton->setEnabled(isRunning || isPaused);
    basicDeleteButton->setEnabled(exists);
    basicDashboardButton->setEnabled(isRunning);
#if __linux__ || __APPLE__
    basicSSHButton->setEnabled(exists);
#else
    basicSSHButton->setEnabled(false);
#endif
    basicPauseButton->setText(getPauseLabel(isPaused));
    basicStartButton->setText(getStartLabel(isRunning));
}

QString Window::getPauseLabel(bool isPaused)
{
    if (isPaused) {
        return tr("Unpause");
    }
    return tr("Pause");
}

QString Window::getStartLabel(bool isRunning)
{
    if (isRunning) {
        return tr("Reload");
    }
    return tr("Start");
}

void Window::pauseOrUnpauseMinikube()
{
    Cluster cluster = selectedCluster();
    if (cluster.status() == "Paused") {
        unpauseMinikube();
        return;
    }
    pauseMinikube();
}

void Window::updateAdvancedButtons(Cluster cluster)
{
    bool exists = !cluster.isEmpty();
    bool isRunning = cluster.status() == "Running";
    bool isPaused = cluster.status() == "Paused";
    startButton->setEnabled(exists);
    stopButton->setEnabled(isRunning || isPaused);
    pauseButton->setEnabled(isRunning || isPaused);
    deleteButton->setEnabled(exists);
    dashboardButton->setEnabled(isRunning);
#if __linux__ || __APPLE__
    sshButton->setEnabled(exists);
#else
    sshButton->setEnabled(false);
#endif
    pauseButton->setText(getPauseLabel(isPaused));
    startButton->setText(getStartLabel(isRunning));
}

bool Window::sendMinikubeCommand(QStringList cmds)
{
    QString text;
    return sendMinikubeCommand(cmds, text);
}

bool Window::sendMinikubeCommand(QStringList cmds, QString &text)
{
    QString program = minikubePath();
    if (program.isEmpty()) {
        return false;
    }
    QStringList arguments = { "--user", "minikube-gui" };
    arguments << cmds;

    QProcess *process = new QProcess(this);
#if __APPLE__
    if (env.isEmpty()) {
        env = setMacEnv();
    }
    process->setProcessEnvironment(env);
#endif
    process->start(program, arguments);
    this->setCursor(Qt::WaitCursor);
    bool timedOut = process->waitForFinished(300 * 1000);
    int exitCode = process->exitCode();
    bool success = !timedOut && exitCode == 0;
    this->unsetCursor();

    text = process->readAllStandardOutput();
    if (success) {
    } else {
        qDebug() << text;
        qDebug() << process->readAllStandardError();
    }
    delete process;
    emit updateCheck();
    return success;
}

bool Window::sendMinikubeStart(QStringList cmds, QString &text)
{
    QString program = minikubePath();
    if (program.isEmpty()) {
        return false;
    }
    QStringList arguments = { "--user", "minikube-gui" };
    arguments << cmds;

    QProcess *process = new QProcess(this);
    connect(process, &QProcess::readyReadStandardOutput,
            [process, this]() { startStep(process->readAllStandardOutput()); });
    startProcess = process;
#if __APPLE__
    if (env.isEmpty()) {
        env = setMacEnv();
    }
    startProcess->setProcessEnvironment(env);
#endif
    this->setCursor(Qt::WaitCursor);
    startProcess->start(program, arguments);
    startProgress();
    while (startProcess->state() != QProcess::NotRunning) {
        delay();
    }
    endProgress();
    int exitCode = startProcess->exitCode();
    bool success = exitCode == 0;
    this->unsetCursor();

    text = startProcess->readAllStandardOutput();
    if (success) {
    } else {
        qDebug() << text;
        qDebug() << startProcess->readAllStandardError();
    }
    delete startProcess;
    return success;
}

void Window::startProgress()
{
    progressDialog = new QDialog(this);
    progressDialog->resize(300, 150);
    progressDialog->setWindowTitle(tr("minikube start Progress"));
    progressDialog->setWindowIcon(*trayIconIcon);
    progressDialog->setWindowFlags(Qt::FramelessWindowHint);
    progressDialog->setModal(true);

    QVBoxLayout form(progressDialog);
    progressText = new QLabel();
    progressText->setText("Starting...");
    progressText->setWordWrap(true);
    form.addWidget(progressText);
    progressBar.setMaximum(19);
    form.addWidget(&progressBar);
    QPushButton *cancel = new QPushButton(tr("Cancel"));
    connect(cancel, &QAbstractButton::clicked, startProcess, &QProcess::kill);
    form.addWidget(cancel);

    progressDialog->open();
}

void Window::endProgress()
{
    progressDialog->hide();
    progressBar.setValue(0);
}

void Window::startStep(QString step)
{
    QStringList lines;
#if QT_VERSION >= QT_VERSION_CHECK(5, 14, 0)
    lines = step.split("\n", Qt::SkipEmptyParts);
#else
    lines = step.split("\n", QString::SkipEmptyParts);
#endif
    for (int i = 0; i < lines.size(); i++) {
        QJsonDocument json = QJsonDocument::fromJson(lines[i].toUtf8());
        QJsonObject object = json.object();
        QString type = object["type"].toString();
        if (type != "io.k8s.sigs.minikube.step") {
            return;
        }
        QJsonObject data = object["data"].toObject();
        QString stringStep = data["currentstep"].toString();
        int currStep = stringStep.toInt();
        QString message = data["message"].toString();
        progressBar.setValue(currStep);
        progressText->setText(message);
    }
}

static QString profile = "minikube";
static int cpus = 2;
static int memory = 2400;
static QString driver = "";
static QString containerRuntime = "";
static QString k8sVersion = "";

void Window::askName()
{
    QDialog dialog;
    dialog.setWindowTitle(tr("Create minikube Cluster"));
    dialog.setWindowIcon(*trayIconIcon);
    dialog.setModal(true);

    QFormLayout form(&dialog);
    QDialogButtonBox buttonBox(Qt::Horizontal, &dialog);
    QLineEdit profileField(profile, &dialog);
    form.addRow(new QLabel(tr("Profile")), &profileField);
    buttonBox.addButton(QString(tr("Use Default Values")), QDialogButtonBox::AcceptRole);
    connect(&buttonBox, &QDialogButtonBox::accepted, &dialog, &QDialog::accept);
    buttonBox.addButton(QString(tr("Set Custom Values")), QDialogButtonBox::RejectRole);
    connect(&buttonBox, &QDialogButtonBox::rejected, &dialog, &QDialog::reject);
    form.addRow(&buttonBox);

    int code = dialog.exec();
    profile = profileField.text();
    if (code == QDialog::Accepted) {
        QStringList args = { "-p", profile };
        startMinikube(args);
    } else if (code == QDialog::Rejected) {
        askCustom();
    }
}

void Window::askCustom()
{
    QDialog dialog;
    dialog.setWindowTitle(tr("Set Cluster Values"));
    dialog.setWindowIcon(*trayIconIcon);
    dialog.setModal(true);

    QFormLayout form(&dialog);
    driverComboBox = new QComboBox;
    driverComboBox->addItems({ "docker", "virtualbox", "vmware", "podman" });
#if __linux__
    driverComboBox->addItem("kvm2");
#elif __APPLE__
    driverComboBox->addItems({ "hyperkit", "parallels" });
#else
    driverComboBox->addItem("hyperv");
#endif
    form.addRow(new QLabel(tr("Driver")), driverComboBox);
    containerRuntimeComboBox = new QComboBox;
    containerRuntimeComboBox->addItems({ "docker", "containerd", "crio" });
    form.addRow(new QLabel(tr("Container Runtime")), containerRuntimeComboBox);
    k8sVersionComboBox = new QComboBox;
    k8sVersionComboBox->addItems({ "stable", "latest", "none" });
    form.addRow(new QLabel(tr("Kubernetes Version")), k8sVersionComboBox);
    QLineEdit cpuField(QString::number(cpus), &dialog);
    form.addRow(new QLabel(tr("CPUs")), &cpuField);
    QLineEdit memoryField(QString::number(memory), &dialog);
    form.addRow(new QLabel(tr("Memory")), &memoryField);

    QDialogButtonBox buttonBox(Qt::Horizontal, &dialog);
    buttonBox.addButton(QString(tr("Create")), QDialogButtonBox::AcceptRole);
    connect(&buttonBox, &QDialogButtonBox::accepted, &dialog, &QDialog::accept);
    buttonBox.addButton(QString(tr("Cancel")), QDialogButtonBox::RejectRole);
    connect(&buttonBox, &QDialogButtonBox::rejected, &dialog, &QDialog::reject);
    form.addRow(&buttonBox);

    int code = dialog.exec();
    if (code == QDialog::Accepted) {
        driver = driverComboBox->itemText(driverComboBox->currentIndex());
        containerRuntime =
                containerRuntimeComboBox->itemText(containerRuntimeComboBox->currentIndex());
        k8sVersion = k8sVersionComboBox->itemText(k8sVersionComboBox->currentIndex());
        if (k8sVersion == "none") {
            k8sVersion = "v0.0.0";
        }
        cpus = cpuField.text().toInt();
        memory = memoryField.text().toInt();
        QStringList args = { "-p",
                             profile,
                             "--driver",
                             driver,
                             "--container-runtime",
                             containerRuntime,
                             "--kubernetes-version",
                             k8sVersion,
                             "--cpus",
                             QString::number(cpus),
                             "--memory",
                             QString::number(memory) };
        startMinikube(args);
    }
}

void Window::outputFailedStart(QString text)
{
    QStringList lines;
#if QT_VERSION >= QT_VERSION_CHECK(5, 14, 0)
    lines = text.split("\n", Qt::SkipEmptyParts);
#else
    lines = text.split("\n", QString::SkipEmptyParts);
#endif
    for (int i = 0; i < lines.size(); i++) {
        QString line = lines.at(i);
        QJsonParseError error;
        QJsonDocument json = QJsonDocument::fromJson(line.toUtf8(), &error);
        if (json.isNull() || !json.isObject()) {
            continue;
        }
        QJsonObject par = json.object();
        QJsonObject data = par["data"].toObject();
        if (!data.contains("exitcode")) {
            continue;
        }
        QString advice = data["advice"].toString();
        QString message = data["message"].toString();
        QString name = data["name"].toString();
        QString url = data["url"].toString();
        QString issues = data["issues"].toString();

        QDialog dialog;
        dialog.setWindowTitle(tr("minikube start failed"));
        dialog.setWindowIcon(*trayIconIcon);
        dialog.setFixedWidth(600);
        dialog.setModal(true);
        QFormLayout form(&dialog);
        createLabel("Error Code", name, &form, false);
        createLabel("Advice", advice, &form, false);
        QTextEdit *errorMessage = new QTextEdit();
        errorMessage->setText(message);
        errorMessage->setWordWrapMode(QTextOption::WrapAnywhere);
        int pointSize = errorMessage->font().pointSize();
        errorMessage->setFont(QFont("Courier", pointSize));
        errorMessage->setAutoFillBackground(true);
        errorMessage->setReadOnly(true);
        form.addRow(errorMessage);
        createLabel("Link to documentation", url, &form, true);
        createLabel("Link to related issue", issues, &form, true);
        QLabel *fileLabel = new QLabel(this);
        fileLabel->setOpenExternalLinks(true);
        fileLabel->setWordWrap(true);
        QString logFile = QDir::homePath() + "/.minikube/logs/lastStart.txt";
        fileLabel->setText("<a href='file:///" + logFile + "'>View log file</a>");
        form.addRow(fileLabel);
        QDialogButtonBox buttonBox(Qt::Horizontal, &dialog);
        buttonBox.addButton(QString(tr("OK")), QDialogButtonBox::AcceptRole);
        connect(&buttonBox, &QDialogButtonBox::accepted, &dialog, &QDialog::accept);
        form.addRow(&buttonBox);
        dialog.exec();
    }
}

QLabel *Window::createLabel(QString title, QString text, QFormLayout *form, bool isLink)
{
    QLabel *label = new QLabel(this);
    if (!text.isEmpty()) {
        form->addRow(label);
    }
    if (isLink) {
        label->setOpenExternalLinks(true);
        text = "<a href='" + text + "'>" + text + "</a>";
    }
    label->setWordWrap(true);
    label->setText(title + ": " + text);
    return label;
}

void Window::initMachine()
{
    askName();
    updateClustersTable();
}

void Window::sshConsole()
{
    QString program = minikubePath();
#ifndef QT_NO_TERMWIDGET
    QMainWindow *mainWindow = new QMainWindow();
    int startnow = 0; // set shell program first

    QTermWidget *console = new QTermWidget(startnow);

    QFont font = QApplication::font();
    font.setFamily("Monospace");
    font.setPointSize(10);

    console->setTerminalFont(font);
    console->setColorScheme("Tango");
    console->setShellProgram(program);
    QStringList args = { "ssh", "-p", selectedClusterName() };
    console->setArgs(args);
    console->startShellProgram();

    QObject::connect(console, SIGNAL(finished()), mainWindow, SLOT(close()));

    mainWindow->setWindowTitle(nameLabel->text());
    mainWindow->resize(800, 400);
    mainWindow->setCentralWidget(console);
    mainWindow->show();
#elif __APPLE__
    QString command = program + " ssh -p " + selectedClusterName();
    QStringList arguments = { "-e", "tell app \"Terminal\"",         "-e", "activate",
                              "-e", "do script \"" + command + "\"", "-e", "end tell" };
    QProcess *process = new QProcess(this);
    process->start("/usr/bin/osascript", arguments);
#else
    QString terminal = qEnvironmentVariable("TERMINAL");
    if (terminal.isEmpty()) {
        terminal = "x-terminal-emulator";
        if (QStandardPaths::findExecutable(terminal).isEmpty()) {
            terminal = "xterm";
        }
    }

    QStringList arguments = { "-e", QString("%1 ssh -p %2").arg(program, selectedClusterName()) };
    QProcess *process = new QProcess(this);
    process->start(QStandardPaths::findExecutable(terminal), arguments);
#endif
}

#if __APPLE__
bool Window::hyperkitPermissionFix(QStringList args, QString text)
{
    if (!text.contains("docker-machine-driver-hyperkit needs to run with elevated permissions")) {
        return false;
    }
    if (!showHyperKitMessage()) {
        return false;
    }

    hyperkitPermission();
    return sendMinikubeCommand(args, text);
}

void Window::hyperkitPermission()
{
    QString command = "sudo chown root:wheel ~/.minikube/bin/docker-machine-driver-hyperkit && "
                      "sudo chmod u+s ~/.minikube/bin/docker-machine-driver-hyperkit && exit";
    QStringList arguments = { "-e", "tell app \"Terminal\"",
                              "-e", "set w to do script \"" + command + "\"",
                              "-e", "activate",
                              "-e", "repeat",
                              "-e", "delay 0.1",
                              "-e", "if not busy of w then exit repeat",
                              "-e", "end repeat",
                              "-e", "end tell" };
    QProcess *process = new QProcess(this);
    process->start("/usr/bin/osascript", arguments);
    process->waitForFinished(-1);
}

bool Window::showHyperKitMessage()
{
    QMessageBox msgBox;
    msgBox.setWindowTitle(tr("HyperKit Permissions Required"));
    msgBox.setWindowIcon(*trayIconIcon);
    msgBox.setModal(true);
    msgBox.setText(tr("The HyperKit driver requires a one-time sudo permission.\n\nIf you'd like "
                      "to proceed, press OK and then enter your password into the terminal prompt, "
                      "the start will resume after."));
    msgBox.setStandardButtons(QMessageBox::Ok | QMessageBox::Cancel);
    msgBox.setDefaultButton(QMessageBox::Ok);
    int code = msgBox.exec();
    return code == QMessageBox::Ok;
}
#endif

void Window::dashboardBrowser()
{
    dashboardClose();

    QString program = minikubePath();
    QProcess *process = new QProcess(this);
    QStringList arguments = { "dashboard", "-p", selectedClusterName() };
    process->start(program, arguments);

    dashboardProcess = process;
    dashboardProcess->waitForStarted();
}

void Window::dashboardClose()
{
    if (dashboardProcess) {
        dashboardProcess->terminate();
        dashboardProcess->waitForFinished();
    }
}

void Window::checkForMinikube()
{
    QString program = minikubePath();
    if (!program.isEmpty()) {
        return;
    }

    QDialog dialog;
    dialog.setWindowTitle(tr("minikube"));
    dialog.setWindowIcon(*trayIconIcon);
    dialog.setModal(true);
    QFormLayout form(&dialog);
    QLabel *message = new QLabel(this);
    message->setText("minikube was not found on the path.\nPlease follow the install instructions "
                     "below to install minikube first.\n");
    form.addWidget(message);
    QLabel *link = new QLabel(this);
    link->setOpenExternalLinks(true);
    link->setText("<a "
                  "href='https://minikube.sigs.k8s.io/docs/start/'>https://minikube.sigs.k8s.io/"
                  "docs/start/</a>");
    form.addWidget(link);
    QDialogButtonBox buttonBox(Qt::Horizontal, &dialog);
    buttonBox.addButton(QString(tr("OK")), QDialogButtonBox::AcceptRole);
    connect(&buttonBox, &QDialogButtonBox::accepted, &dialog, &QDialog::accept);
    form.addRow(&buttonBox);
    dialog.exec();
    exit(EXIT_FAILURE);
}

static bool checkedForUpdateRecently()
{
    QString filePath = QStandardPaths::locate(QStandardPaths::HomeLocation, "/.minikube-gui/last_update_check");
    if (filePath == "") {
        return false;
    }
    QFile file(filePath);
    if (!file.open(QIODevice::ReadOnly)) {
        return false;
    }
    QTextStream in(&file);
    QString line = in.readLine();
    QDateTime nextCheck = QDateTime::fromString(line).addSecs(60*60*24);
    QDateTime now = QDateTime::currentDateTime();
    return nextCheck > now;
}

static void logUpdateCheck()
{
    QDir dir = QDir(QDir::homePath() +  "/.minikube-gui");
    if (!dir.exists()) {
        dir.mkpath(".");
    }
    QString filePath = dir.filePath("last_update_check");
    QFile file(filePath);
    if (!file.open(QIODevice::WriteOnly)) {
        return;
    }
    QTextStream stream(&file);
    stream << QDateTime::currentDateTime().toString() << Qt::endl;
}

void Window::checkForUpdates()
{
    if (checkedForUpdateRecently()) {
        return;
    }
    logUpdateCheck();
    QString releases = getRequest("https://storage.googleapis.com/minikube-gui/releases.json");
    QJsonObject latestRelease =
            QJsonDocument::fromJson(releases.toUtf8()).array().first().toObject();
    QString latestReleaseVersion = latestRelease["name"].toString();
    QVersionNumber latestReleaseVersionNumber = QVersionNumber::fromString(latestReleaseVersion);
    if (version >= latestReleaseVersionNumber) {
        return;
    }
    QJsonObject links = latestRelease["links"].toObject();
    QString key;
#if __linux__
    key = "linux";
#elif __APPLE__
    key = "darwin";
#else
    key = "windows";
#endif
    QString link = links[key].toString();
    notifyUpdate(latestReleaseVersion, link);
}

void Window::notifyUpdate(QString latest, QString link)
{
    QDialog dialog;
    dialog.setWindowTitle(tr("minikube GUI Update Available"));
    dialog.setWindowIcon(*trayIconIcon);
    dialog.setModal(true);
    QFormLayout form(&dialog);
    QLabel *msgLabel = new QLabel(this);
    msgLabel->setText("Version " + latest
                      + " of minikube GUI is now available!\n\nDownload the update from:");
    form.addWidget(msgLabel);
    QLabel *linkLabel = new QLabel(this);
    linkLabel->setOpenExternalLinks(true);
    linkLabel->setText("<a href=\"" + link + "\">" + link + "</a>");
    form.addWidget(linkLabel);
    QDialogButtonBox buttonBox(Qt::Horizontal, &dialog);
    buttonBox.addButton(QString(tr("OK")), QDialogButtonBox::AcceptRole);
    connect(&buttonBox, &QDialogButtonBox::accepted, &dialog, &QDialog::accept);
    form.addRow(&buttonBox);
    dialog.exec();
}

QString Window::getRequest(QString url)
{
    QNetworkAccessManager *manager = new QNetworkAccessManager();
    QObject::connect(manager, &QNetworkAccessManager::finished, this, [=](QNetworkReply *reply) {
        if (reply->error()) {
            qDebug() << reply->errorString();
        }
    });
    QNetworkReply *resp = manager->get(QNetworkRequest(QUrl(url)));
    QEventLoop loop;
    connect(resp, &QNetworkReply::finished, &loop, &QEventLoop::quit);
    loop.exec();
    return resp->readAll();
}

#endif
