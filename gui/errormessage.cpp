#include "errormessage.h"

#include <QFormLayout>
#include <QDialogButtonBox>
#include <QLabel>
#include <QDir>
#include <QTextOption>
#include <QTextEdit>

ErrorMessage::ErrorMessage(QDialog *parent, QIcon icon)
{
    m_parent = parent;
    m_icon = icon;
}

void ErrorMessage::error(QString errorCode, QString advice, QString message, QString url, QString issues)
{

    m_dialog = new QDialog(m_parent);
    m_dialog->setWindowTitle(tr("minikube start failed"));
    m_dialog->setWindowIcon(m_icon);
    m_dialog->setFixedWidth(600);
    m_dialog->setModal(true);
    QFormLayout form(m_dialog);
    createLabel("Error Code", errorCode, &form, false);
    createLabel("Advice", advice, &form, false);
    QTextEdit *errorMessage = new QTextEdit();
    errorMessage->setText(message);
    errorMessage->setWordWrapMode(QTextOption::WrapAnywhere);
    int pointSize = errorMessage->font().pointSize();
    errorMessage->setFont(QFont("Courier", pointSize));
    errorMessage->setAutoFillBackground(true);
    errorMessage->setReadOnly(true);
    form.addRow(errorMessage);
    createLabel("Link to documentation", url, &form, true);
    createLabel("Link to related issue", issues, &form, true);
    QLabel *fileLabel = new QLabel();
    fileLabel->setOpenExternalLinks(true);
    fileLabel->setWordWrap(true);
    QString logFile = QDir::homePath() + "/.minikube/logs/lastStart.txt";
    fileLabel->setText("<a href='file:///" + logFile + "'>View log file</a>");
    form.addRow(fileLabel);
    QDialogButtonBox buttonBox(Qt::Horizontal, m_dialog);
    buttonBox.addButton(QString(tr("OK")), QDialogButtonBox::AcceptRole);
    connect(&buttonBox, &QDialogButtonBox::accepted, m_dialog, &QDialog::accept);
    form.addRow(&buttonBox);
    m_dialog->exec();
}

QLabel *ErrorMessage::createLabel(QString title, QString text, QFormLayout *form, bool isLink)
{
    QLabel *label = new QLabel();
    if (!text.isEmpty()) {
        form->addRow(label);
    }
    if (isLink) {
        label->setOpenExternalLinks(true);
        text = "<a href='" + text + "'>" + text + "</a>";
    }
    label->setWordWrap(true);
    label->setText(title + ": " + text);
    return label;
}
