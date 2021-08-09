package dhhclient

import ("strconv")

type DHHError struct {
	errCode int
	verbose string
}

func (this *DHHError) Error() string {
	return "Errcode(" + strconv.Itoa(this.errCode) + "), verbose: " + this.verbose
}

func (this *DHHError) Code() int {
	return this.errCode
}

func NewError(errCode int, verbose string) *DHHError {
	return &DHHError{errCode, verbose}
}