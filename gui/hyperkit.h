#ifndef HYPERKIT_H
#define HYPERKIT_H

#include <QStringList>
#include <QObject>
#include <QIcon>

class HyperKit : public QObject
{
    Q_OBJECT

public:
    explicit HyperKit(QIcon icon);
    bool hyperkitPermissionFix(QStringList args, QString text);

signals:
    void rerun(QStringList args);

private:
    void hyperkitPermission();
    bool showHyperKitMessage();
    QIcon m_icon;
};

#endif // HYPERKIT_H
