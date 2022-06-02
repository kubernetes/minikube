#include "progresswindow.h"

#include <QPushButton>
#include <QVBoxLayout>

ProgressWindow::ProgressWindow(QWidget *parent, QIcon icon)
{
    m_icon = icon;

    m_dialog = new QDialog(parent);
    m_dialog->setWindowIcon(m_icon);
    m_dialog->resize(300, 150);
    m_dialog->setWindowFlags(Qt::FramelessWindowHint);
    m_dialog->setModal(true);

    QVBoxLayout form(m_dialog);

    m_text = new QLabel();
    m_text->setWordWrap(true);
    form.addWidget(m_text);

    m_progressBar = new QProgressBar();
    form.addWidget(m_progressBar);

    m_cancelButton = new QPushButton();
    m_cancelButton->setText(tr("Cancel"));
    connect(m_cancelButton, &QAbstractButton::clicked, this, &ProgressWindow::cancel);
    form.addWidget(m_cancelButton);

    // if the dialog isn't opened now it breaks formatting
    m_dialog->open();
    m_dialog->hide();
}

void ProgressWindow::setBarMaximum(int max)
{
    m_progressBar->setMaximum(max);
}

void ProgressWindow::setBarValue(int value)
{
    m_progressBar->setValue(value);
}

void ProgressWindow::setText(QString text)
{
    m_text->setText(text);
}

void ProgressWindow::show()
{
    m_dialog->open();
}

void ProgressWindow::cancel()
{
    done();
    emit cancelled();
}

void ProgressWindow::done()
{
    m_dialog->hide();
    m_progressBar->setValue(0);
}
