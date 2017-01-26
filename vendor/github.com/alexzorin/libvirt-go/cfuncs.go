package libvirt

/*
#cgo CFLAGS: -Wno-implicit-function-declaration
#cgo LDFLAGS: -lvirt
#include <libvirt/libvirt.h>
#include <libvirt/virterror.h>
#include <stdlib.h>
#include <string.h>
#include "go_libvirt.h"

void domainEventLifecycleCallback_cgo(virConnectPtr c, virDomainPtr d,
                                     int event, int detail, void *data)
{
    domainEventLifecycleCallback(c, d, event, detail, data);
}

void domainEventGenericCallback_cgo(virConnectPtr c, virDomainPtr d, void *data)
{
    domainEventGenericCallback(c, d, data);
}

void domainEventRTCChangeCallback_cgo(virConnectPtr c, virDomainPtr d,
                                     long long utcoffset, void *data)
{
    domainEventRTCChangeCallback(c, d, utcoffset, data);
}

void domainEventWatchdogCallback_cgo(virConnectPtr c, virDomainPtr d,
                                    int action, void *data)
{
    domainEventWatchdogCallback(c, d, action, data);
}

void domainEventIOErrorCallback_cgo(virConnectPtr c, virDomainPtr d,
                                   const char *srcPath, const char *devAlias,
                                   int action, void *data)
{
    domainEventIOErrorCallback(c, d, srcPath, devAlias, action, data);
}

void domainEventGraphicsCallback_cgo(virConnectPtr c, virDomainPtr d,
                                    int phase, const virDomainEventGraphicsAddress *local,
                                    const virDomainEventGraphicsAddress *remote,
                                    const char *authScheme,
                                    const virDomainEventGraphicsSubject *subject, void *data)
{
    domainEventGraphicsCallback(c, d, phase, local, remote, authScheme, subject, data);
}

void domainEventIOErrorReasonCallback_cgo(virConnectPtr c, virDomainPtr d,
                                         const char *srcPath, const char *devAlias,
                                         int action, const char *reason, void *data)
{
    domainEventIOErrorReasonCallback(c, d, srcPath, devAlias, action, reason, data);
}

void domainEventBlockJobCallback_cgo(virConnectPtr c, virDomainPtr d,
                                    const char *disk, int type, int status, void *data)
{
    domainEventBlockJobCallback(c, d, disk, type, status, data);
}

void domainEventDiskChangeCallback_cgo(virConnectPtr c, virDomainPtr d,
                                      const char *oldSrcPath, const char *newSrcPath,
                                      const char *devAlias, int reason, void *data)
{
    domainEventDiskChangeCallback(c, d, oldSrcPath, newSrcPath, devAlias, reason, data);
}

void domainEventTrayChangeCallback_cgo(virConnectPtr c, virDomainPtr d,
                                      const char *devAlias, int reason, void *data)
{
    domainEventTrayChangeCallback(c, d, devAlias, reason, data);
}

void domainEventReasonCallback_cgo(virConnectPtr c, virDomainPtr d,
                                  int reason, void *data)
{
    domainEventReasonCallback(c, d, reason, data);
}

void domainEventBalloonChangeCallback_cgo(virConnectPtr c, virDomainPtr d,
                                         unsigned long long actual, void *data)
{
    domainEventBalloonChangeCallback(c, d, actual, data);
}

void domainEventDeviceRemovedCallback_cgo(virConnectPtr c, virDomainPtr d,
                                         const char *devAlias, void *data)
{
    domainEventDeviceRemovedCallback(c, d, devAlias, data);
}

void freeGoCallback_cgo(void* goCallbackId) {
   freeCallbackId((long)goCallbackId);
}

int virConnectDomainEventRegisterAny_cgo(virConnectPtr c,  virDomainPtr d,
                                         int eventID, virConnectDomainEventGenericCallback cb,
                                         long goCallbackId) {
    void* id = (void*)goCallbackId;
    return virConnectDomainEventRegisterAny(c, d, eventID, cb, id, freeGoCallback_cgo);
}

void errorGlobalCallback_cgo(void *userData, virErrorPtr error)
{
    globalErrorCallback(error);
}

void errorConnCallback_cgo(void *userData, virErrorPtr error)
{
    connErrorCallback((long)userData, error);
}

void virConnSetErrorFunc_cgo(virConnectPtr c, long goCallbackId, virErrorFunc cb)
{
    void* id = (void*)goCallbackId;
    virConnSetErrorFunc(c, id, cb);
}

void closeCallback_cgo(virConnectPtr conn, int reason, void *opaque)
{
    closeCallback(conn, reason, (long)opaque);
}

int authCb(virConnectCredentialPtr cred, unsigned int ncred, void *cbdata)
{
	int i;

    auth_cb_data *data = (auth_cb_data*)cbdata;
    for (i = 0; i < ncred; i++) {
        if (cred[i].type == VIR_CRED_AUTHNAME) {
            cred[i].result = strndup(data->username, data->username_len);
            if (cred[i].result == NULL)
                return -1;
            cred[i].resultlen = strlen(cred[i].result);
        }
        else if (cred[i].type == VIR_CRED_PASSPHRASE) {
            cred[i].result = strndup(data->passphrase, data->passphrase_len);
            if (cred[i].result == NULL)
                return -1;
            cred[i].resultlen = strlen(cred[i].result);
        }
    }
    return 0;
}

auth_cb_data* authData(char* username, uint username_len, char* passphrase, uint passphrase_len) {
    auth_cb_data * data = malloc(sizeof(auth_cb_data));
    data->username = username;
    data->username_len = username_len;
    data->passphrase = passphrase;
    data->passphrase_len = passphrase_len;
    return data;
}

int* authMechs() {
    int* authMechs = malloc(2*sizeof(VIR_CRED_AUTHNAME));
    authMechs[0] = VIR_CRED_AUTHNAME;
    authMechs[1] = VIR_CRED_PASSPHRASE;
    return authMechs;
}

int virConnectRegisterCloseCallback_cgo(virConnectPtr c, virConnectCloseFunc cb, long goCallbackId)
{
    void *id = (void*)goCallbackId;
    return virConnectRegisterCloseCallback(c, cb, id, freeGoCallback_cgo);
}

*/
import "C"
