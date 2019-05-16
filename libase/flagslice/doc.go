/*
Package flagslice defines slice types to be used as custom flag types for the flag library.

Slice types as custom flag types allow a flag to be passed multiple
times, e.g. to accept multiple options.

	var fOpts = &flagslice.FlagStringSlice{}

	flag.Var(fOpts, "o", "Options in key/value form, can be passed multiple times")
	flag.Parse()

	for i, fOpt := range fOpts.Slice() {
		log.Printf("Passed value no. %d: %s", i, fOpt)
	}

*/
package flagslice
