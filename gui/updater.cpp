#include "updater.h"

#include <QJsonObject>
#include <QJsonDocument>
#include <QJsonArray>
#include <QDialog>
#include <QLabel>
#include <QFormLayout>
#include <QDialogButtonBox>
#include <QNetworkAccessManager>
#include <QNetworkReply>
#include <QEventLoop>
#include <QStandardPaths>
#include <QDir>

Updater::Updater(QVersionNumber version, QIcon icon)
{
    m_version = version;
    m_icon = icon;
}

static bool checkedForUpdateRecently()
{
    QString filePath = QStandardPaths::locate(QStandardPaths::HomeLocation,
                                              "/.minikube-gui/last_update_check");
    if (filePath == "") {
        return false;
    }
    QFile file(filePath);
    if (!file.open(QIODevice::ReadOnly)) {
        return false;
    }
    QTextStream in(&file);
    QString line = in.readLine();
    QDateTime nextCheck = QDateTime::fromString(line).addSecs(60 * 60 * 24);
    QDateTime now = QDateTime::currentDateTime();
    return nextCheck > now;
}

static void logUpdateCheck()
{
    QDir dir = QDir(QDir::homePath() + "/.minikube-gui");
    if (!dir.exists()) {
        dir.mkpath(".");
    }
    QString filePath = dir.filePath("last_update_check");
    QFile file(filePath);
    if (!file.open(QIODevice::WriteOnly)) {
        return;
    }
    QTextStream stream(&file);
    stream << QDateTime::currentDateTime().toString() << Qt::endl;
}

void Updater::checkForUpdates()
{
    if (checkedForUpdateRecently()) {
        return;
    }
    logUpdateCheck();
    QString releases = getRequest("https://storage.googleapis.com/minikube-gui/releases.json");
    QJsonObject latestRelease =
            QJsonDocument::fromJson(releases.toUtf8()).array().first().toObject();
    QString latestReleaseVersion = latestRelease["name"].toString();
    QVersionNumber latestReleaseVersionNumber = QVersionNumber::fromString(latestReleaseVersion);
    if (m_version >= latestReleaseVersionNumber) {
        return;
    }
    QJsonObject links = latestRelease["links"].toObject();
    QString key;
#if __linux__
    key = "linux";
#elif __APPLE__
    key = "darwin";
#else
    key = "windows";
#endif
    QString link = links[key].toString();
    notifyUpdate(latestReleaseVersion, link);
}

void Updater::notifyUpdate(QString latest, QString link)
{
    QDialog dialog;
    dialog.setWindowTitle(tr("minikube GUI Update Available"));
    dialog.setWindowIcon(m_icon);
    dialog.setModal(true);
    QFormLayout form(&dialog);
    QLabel *msgLabel = new QLabel();
    msgLabel->setText("Version " + latest
                      + " of minikube GUI is now available!\n\nDownload the update from:");
    form.addWidget(msgLabel);
    QLabel *linkLabel = new QLabel();
    linkLabel->setOpenExternalLinks(true);
    linkLabel->setText("<a href=\"" + link + "\">" + link + "</a>");
    form.addWidget(linkLabel);
    QDialogButtonBox buttonBox(Qt::Horizontal, &dialog);
    buttonBox.addButton(QString(tr("OK")), QDialogButtonBox::AcceptRole);
    connect(&buttonBox, &QDialogButtonBox::accepted, &dialog, &QDialog::accept);
    form.addRow(&buttonBox);
    dialog.exec();
}

QString Updater::getRequest(QString url)
{
    QNetworkAccessManager *manager = new QNetworkAccessManager();
    QObject::connect(manager, &QNetworkAccessManager::finished, this, [=](QNetworkReply *reply) {
        if (reply->error()) {
            qDebug() << reply->errorString();
        }
    });
    QNetworkReply *resp = manager->get(QNetworkRequest(QUrl(url)));
    QEventLoop loop;
    connect(resp, &QNetworkReply::finished, &loop, &QEventLoop::quit);
    loop.exec();
    return resp->readAll();
}
