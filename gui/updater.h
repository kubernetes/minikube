#ifndef UPDATER_H
#define UPDATER_H

#include <QObject>
#include <QVersionNumber>
#include <QIcon>

class Updater : public QObject
{
    Q_OBJECT

public:
    explicit Updater(QVersionNumber version, QIcon icon);
    void checkForUpdates();

private:
    void notifyUpdate(QString latest, QString link);
    QString getRequest(QString url);
    QVersionNumber m_version;
    QIcon m_icon;
};

#endif // UPDATER_H
