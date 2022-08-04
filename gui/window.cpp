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

    stackedWidget = new QStackedWidget;
    logger = new Logger();
    commandRunner = new CommandRunner(this, logger);
    basicView = new BasicView();
    advancedView = new AdvancedView(*trayIconIcon);
    errorMessage = new ErrorMessage(this, *trayIconIcon);
    progressWindow = new ProgressWindow(this, *trayIconIcon);
    tray = new Tray(*trayIconIcon);
    hyperKit = new HyperKit(*trayIconIcon);
    updater = new Updater(version, *trayIconIcon);

    op = new Operator(advancedView, basicView, commandRunner, errorMessage, progressWindow, tray,
                      hyperKit, updater, stackedWidget, this);

    stackedWidget->addWidget(basicView->basicView);
    stackedWidget->addWidget(advancedView->advancedView);
    layout = new QVBoxLayout;
    layout->addWidget(stackedWidget);
    setLayout(layout);
    resize(200, 300);
    setWindowTitle(tr("minikube"));
    setWindowIcon(*trayIconIcon);
}

void Window::setVisible(bool visible)
{
    tray->setVisible(visible);
    QDialog::setVisible(visible);
}

void Window::closeEvent(QCloseEvent *event)
{
#if __APPLE__
    if (!event->spontaneous() || !isVisible()) {
        return;
    }
#endif
    if (tray->isVisible()) {
        QMessageBox::information(this, tr("Systray"),
                                 tr("The program will keep running in the "
                                    "system tray. To terminate the program, "
                                    "choose <b>Quit</b> in the context menu "
                                    "of the system tray entry."));
        hide();
        event->ignore();
    }
}

static QString minikubePath()
{
    QString minikubePath = QStandardPaths::findExecutable("minikube");
    if (!minikubePath.isEmpty()) {
        return minikubePath;
    }
    QStringList path = { "/usr/local/bin" };
    return QStandardPaths::findExecutable("minikube", path);
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
