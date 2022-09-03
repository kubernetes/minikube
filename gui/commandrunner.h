#ifndef COMMANDRUNNER_H
#define COMMANDRUNNER_H

#include "cluster.h"
#include "logger.h"

#include <QString>
#include <QDialog>
#include <QStringList>
#include <QObject>
#include <QProcess>
#include <QProcessEnvironment>
#include <QIcon>

class CommandRunner : public QObject
{
    Q_OBJECT

public:
    CommandRunner(QDialog *parent, Logger *logger);

    void executeCommand(QString program, QStringList args);
    void startMinikube(QStringList args);
    void stopMinikube(QStringList args);
    void pauseMinikube(QStringList args);
    void unpauseMinikube(QStringList args);
    void deleteMinikube(QStringList args);
    void stopCommand();
    void requestClusters();
    bool isRunning();

signals:
    void startingExecution();
    void executionEnded();
    void output(QString text);
    void error(QStringList args, QString text);
    void updatedClusters(ClusterList clusterList);
    void startCommandStarting();

private slots:
    void executionCompleted();
    void outputReady();
    void errorReady();

private:
    void executeMinikubeCommand(QStringList args);
    void minikubePath();
#if __APPLE__
    void setMinikubePath();
#endif

    QProcess *m_process;
    QProcessEnvironment m_env;
    QString m_output;
    QString m_minikubePath;
    QString m_command;
    QDialog *m_parent;
    Logger *m_logger;
    QStringList m_args;
    bool m_isRunning;
};

#endif // COMMANDRUNNER_H
