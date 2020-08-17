// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

/*

The package libtest contains helpers to setup and teardown database for
various connection types, as well as shared code to run the same tests
for types and database/sql methods for the cgo and go implementation.

Files starting with `type_` contain functions to test a specific data type.
These tests should follow these constraints:
	1. The underlying test should be run for each connection type once.
	2. Each separate test run must create its own table.
		The table must not be deleted after the test.
	3. If a type can be nulled the handling of the null value must be tested.

Files starting with `sql_` contain functions to test a function group from database/sql.
These tests should follow these constraints:
	1. Methods accepting a context must be tested for their handling of the context.
	2. All execution paths must be tested.
		E.g. .Begin returns a transaction - which can be commmited or rolled back.
		In that case both the .Commit and the .Rollback must be tested.

*/
package libtest
