#include "commandrunner.h"

#include <QStandardPaths>
#include <QJsonObject>
#include <QJsonArray>
#include <QJsonDocument>
#include <QJsonParseError>
#include <QDebug>

CommandRunner::CommandRunner(QDialog *parent, Logger *logger)
{
    m_env = QProcessEnvironment::systemEnvironment();
    m_parent = parent;
    m_logger = logger;
    minikubePath();
#if __APPLE__
    setMinikubePath();
#endif
}

void CommandRunner::executeCommand(QString program, QStringList args)
{
    QProcess *process = new QProcess(this);
    process->setProcessEnvironment(m_env);
    process->start(program, args);
    process->waitForFinished(-1);
    if (process->exitCode() == 0) {
        return;
    }
    QString out = process->readAllStandardOutput();
    QString err = process->readAllStandardError();
    QString log = QString("The following command failed:\n%1 %2\n\nStdout:\n%3\n\nStderr:\n%4\n\n")
                          .arg(program, args.join(" "), out, err);
    m_logger->log(log);
    delete process;
}

void CommandRunner::executeMinikubeCommand(QStringList args)
{
    m_isRunning = true;
    m_output = "";
    QStringList userArgs = { "--user", "minikube-gui" };
    args << userArgs;
    m_process = new QProcess(m_parent);
    connect(m_process, QOverload<int, QProcess::ExitStatus>::of(&QProcess::finished), this,
            &CommandRunner::executionCompleted);
    connect(m_process, &QProcess::readyReadStandardError, this, &CommandRunner::errorReady);
    connect(m_process, &QProcess::readyReadStandardOutput, this, &CommandRunner::outputReady);
    m_process->setProcessEnvironment(m_env);
    m_process->start(m_minikubePath, args);
    emit CommandRunner::startingExecution();
}

void CommandRunner::startMinikube(QStringList args)
{
    m_command = "start";
    QStringList baseArgs = { "start", "-o", "json" };
    baseArgs << args;
    m_args = baseArgs;
    executeMinikubeCommand(baseArgs);
    emit startCommandStarting();
}

void CommandRunner::stopMinikube(QStringList args)
{
    QStringList baseArgs = { "stop" };
    baseArgs << args;
    executeMinikubeCommand(baseArgs);
}

void CommandRunner::pauseMinikube(QStringList args)
{
    QStringList baseArgs = { "pause" };
    baseArgs << args;
    executeMinikubeCommand(baseArgs);
}

void CommandRunner::unpauseMinikube(QStringList args)
{
    QStringList baseArgs = { "unpause" };
    baseArgs << args;
    executeMinikubeCommand(baseArgs);
}

void CommandRunner::deleteMinikube(QStringList args)
{
    m_command = "delete";
    QStringList baseArgs = { "delete" };
    baseArgs << args;
    executeMinikubeCommand(baseArgs);
}

void CommandRunner::stopCommand()
{
    m_process->terminate();
}

static Cluster createClusterObject(QJsonObject obj)
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

static ClusterList jsonToClusterList(QString text)
{
    ClusterList clusters;
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
    return clusters;
}

void CommandRunner::requestClusters()
{
    m_command = "cluster";
    QStringList args = { "profile", "list", "-o", "json" };
    executeMinikubeCommand(args);
}

void CommandRunner::executionCompleted()
{
    m_isRunning = false;
    QString cmd = m_command;
    m_command = "";
    QString output = m_output;
    int exitCode = m_process->exitCode();
    delete m_process;
    if (cmd != "cluster") {
        emit executionEnded();
    }
    if (cmd == "start" && exitCode != 0) {
        emit error(m_args, output);
    }
    if (cmd == "cluster") {
        ClusterList clusterList = jsonToClusterList(output);
        emit updatedClusters(clusterList);
    }
}

void CommandRunner::errorReady()
{
    QString text = m_process->readAllStandardError();
    m_output.append(text);
    emit output(text);
}

void CommandRunner::outputReady()
{
    QString text = m_process->readAllStandardOutput();
    m_output.append(text);
    emit output(text);
}

#if __APPLE__
void CommandRunner::setMinikubePath()
{
    m_env = QProcessEnvironment::systemEnvironment();
    QString path = m_env.value("PATH") + ":/usr/local/bin";
    m_env.insert("PATH", path);
}
#endif

void CommandRunner::minikubePath()
{
    m_minikubePath = QStandardPaths::findExecutable("minikube");
    if (!m_minikubePath.isEmpty()) {
        return;
    }
    QStringList path = { "/usr/local/bin" };
    m_minikubePath = QStandardPaths::findExecutable("minikube", path);
}

bool CommandRunner::isRunning()
{
    return m_isRunning;
}
