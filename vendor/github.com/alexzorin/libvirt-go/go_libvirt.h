#ifndef GO_LIBVIRT_H
#define GO_LIBVIRT_H
void closeCallback_cgo(virConnectPtr conn, int reason, void *opaque);
int virConnectRegisterCloseCallback_cgo(virConnectPtr c, virConnectCloseFunc cb, long goCallbackId);

typedef struct auth_cb_data {
    char* username;
	uint  username_len;
    char* passphrase;
	uint  passphrase_len;
} auth_cb_data;

int* authMechs();
int authCb(virConnectCredentialPtr cred, unsigned int ncred, void *cbdata);
auth_cb_data* authData(char* username, uint username_len, char* passphrase, uint passphrase_len);

#endif /* GO_LIBVIRT_H */
