test:
	@printf "\033[0;32mRUNNING TESTS\033[0m\n"
	@printf "\033[1;30m..................................\033[0m\n"
	@GOPATH=$(GOPATH):`pwd` go test -gocheck.vv