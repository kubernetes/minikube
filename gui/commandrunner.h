#ifndef COMMANDRUNNER_H
#define COMMANDRUNNER_H

#include "cluster.h"

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
    CommandRunner(QDialog *parent);

    void startMinikube(QStringList args);
    void stopMinikube(QStringList args);
    void pauseMinikube(QStringList args);
    void unpauseMinikube(QStringList args);
    void deleteMinikube(QStringList args);
    void stopCommand();
    void requestClusters();

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
    QStringList m_args;
};

#endif // COMMANDRUNNER_H
