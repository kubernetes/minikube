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

//! [0]
Window::Window()
{
    createInstanceGroupBox();

    createActions();
    createTrayIcon();

    connect(startButton, &QAbstractButton::clicked, this, &Window::startMinikube);
    connect(stopButton, &QAbstractButton::clicked, this, &Window::stopMinikube);
    connect(deleteButton, &QAbstractButton::clicked, this, &Window::deleteMinikube);
    connect(refreshButton, &QAbstractButton::clicked, this, &Window::updateInstances);
    connect(createButton, &QAbstractButton::clicked, this, &Window::initMachine);
    connect(trayIcon, &QSystemTrayIcon::messageClicked, this, &Window::messageClicked);

    QVBoxLayout *mainLayout = new QVBoxLayout;
    mainLayout->addWidget(instanceGroupBox);
    setLayout(mainLayout);

    trayIcon->show();

    setWindowTitle(tr("minikube"));
    setWindowIcon(*trayIconIcon);
    resize(600, 400);
}
//! [0]

//! [1]
void Window::setVisible(bool visible)
{
    minimizeAction->setEnabled(visible);
    restoreAction->setEnabled(!visible);
    QDialog::setVisible(visible);
}
//! [1]

//! [2]
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
//! [2]

//! [6]
void Window::messageClicked()
{
    QMessageBox::information(0, tr("Systray"),
                             tr("Sorry, I already gave what help I could.\n"
                                "Maybe you should try asking a human?"));
}
//! [6]

void Window::createActions()
{
    minimizeAction = new QAction(tr("Mi&nimize"), this);
    connect(minimizeAction, &QAction::triggered, this, &QWidget::hide);

    restoreAction = new QAction(tr("&Restore"), this);
    connect(restoreAction, &QAction::triggered, this, &QWidget::showNormal);

    quitAction = new QAction(tr("&Quit"), this);
    connect(quitAction, &QAction::triggered, qApp, &QCoreApplication::quit);
}

void Window::createTrayIcon()
{
    trayIconMenu = new QMenu(this);
    trayIconMenu->addAction(minimizeAction);
    trayIconMenu->addAction(restoreAction);
    trayIconMenu->addSeparator();
    trayIconMenu->addAction(quitAction);

    trayIconIcon = new QIcon(":/images/minikube.png");

    trayIcon = new QSystemTrayIcon(this);
    trayIcon->setContextMenu(trayIconMenu);
    trayIcon->setIcon(*trayIconIcon);
}

void Window::startMinikube()
{
    QStringList args = {"start", "-p", selectedInstance()};
    sendMinikubeCommand(args);
    updateInstances();
}

void Window::stopMinikube()
{
    QStringList args = {"stop", "-p", selectedInstance()};
    sendMinikubeCommand(args);
    updateInstances();
}

void Window::deleteMinikube()
{
    QStringList args = {"delete", "-p", selectedInstance()};
    sendMinikubeCommand(args);
    updateInstances();
}

void Window::updateInstances()
{
    QString instance = selectedInstance();
    instanceModel->setInstances(getInstances());
    setSelectedInstance(instance);
    updateButtons();
}

InstanceList Window::getInstances()
{
    InstanceList instances;
    QStringList args = {"profile", "list", "-o", "json"};
    QString text;
    bool success = sendMinikubeCommand(args, text);
    QStringList lines;
    if (success) {
#if QT_VERSION >= QT_VERSION_CHECK(5, 14, 0)
        lines = text.split("\n", Qt::SkipEmptyParts);
#else
        lines = text.split("\n", QString::SkipEmptyParts);
#endif
    }
    for (int i = 0; i < lines.size(); i++) {
        QString line = lines.at(i);
        QJsonParseError error;
        QJsonDocument json = QJsonDocument::fromJson(line.toUtf8(), &error);
        if (json.isNull()) {
            qDebug() << error.errorString();
            continue;
        }
        if (json.isObject()) {
            QJsonObject par = json.object();
            QJsonArray a = par["valid"].toArray();
            for (int j = 0; j < a.size(); j++) {
                QJsonObject obj = a[j].toObject();
                QString name;
                if (obj.contains("Name")) {
                    name = obj["Name"].toString();
                }
                if (name.isEmpty()) {
                    continue;
                }
                Instance instance(name);
                if (obj.contains("Status")) {
                    QString status = obj["Status"].toString();
                    instance.setStatus(status);
                }
                if (!obj.contains("Config")) {
                    instances << instance;
                    continue;
                }
                QJsonObject config = obj["Config"].toObject();
                if (config.contains("CPUs")) {
                    int cpus = config["CPUs"].toInt();
                    instance.setCpus(cpus);
                }
                if (config.contains("Memory")) {
                    int memory = config["Memory"].toInt();
                    instance.setMemory(memory);
                }
                if (config.contains("Driver")) {
                    QString driver = config["Driver"].toString();
                    instance.setDriver(driver);
                }
                if (!config.contains("KubernetesConfig")) {
                    instances << instance;
                    continue;
                }
                QJsonObject k8sConfig = config["KubernetesConfig"].toObject();
                if (k8sConfig.contains("ContainerRuntime")) {
                    QString containerRuntime = k8sConfig["ContainerRuntime"].toString();
                    instance.setContainerRuntime(containerRuntime);
                }
                instances << instance;
            }
        }
    }
    return instances;
}

QString Window::selectedInstance()
{
    QModelIndex index = instanceListView->currentIndex();
    QVariant variant = index.data(Qt::DisplayRole);
    if (variant.isNull()) {
        return QString();
    }
    return variant.toString();
}

