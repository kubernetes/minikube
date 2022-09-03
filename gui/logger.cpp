#include "logger.h"

#include <QStandardPaths>
#include <QFile>
#include <QDir>
#include <QTextStream>

Logger::Logger()
{
    QDir dir = QDir(QDir::homePath() + "/.minikube-gui");
    if (!dir.exists()) {
        dir.mkpath(".");
    }
    m_logPath = dir.filePath("logs.txt");
}

void Logger::log(QString message)
{
    QFile file(m_logPath);
    if (!file.open(QIODevice::Append)) {
        return;
    }
    QTextStream stream(&file);
    stream << message << "\n";
    file.close();
}
