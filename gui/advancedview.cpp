#include "advancedview.h"

#include <QHeaderView>
#include <QTableView>
#include <QHBoxLayout>
#include <QLabel>
#include <QDialog>
#include <QFormLayout>
#include <QDialogButtonBox>
#include <QLineEdit>
#include <QComboBox>

AdvancedView::AdvancedView(QIcon icon)
{
    m_icon = icon;

    advancedView = new QWidget();
    advancedView->setSizePolicy(QSizePolicy::Ignored, QSizePolicy::Ignored);

    ClusterList clusters;
    m_clusterModel = new ClusterModel(clusters);

    clusterListView = new QTableView();
    clusterListView->setModel(m_clusterModel);
    clusterListView->setSelectionMode(QAbstractItemView::SingleSelection);
    clusterListView->setSelectionBehavior(QAbstractItemView::SelectRows);
    clusterListView->horizontalHeader()->setSectionResizeMode(0, QHeaderView::Stretch);
    clusterListView->horizontalHeader()->setSectionResizeMode(1, QHeaderView::ResizeToContents);
    clusterListView->horizontalHeader()->setSectionResizeMode(2, QHeaderView::ResizeToContents);
    clusterListView->horizontalHeader()->setSectionResizeMode(3, QHeaderView::ResizeToContents);
    clusterListView->horizontalHeader()->setSectionResizeMode(4, QHeaderView::ResizeToContents);
    clusterListView->horizontalHeader()->setSectionResizeMode(5, QHeaderView::ResizeToContents);
    clusterListView->horizontalHeader()->setSectionResizeMode(6, QHeaderView::ResizeToContents);
    setSelectedClusterName("default");

    startButton = new QPushButton(tr("Start"));
    stopButton = new QPushButton(tr("Stop"));
    pauseButton = new QPushButton(tr("Pause"));
    deleteButton = new QPushButton(tr("Delete"));
    refreshButton = new QPushButton(tr("Refresh"));
    createButton = new QPushButton(tr("Create"));
    dockerEnvButton = new QPushButton(tr("docker-env"));
    sshButton = new QPushButton(tr("SSH"));
    dashboardButton = new QPushButton(tr("Dashboard"));
    basicButton = new QPushButton(tr("Basic View"));

    disableButtons();

    QHBoxLayout *topButtonLayout = new QHBoxLayout;
    topButtonLayout->addWidget(createButton);
    topButtonLayout->addWidget(refreshButton);
    topButtonLayout->addWidget(basicButton);
    topButtonLayout->addSpacing(340);

    QHBoxLayout *bottomButtonLayout = new QHBoxLayout;
    bottomButtonLayout->addWidget(startButton);
    bottomButtonLayout->addWidget(stopButton);
    bottomButtonLayout->addWidget(pauseButton);
    bottomButtonLayout->addWidget(deleteButton);
    bottomButtonLayout->addWidget(dockerEnvButton);
    bottomButtonLayout->addWidget(sshButton);
    bottomButtonLayout->addWidget(dashboardButton);

    QVBoxLayout *clusterLayout = new QVBoxLayout;
    clusterLayout->addLayout(topButtonLayout);
    clusterLayout->addWidget(clusterListView);
    clusterLayout->addLayout(bottomButtonLayout);
    advancedView->setLayout(clusterLayout);

    QFont *loadingFont = new QFont();
    loadingFont->setPointSize(30);
    loading = new QLabel("Loading...");
    loading->setFont(*loadingFont);
    loading->setParent(clusterListView);
    loading->setHidden(true);

    connect(startButton, &QAbstractButton::clicked, this, &AdvancedView::start);
    connect(stopButton, &QAbstractButton::clicked, this, &AdvancedView::stop);
    connect(pauseButton, &QAbstractButton::clicked, this, &AdvancedView::pause);
    connect(deleteButton, &QAbstractButton::clicked, this, &AdvancedView::delete_);
    connect(refreshButton, &QAbstractButton::clicked, this, &AdvancedView::refresh);
    connect(createButton, &QAbstractButton::clicked, this, &AdvancedView::askName);
    connect(dockerEnvButton, &QAbstractButton::clicked, this, &AdvancedView::dockerEnv);
    connect(sshButton, &QAbstractButton::clicked, this, &AdvancedView::ssh);
    connect(dashboardButton, &QAbstractButton::clicked, this, &AdvancedView::dashboard);
    connect(basicButton, &QAbstractButton::clicked, this, &AdvancedView::basic);
}

