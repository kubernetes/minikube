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

#ifndef QT_NO_TERMWIDGET
#include <QApplication>
#include <QMainWindow>
#include "qtermwidget.h"
#endif

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
    resize(200, 250);

    setWindowTitle(tr("minikube"));
    setWindowIcon(*trayIconIcon);
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
    basicStartButton = new QPushButton(tr("Start"));
    basicStopButton = new QPushButton(tr("Stop"));
    basicPauseButton = new QPushButton(tr("Pause"));
    basicDeleteButton = new QPushButton(tr("Delete"));
    basicRefreshButton = new QPushButton(tr("Refresh"));
    basicSSHButton = new QPushButton(tr("SSH"));
    basicDashboardButton = new QPushButton(tr("Dashboard"));
    QPushButton *advancedViewButton = new QPushButton(tr("Advanced View"));

    QVBoxLayout *buttonLayout = new QVBoxLayout;
    QGroupBox *catBox = new QGroupBox();
    catBox->setLayout(buttonLayout);
    buttonLayout->addWidget(basicStartButton);
    buttonLayout->addWidget(basicStopButton);
    buttonLayout->addWidget(basicPauseButton);
    buttonLayout->addWidget(basicDeleteButton);
    buttonLayout->addWidget(basicRefreshButton);
    buttonLayout->addWidget(basicSSHButton);
    buttonLayout->addWidget(basicDashboardButton);
    buttonLayout->addWidget(advancedViewButton);
    catBox->setSizePolicy(QSizePolicy::Ignored, QSizePolicy::Ignored);
    stackedWidget->addWidget(catBox);

    connect(basicSSHButton, &QAbstractButton::clicked, this, &Window::sshConsole);
    connect(basicDashboardButton, &QAbstractButton::clicked, this, &Window::dashboardBrowser);
    connect(basicStartButton, &QAbstractButton::clicked, this, &Window::startSelectedMinikube);
    connect(basicStopButton, &QAbstractButton::clicked, this, &Window::stopMinikube);
    connect(basicPauseButton, &QAbstractButton::clicked, this, &Window::pauseOrUnpauseMinikube);
    connect(basicDeleteButton, &QAbstractButton::clicked, this, &Window::deleteMinikube);
    connect(basicRefreshButton, &QAbstractButton::clicked, this, &Window::updateClusters);
    connect(advancedViewButton, &QAbstractButton::clicked, this, &Window::toAdvancedView);
}

void Window::toAdvancedView()
{
    isBasicView = false;
    stackedWidget->setCurrentIndex(1);
    resize(600, 400);
}

void Window::toBasicView()
{
    isBasicView = true;
    stackedWidget->setCurrentIndex(0);
    resize(200, 250);
}

void Window::createAdvancedView()
{
    connect(sshButton, &QAbstractButton::clicked, this, &Window::sshConsole);
    connect(dashboardButton, &QAbstractButton::clicked, this, &Window::dashboardBrowser);
    connect(startButton, &QAbstractButton::clicked, this, &Window::startSelectedMinikube);
    connect(stopButton, &QAbstractButton::clicked, this, &Window::stopMinikube);
    connect(pauseButton, &QAbstractButton::clicked, this, &Window::pauseOrUnpauseMinikube);
    connect(deleteButton, &QAbstractButton::clicked, this, &Window::deleteMinikube);
    connect(refreshButton, &QAbstractButton::clicked, this, &Window::updateClusters);
    connect(createButton, &QAbstractButton::clicked, this, &Window::initMachine);
    connect(trayIcon, &QSystemTrayIcon::messageClicked, this, &Window::messageClicked);

    clusterGroupBox->setSizePolicy(QSizePolicy::Ignored, QSizePolicy::Ignored);
    stackedWidget->addWidget(clusterGroupBox);
}

void Window::setVisible(bool visible)
{
    minimizeAction->setEnabled(visible);
    restoreAction->setEnabled(!visible);
    QDialog::setVisible(visible);
}

void Window::closeEvent(QCloseEvent *event)
{
#ifdef Q_OS_OSX
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

    quitAction = new QAction(tr("&Quit"), this);
    connect(quitAction, &QAction::triggered, qApp, &QCoreApplication::quit);
}

void Window::restoreWindow()
{
    QWidget::showNormal();
    updateClusters();
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
    trayIconMenu->addAction(minimizeAction);
    trayIconMenu->addAction(restoreAction);
    trayIconMenu->addSeparator();
    trayIconMenu->addAction(quitAction);

    trayIcon = new QSystemTrayIcon(this);
    trayIcon->setContextMenu(trayIconMenu);
    trayIcon->setIcon(*trayIconIcon);
}

