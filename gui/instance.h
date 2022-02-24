#ifndef INSTANCE_H
#define INSTANCE_H

#include <QAbstractListModel>
#include <QString>
#include <QList>
#include <QMap>

//! [0]
class Instance
{
public:
    Instance() : Instance("") { }
    Instance(const QString &name)
        : m_name(name), m_status(""), m_driver(""), m_container_runtime(""), m_cpus(0), m_memory(0)
    {
    }

    QString name() const { return m_name; }
    QString status() const { return m_status; }
    void setStatus(QString status) { m_status = status; }
    QString driver() const { return m_driver; }
    void setDriver(QString driver) { m_driver = driver; }
    QString containerRuntime() const { return m_container_runtime; }
    void setContainerRuntime(QString containerRuntime) { m_container_runtime = containerRuntime; }
    int cpus() const { return m_cpus; }
    void setCpus(int cpus) { m_cpus = cpus; }
    int memory() const { return m_memory; }
    void setMemory(int memory) { m_memory = memory; }

private:
    QString m_name;
    QString m_status;
    QString m_driver;
    QString m_container_runtime;
    int m_cpus;
    int m_memory;
};
//! [0]

typedef QList<Instance> InstanceList;
typedef QHash<QString, Instance> InstanceHash;

//! [1]
class InstanceModel : public QAbstractListModel
{
    Q_OBJECT

public:
    InstanceModel(const InstanceList &instances, QObject *parent = nullptr)
        : QAbstractListModel(parent), instanceList(instances)
    {
    }

    void setInstances(const InstanceList &instances);
    int rowCount(const QModelIndex &parent = QModelIndex()) const override;
    int columnCount(const QModelIndex &parent = QModelIndex()) const override;
    QVariant data(const QModelIndex &index, int role) const override;
    QVariant headerData(int section, Qt::Orientation orientation,
                        int role = Qt::DisplayRole) const override;

private:
    InstanceList instanceList;
};
//! [1]

#endif // INSTANCE_H
