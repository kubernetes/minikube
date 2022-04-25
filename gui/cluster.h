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

#ifndef CLUSTER_H
#define CLUSTER_H

#include <QAbstractListModel>
#include <QString>
#include <QList>
#include <QMap>

class Cluster
{
public:
    Cluster() : Cluster("") { }
    Cluster(const QString &name)
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
    bool isEmpty() { return m_name.isEmpty(); }

private:
    QString m_name;
    QString m_status;
    QString m_driver;
    QString m_container_runtime;
    int m_cpus;
    int m_memory;
};

typedef QList<Cluster> ClusterList;
typedef QHash<QString, Cluster> ClusterHash;

class ClusterModel : public QAbstractListModel
{
    Q_OBJECT

public:
    ClusterModel(const ClusterList &clusters, QObject *parent = nullptr)
        : QAbstractListModel(parent), clusterList(clusters)
    {
    }

    void setClusters(const ClusterList &clusters);
    int rowCount(const QModelIndex &parent = QModelIndex()) const override;
    int columnCount(const QModelIndex &parent = QModelIndex()) const override;
    QVariant data(const QModelIndex &index, int role) const override;
    QVariant headerData(int section, Qt::Orientation orientation,
                        int role = Qt::DisplayRole) const override;

private:
    ClusterList clusterList;
};

#endif // CLUSTER_H
