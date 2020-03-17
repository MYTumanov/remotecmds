package download

// ErrStopDownloadSignal error signal stop download
type ErrStopDownloadSignal struct {
	message string
}

// NewErrStopDownloadSignal return new error type ErrStopDownloadSignal
func NewErrStopDownloadSignal(msg string) *ErrStopDownloadSignal {
	return &ErrStopDownloadSignal{
		message: msg,
	}
}

func (e *ErrStopDownloadSignal) Error() string {
	return e.message
}
