#include <stdio.h>
#include "ctpublic.h"
#include "bridge.h"

CS_CONNECTION_WRAPPER ct_con_alloc_wrapper(CS_CONTEXT* ctx) {
	CS_CONNECTION* conn;
	CS_RETCODE rc;
	rc = ct_con_alloc(ctx, &conn);
	CS_CONNECTION_WRAPPER w = { conn, rc };
	return w;
}

CS_RETCODE ct_callback_wrapper_for_server_messages(CS_CONTEXT* ctx) {
	CS_RETCODE rc;
	rc = ct_callback(ctx, NULL, CS_SET, CS_SERVERMSG_CB, ct_callback_server_message);
	return rc;
}

CS_RETCODE ct_callback_server_message(CS_CONTEXT* ctx, CS_CONNECTION* con, CS_SERVERMSG* msg) {
	srvMsg(msg);
}

CS_RETCODE ct_callback_wrapper_for_client_messages(CS_CONTEXT* ctx) {
	CS_RETCODE rc;
	rc = ct_callback(ctx, NULL, CS_SET, CS_CLIENTMSG_CB, ct_callback_client_message);
	return rc;
}

CS_RETCODE ct_callback_client_message(CS_CONTEXT* ctx, CS_CONNECTION* con, CS_CLIENTMSG* msg) {
	ctlMsg(msg);
}
