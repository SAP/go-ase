// Copyright (c) 2013 SAP AG or an SAP affiliate company.  All rights reserved.
// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

#ifndef __SQLDA_H__
#define __SQLDA_H__

typedef struct _sqlda
{
	CS_SMALLINT sd_sqln;
	CS_SMALLINT sd_sqld;
	struct _sd_column
	{
		CS_DATAFMT sd_datafmt;
		CS_VOID *sd_sqldata;
		CS_SMALLINT sd_sqlind;
		CS_INT sd_sqllen;
		CS_VOID	*sd_sqlmore;
	} sd_column[1];
} syb_sqlda;

typedef syb_sqlda SQLDA;

#define SYB_SQLDA_SIZE(n) (sizeof(SQLDA) \
		- sizeof(struct _sd_column) \
		+ (n) * sizeof(struct _sd_column))

#ifndef SQLDA_DECL
#define SQLDA_DECL(name, size) \
struct { \
	CS_SMALLINT sd_sqln; \
	CS_SMALLINT sd_sqld; \
	struct { \
		CS_DATAFMT sd_datafmt; \
		CS_VOID *sd_sqldata; \
		CS_SMALLINT sd_sqlind; \
		CS_INT sd_sqllen; \
		CS_VOID	*sd_sqlmore; \
	} sd_column[(size)]; \
} name
#endif /* SQLDA_DECL */

#endif /* __SQLDA_H__ */
