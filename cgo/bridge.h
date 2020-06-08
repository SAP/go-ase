#ifndef BRIDGE_H
#define BRIDGE_H

CS_RETCODE ct_callback_server_message(CS_CONTEXT*, CS_CONNECTION*, CS_SERVERMSG*);
CS_RETCODE ct_callback_client_message(CS_CONTEXT*, CS_CONNECTION*, CS_CLIENTMSG*);

CS_RETCODE srvMsg(CS_SERVERMSG*);
CS_RETCODE cltMsg(CS_CLIENTMSG*);

#endif
