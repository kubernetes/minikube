#ifndef OPERATOR_H
#define OPERATOR_H

#include "advancedview.h"
#include "basicview.h"
#include "cluster.h"
#include "commandrunner.h"
#include "errormessage.h"
#include "progresswindow.h"
#include "tray.h"
#include "hyperkit.h"
#include "updater.h"

#include <QStackedWidget>

class Operator : public QObject
{
    Q_OBJECT

public:
    Operator(AdvancedView *advancedView, BasicView *basicView, CommandRunner *commandRunner,
             ErrorMessage *errorMessage, ProgressWindow *progressWindow, Tray *tray,
             HyperKit *hyperKit, Updater *updater, QStackedWidget *stackedWidget, QDialog *parent);

public slots:
    void startMinikube();
    void stopMinikube();
    void pauseOrUnpauseMinikube();
    void deleteMinikube();

private slots:
    void commandStarting();
    void commandEnding();
    void commandOutput(QString text);
    void commandError(QStringList args, QString text);
    void cancelCommand();
    void toBasicView();
    void toAdvancedView();
    void createCluster(QStringList args);
    void updateButtons();
    void clustersReceived(ClusterList clusterList);
    void startCommandStarting();

private:
    QStringList getCurrentClusterFlags();
    void updateClusters();
    QString selectedClusterName();
    Cluster selectedCluster();
    void sshConsole();
    void dockerEnv();
    void dashboardBrowser();
    void dashboardClose();
    void pauseMinikube();
    void unpauseMinikube();
    void restoreWindow();
    void hideWindow();
    void disableButtons();

    AdvancedView *m_advancedView;
    BasicView *m_basicView;
    CommandRunner *m_commandRunner;
    ErrorMessage *m_errorMessage;
    ProgressWindow *m_progressWindow;
    ClusterList m_clusterList;
    Tray *m_tray;
    HyperKit *m_hyperKit;
    Updater *m_updater;
    bool m_isBasicView;
    QProcess *dashboardProcess;
    QStackedWidget *m_stackedWidget;
    QDialog *m_parent;
};

#endif // OPERATOR_H
