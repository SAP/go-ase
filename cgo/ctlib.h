#ifndef CTLIB_H
#define CTLIB_H

// _LP64 and __LP64__ are the correct defines for the LP64 ABI - in older Open
// Server headers only SYB_LP64 is being checked. Hence SYB_LP64 is defined here
// manually when on a 64bit system.
#if defined(_LP64) || defined(__LP64__)
#  define SYB_LP64
#endif

#include <ctpublic.h>

#endif
