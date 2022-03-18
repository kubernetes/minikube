/****************************************************************************
**
** Copyright 2022 The Kubernetes Authors All rights reserved.
**
** Copyright (C) 2021 Anders F Björklund
** Copyright (C) 2016 The Qt Company Ltd.
** Contact: https://www.qt.io/licensing/
**
** This file is part of the examples of the Qt Toolkit.
**
** $QT_BEGIN_LICENSE:BSD$
** Commercial License Usage
** Licensees holding valid commercial Qt licenses may use this file in
** accordance with the commercial license agreement provided with the
** Software or, alternatively, in accordance with the terms contained in
** a written agreement between you and The Qt Company. For licensing terms
** and conditions see https://www.qt.io/terms-conditions. For further
** information use the contact form at https://www.qt.io/contact-us.
**
** BSD License Usage
** Alternatively, you may use this file under the terms of the BSD license
** as follows:
**
** "Redistribution and use in source and binary forms, with or without
** modification, are permitted provided that the following conditions are
** met:
**   * Redistributions of source code must retain the above copyright
**     notice, this list of conditions and the following disclaimer.
**   * Redistributions in binary form must reproduce the above copyright
**     notice, this list of conditions and the following disclaimer in
**     the documentation and/or other materials provided with the
**     distribution.
**   * Neither the name of The Qt Company Ltd nor the names of its
**     contributors may be used to endorse or promote products derived
**     from this software without specific prior written permission.
**
**
** THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
** "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
** LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
** A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
** OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
** SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
** LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
** DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
** THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
** (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
** OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE."
**
** $QT_END_LICENSE$
**
****************************************************************************/

#include "cluster.h"

#include <QStringList>

void ClusterModel::setClusters(const ClusterList &clusters)
{
    beginResetModel();
    clusterList = clusters;
    endResetModel();
}

int ClusterModel::rowCount(const QModelIndex &) const
{
    return clusterList.count();
}

int ClusterModel::columnCount(const QModelIndex &) const
{
    return 6;
}

static QStringList binaryAbbrs = { "B", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB", "ZiB", "YiB" };

QVariant ClusterModel::data(const QModelIndex &index, int role) const
{
    if (!index.isValid())
        return QVariant();

    if (index.row() >= clusterList.size())
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
        Cluster cluster = clusterList.at(index.row());
        switch (index.column()) {
        case 0:
            return cluster.name();
        case 1:
            return cluster.status();
        case 2:
            return cluster.driver();
        case 3:
            return cluster.containerRuntime();
        case 4:
            return QString::number(cluster.cpus());
        case 5:
            return QString::number(cluster.memory());
        }
    }
    return QVariant();
}

QVariant ClusterModel::headerData(int section, Qt::Orientation orientation, int role) const
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