void Window::startMinikube(QStringList moreArgs)
{
    QString text;
    QStringList args = { "start", "-o", "json" };
    args << moreArgs;
    bool success = sendMinikubeCommand(args, text);
    updateClusters();
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
    updateClusters();
}

void Window::pauseMinikube()
{
    QStringList args = { "pause", "-p", selectedClusterName() };
    sendMinikubeCommand(args);
    updateClusters();
}

void Window::unpauseMinikube()
{
    QStringList args = { "unpause", "-p", selectedClusterName() };
    sendMinikubeCommand(args);
    updateClusters();
}

void Window::deleteMinikube()
{
    QStringList args = { "delete", "-p", selectedClusterName() };
    sendMinikubeCommand(args);
    updateClusters();
}

void Window::updateClusters()
{
    QString cluster = selectedClusterName();
    clusterModel->setClusters(getClusters());
    setSelectedClusterName(cluster);
    updateButtons();
}

ClusterList Window::getClusters()
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
        QJsonArray a = par["valid"].toArray();
        QJsonArray b = par["invalid"].toArray();
        for (int i = 0; i < b.size(); i++) {
            a.append(b[i]);
        }
        for (int i = 0; i < a.size(); i++) {
            QJsonObject obj = a[i].toObject();
            Cluster cluster = createClusterObject(obj);
            clusters << cluster;
        }
    }
    return clusters;
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
    return cluster;
}

QString Window::selectedClusterName()
{
    if (isBasicView) {
        return "minikube";
    }
    QModelIndex index = clusterListView->currentIndex();
    QVariant variant = index.data(Qt::DisplayRole);
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
    clusterGroupBox = new QGroupBox(tr("Clusters"));

    ClusterList clusters = getClusters();
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
    clusterGroupBox->setLayout(clusterLayout);
}

void Window::updateButtons()
{
    if (isBasicView) {
        updateBasicButtons();
    } else {
        updateAdvancedButtons();
    }
}

Cluster Window::selectedCluster()
{
    QString clusterName = selectedClusterName();
    if (clusterName.isEmpty()) {
        return Cluster();
    }
    ClusterList clusters = getClusters();
    ClusterHash clusterHash;
    for (int i = 0; i < clusters.size(); i++) {
        Cluster cluster = clusters.at(i);
        clusterHash[cluster.name()] = cluster;
    }
    return clusterHash[clusterName];
}

void Window::updateBasicButtons()
{
    Cluster cluster = selectedCluster();
    bool exists = !cluster.isEmpty();
    bool isRunning = cluster.status() == "Running";
    bool isPaused = cluster.status() == "Paused";
    basicStopButton->setEnabled(isRunning || isPaused);
    basicPauseButton->setEnabled(isRunning || isPaused);
    basicDeleteButton->setEnabled(exists);
    basicDashboardButton->setEnabled(isRunning);
#if __linux__
    basicSSHButton->setEnabled(exists);
#else
    basicSSHButton->setEnabled(false);
#endif
    QString pauseLabel = tr("Pause");
    if (isPaused) {
        pauseLabel = tr("Unpause");
    }
    basicPauseButton->setText(pauseLabel);
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

void Window::updateAdvancedButtons()
{
    Cluster cluster = selectedCluster();
    bool exists = !cluster.isEmpty();
    bool isRunning = cluster.status() == "Running";
    bool isPaused = cluster.status() == "Paused";
    startButton->setEnabled(exists);
    stopButton->setEnabled(isRunning || isPaused);
    pauseButton->setEnabled(isRunning || isPaused);
    deleteButton->setEnabled(exists);
    dashboardButton->setEnabled(isRunning);
#if __linux__
    sshButton->setEnabled(exists);
#else
    sshButton->setEnabled(false);
#endif
    QString pauseLabel = tr("Pause");
    if (isPaused) {
        pauseLabel = tr("Unpause");
    }
    pauseButton->setText(pauseLabel);
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
    return success;
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
        QLabel *errorMessage = createLabel("Error Message", message, &form, false);
        errorMessage->setFont(QFont("Courier", 10));
        errorMessage->setStyleSheet("background-color:white;");
        createLabel("Link to documentation", url, &form, true);
        createLabel("Link to related issue", issues, &form, true);
        // Enabling once https://github.com/kubernetes/minikube/issues/13925 is fixed
        // QLabel *fileLabel = new QLabel(this);
        // fileLabel->setOpenExternalLinks(true);
        // fileLabel->setWordWrap(true);
        // QString logFile = QDir::homePath() + "/.minikube/logs/lastStart.txt";
        // fileLabel->setText("<a href='file:///" + logFile + "'>View log file</a>");
        // form.addRow(fileLabel);
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
    updateClusters();
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

#endif
