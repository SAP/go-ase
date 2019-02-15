#ifndef BRIDGE_H
#define BRIDGE_H

typedef struct CS_CONNECTION_WRAPPER {
	CS_CONNECTION *conn;
	CS_RETCODE rc;
} CS_CONNECTION_WRAPPER;
CS_CONNECTION_WRAPPER ct_con_alloc_wrapper(CS_CONTEXT*);

CS_RETCODE ct_callback_wrapper_for_server_messages(CS_CONTEXT*);
CS_RETCODE ct_callback_server_message(CS_CONTEXT*, CS_CONNECTION*, CS_SERVERMSG*);
CS_RETCODE ct_callback_wrapper_for_client_messages(CS_CONTEXT*);
CS_RETCODE ct_callback_client_message(CS_CONTEXT*, CS_CONNECTION*, CS_CLIENTMSG*);

#endif
