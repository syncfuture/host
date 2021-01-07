package main

func (x *basePage) IsAuthenticated() bool {
	return x.Username != ""
}
