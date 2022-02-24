#include "instance.h"

#include <QStringList>

void InstanceModel::setInstances(const InstanceList &instances)
{
    beginResetModel();
    instanceList = instances;
    endResetModel();
}

int InstanceModel::rowCount(const QModelIndex &) const
{
    return instanceList.count();
}

int InstanceModel::columnCount(const QModelIndex &) const
{
    return 6;
}

static QStringList binaryAbbrs = { "B", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB", "ZiB", "YiB" };

QVariant InstanceModel::data(const QModelIndex &index, int role) const
{
    if (!index.isValid())
        return QVariant();

    if (index.row() >= instanceList.size())
        return QVariant();
    if (index.column() >= 6)
        return QVariant();

    if (role == Qt::TextAlignmentRole) {
        switch (index.column()) {
        case 0:
            return QVariant(Qt::AlignLeft | Qt::AlignVCenter);
        case 1:
            return QVariant(Qt::AlignRight | Qt::AlignVCenter);
        case 2:
            // fall-through
        case 3:
            // fall-through
        case 4:
            // fall-through
        case 5:
            return QVariant(Qt::AlignHCenter | Qt::AlignVCenter);
        }
    }
    if (role == Qt::DisplayRole) {
        Instance instance = instanceList.at(index.row());
        switch (index.column()) {
        case 0:
            return instance.name();
        case 1:
            return instance.status();
        case 2:
            return instance.driver();
        case 3:
            return instance.containerRuntime();
        case 4:
            return QString::number(instance.cpus());
        case 5:
            return QString::number(instance.memory());
        }
    }
    return QVariant();
}

QVariant InstanceModel::headerData(int section, Qt::Orientation orientation, int role) const
{
    if (role != Qt::DisplayRole)
        return QVariant();

    if (orientation == Qt::Horizontal) {
        switch (section) {
        case 0:
            return tr("Name");
        case 1:
            return tr("Status");
        case 2:
            return tr("Driver");
        case 3:
            return tr("Container Runtime");
        case 4:
            return tr("CPUs");
        case 5:
            return tr("Memory (MB)");
        }
    }
    return QVariant(); // QStringLiteral("Row %1").arg(section);
}
