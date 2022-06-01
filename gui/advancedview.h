#ifndef ADVANCEDVIEW_H
#define ADVANCEDVIEW_H

#include "cluster.h"

#include <QObject>
#include <QPushButton>
#include <QLabel>
#include <QTableView>

class AdvancedView : public QObject
{
    Q_OBJECT

public:
    explicit AdvancedView(QIcon icon);
    QWidget *advancedView;
    QTableView *clusterListView;

    QString selectedClusterName();
    void updateClustersTable(ClusterList clusters);
    void showLoading();
    void hideLoading();
    void disableButtons();

public slots:
    void update(Cluster cluster);

signals:
    void start();
    void stop();
    void pause();
    void delete_();
    void refresh();
    void ssh();
    void dashboard();
    void basic();
    void createCluster(QStringList args);

private:
    void setSelectedClusterName(QString cluster);
    void askName();
    void askCustom();

    QPushButton *startButton;
    QPushButton *stopButton;
    QPushButton *pauseButton;
    QPushButton *deleteButton;
    QPushButton *refreshButton;
    QPushButton *sshButton;
    QPushButton *dashboardButton;
    QPushButton *basicButton;
    QPushButton *createButton;
    QLabel *loading;
    ClusterModel *m_clusterModel;
    QIcon m_icon;
};

#endif // ADVANCEDVIEW_H