static QString getPauseLabel(bool isPaused)
{
    if (isPaused) {
        return "Unpause";
    }
    return "Pause";
}

static QString getStartLabel(bool isRunning)
{
    if (isRunning) {
        return "Reload";
    }
    return "Start";
}

void AdvancedView::update(Cluster cluster)
{
    basicButton->setEnabled(true);
    createButton->setEnabled(true);
    refreshButton->setEnabled(true);
    bool exists = !cluster.isEmpty();
    bool isRunning = cluster.status() == "Running";
    bool isPaused = cluster.status() == "Paused";
    startButton->setEnabled(exists);
    stopButton->setEnabled(isRunning || isPaused);
    pauseButton->setEnabled(isRunning || isPaused);
    deleteButton->setEnabled(exists);
    dashboardButton->setEnabled(isRunning);
#if __linux__ || __APPLE__
    dockerEnvButton->setEnabled(isRunning);
    sshButton->setEnabled(exists);
#else
    dockerEnvButton->setEnabled(false);
    sshButton->setEnabled(false);
#endif
    pauseButton->setText(getPauseLabel(isPaused));
    startButton->setText(getStartLabel(isRunning));
}

void AdvancedView::setSelectedClusterName(QString cluster)
{
    QAbstractItemModel *model = clusterListView->model();
    QModelIndex start = model->index(0, 0);
    QModelIndexList index = model->match(start, Qt::DisplayRole, cluster);
    if (index.size() == 0) {
        return;
    }
    clusterListView->setCurrentIndex(index[0]);
}

QString AdvancedView::selectedClusterName()
{
    QModelIndex index = clusterListView->currentIndex();
    QVariant variant = index.siblingAtColumn(0).data(Qt::DisplayRole);
    if (variant.isNull()) {
        return QString();
    }
    return variant.toString();
}

void AdvancedView::updateClustersTable(ClusterList clusterList)
{
    QString cluster = selectedClusterName();
    m_clusterModel->setClusters(clusterList);
    setSelectedClusterName(cluster);
}

static int getCenter(int widgetSize, int parentSize)
{
    return parentSize / 2 - widgetSize / 2;
}

void AdvancedView::showLoading()
{
    clusterListView->setEnabled(false);
    loading->setHidden(false);
    loading->raise();
    int width = getCenter(loading->width(), clusterListView->width());
    int height = getCenter(loading->height(), clusterListView->height());
    loading->move(width, height);
}

void AdvancedView::hideLoading()
{
    loading->setHidden(true);
    clusterListView->setEnabled(true);
}

static QString profile = "minikube";
static int cpus = 2;
static int memory = 2400;
static QString driver = "";
static QString containerRuntime = "";
static QString k8sVersion = "";

void AdvancedView::askName()
{
    QDialog dialog;
    dialog.setWindowTitle(tr("Create minikube Cluster"));
    dialog.setWindowIcon(m_icon);
    dialog.setModal(true);

    QFormLayout form(&dialog);
    QDialogButtonBox buttonBox(Qt::Horizontal, &dialog);
    QLineEdit profileField(profile, &dialog);
    form.addRow(new QLabel(tr("Profile")), &profileField);
    buttonBox.addButton(QString(tr("Use Default Values")), QDialogButtonBox::AcceptRole);
    connect(&buttonBox, &QDialogButtonBox::accepted, &dialog, &QDialog::accept);
    buttonBox.addButton(QString(tr("Set Custom Values")), QDialogButtonBox::RejectRole);
    connect(&buttonBox, &QDialogButtonBox::rejected, &dialog, &QDialog::reject);
    form.addRow(&buttonBox);

    int code = dialog.exec();
    profile = profileField.text();
    if (code == QDialog::Accepted) {
        QStringList args = { "-p", profile };
        emit createCluster(args);
    } else if (code == QDialog::Rejected) {
        askCustom();
    }
}

