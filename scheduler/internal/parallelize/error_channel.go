package parallelize

import "context"

type ErrorChannel struct {
	errCh chan error
}

func (e *ErrorChannel) SendError(err error) {
	select {
	case e.errCh <- err:
	default:
	}
}

func (e *ErrorChannel) SendErrorWithCancel(err error, cancel context.CancelFunc) {
	e.SendError(err)
	cancel()
}

func (e *ErrorChannel) ReceiveError() error {
	select {
	case err := <-e.errCh:
		return err
	default:
		return nil
	}
}

func NewErrorChannel() *ErrorChannel {
	return &ErrorChannel{
		errCh: make(chan error, 1),
	}
}
