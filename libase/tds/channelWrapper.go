package tds

import "errors"

type channelWrapper struct {
	ch   *channel
	err  error
	done bool
}

func (chw *channelWrapper) Finish() {
	chw.done = true
	chw.ch = nil
}

func (chw channelWrapper) Error() error {
	if errors.Is(chw.err, ErrChannelExhausted) {
		return nil
	}

	return chw.err
}

func (chw channelWrapper) Finished() bool {
	return chw.done
}