void AdvancedView::askCustom()
{
    QDialog dialog;
    dialog.setWindowTitle(tr("Set Cluster Values"));
    dialog.setWindowIcon(m_icon);
    dialog.setModal(true);

    QFormLayout form(&dialog);
    QComboBox *driverComboBox = new QComboBox;
    driverComboBox->addItems({ "docker", "virtualbox", "vmware", "podman" });
#if __linux__
    driverComboBox->addItems({ "kvm2", "qemu" });
#elif __APPLE__
    driverComboBox->addItems({ "hyperkit", "qemu", "parallels" });
#else
    driverComboBox->addItem("hyperv");
#endif
    form.addRow(new QLabel(tr("Driver")), driverComboBox);
    QComboBox *containerRuntimeComboBox = new QComboBox;
    containerRuntimeComboBox->addItems({ "docker", "containerd", "crio" });
    form.addRow(new QLabel(tr("Container Runtime")), containerRuntimeComboBox);
    QComboBox *k8sVersionComboBox = new QComboBox;
    k8sVersionComboBox->addItems({ "stable", "latest", "none" });
    form.addRow(new QLabel(tr("Kubernetes Version")), k8sVersionComboBox);
    QLineEdit cpuField(QString::number(cpus), &dialog);
    form.addRow(new QLabel(tr("CPUs")), &cpuField);
    QLineEdit memoryField(QString::number(memory), &dialog);
    form.addRow(new QLabel(tr("Memory")), &memoryField);

    QDialogButtonBox buttonBox(Qt::Horizontal, &dialog);
    buttonBox.addButton(QString(tr("Create")), QDialogButtonBox::AcceptRole);
    connect(&buttonBox, &QDialogButtonBox::accepted, &dialog, &QDialog::accept);
    buttonBox.addButton(QString(tr("Cancel")), QDialogButtonBox::RejectRole);
    connect(&buttonBox, &QDialogButtonBox::rejected, &dialog, &QDialog::reject);
    form.addRow(&buttonBox);

    int code = dialog.exec();
    if (code == QDialog::Accepted) {
        driver = driverComboBox->itemText(driverComboBox->currentIndex());
        containerRuntime =
                containerRuntimeComboBox->itemText(containerRuntimeComboBox->currentIndex());
        k8sVersion = k8sVersionComboBox->itemText(k8sVersionComboBox->currentIndex());
        if (k8sVersion == "none") {
            k8sVersion = "v0.0.0";
        }
        cpus = cpuField.text().toInt();
        memory = memoryField.text().toInt();
        QStringList args = { "-p",
                             profile,
                             "--driver",
                             driver,
                             "--container-runtime",
                             containerRuntime,
                             "--kubernetes-version",
                             k8sVersion,
                             "--cpus",
                             QString::number(cpus),
                             "--memory",
                             QString::number(memory) };
        emit createCluster(args);
    }
}

void AdvancedView::disableButtons()
{
    startButton->setEnabled(false);
    stopButton->setEnabled(false);
    pauseButton->setEnabled(false);
    deleteButton->setEnabled(false);
    dockerEnvButton->setEnabled(false);
    sshButton->setEnabled(false);
    dashboardButton->setEnabled(false);
    basicButton->setEnabled(false);
    createButton->setEnabled(false);
    refreshButton->setEnabled(false);
}
