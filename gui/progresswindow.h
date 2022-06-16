#ifndef PROGRESSWINDOW_H
#define PROGRESSWINDOW_H

#include <QDialog>
#include <QObject>
#include <QLabel>
#include <QProcess>
#include <QProgressBar>
#include <QString>
#include <QIcon>

class ProgressWindow : public QObject
{
    Q_OBJECT

public:
    explicit ProgressWindow(QWidget *parent, QIcon icon);

    void setBarMaximum(int max);
    void setBarValue(int value);
    void setText(QString text);
    void show();
    void done();

signals:
    void cancelled();

private:
    void cancel();

    QDialog *m_dialog;
    QLabel *m_text;
    QProgressBar *m_progressBar;
    QPushButton *m_cancelButton;
    QIcon m_icon;
};

#endif // PROGRESSWINDOW_H
