#include "operator.h"

#include <QPushButton>
#include <QJsonObject>
#include <QJsonDocument>
#include <QStandardPaths>

Operator::Operator(AdvancedView *advancedView, BasicView *basicView, CommandRunner *commandRunner,
                   ErrorMessage *errorMessage, ProgressWindow *progressWindow, Tray *tray,
                   HyperKit *hyperKit, Updater *updater, QStackedWidget *stackedWidget,
                   QDialog *parent)
{
    m_advancedView = advancedView;
    m_basicView = basicView;
    m_commandRunner = commandRunner;
    m_errorMessage = errorMessage;
    m_progressWindow = progressWindow;
    m_tray = tray;
    m_hyperKit = hyperKit;
    m_updater = updater;
    m_stackedWidget = stackedWidget;
    m_parent = parent;
    m_isBasicView = true;
    dashboardProcess = NULL;

    connect(m_basicView, &BasicView::start, this, &Operator::startMinikube);
    connect(m_basicView, &BasicView::stop, this, &Operator::stopMinikube);
    connect(m_basicView, &BasicView::pause, this, &Operator::pauseOrUnpauseMinikube);
    connect(m_basicView, &BasicView::delete_, this, &Operator::deleteMinikube);
    connect(m_basicView, &BasicView::refresh, this, &Operator::updateClusters);
    connect(m_basicView, &BasicView::dockerEnv, this, &Operator::dockerEnv);
    connect(m_basicView, &BasicView::ssh, this, &Operator::sshConsole);
    connect(m_basicView, &BasicView::dashboard, this, &Operator::dashboardBrowser);
    connect(m_basicView, &BasicView::advanced, this, &Operator::toAdvancedView);

    connect(m_advancedView, &AdvancedView::start, this, &Operator::startMinikube);
    connect(m_advancedView, &AdvancedView::stop, this, &Operator::stopMinikube);
    connect(m_advancedView, &AdvancedView::pause, this, &Operator::pauseOrUnpauseMinikube);
    connect(m_advancedView, &AdvancedView::delete_, this, &Operator::deleteMinikube);
    connect(m_advancedView, &AdvancedView::refresh, this, &Operator::updateClusters);
    connect(m_advancedView, &AdvancedView::dockerEnv, this, &Operator::dockerEnv);
    connect(m_advancedView, &AdvancedView::ssh, this, &Operator::sshConsole);
    connect(m_advancedView, &AdvancedView::dashboard, this, &Operator::dashboardBrowser);
    connect(m_advancedView, &AdvancedView::basic, this, &Operator::toBasicView);
    connect(m_advancedView, &AdvancedView::createCluster, this, &Operator::createCluster);
    connect(m_advancedView->clusterListView, SIGNAL(clicked(QModelIndex)), this,
            SLOT(updateButtons()));

    connect(m_commandRunner, &CommandRunner::startingExecution, this, &Operator::commandStarting);
    connect(m_commandRunner, &CommandRunner::executionEnded, this, &Operator::commandEnding);
    connect(m_commandRunner, &CommandRunner::output, this, &Operator::commandOutput);
    connect(m_commandRunner, &CommandRunner::error, this, &Operator::commandError);
    connect(m_commandRunner, &CommandRunner::updatedClusters, this, &Operator::clustersReceived);
    connect(m_commandRunner, &CommandRunner::startCommandStarting, this,
            &Operator::startCommandStarting);

    connect(m_progressWindow, &ProgressWindow::cancelled, this, &Operator::cancelCommand);

    connect(m_tray, &Tray::restoreWindow, this, &Operator::restoreWindow);
    connect(m_tray, &Tray::hideWindow, this, &Operator::hideWindow);
    connect(m_tray, &Tray::start, this, &Operator::startMinikube);
    connect(m_tray, &Tray::stop, this, &Operator::stopMinikube);
    connect(m_tray, &Tray::pauseOrUnpause, this, &Operator::pauseOrUnpauseMinikube);

    connect(m_hyperKit, &HyperKit::rerun, this, &Operator::createCluster);

    updateClusters();
}

