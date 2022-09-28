#ifndef LOGGER_H
#define LOGGER_H

#include <QObject>

class Logger : public QObject
{
    Q_OBJECT

public:
    explicit Logger();
    void log(QString message);

private:
    QString m_logPath;
};

#endif // LOGGER_H
