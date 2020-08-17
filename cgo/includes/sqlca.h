// Copyright (c) 2013 SAP AG or an SAP affiliate company.  All rights reserved.
// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

/*
** sqlca.h - This is the header file for the sqlca structure for precompilers
*/
#ifndef __SQLCA_H__

#define __SQLCA_H__

/*****************************************************************************
**
** sqlca structure used
**
*****************************************************************************/

typedef struct _sqlca
{
	char	sqlcaid[8];
	long	sqlcabc;
	long	sqlcode;
	
	struct
	{
		long		sqlerrml;
		char		sqlerrmc[256];
	} sqlerrm;

	char	sqlerrp[8];
	long	sqlerrd[6];
	char	sqlwarn[8];
	char	sqlext[8];

} SQLCA;

#endif /* __SQLCA_H__ */