QStringList Operator::getCurrentClusterFlags()
{
    return { "-p", selectedClusterName() };
}

void Operator::startMinikube()
{
    m_commandRunner->startMinikube(getCurrentClusterFlags());
}

void Operator::stopMinikube()
{
    m_commandRunner->stopMinikube(getCurrentClusterFlags());
}

void Operator::pauseOrUnpauseMinikube()
{
    Cluster cluster = selectedCluster();
    if (cluster.status() == "Paused") {
        unpauseMinikube();
        return;
    }
    pauseMinikube();
}

void Operator::pauseMinikube()
{
    m_commandRunner->pauseMinikube(getCurrentClusterFlags());
}

void Operator::unpauseMinikube()
{
    m_commandRunner->unpauseMinikube(getCurrentClusterFlags());
}

void Operator::deleteMinikube()
{
    m_commandRunner->deleteMinikube(getCurrentClusterFlags());
}

void Operator::createCluster(QStringList args)
{
    m_commandRunner->startMinikube(args);
}

void Operator::startCommandStarting()
{
    commandStarting();
    m_progressWindow->setText("Starting...");
    m_progressWindow->show();
}

void Operator::commandStarting()
{
    m_advancedView->showLoading();
    m_tray->disableActions();
    m_parent->setCursor(Qt::WaitCursor);
    disableButtons();
}

void Operator::disableButtons()
{
    if (m_isBasicView) {
        m_basicView->disableButtons();
    } else {
        m_advancedView->disableButtons();
    }
}

void Operator::commandEnding()
{
    m_progressWindow->done();
    updateClusters();
}

void Operator::toAdvancedView()
{
    m_isBasicView = false;
    m_stackedWidget->setCurrentIndex(1);
    m_parent->resize(670, 400);
    updateButtons();
}

void Operator::toBasicView()
{
    m_isBasicView = true;
    m_stackedWidget->setCurrentIndex(0);
    m_parent->resize(200, 300);
    updateButtons();
}

void Operator::updateClusters()
{
    m_commandRunner->requestClusters();
}

void Operator::clustersReceived(ClusterList clusterList)
{
    m_clusterList = clusterList;
    m_advancedView->updateClustersTable(m_clusterList);
    updateButtons();
    m_advancedView->hideLoading();
    m_parent->unsetCursor();
    m_updater->checkForUpdates();
}

void Operator::updateButtons()
{
    Cluster cluster = selectedCluster();
    if (m_isBasicView) {
        m_basicView->update(cluster);
    } else {
        m_advancedView->update(cluster);
    }
    m_tray->updateTrayActions(cluster);
    m_tray->updateStatus(cluster);
}

void Operator::restoreWindow()
{
    bool wasVisible = m_parent->isVisible();
    m_parent->showNormal();
    m_parent->activateWindow();
    if (wasVisible) {
        return;
    }
    if (m_commandRunner->isRunning())
        return;
    updateClusters();
}

void Operator::hideWindow()
{
    m_parent->hide();
}

void Operator::commandOutput(QString text)
{
    QStringList lines;
#if QT_VERSION >= QT_VERSION_CHECK(5, 14, 0)
    lines = text.split("\n", Qt::SkipEmptyParts);
#else
    lines = text.split("\n", QString::SkipEmptyParts);
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
        QString totalString = data["totalsteps"].toString();
        int totalSteps = totalString.toInt();
        QString message = data["message"].toString();
        m_progressWindow->setBarMaximum(totalSteps);
        m_progressWindow->setBarValue(currStep);
        m_progressWindow->setText(message);
    }
}

void Operator::commandError(QStringList args, QString text)
{
#if __APPLE__
    if (m_hyperKit->hyperkitPermissionFix(args, text)) {
        return;
    }
#endif
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

        m_errorMessage->error(name, advice, message, url, issues);
        break;
    }
}

void Operator::cancelCommand()
{
    m_commandRunner->stopCommand();
}

