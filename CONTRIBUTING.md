<!--
SPDX-FileCopyrightText: 2020 SAP SE

SPDX-License-Identifier: Apache-2.0
-->

# Contributing to go-ase

- [Report an Issue](#report-an-issue)
  * [Checklist for Bug Reports](#checklist-for-bug-reports)
  * [Report a security issue](#report-a-security-issue)
- [Contribute Code](#contribute-code)
  * [Contributor License Agreement](#contributor-license-agreement)
    - [Company Contributors](#company-contributors)
  * [Contribution Content Guidelines](#contribution-content-guidelines)

## Report an Issue

If you find a bug - behaviour of go-ase code contradicting its
specification - you are welcome to report it.

We can only handle well-reported, actual bugs - so please follow the
guidelines below and the issue template.

You can report issues at our [issue tracker][issues].

### Checklist for Bug Reports

Please only report bugs if all of the following points are true:

* Bug exists in the newest release.
* The bug has not been reported yet.
* The bug is reproducible to an extent (e.g. occurs once every x tries).
* You can provide a minimal example reproducing the bug.

### Report a Security Issue

If you find a security issue, please act responsibly and report it not
in the public issue tracker, but directly to us, so we can fix it before
it can be exploited:
 * SAP Customers: if the found security issue is not covered by
   a published security note, please report it by creating a customer
message at https://service.sap.com/message.
 * Researchers/non-Customers: please send the related information to
   secure@sap.com using [PGP for e-mail
encryption](https://www.sap.com/dmc/policies/pgp/keyblock.txt).
Also refer to the general [SAP security information
page](https://www.sap.com/corporate/en/company/security.html).

## Contribute Code

You are welcome to contribute code to go-ase in order to fix bugs or to
implement new features.

There are three important things to know:

1. You must be aware of the Apache License and agree to the Contributors
   License Agreement. This is common practice in all major Open Source
   projects.
2. The code must fulfill several requirements regarding code style,
   quality and product standards. These are detailed below in the
   [respective section](#contribution-content-guidelines).
3. Not all proposed contributions can be accepted.
   Some features may fit better in a third-party package. The code
   overall must fit into the go-ase project.

### Contributor License Agreement

When you contribute (code, documentation, or anything else), you have to
be aware that your contribution is covered by the same [Apache 2.0
License](http://www.apache.org/licenses/LICENSE-2.0) that is applied to
go-ase itself.

In particular you need to agree to the Developer Certificate of Origin,
which can be [found here][dco].

This applies to all contributors, including those contributing on behalf
of a company. If you agree to its content, you simply have to click on
the link posted by the CLA assistant as a comment to the pull request.
Click it to check the DCO, then accept it on the following screen if you
agree to it. CLA assistant will save this decision for upcoming
contributions and will notify you if there is any change to the DCO in
the meantime.

### Contribution Content Guidelines

Contributions must fulfill following requirements to be accepted:

1. Code must formatted using `gofmt` and adhere to the standards
   explained in [effective go][effective-go]. The following points
   may formulate statements from effective go more strongly.
2. All exported functions must be documented.
3. Code that can be shared between the go and cgo implementations must
   be placed inside the libase/ package.
4. Tests for new features or bugfixes must be added.

[dco]: https://developercertificate.org/
[effective-go]: https://golang.org/doc/effective_go.html
[issues]: https://github.com/SAP/go-ase/issues
