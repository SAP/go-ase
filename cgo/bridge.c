#include <stdio.h>
#include "ctlib.h"
#include "bridge.h"

CS_RETCODE ct_callback_server_message(CS_CONTEXT* ctx, CS_CONNECTION* con, CS_SERVERMSG* msg) {
	return srvMsg(msg);
}

CS_RETCODE ct_callback_client_message(CS_CONTEXT* ctx, CS_CONNECTION* con, CS_CLIENTMSG* msg) {
	return ctlMsg(msg);
}