QString Operator::selectedClusterName()
{
    if (m_isBasicView) {
        return "minikube";
    }
    return m_advancedView->selectedClusterName();
}

Cluster Operator::selectedCluster()
{
    QString clusterName = selectedClusterName();
    if (clusterName.isEmpty()) {
        return Cluster();
    }
    ClusterList clusters = m_clusterList;
    ClusterHash clusterHash;
    for (int i = 0; i < clusters.size(); i++) {
        Cluster cluster = clusters.at(i);
        clusterHash[cluster.name()] = cluster;
    }
    return clusterHash[clusterName];
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

void Operator::sshConsole()
{
    QString program = minikubePath();
    QString commandArgs = QString("ssh -p %1").arg(selectedClusterName());
    QString command = QString("%1 %2").arg(program, commandArgs);
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
    console->setArgs({ commandArgs });
    console->startShellProgram();

    QObject::connect(console, SIGNAL(finished()), mainWindow, SLOT(close()));

    mainWindow->setWindowTitle(nameLabel->text());
    mainWindow->resize(800, 400);
    mainWindow->setCentralWidget(console);
    mainWindow->show();
#elif __APPLE__
    QStringList arguments = { "-e", "tell app \"Terminal\"",
                              "-e", "do script \"" + command + "\"",
                              "-e", "activate",
                              "-e", "end tell" };
    m_commandRunner->executeCommand("/usr/bin/osascript", arguments);
#else
    QString terminal = qEnvironmentVariable("TERMINAL");
    if (terminal.isEmpty()) {
        terminal = "x-terminal-emulator";
        if (QStandardPaths::findExecutable(terminal).isEmpty()) {
            terminal = "xterm";
        }
    }

    m_commandRunner->executeCommand(QStandardPaths::findExecutable(terminal), { "-e", command });
#endif
}

void Operator::dockerEnv()
{
    QString program = minikubePath();
    QString commandArgs = QString("$(%1 -p %2 docker-env)").arg(program, selectedClusterName());
    QString command = QString("eval %1").arg(commandArgs);
#ifndef QT_NO_TERMWIDGET
    QMainWindow *mainWindow = new QMainWindow();
    int startnow = 0; // set shell program first

    QTermWidget *console = new QTermWidget(startnow);

    QFont font = QApplication::font();
    font.setFamily("Monospace");
    font.setPointSize(10);

    console->setTerminalFont(font);
    console->setColorScheme("Tango");
    console->setShellProgram("eval");
    console->setArgs({ commandArgs });
    console->startShellProgram();

    QObject::connect(console, SIGNAL(finished()), mainWindow, SLOT(close()));

    mainWindow->setWindowTitle(nameLabel->text());
    mainWindow->resize(800, 400);
    mainWindow->setCentralWidget(console);
    mainWindow->show();
#elif __APPLE__
    QStringList arguments = { "-e", "tell app \"Terminal\"",
                              "-e", "do script \"" + command + "\"",
                              "-e", "activate",
                              "-e", "end tell" };
    m_commandRunner->executeCommand("/usr/bin/osascript", arguments);
#else
    QString terminal = qEnvironmentVariable("TERMINAL");
    if (terminal.isEmpty()) {
        terminal = "x-terminal-emulator";
        if (QStandardPaths::findExecutable(terminal).isEmpty()) {
            terminal = "xterm";
        }
    }

    m_commandRunner->executeCommand(QStandardPaths::findExecutable(terminal), { "-e", command });
#endif
}

void Operator::dashboardBrowser()
{
    dashboardClose();

    QString program = minikubePath();
    QProcess *process = new QProcess(this);
    QStringList arguments = { "dashboard", "-p", selectedClusterName() };
    process->start(program, arguments);

    dashboardProcess = process;
    dashboardProcess->waitForStarted();
}

void Operator::dashboardClose()
{
    if (dashboardProcess) {
        dashboardProcess->terminate();
        dashboardProcess->waitForFinished();
    }
}
