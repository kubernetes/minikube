#ifndef ERRORMESSAGE_H
#define ERRORMESSAGE_H

#include <QObject>
#include <QDialog>
#include <QFormLayout>
#include <QLabel>
#include <QIcon>

class ErrorMessage : public QObject
{
    Q_OBJECT

public:
    explicit ErrorMessage(QDialog *parent, QIcon icon);

    void error(QString errorCode, QString advice, QString errorMessage, QString url, QString issues);
    QLabel *createLabel(QString title, QString text, QFormLayout *form, bool isLink);

private:
    QDialog *m_dialog;
    QIcon m_icon;
    QDialog *m_parent;
};

#endif // ERRORMESSAGE_H
