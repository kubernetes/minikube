#ifndef BASICVIEW_H
#define BASICVIEW_H

#include "cluster.h"

#include <QObject>
#include <QPushButton>

class BasicView : public QObject
{
    Q_OBJECT

public:
    explicit BasicView();
    QWidget *basicView;
    void update(Cluster cluster);
    void disableButtons();

signals:
    void start();
    void stop();
    void pause();
    void delete_();
    void refresh();
    void dockerEnv();
    void ssh();
    void dashboard();
    void advanced();

private:
    QPushButton *startButton;
    QPushButton *stopButton;
    QPushButton *pauseButton;
    QPushButton *deleteButton;
    QPushButton *refreshButton;
    QPushButton *dockerEnvButton;
    QPushButton *sshButton;
    QPushButton *dashboardButton;
    QPushButton *advancedButton;
};

#endif // BASICVIEW_H