void Window::setSelectedInstance(QString instance)
{
    QAbstractItemModel *model = instanceListView->model();
    QModelIndex start = model->index(0, 0);
    QModelIndexList index = model->match(start, Qt::DisplayRole, instance);
    if (index.size() == 0) {
        return;
    }
    instanceListView->setCurrentIndex(index[0]);
}

void Window::createInstanceGroupBox()
{
    instanceGroupBox = new QGroupBox(tr("Instances"));

    InstanceList instances = getInstances();
    instanceModel = new InstanceModel(instances);

    instanceListView = new QTableView();
    instanceListView->setModel(instanceModel);
    instanceListView->setSelectionMode(QAbstractItemView::SingleSelection);
    instanceListView->setSelectionBehavior(QAbstractItemView::SelectRows);
    instanceListView->horizontalHeader()->setSectionResizeMode(0, QHeaderView::Stretch);
    instanceListView->horizontalHeader()->setSectionResizeMode(1, QHeaderView::ResizeToContents);
    instanceListView->horizontalHeader()->setSectionResizeMode(2, QHeaderView::ResizeToContents);
    instanceListView->horizontalHeader()->setSectionResizeMode(3, QHeaderView::ResizeToContents);
    instanceListView->horizontalHeader()->setSectionResizeMode(4, QHeaderView::ResizeToContents);
    instanceListView->horizontalHeader()->setSectionResizeMode(5, QHeaderView::ResizeToContents);
    setSelectedInstance("default");

    connect(instanceListView, SIGNAL(clicked(QModelIndex)), this, SLOT(updateButtons()));

    startButton = new QPushButton(tr("Start"));
    stopButton = new QPushButton(tr("Stop"));
    deleteButton = new QPushButton(tr("Delete"));
    refreshButton = new QPushButton(tr("Refresh"));
    createButton = new QPushButton(tr("Create"));

    updateButtons();

    QHBoxLayout *instanceButtonLayout = new QHBoxLayout;
    instanceButtonLayout->addWidget(startButton);
    instanceButtonLayout->addWidget(stopButton);
    instanceButtonLayout->addWidget(deleteButton);
    instanceButtonLayout->addWidget(refreshButton);
    instanceButtonLayout->addWidget(createButton);

    QVBoxLayout *instanceLayout = new QVBoxLayout;
    instanceLayout->addWidget(instanceListView);
    instanceLayout->addLayout(instanceButtonLayout);
    instanceGroupBox->setLayout(instanceLayout);
}

void Window::updateButtons()
{
    QString inst = selectedInstance();
    if (inst.isEmpty()) {
        startButton->setEnabled(false);
        stopButton->setEnabled(false);
        deleteButton->setEnabled(false);
        return;
    }
    deleteButton->setEnabled(true);
    Instance instance = getInstanceHash()[inst];
    if (instance.status() == "Running") {
        startButton->setEnabled(false);
        stopButton->setEnabled(true);
    } else {
        startButton->setEnabled(true);
        stopButton->setEnabled(false);
    }
}

InstanceHash Window::getInstanceHash()
{
    InstanceList instances = getInstances();
    InstanceHash instanceHash;
    for (int i = 0; i < instances.size(); i++) {
        Instance instance = instances.at(i);
        instanceHash[instance.name()] = instance;
    }
    return instanceHash;
}

bool Window::sendMinikubeCommand(QStringList cmds)
{
    QString text;
    return sendMinikubeCommand(cmds, text);
}

bool Window::sendMinikubeCommand(QStringList cmds, QString &text)
{
    QString program = "minikube";
    QStringList arguments;
    arguments << cmds;
    bool success;

    QProcess *process = new QProcess(this);
    process->start(program, arguments);
    this->setCursor(Qt::WaitCursor);
    success = process->waitForFinished(300 * 1000);
    this->unsetCursor();

    if (success) {
        text = process->readAllStandardOutput();
    } else {
        qDebug() << process->readAllStandardOutput();
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

void Window::askName()
{
    QDialog dialog;
    dialog.setWindowTitle(tr("Create minikube Instance"));
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
    if (code == QDialog::Accepted) {
        QStringList arg = {"start", "-p", profile};
        sendMinikubeCommand(arg);
    } else if (code == QDialog::Rejected) {
        askCustom();
    }
}

void Window::askCustom()
{
    QDialog dialog;
    dialog.setWindowTitle(tr("Create minikube Instance"));
    dialog.setModal(true);

    QFormLayout form(&dialog);
    driverComboBox = new QComboBox;
    driverComboBox->addItem("docker");
    driverComboBox->addItem("hyperv");
    driverComboBox->addItem("hyperkit");
    form.addRow(new QLabel(tr("Driver")), driverComboBox);
    containerRuntimeComboBox = new QComboBox;
    containerRuntimeComboBox->addItem("docker");
    containerRuntimeComboBox->addItem("containerd");
    containerRuntimeComboBox->addItem("crio");
    form.addRow(new QLabel(tr("Container Runtime")), containerRuntimeComboBox);
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
        containerRuntime = containerRuntimeComboBox->itemText(containerRuntimeComboBox->currentIndex());
        cpus = cpuField.text().toInt();
        memory = memoryField.text().toInt();
        QStringList args = {"start", "-p", profile, "--driver", driver, "--container-runtime", containerRuntime, "--cpus", QString::number(cpus), "--memory", QString::number(memory)};
        sendMinikubeCommand(args);
    }
}

void Window::initMachine()
{
    askName();
    updateInstances();
}

#endif
